package network

import (
	"bytes"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"tools/flexrange"
	"tools/utils"
	"tools/validator"
)

type IPNet struct {
	// global.GVA_MODEL `mapstructure:",squash" json:"-"`
	IP   IP     `gorm:"type:ip"`
	Mask IPMask `gorm:"tpye:mask"`
}

type AbbrNet interface {
	Copy() utils.CopyAble
	Type() IPFamily
	First() *IP
	Last() *IP
	Size() int
	Count() *big.Int
	MatchIPNet(n *IPNet) bool
	MatchIPRange(n *IPRange) bool
	IPNetList() ([]*IPNet, error)
	SuperNet() (*IPNet, error)
	DataRange() flexrange.DataRangeInf
	String() string
	AddressType() AddressType
	IPNet() (*IPNet, bool)
}

func (m *IPNet) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New(fmt.Sprint("Failed to unmarshal IPNet value:", value))
	}

	err := json.Unmarshal(bytes, m)
	return err
}

func (m IPNet) Value() (driver.Value, error) {
	return json.Marshal(&m)
}

func (IPNet) GormDataType() string {
	return "ipnet"
}

func (n *IPNet) AddressType() AddressType {
	if n.Type() == IPv4 && n.Mask.Prefix() == 32 {
		return HOST
	}
	if n.Type() == IPv6 && n.Mask.Prefix() == 128 {
		return HOST
	}
	return SUBNET
}

func (n *IPNet) Type() IPFamily {
	return n.IP.Type()
}

func (n *IPNet) First() *IP {
	ip := make(IP, n.IP.Len())
	for index, b := range n.IP {
		ip[index] = b & n.Mask[index]
	}
	return &ip
}

func (n *IPNet) Size() int {
	return n.IP.Size()
}

func (n *IPNet) Last() *IP {
	ip := make(IP, n.IP.Len())
	copy(ip, *n.First())

	rm := n.Mask.Reverse()

	for index, b := range ip {
		ip[index] = b ^ (*rm)[index]
	}
	return &ip
}

func (n *IPNet) Count() *big.Int {
	if n.Mask.Prefix() != -1 {
		prefix := n.IP.Size() - n.Mask.Prefix()
		z := new(big.Int)
		z.Lsh(big.NewInt(1), uint(prefix))
		return z
	} else {
		_, bit0s := n.Mask.BitsLen()
		z := new(big.Int)
		z.Lsh(big.NewInt(1), uint(bit0s))
		return z
	}
}

func (n *IPNet) ToRange() *IPRange {
	if n.Mask.Prefix() != -1 {

		first := n.First()
		last := n.Last()
		return &IPRange{
			*first,
			*last,
		}
	} else {
		return nil
	}

}

func (n *IPNet) Copy() utils.CopyAble {
	net := &IPNet{
		IP:   *n.IP.Copy(),
		Mask: *n.Mask.Copy(),
	}

	return net
}

func (n *IPNet) SuperNet() (*IPNet, error) {
	l := n.Mask.Prefix()
	if l == 0 {
		return nil, fmt.Errorf("not support zero prefix, l = %d", l)
	}

	if l == -1 {
		l = n.Mask.FirstBit(true)
		if l == -1 {
			mask, err := NewIPMask(uint(n.Size()-1), n.Type())
			if err != nil {
				return nil, err
			}
			return &IPNet{
				IP:   n.IP.AfterMask(mask),
				Mask: *mask,
			}, nil
		} else {
			if l == 0 {
				//return nil
				return nil, fmt.Errorf("not support zero prefix, l = %d", l)
			} else {
				mask, err := NewIPMask(uint(l-1), n.Type())
				if err != nil {
					return nil, err
				}
				return &IPNet{
					IP:   n.IP.AfterMask(mask),
					Mask: *mask,
				}, nil
			}
		}
	} else {
		mask, err := NewIPMask(uint(l-1), n.Type())
		if err != nil {
			return nil, err
		}
		return &IPNet{
			IP:   n.IP.AfterMask(mask),
			Mask: *mask,
		}, nil

	}
}

func (r *IPNet) MatchIP(ip *IP) bool {
	if r.IP.Type() != ip.Type() {
		return false
	}

	if bytes.Compare(*r.First(), *ip) > 0 || bytes.Compare(*r.Last(), *ip) < 0 {
		return false
	}

	prefix := r.Mask.Prefix()
	if prefix != -1 {
		return true
	} else {
		if bytes.Compare(r.IP.AfterMask(&r.Mask), ip.AfterMask(&r.Mask)) == 0 {
			return true
		} else {
			return false
		}
	}
}

func (r *IPNet) MatchIPNet(n *IPNet) bool {
	if r.IP.Type() != n.IP.Type() {
		return false
	}

	if r.Mask.Prefix() == -1 {
		for it := n.Mask.Iterator(&n.IP); it.HasNext(); {
			e := it.Next()
			if !r.MatchIP(e) {
				return false
			}
		}
		return true
	} else {
		if bytes.Compare(*r.First(), *n.First()) > 0 || bytes.Compare(*r.Last(), *n.Last()) < 0 {
			return false
		}
	}

	return true
}

func (r *IPNet) Prefix() int {
	return r.Mask.Prefix()
}

func (r *IPNet) MatchIPRange(n *IPRange) bool {
	if r.IP.Type() != n.Start.Type() {
		return false
	}

	if bytes.Compare(*r.First(), n.Start) > 0 || bytes.Compare(*r.Last(), n.End) < 0 {
		return false
	}

	if r.Mask.Prefix() == -1 {
		nl, err := n.IPNetList()
		if err != nil {
			return false
		}

		for _, n := range nl {
			if !r.MatchIPNet(n) {
				return false
			}
		}
		return true
	} else {
		return true
	}
}

func NewIPNet(s string) (*IPNet, error) {
	return ParseIPNet(s)
}

func ParseIPNet(s string) (*IPNet, error) {
	//fmt.Printf("ParseIPNet: s = %+v\n", s)
	if validator.IsIPv4Address(s) {
		return parseIPNet(s + "/32")
	} else if validator.IsIPv6Address(s) {
		return parseIPNet(s + "/128")
	} else if validator.IsIPv4AddressWithMask(s) || validator.IsIPv6AddressWithMask(s) {
		return parseIPNet(s)
	} else {
		return nil, fmt.Errorf("s:%s format error", s)
	}
}

func parseIPNet(s string) (*IPNet, error) {
	//fmt.Printf("parseIPNet: s = %+v\n", s)
	tokens := strings.Split(s, "/")
	if len(tokens) != 2 {
		return nil, fmt.Errorf("s:%s format error", s)
	}

	ip, err := ParseIP(tokens[0])
	if err != nil {
		return nil, err
	}
	mask, err := ParseIP(tokens[1])

	if err != nil {
		prefix, _ := strconv.Atoi(tokens[1])
		_ip, err := ParseIP(tokens[0])
		if err != nil {
			return nil, err
		}

		mask, err := NewIPMask(uint(prefix), ip.Type())
		if err != nil {
			return nil, err
		}

		return &IPNet{
			IP:   *_ip,
			Mask: *mask,
		}, nil

	} else {
		_ip, err := ParseIP(tokens[0])
		if err != nil {
			return nil, err
		}

		_mask, err := IPtoMask(mask)
		if err != nil {
			return nil, err
		}

		return &IPNet{
			IP:   *_ip,
			Mask: *_mask,
		}, nil
	}

	return nil, fmt.Errorf("unknown error")

}

func (n *IPNet) IPNetList() ([]*IPNet, error) {
	var result []*IPNet
	result = append(result, n)
	return result, nil
}

func (n *IPNet) DataRange() flexrange.DataRangeInf {
	dr := flexrange.NewDataRange(uint32(n.Size()), big.NewInt(0))
	if n.Mask.Prefix() != -1 {
		low := n.First().Int()
		high := n.Last().Int()
		dr.Push(low, high, nil)
		return dr
	} else {
		dr := flexrange.NewDataRange(uint32(n.Size()), big.NewInt(0))
		for it := n.Iterator(); it.HasNext(); {
			ip := it.Next()
			dr.Push(ip.Int(), ip.Int(), nil)
		}
		return dr
	}
}

func (n *IPNet) String() string {
	prefix := n.Mask.Prefix()
	if prefix == -1 {
		m := MasktoIP(n.Mask)
		return fmt.Sprintf("%s/%s", n.IP, m)
	} else {
		return fmt.Sprintf("%s/%d", n.IP, prefix)
	}
}

func (net *IPNet) Iterator() *IPMaskIterator {
	return net.Mask.Iterator(&net.IP)
}

func (net *IPNet) IPNet() (*IPNet, bool) {
	return net.Copy().(*IPNet), true
}

func (net *IPNet) AllIP() []*IP {
	ips := []*IP{}
	for t := net.Iterator(); t.HasNext(); {
		ips = append(ips, t.Next())
	}
	return ips
}

func (net *IPNet) AllIPToString() []string {
	ips := []string{}
	for t := net.Iterator(); t.HasNext(); {
		ip := t.Next().String()
		ips = append(ips, ip)
	}
	return ips
}
