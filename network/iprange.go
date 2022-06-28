package network

import (
	"bytes"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/big"
	"strings"
	"tools/flexrange"
	"tools/utils"
	"tools/validator"
)

func NewIPRange(ip string) (*IPRange, error) {
	if validator.IsIPRange(ip) {
		tokens := strings.Split(ip, "-")
		_ip1, err := ParseIP(tokens[0])
		if err != nil {
			return nil, err
		}
		_ip2, err := ParseIP(tokens[1])
		if err != nil {
			return nil, err
		}
		return &IPRange{
			*_ip1,
			*_ip2,
		}, nil
	}
	return nil, fmt.Errorf("ip:%s format error", ip)
}

func NewIPRangeFromInt(low *big.Int, high *big.Int, fa IPFamily) *IPRange {
	if low.Cmp(high) > 0 {
		return nil
	}
	if low.Cmp(IPMaxInt(fa)) > 0 || high.Cmp(IPMaxInt(fa)) > 0 {
		return nil
	}

	ip1 := NewIPFromInt(low, fa)
	ip2 := NewIPFromInt(high, fa)
	return &IPRange{
		*ip1,
		*ip2,
	}
}

func NewIPRangeFromEtnry(entry flexrange.EntryInt, fa IPFamily) *IPRange {
	return NewIPRangeFromInt(entry.Low(), entry.High(), fa)
}

type IPRange struct {
	Start IP
	End   IP
}

func (r *IPRange) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New(fmt.Sprint("Failed to unmarshal IPNet value:", value))
	}

	err := json.Unmarshal(bytes, r)
	return err
}

func (r IPRange) Value() (driver.Value, error) {
	return json.Marshal(&r)
}

func (IPRange) GormDataType() string {
	return "iprange"
}

func (r *IPRange) AddressType() AddressType {
	if r.Start.Int().Cmp(r.End.Int()) == 0 {
		return HOST
	}

	ipNets := r.CIDRs()
	if len(ipNets) == 1 {
		return SUBNET
	}
	return RANGE
}

func (r *IPRange) String() string {
	return fmt.Sprintf("%s-%s", r.Start, r.End)
}

func (r *IPRange) Copy() utils.CopyAble {
	return &IPRange{
		*r.Start.Copy(),
		*r.End.Copy(),
	}
}

func (r *IPRange) Type() IPFamily {
	return r.Start.Type()
}

func (r *IPRange) First() *IP {
	return &r.Start
}

func (r *IPRange) Size() int {
	return r.Start.Size()
}

func (r *IPRange) Count() *big.Int {
	start := r.Start.Int()
	end := r.End.Int()
	count := new(big.Int)
	count.Sub(end, start)
	count = utils.AddInt(count, 1)
	return count
}

func (r *IPRange) Last() *IP {
	return &r.End
}

func (r *IPRange) MatchIPNet(n *IPNet) bool {
	if r.Start.Type() != n.IP.Type() {
		return false
	}
	if bytes.Compare(r.Start, *n.First()) > 0 || bytes.Compare(r.End, *n.Last()) < 0 {
		return false
	}

	return true
}

func (r *IPRange) MatchIP(ip *IP) bool {
	if r.Start.Type() != ip.Type() {
		return false
	}

	if bytes.Compare(r.Start, *ip) > 0 || bytes.Compare(r.End, *ip) < 0 {
		return false
	}

	return true

}

func (r *IPRange) StartShift(i *big.Int) (*IPRange, error) {
	start, err := r.Start.Add(i)
	if err != nil {
		return nil, err
	}
	if start.Int().Cmp(r.End.Int()) > 0 {
		return nil, fmt.Errorf("%d > %d", start.Int(), r.End.Int())
	}
	if start.Int().Cmp(big.NewInt(0)) < 0 {
		return nil, fmt.Errorf("%d < %d", start.Int(), 0)
	}

	return &IPRange{
		*start,
		*r.End.Copy(),
	}, nil
}

func (r *IPRange) EndShift(i *big.Int) (*IPRange, error) {
	end, err := r.End.Add(i)
	if err != nil {
		return nil, err
	}

	if end.Int().Cmp(r.Start.Int()) < 0 {
		return nil, fmt.Errorf("%d < %d", end.Int(), r.Start.Int())
	}
	if end.Int().Cmp(IPMaxInt(r.Start.Type())) > 0 {
		return nil, fmt.Errorf("%d > %d", end.Int(), IPMaxInt(r.Start.Type()))
	}

	return &IPRange{
		*r.Start.Copy(),
		*end,
	}, nil
}

func (r *IPRange) MatchIPRange(n *IPRange) bool {
	if r.Start.Type() != n.Start.Type() {
		return false
	}

	if bytes.Compare(r.Start, n.Start) > 0 || bytes.Compare(r.End, n.End) < 0 {
		return false
	}

	return true
}

func (r *IPRange) IPNetList() (result []*IPNet, err error) {
	var mask *IPMask
	mask, err = NewIPMask(uint(r.Start.Size()), r.Start.Type())
	if err != nil {
		return
	}

	//net := &IPNet{r.Start, NewIPMask(uint(r.Start.Size()), r.Start.Type())}
	net := &IPNet{IP: r.Start, Mask: *mask}

	count := new(big.Int)
	count = count.Sub(r.End.Int(), r.Start.Int())
	count = count.Add(count, big.NewInt(1))
	zero := big.NewInt(0)

	nr := &IPRange{
		r.Start,
		r.End,
	}

	result = []*IPNet{}

	tmp := net
	test_count := new(big.Int)
	test_count.Set(count)
	for test_count.Cmp(zero) > 0 || test_count.Cmp(zero) == 0 {
		if nr.MatchIPNet(net) {
			test_count = test_count.Sub(count, net.Count())
			tmp = net.Copy().(*IPNet)
			if net.MatchIPRange(nr) {
				result = append(result, tmp)
				return
			}
			net, err = net.SuperNet()
			if err != nil {
				return
			}

		} else {
			result = append(result, tmp)
			nr, err = nr.StartShift(tmp.Count())
			if err != nil {
				return
			}

			count = count.Sub(nr.End.Int(), nr.Start.Int())
			count = count.Add(count, big.NewInt(1))
			test_count.Set(count)
			mask, err = NewIPMask(uint(nr.Start.Size()), nr.Start.Type())
			if err != nil {
				return
			}
			//net = &IPNet{nr.Start, NewIPMask(uint(nr.Start.Size()), nr.Start.Type())}
			net = &IPNet{IP: nr.Start, Mask: *mask}
		}

	}
	return
}

func (r *IPRange) CIDRs() []*IPNet {
	start := r.First().Int()
	end := r.Last().Int()
	var max int
	if r.Type() == IPv4 {
		max = 32
	} else {
		max = 128
	}

	var prev, current *IPNet
	var result []*IPNet
	for i := start; i.Cmp(end) <= 0; {
		ip := NewIPFromInt(i, r.Type())
		mask, _ := NewIPMask(uint(max), r.Type())
		prev = &IPNet{
			IP:   *ip,
			Mask: *mask,
		}
		for m := max - 1; m >= 0; m-- {
			mask, _ := NewIPMask(uint(m), r.Type())
			current = &IPNet{
				IP:   *ip,
				Mask: *mask,
			}

			step := max - m - 1
			nr := NewIPRangeFromInt(i, end, r.Type())

			if !nr.MatchIPNet(current) {
				result = append(result, prev)
				i = i.Add(i, big.NewInt(int64(math.Pow(2, float64(step)))))
				break
			}
			if m == 0 {
				result = append(result, current)
				i = i.Add(i, big.NewInt(int64(math.Pow(2, float64(step+1)))))
				break
			}
			prev = current

			// if
		}

	}

	return result
}

func (r *IPRange) IPNet() (*IPNet, bool) {
	cidrs := r.CIDRs()
	if len(cidrs) == 1 {
		return cidrs[0], true
	} else {
		return nil, false
	}
}

func (r *IPRange) SuperNet() (*IPNet, error) {
	ip := r.Start.Copy()
	mask, err := NewIPMask(uint(r.Start.Size()), r.Start.Type())
	if err != nil {
		return nil, err
	}
	n := &IPNet{
		IP:   *ip,
		Mask: *mask,
	}

	for n.Mask.Prefix() >= 0 {
		if n.MatchIP(&r.Start) && n.MatchIP(&r.End) {
			return n, nil
		} else {
			n, err = n.SuperNet()
			if err != nil {
				return nil, err
			}
		}
	}

	return nil, fmt.Errorf("unknown error")

}

func (r *IPRange) DataRange() flexrange.DataRangeInf {
	dr := flexrange.NewDataRange(uint32(r.Size()), big.NewInt(0))
	low := r.First().Int()
	high := r.Last().Int()
	dr.Push(low, high, nil)

	return dr
}

type IPRangeIterator struct {
	start *IP
	end   *IP
	index *big.Int
}

func (ir *IPRange) Iterator() *IPRangeIterator {

	return &IPRangeIterator{
		&ir.Start,
		&ir.End,
		big.NewInt(0),
	}
}

func (it *IPRangeIterator) HasNext() bool {
	s1 := it.start.Int()
	s2 := it.end.Int()

	if s1.Add(s1, it.index); !(s1.Cmp(s2) > 0) {
		return true
	}
	return false
}

func (it *IPRangeIterator) Next() *IP {
	s1 := it.start.Int()
	s2 := it.end.Int()

	r := s1.Add(s1, it.index)
	if r.Cmp(s2) > 0 {
		return nil
	}
	it.index = utils.AddInt(it.index, 1)
	return NewIPFromInt(r, it.start.Type())

}
