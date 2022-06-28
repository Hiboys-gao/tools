package network

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/big"
	"net"
	"strconv"
	"strings"
	"tools/validator"
)

type AddressType int

const (
	ADDRESS_NONE AddressType = iota
	HOST
	SUBNET
	RANGE
	MASK
	LIST
	MIXED
)

type IPFamily int

const (
	IPv4 IPFamily = iota
	IPv6
	//V4AndV6
	//Empty
)

const (
	IPv4Len = 4
	IPv6Len = 16
)

func (t IPFamily) String() string {
	return [...]string{"IPv4", "IPv6"}[t]
}

type IPMask []byte
type IP []byte
type RG interface {
	First() IP
	Last() IP
}

func (m IP) MarshalJSON() (b []byte, err error) {
	var data []string
	for _, d := range m {
		data = append(data, fmt.Sprintf("%d", d))
	}

	return json.Marshal(data)
}

func (m *IP) UnmarshalJSON(b []byte) error {
	var data []string

	err := json.Unmarshal(b, &data)
	if err != nil {
		return err
	}

	for _, d := range data {
		byteData, err := strconv.Atoi(d)
		if err != nil {
			return err
		}

		*m = append(*m, byte(byteData))
	}

	return nil
}

func (m *IP) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New(fmt.Sprint("Failed to unmarshal IP value:", value))
	}

	err := m.UnmarshalJSON(bytes)
	return err
}

func (m IP) Value() (driver.Value, error) {
	if len(m) == 0 {
		return nil, nil
	}

	return m.MarshalJSON()
}

func (IP) GormDataType() string {
	return "ip"
}

func (m *IPMask) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New(fmt.Sprint("Failed to unmarshal IPMask value:", value))
	}

	err := m.UnmarshalJSON(bytes)
	return err
}

func (m IPMask) Value() (driver.Value, error) {
	if len(m) == 0 {
		return nil, nil
	}

	return m.MarshalJSON()
}

func (IPMask) GormDataType() string {
	return "mask"
}

// func (m IPMask) MarshalJSON() (b []byte, err error) {
// if len(b) == 0 {
// return nil, nil
// }
// s := m.String()
//
// return json.Marshal(&s)
// }
//
// func (m *IPMask) UnmarshalJSON(b []byte) error {
// var s *string
// err := json.Unmarshal(b, s)
// if err != nil {
// return err
// }
//
// ip, err := ParseIP(*s)
// if err != nil {
// return err
// }
//
// m, err = IPtoMask(ip)
// if err != nil {
// return err
// }
//
// return nil
// }

func (m IPMask) MarshalJSON() (b []byte, err error) {
	var data []string
	for _, d := range m {
		data = append(data, fmt.Sprintf("%d", d))
	}

	return json.Marshal(data)

}

func (m *IPMask) UnmarshalJSON(b []byte) error {
	var data []string

	err := json.Unmarshal(b, &data)
	if err != nil {
		return err
	}

	for _, d := range data {
		byteData, err := strconv.Atoi(d)
		if err != nil {
			return err
		}

		*m = append(*m, byte(byteData))
	}

	return nil
}

func (m *IP) Size() int {
	if m.Type() == IPv4 {
		return 32
	}

	if m.Type() == IPv6 {
		return 128
	}

	return -1
}

func (ip *IP) AddressType() AddressType {
	return HOST
}

func (ip *IP) Type() IPFamily {
	if ip.Len() == 4 {
		return IPv4
	} else if ip.Len() == 16 {
		return IPv6
	} else {
		panic("IP Type() failed")
	}

	//if len(ip) == 4 {
	//return IPv4
	//} else if len(ip) == 16 {
	//return IPv6
	//} else {
	////panic(fmt.Sprintf("IPMask ip: %+v is invalid", ip))
	//fmt.Printf("len(ip) = %+v\n", len(ip))
	//panic("IPMask Type() failed")
	//}

}

func (ip *IP) Len() int {
	return len(*ip)
}

func (ip *IP) Copy() *IP {
	//func (ip IP) Copy() utils.CopyAble {
	p := make(IP, ip.Len())
	copy(p, *ip)
	return &p
}

func (ip IP) String() string {

	if ip.Type() == IPv4 {
		return fmt.Sprintf("%d.%d.%d.%d", ip[0], ip[1], ip[2], ip[3])
	} else {
		e0 := -1
		e1 := -1
		for i := 0; i < IPv6Len; i += 2 {
			j := i
			for j < IPv6Len && ip[j] == 0 && ip[j+1] == 0 {
				j += 2
			}
			if j > i && j-i > e1-e0 {
				e0 = i
				e1 = j
				i = j
			}
		}
		if e1-e0 <= 2 {
			e0 = -1
			e1 = -1
		}

		//const maxLen = len("ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff")
		//b := make([]byte, 0, maxLen)
		s := ""
		//fmt.Printf("e0 = %+v\n", e0)
		//fmt.Printf("e1 = %+v\n", e1)
		// Print with possible :: in place of run of zeros
		for i := 0; i < IPv6Len; i += 2 {
			if i == e0 {
				s = s + "::"
				i = e1
				if i >= IPv6Len {
					break
				}
			} else if i > 0 {
				s = s + ":"
				//b = append(b, ':')
			}
			//fmt.Printf("ip[i] = %x,  %x\n", ip[i], uint32(ip[i])<<8)
			//fmt.Printf("ip[i+1] = %x\n", ip[i+1])
			//fmt.Printf("%x\n", (uint32(ip[i])<<8)|uint32(ip[i+1]))
			s = s + fmt.Sprintf("%x", (uint32(ip[i])<<8)|uint32(ip[i+1]))
			//fmt.Printf("s = %+v\n", s)

			//b = appendHex(b, (uint32(p[i])<<8)|uint32(p[i+1]))
		}
		return s
	}
}

func (m *IPMask) AddressType() AddressType {
	return MASK
}

func (m IPMask) GetBit(index int) int {
	if index < 0 || index >= m.Size() {
		return -1
		//return byte(0), fmt.Errorf("index: %d,  m.Size: %d", index, m.Size())
	}

	b := int(math.Pow(2, float64((m.Size()-index-1)%8)))
	//fmt.Printf("b = %+v\n", b)
	//fmt.Printf("m[index/8] = %+v\n", m[index/8])
	//fmt.Printf("m[index/8]&b = %+v\n", m[index/8]&byte(b))
	if m[index/8]&byte(b) > 0 {
		return 1
	} else {
		return 0
	}
}

func (m *IPMask) BitsLen() (bit0s int, bit1s int) {
	bit0s = 0
	bit1s = 0
	for i := 0; i < m.Size(); i++ {
		if m.GetBit(i) == 0 {
			bit0s++
		} else {
			bit1s++
		}
	}

	return
}

func (m *IPMask) FirstBit(bit0 bool) int {
	if bit0 {
		for i := 0; i < m.Size(); i++ {
			if m.GetBit(i) == 0 {
				return i
			}
		}
		return -1
	} else {
		for i := 0; i < m.Size(); i++ {
			if m.GetBit(i) == 1 {
				return i
			}
		}
		return -1
	}
}

func (m *IPMask) Size() int {

	if m.Type() == IPv4 {
		return 32
	}

	if m.Type() == IPv6 {
		return 128
	}

	return -1

}

func (ip *IPMask) Type() IPFamily {
	if ip.Len() == 4 {
		return IPv4
	} else if ip.Len() == 16 {
		return IPv6
	} else {
		fmt.Printf("ip.Len() = %d\n", ip.Len())
		panic("IPMask Type() failed")
	}

	//if len(ip) == 4 {
	//return IPv4
	//} else if len(ip) == 16 {
	//return IPv6
	//} else {
	//fmt.Printf("len(ip) = %+v\n", len(ip))
	//panic("ip Type() failed")
	////panic(fmt.Sprintf("IPMask ip: %s is invalid", ip))
	//}
}

func (ip *IPMask) Len() int {
	return len(*ip)
}

func (ip *IPMask) Copy() *IPMask {
	p := make(IPMask, ip.Len())
	copy(p, *ip)
	return &p
}

func (ip *IP) AfterMask(mask *IPMask) IP {
	//ip地址与掩码取and以后的结果
	if ip.Type() != mask.Type() {
		return nil
	}
	i := ip.Copy()
	for index, _ := range *i {
		(*i)[index] = (*i)[index] & (*mask)[index]
	}
	return *i
}

func NewIPMaskFromString(s string) (*IPMask, error) {
	ip, err := ParseIP(s)
	if err != nil {
		return nil, err
	}
	//if ip == nil {
	//return nil
	//}
	return IPtoMask(ip)
}

func NewIPMask(prefix uint, fa IPFamily) (*IPMask, error) {
	//fmt.Printf("fa = %+v\n", fa)
	var mask IPMask
	var l uint
	if fa == IPv4 {
		mask = make(IPMask, IPv4Len)
		l = 32
	} else if fa == IPv6 {
		mask = make(IPMask, IPv6Len)
		l = 128
	} else {
		return nil, fmt.Errorf("parameter error, fa = %d", fa)
	}

	for i := uint(1); i <= prefix; i++ {
		index := (i - 1) / 8
		mask[index] = mask[index] | byte(math.Pow(2, float64((l-i)%8)))
	}

	return &mask, nil

}

func (m *IPMask) Iterator(ip *IP) *IPMaskIterator {

	var zero []int
	//for i := 0; i < m.Size(); i++ {
	for i := m.Size() - 1; i >= 0; i-- {
		b := m.GetBit(i)
		if b == 0 {
			zero = append(zero, i)
		}
	}

	if ip.Type() != m.Type() {
		return nil
	}

	size := len(zero)
	count := new(big.Int)
	count.Lsh(big.NewInt(1), uint(size))
	if count.Cmp(big.NewInt(65536)) > 0 {
		return nil
	}
	//fmt.Printf("count = %+v\n", count)
	return &IPMaskIterator{
		ip:    ip,
		zero:  zero,
		mask:  m,
		count: count,
		index: 0,
	}
}

type IPMaskIterator struct {
	ip    *IP
	zero  []int
	mask  *IPMask
	count *big.Int
	index int
}

func (it *IPMaskIterator) HasNext() bool {
	//size := len(i.zero)
	//count := int(math.Pow(2, float64(size)))
	if big.NewInt(int64(it.index)).Cmp(it.count) < 0 {
		return true
	}
	return false
}

func (it *IPMaskIterator) Next() *IP {
	ip := it.ip.Copy()
	for index, _ := range *ip {
		(*ip)[index] = (*ip)[index] & (*it.mask)[index]
	}

	size := len(it.zero)
	//count := int(math.Pow(2, float64(size)))
	for i := it.index; big.NewInt(int64(i)).Cmp(it.count) < 0; i++ {
		//fmt.Printf("i = %+v\n", i)
		//fmt.Printf("size = %+v\n", size)
		if i == 0 {
			it.index++
			return ip.Copy()
		}
		value := ip.Int()
		for j := 0; j < size; j++ {
			mask_bit := 0x1
			mask_bit = mask_bit << uint(j)
			if i&mask_bit > 0 {
				l := new(big.Int)
				l.Lsh(big.NewInt(1), uint(it.mask.Size()-it.zero[j]-1))
				//fmt.Printf("value = %x\n", value)
				//fmt.Printf("l = %x\n", l)
				value = value.Add(value, l)
			}
		}
		it.index++
		return NewIPFromInt(value, it.mask.Type())
	}
	return nil
}

func (m *IPMask) Prefix() int {
	begin := false
	prefix := 0
	for i := 0; i < m.Len(); i++ {
		tmp := byte(0x80)
		for j := 0; j < 8; j++ {
			if !begin {
				if tmp&(*m)[i] == tmp {
					prefix++
				} else {
					begin = true
				}
			} else {
				if tmp&(*m)[i] == tmp {
					return -1
				}
			}
			tmp = tmp >> 1
		}
	}
	return prefix
}

func (m *IPMask) Reverse() *IPMask {
	mask := make(IPMask, m.Len())
	for index, b := range *m {
		mask[index] = b ^ 0xFF
	}
	return &mask
}

func (m *IPMask) String() string {
	return MasktoIP(*m).String()
}

func (ip *IP) Int() *big.Int {
	z := new(big.Int)
	z.SetBytes(*ip)
	return z
}

func (ip *IP) Equal(o IP) bool {
	if ip.Type() != o.Type() {
		return false
	}

	if ip.Int().Cmp(o.Int()) == 0 {
		return true
	}

	return false
}

func (ip *IP) Add(i *big.Int) (*IP, error) {
	high := IPMaxInt(ip.Type())

	add := new(big.Int)
	add = add.Add(ip.Int(), i)
	if add.Cmp(high) > 0 {
		return nil, fmt.Errorf("%d > %d", add, high)
	}
	if add.Cmp(big.NewInt(0)) < 0 {
		return nil, fmt.Errorf("%d < %d", add, 0)
	}

	return NewIPFromInt(add, ip.Type()), nil
}

func NewIPFromInt(i *big.Int, fa IPFamily) *IP {
	var ip IP
	if fa == IPv4 {
		ip = make(IP, IPv4Len)
	} else if fa == IPv6 {
		ip = make(IP, IPv6Len)
	}
	high := IPMaxInt(fa)
	low := big.NewInt(0)
	if i.Cmp(low) < 0 || i.Cmp(high) > 0 {
		return nil
	}

	tmp := i.Bytes()
	last := len(ip) - 1
	for index := len(tmp) - 1; index >= 0; index-- {
		ip[last] = tmp[index]
		last--
	}

	return &ip

}

func ParseIP(s string) (*IP, error) {
	if validator.IsIPv4Address(s) {
		ip := make(IP, IPv4Len)
		tokens := strings.Split(s, ".")

		for index, t := range tokens {
			i, err := strconv.Atoi(t)
			if err != nil {
				return nil, err
			}
			ip[index] = byte(i)
		}
		return &ip, nil
	} else if validator.IsIPv6Address(s) {
		ip := make(IP, IPv6Len)
		ip6 := net.ParseIP(s)
		copy(ip, ip6)
		return &ip, nil

	} else {
		return nil, fmt.Errorf("s:'%s' is not valid ip string", s)
	}

}

func MasktoIP(m IPMask) IP {
	ip := make(IP, len(m))
	copy(ip, m)
	return ip
}

func IPtoMask(ip *IP) (*IPMask, error) {
	mask := make(IPMask, len(*ip))
	copy(mask, *ip)
	return &mask, nil
}

func IPMaxInt(fa IPFamily) *big.Int {
	var ip IP
	if fa == IPv4 {
		ip = make(IP, IPv4Len)
	} else {
		ip = make(IP, IPv6Len)
	}
	for index, _ := range ip {
		ip[index] = 0xFF
	}

	i := new(big.Int)
	i.SetBytes(ip)
	return i
}
