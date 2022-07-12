package network

import (
	"encoding/json"
	"fmt"
	"math/big"

	//"net"
	"testing"

	"github.com/google/go-cmp/cmp"
)

var validIPRangeTests = []struct {
	ip string
	//want *IPRange
	ip1 string
	ip2 string
}{
	//{"127.0.0.1-127.0.0.1", &IPRange{ParseIP("127.0.0.1"), ParseIP("127.0.0.1")}},
	//{"127.0.0.1-127.0.0.2", &IPRange{ParseIP("127.0.0.1"), ParseIP("127.0.0.2")}},
	//{"127.0.0.2-127.0.0.1", nil},
	//{"::127.0.0.1-::128.0.0.2", &IPRange{ParseIP("::127.0.0.1"), ParseIP("::128.0.0.2")}},
	//{"::127.0.0.1-::126.0.0.2", nil},

	{"127.0.0.1-127.0.0.1", "127.0.0.1", "127.0.0.1"},
	{"127.0.0.1-127.0.0.2", "127.0.0.1", "127.0.0.2"},
	{"127.0.0.2-127.0.0.1", "", ""},
	{"::127.0.0.1-::128.0.0.2", "::127.0.0.1", "::128.0.0.2"},
	{"::127.0.0.1-::126.0.0.2", "", ""},
}

func TestIPRange(t *testing.T) {
	for _, tt := range validIPRangeTests {
		var want *IPRange
		if tt.ip1 == "" || tt.ip2 == "" {
			want = nil
		} else {
			ip1, _ := ParseIP(tt.ip1)
			ip2, _ := ParseIP(tt.ip2)

			out, _ := NewIPRange(tt.ip)
			// want = &IPRange{*ip1, *ip2}
			want, _ = NewIPRange(fmt.Sprintf("%s-%s", ip1, ip2))

			if !cmp.Equal(want, out) {
				t.Errorf("ip: %s, got '%+v' want '%+v'", tt.ip, out, want)
			}
		}

	}
}

var validParseIPNetTests = []struct {
	ip   string
	want string
	mask uint
}{
	{"127.0.0.1", "127.0.0.1", 32},
	{"127.0.0.1/1", "127.0.0.1", 1},
	{"127.0.0.1/128.0.0.0", "127.0.0.1", 1},
	{"127.0.0.1/2", "127.0.0.1", 2},
	{"127.0.0.1/192.0.0.0", "127.0.0.1", 2},
	{"127.0.0.1/3", "127.0.0.1", 3},
	{"127.0.0.1/224.0.0.0", "127.0.0.1", 3},
	{"127.0.0.1/4", "127.0.0.1", 4},
	{"127.0.0.1/240.0.0.0", "127.0.0.1", 4},
	{"327.0.0.1/4", "", 0},
}

func TestParseIPNet(t *testing.T) {
	for _, tt := range validParseIPNetTests {
		if tt.want == "" {
		} else {
			out, _ := ParseIPNet(tt.ip)
			ip, _ := ParseIP(tt.want)
			mask, _ := NewIPMask(tt.mask, IPv4)

			want := &IPNet{IP: *ip, Mask: *mask}

			if !cmp.Equal(want, out) {
				t.Errorf("ip: %s, got '%+v' want '%+v'", tt.ip, out, want)
			}

		}

	}
}

var validParseIPTests = []struct {
	ip   string
	want IP
	//ref  net.IP
}{
	{"1.1.1.1",
		IP{0x1, 0x1, 0x1, 0x1}},
	{"255.255.255.1",
		IP{0xff, 0xff, 0xff, 0x1}},
	{"255.255.255.256",
		nil},
	{"255..255.255",
		nil},
	{"255.-1.255.255",
		nil},
	{"::",
		IP{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}},
	{"::1",
		IP{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1}},
	{"::2:1",
		IP{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x2, 0x0, 0x1}},
	{"::2:FF1",
		IP{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x2, 0xF, 0xF1}},
	{"::2:1.1.1.1",
		IP{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x2, 0x1, 0x1, 0x1, 0x1}},
	{"1::2:1.1.1.1",
		IP{0x0, 0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x2, 0x1, 0x1, 0x1, 0x1}},
	{"1::2::1.1.1.1", nil},
	{"1.1.1.1:2::1.1.1.1", nil},
	{"1:2::1231.1.1.1", nil},
	{"1:2FFF::1.1.1.1",
		IP{0x0, 0x1, 0x2F, 0xFF, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1, 0x1, 0x1, 0x1}},
	{"1::1.1.1.1.",
		nil},
	{"G::1.1.1.1",
		nil},
}

func TestParseIP(t *testing.T) {
	for _, tt := range validParseIPTests {
		out, err := ParseIP(tt.ip)
		//var ref net.IP
		//if validator.IsIPv4Address(tt.ip) {
		//ref = tt.ref.To4()
		//} else if validator.IsIPv6Address(tt.ip) {
		//ref = tt.ref.To16()
		//} else {
		//if out != nil {
		//t.Errorf("ip: %s, got '%+v' want '%+v', ref '%+v'", tt.ip, out, tt.want, []byte(ref))
		//}
		//}

		//if bytes.Compare(tt.want, ref) != 0 {
		//t.Errorf("ip: %s, got '%+v' want '%+v', ref '%+v'", tt.ip, out, tt.want, []byte(ref))
		//}
		//if out == nil && tt.want == nil {
		//continue
		//}
		//if out == nil {
		//t.Errorf("ip: %s, got '%+v' want '%+v', ref '%+v'", tt.ip, out, tt.want, []byte(ref))
		//}
		//if tt.want == nil {
		//t.Errorf("ip: %s, got '%+v' want '%+v', ref '%+v'", tt.ip, out, tt.want, []byte(ref))
		//}

		if err != nil {
			if tt.want != nil {
				t.Errorf("ip: %s, got '%+v' want '%+v'", tt.ip, out, tt.want)
			}
		} else {
			if !cmp.Equal(tt.want, *out) {
				t.Errorf("ip: %s, got '%+v' want '%+v'", tt.ip, out, tt.want)
			}
		}
	}
}

var validIPMaskTests = []struct {
	prefix uint
	fa     IPFamily
	want   IPMask
}{
	{1, IPv4, IPMask{0x80, 0x00, 0x00, 0x00}},
	{2, IPv4, IPMask{0xc0, 0x00, 0x00, 0x00}},
	{3, IPv4, IPMask{0xe0, 0x00, 0x00, 0x00}},
	{4, IPv4, IPMask{0xf0, 0x00, 0x00, 0x00}},
	{5, IPv4, IPMask{0xf8, 0x00, 0x00, 0x00}},
	{6, IPv4, IPMask{0xfc, 0x00, 0x00, 0x00}},
	{7, IPv4, IPMask{0xfe, 0x00, 0x00, 0x00}},
	{8, IPv4, IPMask{0xff, 0x00, 0x00, 0x00}},
	{9, IPv4, IPMask{0xff, 0x80, 0x00, 0x00}},
	{10, IPv4, IPMask{0xff, 0xc0, 0x00, 0x00}},
	{11, IPv4, IPMask{0xff, 0xe0, 0x00, 0x00}},
	{12, IPv4, IPMask{0xff, 0xf0, 0x00, 0x00}},
	{13, IPv4, IPMask{0xff, 0xf8, 0x00, 0x00}},
	{14, IPv4, IPMask{0xff, 0xfc, 0x00, 0x00}},
	{15, IPv4, IPMask{0xff, 0xfe, 0x00, 0x00}},
	{16, IPv4, IPMask{0xff, 0xff, 0x00, 0x00}},
	{17, IPv4, IPMask{0xff, 0xff, 0x80, 0x00}},
	{18, IPv4, IPMask{0xff, 0xff, 0xc0, 0x00}},
	{19, IPv4, IPMask{0xff, 0xff, 0xe0, 0x00}},
	{20, IPv4, IPMask{0xff, 0xff, 0xf0, 0x00}},
	{21, IPv4, IPMask{0xff, 0xff, 0xf8, 0x00}},
	{22, IPv4, IPMask{0xff, 0xff, 0xfc, 0x00}},
	{23, IPv4, IPMask{0xff, 0xff, 0xfe, 0x00}},
	{24, IPv4, IPMask{0xff, 0xff, 0xff, 0x00}},
	{25, IPv4, IPMask{0xff, 0xff, 0xff, 0x80}},
	{26, IPv4, IPMask{0xff, 0xff, 0xff, 0xc0}},
	{27, IPv4, IPMask{0xff, 0xff, 0xff, 0xe0}},
	{28, IPv4, IPMask{0xff, 0xff, 0xff, 0xf0}},
	{29, IPv4, IPMask{0xff, 0xff, 0xff, 0xf8}},
	{30, IPv4, IPMask{0xff, 0xff, 0xff, 0xfc}},
	{31, IPv4, IPMask{0xff, 0xff, 0xff, 0xfe}},
	{32, IPv4, IPMask{0xff, 0xff, 0xff, 0xff}},
	{1, IPv6, IPMask{0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
	{2, IPv6, IPMask{0xc0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
	{3, IPv6, IPMask{0xe0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
	{4, IPv6, IPMask{0xf0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
	{5, IPv6, IPMask{0xf8, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
	{6, IPv6, IPMask{0xfc, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
	{7, IPv6, IPMask{0xfe, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
	{8, IPv6, IPMask{0xff, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
	{9, IPv6, IPMask{0xff, 0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
	{10, IPv6, IPMask{0xff, 0xc0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
	{11, IPv6, IPMask{0xff, 0xe0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
	{12, IPv6, IPMask{0xff, 0xf0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
	{13, IPv6, IPMask{0xff, 0xf8, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
	{14, IPv6, IPMask{0xff, 0xfc, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
	{15, IPv6, IPMask{0xff, 0xfe, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
	{16, IPv6, IPMask{0xff, 0xff, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
	{17, IPv6, IPMask{0xff, 0xff, 0x80, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
	{127, IPv6, IPMask{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xfe}},
	{128, IPv6, IPMask{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}},
}

func TestIPMask(t *testing.T) {
	for _, tt := range validIPMaskTests {
		out, _ := NewIPMask(tt.prefix, tt.fa)
		if !cmp.Equal(&tt.want, out) {
			t.Errorf("prefix: %d, got '%+v' want '%+v'", tt.prefix, out, tt.want)
		}
	}
}

var validIPMaskGetBitTests = []struct {
	mask  string
	index int
	want  int
}{
	{"255.255.0.255", -1, -1},
	{"255.255.0.255", 0, 1},
	{"255.255.0.255", 1, 1},
	{"255.255.0.255", 2, 1},
	{"255.255.0.255", 3, 1},
	{"255.255.0.255", 4, 1},
	{"255.255.0.255", 5, 1},
	{"255.255.0.255", 6, 1},
	{"255.255.0.255", 7, 1},
	{"255.255.0.255", 8, 1},
	{"255.255.0.255", 9, 1},
	{"255.255.0.255", 10, 1},
	{"255.255.0.255", 11, 1},
	{"255.255.0.255", 11, 1},
	{"255.255.0.255", 12, 1},
	{"255.255.0.255", 13, 1},
	{"255.255.0.255", 14, 1},
	{"255.255.0.255", 15, 1},
	{"255.255.0.255", 16, 0},
	{"255.255.0.255", 17, 0},
	{"255.255.0.255", 18, 0},
	{"255.255.0.255", 19, 0},
	{"255.255.0.255", 20, 0},
	{"255.255.0.255", 21, 0},
	{"255.255.0.255", 22, 0},
	{"255.255.0.255", 23, 0},
	{"255.255.0.255", 24, 1},
	{"255.255.0.255", 25, 1},
	{"255.255.0.255", 26, 1},
	{"255.255.0.255", 27, 1},
	{"255.255.0.255", 28, 1},
	{"255.255.0.255", 29, 1},
	{"255.255.0.255", 30, 1},
	{"255.255.0.255", 31, 1},
	{"255.255.0.255", 32, -1},
}

func TestIPMaskGetBit(t *testing.T) {
	for _, tt := range validIPMaskGetBitTests {
		ipmask, err := NewIPMaskFromString(tt.mask)
		if err != nil {
			t.Errorf("mask: %s, index: %d, want '%+v'", tt.mask, tt.index, tt.want)
		}
		if out := ipmask.GetBit(tt.index); tt.want != out {
			t.Errorf("mask: %s, index: %d, got '%+v' want '%+v'", tt.mask, tt.index, out, tt.want)
		}
	}
}

var validNetworkTests = []struct {
	ip  string
	ip1 string
	ip2 string
}{
	//{"127.0.0.1-127.0.0.1", &Network{&IPRange{ParseIP("127.0.0.1"), ParseIP("127.0.0.1")}}},
	//{"127.0.0.1/16", &Network{ParseIPNet("127.0.0.1/16")}},
	//{"127.0.0.1-255.255.255.255", &Network{&IPRange{ParseIP("127.0.0.1"), ParseIP("255.255.255.255")}}},

	//{"127.0.0.1-127.0.0.2", &IPRange{ParseIP("127.0.0.1"), ParseIP("127.0.0.2")}},
	//{"127.0.0.2-127.0.0.1", nil},
	//{"::127.0.0.1-::128.0.0.2", &IPRange{ParseIP("::127.0.0.1"), ParseIP("::128.0.0.2")}},
	//{"::127.0.0.1-::126.0.0.2", nil},

	{"127.0.0.1-127.0.0.1", "127.0.0.1", "127.0.0.1"},
	{"127.0.0.1/16", "127.0.0.1/16", ""},
	{"127.0.0.1-255.255.255.255", "127.0.0.1", "255.255.255.255"},
}

func TestParseNetwork(t *testing.T) {
	for _, tt := range validNetworkTests {
		out, _ := NewNetworkFromString(tt.ip)
		var want *Network
		if tt.ip2 == "" {
			net, _ := ParseIPNet(tt.ip1)
			want = &Network{net}
		} else {
			ip1, _ := ParseIP(tt.ip1)
			ip2, _ := ParseIP(tt.ip2)
			want = &Network{&IPRange{*ip1, *ip2}}
		}

		if !cmp.Equal(want, out) {
			t.Errorf("ip: %s, got '%+v' want '%+v'", tt.ip, out, want)
		}
	}
}

var validReverseIPMaskTests = []struct {
	mask string
	want IPMask
}{
	{"255.255.255.0", IPMask{0x00, 0x00, 0x00, 0xFF}},
	{"255.255.0.255", IPMask{0x00, 0x00, 0xFF, 0x00}},
}

func TestReverseIPMask(t *testing.T) {
	for _, tt := range validReverseIPMaskTests {
		mask, err := NewIPMaskFromString(tt.mask)
		if err != nil {
			t.Errorf("mask: %s, want '%+v'", tt.mask, tt.want)
		}
		if out := mask.Reverse(); !cmp.Equal(&tt.want, out) {
			t.Errorf("mask: %s, got '%+v' want '%+v'", tt.mask, out, tt.want)
		}
	}
}

var validIPNetFirstTests = []struct {
	ip   string
	want IP
}{
	{"127.0.0.1/255.255.255.0", IP{0x7F, 0x00, 0x00, 0x00}},
	{"127.0.0.1/255.255.255.255", IP{0x7F, 0x00, 0x00, 0x01}},
	{"128.0.100.1/255.255.0.255", IP{0x80, 0x00, 0x00, 0x01}},
	{"128.1.1.250/16", IP{0x80, 0x01, 0x00, 0x00}},
	{"128.1.1.250/0", IP{0x00, 0x00, 0x00, 0x00}},
	{"0.0.0.0/1", IP{0x00, 0x00, 0x00, 0x00}},
}

func TestIPNetFirst(t *testing.T) {
	for _, tt := range validIPNetFirstTests {
		n, err := ParseIPNet(tt.ip)
		if err != nil {
			t.Errorf("ip: %s, want '%+v', err: %s", tt.ip, tt.want, err)
		}
		if out := n.First(); !cmp.Equal(&tt.want, out) {
			t.Errorf("ip: %s, got '%+v' want '%+v'", tt.ip, out, tt.want)
		}
	}
}

var validIPNetLastTests = []struct {
	ip   string
	want IP
}{
	{"127.0.0.1/255.255.255.0", IP{0x7F, 0x00, 0x00, 0xFF}},
	{"127.0.0.1/255.255.255.255", IP{0x7F, 0x00, 0x00, 0x01}},
	{"128.1.1.250/16", IP{0x80, 0x01, 0xFF, 0xFF}},
	{"128.0.100.1/255.255.0.255", IP{0x80, 0x00, 0xFF, 0x01}},
	{"0.1.1.250/0", IP{0xFF, 0xFF, 0xFF, 0xFF}},
	{"0.0.0.0/0", IP{0xFF, 0xFF, 0xFF, 0xFF}},
	{"0.0.0.0/1", IP{0x7F, 0xFF, 0xFF, 0xFF}},
}

func TestIPNetLast(t *testing.T) {
	for _, tt := range validIPNetLastTests {
		n, err := ParseIPNet(tt.ip)
		if err != nil {
			t.Errorf("ip: %s, want '%+v'", tt.ip, tt.want)
		}
		if out := n.Last(); !cmp.Equal(&tt.want, out) {
			t.Errorf("ip: %s, got '%+v' want '%+v'", tt.ip, out, tt.want)
		}
	}
}

var validIPtoIntTests = []struct {
	ip string
	i  string
}{
	{"2.0.0.1", "0x02000001"},
	{"1::2.0.0.1", "0x00010000000000000000000002000001"},
	{"F1::2.0.0.1", "0x00F10000000000000000000002000001"},
	{"F1F1::2.0.0.1", "0xF1F10000000000000000000002000001"},
}

func TestIPtoInit(t *testing.T) {
	for _, tt := range validIPtoIntTests {
		ip, err := ParseIP(tt.ip)

		if err != nil {
			t.Errorf("ip: %s,  i: %s", tt.ip, tt.i)
		}
		//fmt.Printf("ip = %+v\n", ip)
		i := new(big.Int)
		fmt.Sscan(tt.i, i)

		if out := ip.Int(); i.Cmp(out) != 0 {
			t.Errorf("ip: %s, got '%+v' want '%+v'", tt.ip, out, i)
		}
	}
}

var validIPNettoRangeTests = []struct {
	ip  string
	ip1 string
	ip2 string
}{
	{"127.0.0.1/24", "127.0.0.0", "127.0.0.255"},
	{"::127.0.0.1/120", "::127.0.0.0", "::127.0.0.255"},
}

func TestIPNettoRange(t *testing.T) {
	for _, tt := range validIPNettoRangeTests {
		n, _ := ParseIPNet(tt.ip)
		var want *IPRange
		ip1, _ := ParseIP(tt.ip1)
		ip2, _ := ParseIP(tt.ip2)
		want = &IPRange{*ip1, *ip2}
		if out := n.ToRange(); !cmp.Equal(want, out) {
			t.Errorf("ip: %s, got '%+v' want '%+v'", tt.ip, out, want)
		}
	}
}

var validIntToIPTests = []struct {
	int  string
	want IP
	fa   IPFamily
}{
	{"0x80000001", IP{0x80, 0x00, 0x00, 0x01}, IPv4},
	{"0xF0000001", IP{0xF0, 0x00, 0x00, 0x01}, IPv4},
	{"0xF0000001000000000000000000001100", IP{0xF0, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x11, 0x00}, IPv6},
}

func TestIntToIP(t *testing.T) {
	for _, tt := range validIntToIPTests {
		var bi big.Int
		fmt.Sscan(tt.int, &bi)

		if out := NewIPFromInt(&bi, tt.fa); !cmp.Equal(&tt.want, out) {
			t.Errorf("int: %s, got '%+v' want '%+v'", tt.int, out, tt.want)
		}
	}
}

var validIPMaskPrefixTests = []struct {
	mask string
	want int
}{
	{"0.0.0.0", 0},
	{"128.0.0.0", 1},
	{"192.0.0.0", 2},
	{"224.0.0.0", 3},
	{"240.0.0.0", 4},
	{"248.0.0.0", 5},
	{"252.0.0.0", 6},
	{"254.0.0.0", 7},
	{"255.0.0.0", 8},
	{"0.255.0.0", -1},
	{"254.255.0.0", -1},
	{"255.0.1.0", -1},
	{"255.128.0.0", 9},
	{"255.192.0.0", 10},
	{"255.224.0.0", 11},
	{"255.224.0.255", -1},
	{"255.254.0.0", 15},
	{"255.253.0.0", -1},
	{"FFFF:FFFF::", 32},
	{"FFFF:FFFF:EFFF::", -1},
	{"FFFF:FFFF:FFFF::", 48},
	{"FFFF:FFFF:FF00::", 40},
	{"FFFF:FFFF:FFFF:FFFF:FFFF:FFFF:FFFF:FFFF", 128},
	{"0000:0000:0000:0000:0000:0000:0000:0000", 0},
}

func TestIPMaskPrefix(t *testing.T) {
	for _, tt := range validIPMaskPrefixTests {
		m, _ := NewIPMaskFromString(tt.mask)
		if out := m.Prefix(); !cmp.Equal(tt.want, out) {
			t.Errorf("mask: %s, got '%+v' want '%+v'", tt.mask, out, tt.want)
		}
	}
}

var validSuperNetTests = []struct {
	ip     string
	want   string
	prefix uint
}{
	{"127.0.0.1/24", "127.0.0.0", 23},
	{"127.0.0.1/1", "0.0.0.0", 0},
	{"127.0.0.1/128.0.0.0", "0.0.0.0", 0},
	{"127.0.0.1/2", "0.0.0.0", 1},
	{"127.0.0.1/0", "", 0},
	{"127.0.0.1/0.255.0.255", "", 0},
	{"127.0.0.1/128.255.0.255", "0.0.0.0", 0},
}

func TestSuperNet(t *testing.T) {
	for _, tt := range validSuperNetTests {
		n, _ := ParseIPNet(tt.ip)
		out, _ := n.SuperNet()
		if tt.want != "" {
			var want *IPNet
			wnet, _ := ParseIP(tt.want)
			mask, _ := NewIPMask(tt.prefix, IPv4)
			want = &IPNet{IP: *wnet, Mask: *mask}

			if !cmp.Equal(want, out) {
				t.Errorf("ip: %s, got '%+v' want '%+v'", tt.ip, out, want)
			}
		} else {
		}
	}
}

var validIPNetMatchIPRangeTests = []struct {
	ip   string
	r    string
	want bool
}{
	{"127.0.0.1/24", "127.0.0.0-127.0.0.255", true},
	{"127.0.0.1/24", "127.0.0.1-127.0.0.255", true},
	{"127.0.0.1/24", "127.0.0.1-127.0.1.0", false},
	{"127.0.0.1/25", "127.0.0.1-127.0.0.127", true},
	{"127.0.0.1/25", "127.0.0.1-127.0.0.128", false},
	{"127.0.0.1/25", "127.0.0.100-127.0.0.127", true},
	{"128.0.0.1/255.255.0.255", "128.0.0.1-128.0.0.1", true},
	{"128.0.0.1/255.255.0.255", "128.0.0.1-128.0.0.2", false},
	{"128.0.0.1/255.0.255.0", "128.0.0.0-128.0.0.255", true},
	{"128.0.0.1/255.0.255.0", "128.1.0.0-128.1.0.255", true},
	{"128.0.0.1/255.0.255.0", "128.255.0.0-128.255.0.255", true},
	{"128.0.0.1/0.0.255.0", "128.255.0.0-128.255.0.255", true},
}

func TestIPNetMatchIPRange(t *testing.T) {
	for _, tt := range validIPNetMatchIPRangeTests {
		n, _ := ParseIPNet(tt.ip)
		r, _ := NewIPRange(tt.r)
		if out := n.MatchIPRange(r); !cmp.Equal(tt.want, out) {
			t.Errorf("ip: %s, got '%+v' want '%+v'", tt.ip, out, tt.want)
		}
	}
}

var validIPNetMatchIPNetTests = []struct {
	ip   string
	r    string
	want bool
}{
	{"127.0.0.1/24", "127.0.0.0/24", true},
	{"127.0.0.1/24", "127.0.0.1/24", true},
	{"127.0.0.1/24", "127.0.0.1/23", false},
	{"127.0.0.1/25", "127.0.0.1/25", true},
	{"127.0.0.1/25", "127.0.0.128/25", false},
	{"127.0.0.1/8", "127.0.0.1/255.255.0.128", true},
}

func TestIPNetMatchIPNet(t *testing.T) {
	for _, tt := range validIPNetMatchIPNetTests {
		n, _ := ParseIPNet(tt.ip)
		r, _ := ParseIPNet(tt.r)
		if out := n.MatchIPNet(r); !cmp.Equal(tt.want, out) {
			t.Errorf("ip: %s, got '%+v' want '%+v'", tt.ip, out, tt.want)
		}
	}
}

var validIPRangeMatchIPRangeTests = []struct {
	ip   string
	r    string
	want bool
}{
	{"127.0.0.1-127.0.0.5", "127.0.0.0-127.0.0.5", false},
	{"127.0.0.1-127.0.0.5", "127.0.0.2-127.0.0.5", true},
	{"127.0.0.1-127.0.0.5", "127.0.0.2-127.0.0.6", false},
	//{"127.0.0.1/24", "127.0.0.1-127.0.1.0", false},
	//{"127.0.0.1/25", "127.0.0.1-127.0.0.127", true},
	//{"127.0.0.1/25", "127.0.0.1-127.0.0.128", false},
	//{"127.0.0.1/25", "127.0.0.100-127.0.0.127", true},
}

func TestIPRangeMatchIPRange(t *testing.T) {
	for _, tt := range validIPRangeMatchIPRangeTests {
		n, _ := NewIPRange(tt.ip)
		r, _ := NewIPRange(tt.r)
		if out := n.MatchIPRange(r); !cmp.Equal(tt.want, out) {
			t.Errorf("ip: %s, got '%+v' want '%+v'", tt.ip, out, tt.want)
		}
	}
}

var validIPRangeMatchIPNetTests = []struct {
	ip   string
	r    string
	want bool
}{
	{"127.0.0.1-127.0.0.5", "127.0.0.0/30", false},
	{"127.0.0.0-127.0.0.5", "127.0.0.0/30", true},
	{"127.0.0.0-127.255.255.255", "127.0.0.0/255.0.255.0", true},
	//{"127.0.0.1-127.0.0.5", "127.0.0.2-127.0.0.5", true},
	//{"127.0.0.1-127.0.0.5", "127.0.0.2-127.0.0.6", false},
	//{"127.0.0.1/24", "127.0.0.1-127.0.1.0", false},
	//{"127.0.0.1/25", "127.0.0.1-127.0.0.127", true},
	//{"127.0.0.1/25", "127.0.0.1-127.0.0.128", false},
	//{"127.0.0.1/25", "127.0.0.100-127.0.0.127", true},
}

func TestIPRangeMatchIPNet(t *testing.T) {
	for _, tt := range validIPRangeMatchIPNetTests {
		n, _ := NewIPRange(tt.ip)
		r, _ := ParseIPNet(tt.r)
		if out := n.MatchIPNet(r); !cmp.Equal(tt.want, out) {
			t.Errorf("ip: %s, got '%+v' want '%+v'", tt.ip, out, tt.want)
		}
	}
}

var validIPNetCountTests = []struct {
	ip    string
	count string
}{
	{"127.0.0.1/24", "256"},
	{"127.0.0.1/32", "1"},
	{"0.0.0.0/0", "4294967296"},
}

func TestIPNetCount(t *testing.T) {
	for _, tt := range validIPNetCountTests {
		n, _ := ParseIPNet(tt.ip)
		c := new(big.Int)
		fmt.Sscan(tt.count, c)

		if out := n.Count(); out.Cmp(c) != 0 {
			t.Errorf("ip: %s, got '%+v' want '%+v'", tt.ip, out, tt.count)
		}
	}
}

var validIPAddTests = []struct {
	ip   string
	add  string
	want string
}{
	{"127.0.0.1", "256", "127.0.1.1"},
	{"127.0.0.1", "2", "127.0.0.3"},
	{"0.0.0.1", "-2", "0.0.0.-3"},
	{"0.0.0.1", "-1", "0.0.0.0"},
}

func TestIPAdd(t *testing.T) {
	for _, tt := range validIPAddTests {
		b := new(big.Int)
		fmt.Sscan(tt.add, b)
		ip, _ := ParseIP(tt.ip)
		want, err := ParseIP(tt.want)
		out, err2 := ip.Add(b)
		if err != nil && err2 != nil {

		} else {
			if !cmp.Equal(want, out) {
				t.Errorf("ip: %s, add: %s, got '%+v' want '%+v'", tt.ip, tt.add, out, want)
			}
		}
	}
}

var validIPRangeToIPNetListTests = []struct {
	ip string
	ns []string
}{
	{"127.0.0.0-127.0.1.127", []string{"127.0.0.0/24", "127.0.1.0/25"}},
	{"127.0.0.0-127.0.1.128", []string{"127.0.0.0/24", "127.0.1.0/25", "127.0.1.128/32"}},
	{"127.0.0.1-127.0.1.128", []string{"127.0.0.1/32", "127.0.0.2/31", "127.0.0.4/30", "127.0.0.8/29",
		"127.0.0.16/28", "127.0.0.32/27", "127.0.0.64/26", "127.0.0.128/25", "127.0.1.0/25", "127.0.1.128/32"}},
}

func TestIPRangeToIPNetList(t *testing.T) {
	for _, tt := range validIPRangeToIPNetListTests {
		r, _ := NewIPRange(tt.ip)

		ns := []*IPNet{}
		for _, s := range tt.ns {
			net, _ := ParseIPNet(s)
			ns = append(ns, net)
			//fmt.Printf("ParseIPNet(s) = %+v\n", ParseIPNet(s))
		}

		if out, _ := r.IPNetList(); !cmp.Equal(ns, out) {
			t.Errorf("ip: %s, got '%+v' want '%+v'", tt.ip, out, tt.ns)
		}
	}
}

var validIPRangeStartShiftTests = []struct {
	ip    string
	shift string
	want  string
}{
	{"127.0.0.1-127.0.0.5", "1", "127.0.0.2-127.0.0.5"},
	{"127.0.0.1-127.0.0.5", "10", "127.0.0.20-127.0.0.5"},
	{"127.0.0.1-127.0.0.5", "4", "127.0.0.5-127.0.0.5"},
	//{"127.0.0.0-127.0.0.5", "127.0.0.0/30", true},
	//{"127.0.0.1-127.0.0.5", "127.0.0.2-127.0.0.5", true},
	//{"127.0.0.1-127.0.0.5", "127.0.0.2-127.0.0.6", false},
	//{"127.0.0.1/24", "127.0.0.1-127.0.1.0", false},
	//{"127.0.0.1/25", "127.0.0.1-127.0.0.127", true},
	//{"127.0.0.1/25", "127.0.0.1-127.0.0.128", false},
	//{"127.0.0.1/25", "127.0.0.100-127.0.0.127", true},
}

func TestIPRangeStartShift(t *testing.T) {
	for _, tt := range validIPRangeStartShiftTests {
		r, _ := NewIPRange(tt.ip)
		d := new(big.Int)
		fmt.Sscan(tt.shift, d)
		want, _ := NewIPRange(tt.want)
		if out, _ := r.StartShift(d); !cmp.Equal(want, out) {
			t.Errorf("ip: %s, got '%+v' want '%+v'", tt.ip, out, tt.want)
		}
	}
}

var validIPRangeEndShiftTests = []struct {
	ip    string
	shift string
	want  string
}{
	{"127.0.0.1-127.0.0.5", "1", "127.0.0.1-127.0.0.6"},
	{"127.0.0.1-127.0.0.5", "10", "127.0.0.1-127.0.0.15"},
	{"127.0.0.1-127.0.0.5", "4", "127.0.0.1-127.0.0.9"},
	//{"127.0.0.0-127.0.0.5", "127.0.0.0/30", true},
	//{"127.0.0.1-127.0.0.5", "127.0.0.2-127.0.0.5", true},
	//{"127.0.0.1-127.0.0.5", "127.0.0.2-127.0.0.6", false},
	//{"127.0.0.1/24", "127.0.0.1-127.0.1.0", false},
	//{"127.0.0.1/25", "127.0.0.1-127.0.0.127", true},
	//{"127.0.0.1/25", "127.0.0.1-127.0.0.128", false},
	//{"127.0.0.1/25", "127.0.0.100-127.0.0.127", true},
}

func TestIPRangeEndShift(t *testing.T) {
	for _, tt := range validIPRangeEndShiftTests {
		r, _ := NewIPRange(tt.ip)
		d := new(big.Int)
		fmt.Sscan(tt.shift, d)
		want, _ := NewIPRange(tt.want)
		if out, _ := r.EndShift(d); !cmp.Equal(want, out) {
			t.Errorf("ip: %s, got '%+v' want '%+v'", tt.ip, out, tt.want)
		}
	}
}

var validIPMaskIteratorTests = []struct {
	mask string
	ip   string
	want []string
}{
	{"255.255.255.127", "192.168.1.5", []string{"192.168.1.5", "192.168.1.133"}},
	{"255.127.255.127", "192.168.1.5", []string{"192.40.1.5", "192.40.1.133", "192.168.1.5", "192.168.1.133"}},
	{"254.127.255.127", "192.168.1.5", []string{
		"192.40.1.5", "192.40.1.133", "192.168.1.5", "192.168.1.133",
		"193.40.1.5", "193.40.1.133", "193.168.1.5", "193.168.1.133"}},
	{"FFFF:FFFF:FFFF:FFFF:FFFF:FFFF:255.255.127.255", "1234::192.168.1.5", []string{"1234::192.168.1.5", "1234::192.168.129.5"}},
	{"7FFF:FFFF:FFFF:FFFF:FFFF:FFFF:255.255.127.255", "1234::192.168.1.5",
		[]string{"1234::192.168.1.5", "1234::192.168.129.5", "9234::192.168.1.5", "9234::192.168.129.5"}},
	{"255.255.255.254", "192.168.1.5", []string{"192.168.1.4", "192.168.1.5"}},
	{"255.255.255.252", "192.168.1.5", []string{"192.168.1.4", "192.168.1.5", "192.168.1.6", "192.168.1.7"}},
	{"255.255.255.248", "192.168.1.5",
		[]string{"192.168.1.0", "192.168.1.1", "192.168.1.2", "192.168.1.3", "192.168.1.4", "192.168.1.5", "192.168.1.6", "192.168.1.7"}},
}

func TestIPMaskIterator(t *testing.T) {
	for _, tt := range validIPMaskIteratorTests {
		mask, _ := NewIPMaskFromString(tt.mask)
		ip, _ := ParseIP(tt.ip)
		var wants []*IP
		for _, ss := range tt.want {
			i, _ := ParseIP(ss)
			wants = append(wants, i)
		}

		var gots []*IP
		for it := mask.Iterator(ip); it.HasNext(); {
			t := it.Next()
			gots = append(gots, t)
		}
		if !cmp.Equal(wants, gots) {
			t.Errorf("mask: %s, ip: %s, got '%+v' want '%+v'", tt.mask, tt.ip, gots, wants)
		}
	}
}

func BenchmarkIPMaskIterator(b *testing.B) {
	mask, _ := NewIPMaskFromString("255.0.0.255")
	ip, _ := ParseIP("192.168.1.1")
	for it := mask.Iterator(ip); it.HasNext(); {
		//fmt.Printf("it.index = %+v\n", it.index)
		_ = it.Next()
	}
}

var validIPRangeIteratorTests = []struct {
	ip   string
	want []string
}{
	{"192.168.1.5-192.168.1.6", []string{"192.168.1.5", "192.168.1.6"}},
	{"192.168.1.0-192.168.1.6", []string{"192.168.1.0", "192.168.1.1", "192.168.1.2",
		"192.168.1.3", "192.168.1.4", "192.168.1.5", "192.168.1.6"}},
}

func TestIPRangeIterator(t *testing.T) {
	for _, tt := range validIPRangeIteratorTests {
		ir, _ := NewIPRange(tt.ip)
		var wants []*IP
		for _, ss := range tt.want {
			i, _ := ParseIP(ss)
			wants = append(wants, i)
		}

		var gots []*IP
		for it := ir.Iterator(); it.HasNext(); {
			t := it.Next()
			gots = append(gots, t)
		}
		if !cmp.Equal(wants, gots) {
			t.Errorf("ip: %s, got '%+v' want '%+v'", tt.ip, gots, wants)
		}
	}
}

var validIPNetIteratorTests = []struct {
	ip   string
	want []string
}{
	{"192.168.1.4/31", []string{"192.168.1.4", "192.168.1.5"}},
	{"192.168.1.0/29", []string{"192.168.1.0", "192.168.1.1", "192.168.1.2",
		"192.168.1.3", "192.168.1.4", "192.168.1.5", "192.168.1.6", "192.168.1.7"}},
}

func TestIPNetIterator(t *testing.T) {
	for _, tt := range validIPNetIteratorTests {
		ir, _ := ParseIPNet(tt.ip)
		var wants []*IP
		for _, ss := range tt.want {
			i, _ := ParseIP(ss)
			wants = append(wants, i)
		}

		var gots []*IP
		for it := ir.Iterator(); it.HasNext(); {
			t := it.Next()
			gots = append(gots, t)
		}
		if !cmp.Equal(wants, gots) {
			t.Errorf("ip: %s, got '%+v' want '%+v'", tt.ip, gots, wants)
		}
	}
}

var validIPRangeSuperNetTests = []struct {
	ip   string
	want string
}{
	{"192.168.1.1-192.168.1.3", "192.168.1.0/30"},
	{"192.168.1.0-192.168.1.200", "192.168.1.0/24"},
	{"1.168.1.0-192.168.1.200", "0.0.0.0/0"},
	{"192.168.1.0-192.168.1.255", "192.168.1.0/24"},
}

func TestIPRangeSuperNet(t *testing.T) {
	for _, tt := range validIPRangeSuperNetTests {
		ip, _ := NewIPRange(tt.ip)
		want, _ := ParseIPNet(tt.want)

		if out, _ := ip.SuperNet(); !cmp.Equal(out, want) {
			t.Errorf("ip: %s, got '%+v' want '%+v'", tt.ip, out, want)
		}
	}
}

var validNetworkIteratorTests = []struct {
	ip   string
	want []string
}{
	{"192.168.1.5-192.168.1.6", []string{"192.168.1.5", "192.168.1.6"}},
	{"192.168.1.0-192.168.1.6", []string{"192.168.1.0", "192.168.1.1", "192.168.1.2",
		"192.168.1.3", "192.168.1.4", "192.168.1.5", "192.168.1.6"}},
	{"192.168.1.4/30", []string{"192.168.1.4", "192.168.1.5", "192.168.1.6", "192.168.1.7"}},
}

func TestNetworkIterator(t *testing.T) {
	for _, tt := range validNetworkIteratorTests {
		network, _ := NewNetworkFromString(tt.ip)

		var wants []*IP
		for _, ss := range tt.want {
			i, _ := ParseIP(ss)
			wants = append(wants, i)
		}

		var gots []*IP
		for it := network.Iterator(); it.HasNext(); {
			t := it.Next()
			gots = append(gots, t)
		}

		if !cmp.Equal(wants, gots) {
			t.Errorf("ip: %s, got '%+v' want '%+v'", tt.ip, gots, wants)
		}

	}
}

var validNetworkSuperNetTests = []struct {
	ip   string
	want string
}{
	{"192.168.1.1-192.168.1.2", "192.168.1.0/30"},
	{"192.168.1.0-192.168.1.200", "192.168.1.0/24"},
	{"1.168.1.0-192.168.1.200", "0.0.0.0/0"},
	{"192.168.1.0/24", "192.168.0.0/23"},
}

func TestNetworkSuperNet(t *testing.T) {
	for _, tt := range validNetworkSuperNetTests {
		ip, _ := NewNetworkFromString(tt.ip)
		want, _ := ParseIPNet(tt.want)

		if out, _ := ip.SuperNet(); !cmp.Equal(out, want) {
			t.Errorf("ip: %s, got '%+v' want '%+v'", tt.ip, out, want)
		}
	}
}

var validIPNetStringTests = []struct {
	ip   string
	want string
}{
	{"192.168.1.0/30", "192.168.1.0/30"},
	{"192.168.1.0/255.255.0.255", "192.168.1.0/255.255.0.255"},
	{"::192.168.1.0/::255.255.0.255", "::c0a8:100/::ffff:ff"},
	{"::192.168.1.0/120", "::c0a8:100/120"},
}

func TestIPNetString(t *testing.T) {
	for _, tt := range validIPNetStringTests {
		ip, _ := ParseIPNet(tt.ip)
		//want := ParseIPNet(tt.want)

		if out := ip.String(); !cmp.Equal(out, tt.want) {
			t.Errorf("ip: %s, got '%+v' want '%+v'", tt.ip, out, tt.want)
		}
	}
}

var validNetworkListFirstLastTests = []struct {
	netlist []string
	first   string
	last    string
}{
	{[]string{"192.168.1.1-192.168.1.253", "10.1.1.1/24"}, "10.1.1.0", "192.168.1.253"},
	{[]string{"::192.168.1.1-::192.168.1.253", "::10.1.1.1/120"}, "::10.1.1.0", "::192.168.1.253"},
	{[]string{"192.168.1.1-192.168.1.253", "10.1.1.1/24", "10.1.100.1/16"}, "10.1.0.0", "192.168.1.253"},
}

func TestNetworkListFirstLast(t *testing.T) {
	for _, tt := range validNetworkListFirstLastTests {
		list := []*Network{}
		for _, n := range tt.netlist {
			net, _ := NewNetworkFromString(n)
			list = append(list, net)
		}
		netlist, _ := NewNetworkListFromList(list)

		first, _ := ParseIP(tt.first)
		last, _ := ParseIP(tt.last)

		got_first := netlist.First()
		got_last := netlist.Last()
		if !cmp.Equal(got_first, first) || !cmp.Equal(got_last, last) {
			t.Errorf("ip: %s, want_first '%+v' want_last '%+v', got_first '%+v' got_last '%+v'",
				tt.netlist, tt.first, tt.last, got_first, got_last)
		}
	}
}

var validNetworkListCountTests = []struct {
	netlist []string
	want    *big.Int
}{
	{[]string{"192.168.1.1-192.168.1.253", "10.1.1.1/24"}, big.NewInt(253 + 256)},
	{[]string{"::192.168.1.1-::192.168.1.253", "::10.1.1.1/120"}, big.NewInt(253 + 256)},
	{[]string{"192.168.1.1-192.168.1.253", "10.1.1.1/24", "10.1.100.1/16"}, big.NewInt(253 + 65536)},
}

func TestNetworkListCount(t *testing.T) {
	for _, tt := range validNetworkListCountTests {
		list := []*Network{}
		for _, n := range tt.netlist {
			net, _ := NewNetworkFromString(n)
			list = append(list, net)
		}
		netlist, _ := NewNetworkListFromList(list)

		got := netlist.Count()
		if got.Cmp(tt.want) != 0 {
			t.Errorf("ip: %s, got '%+v' want '%+v'", tt.netlist, got, tt.want)
		}
	}
}

var validNetworkListMatchIPNetTests = []struct {
	netlist []string
	net     string
	want    bool
}{
	{[]string{"192.168.1.1-192.168.1.253", "10.1.1.1/24"}, "192.168.1.128/26", true},
	{[]string{"192.168.1.1-192.168.1.253", "10.1.1.1/24"}, "192.168.1.128/25", false},
	{[]string{"192.168.1.1-192.168.1.255", "10.1.1.1/24"}, "192.168.1.128/25", true},
	{[]string{"192.168.0.0-192.168.1.255", "192.168.2.0/24", "192.168.3.0/24"}, "192.168.0.128/22", true},
	{[]string{"::192.168.1.1-::192.168.1.253", "::10.1.1.1/120"}, "::10.1.1.0/120", true},
	//{[]string{"192.168.1.1-192.168.1.253", "10.1.1.1/24", "10.1.100.1/16"}, big.NewInt(253 + 256 + 65536)},
}

func TestNetworkListMatchIPNet(t *testing.T) {
	for _, tt := range validNetworkListMatchIPNetTests {
		list := []*Network{}
		for _, n := range tt.netlist {
			net, _ := NewNetworkFromString(n)
			list = append(list, net)
		}
		netlist, _ := NewNetworkListFromList(list)

		net, _ := ParseIPNet(tt.net)
		got := netlist.MatchIPNet(net)
		if got != tt.want {
			t.Errorf("ip: %s, got '%+v' want '%+v'", tt.netlist, got, tt.want)
		}
	}
}

var validNetworkListMatchIPRangeTests = []struct {
	netlist []string
	net     string
	want    bool
}{
	{[]string{"192.168.1.1-192.168.1.253", "10.1.1.1/24"}, "192.168.1.1-192.168.1.254", false},
	{[]string{"192.168.1.1-192.168.1.253", "10.1.1.1/24"}, "192.168.1.100-192.168.1.253", true},
	//{[]string{"192.168.1.1-192.168.1.253", "10.1.1.1/24"}, "192.168.1.128/25", false},
	//{[]string{"192.168.1.1-192.168.1.255", "10.1.1.1/24"}, "192.168.1.128/25", true},
	//{[]string{"192.168.0.0-192.168.1.255", "192.168.2.0/24", "192.168.3.0/24"}, "192.168.0.128/22", true},
	//{[]string{"::192.168.1.1-::192.168.1.253", "::10.1.1.1/120"}, "::10.1.1.0/120", true},
	//{[]string{"192.168.1.1-192.168.1.253", "10.1.1.1/24", "10.1.100.1/16"}, big.NewInt(253 + 256 + 65536)},
}

func TestNetworkListMatchIPRange(t *testing.T) {
	for _, tt := range validNetworkListMatchIPRangeTests {
		list := []*Network{}
		for _, n := range tt.netlist {
			net, _ := NewNetworkFromString(n)
			list = append(list, net)
		}
		netlist, _ := NewNetworkListFromList(list)

		net, _ := NewIPRange(tt.net)
		got := netlist.MatchIPRange(net)
		if got != tt.want {
			t.Errorf("ip: %s, got '%+v' want '%+v'", tt.netlist, got, tt.want)
		}
	}
}

var validNetworkListMatchStringTests = []struct {
	netlist []string
	net     string
	want    bool
}{
	{[]string{"192.168.1.1-192.168.1.253", "10.1.1.1/24"}, "192.168.1.1-192.168.1.254", false},
	{[]string{"192.168.1.1-192.168.1.253", "10.1.1.1/24"}, "192.168.1.100-192.168.1.253", true},
}

func TestNetworkListMatchString(t *testing.T) {
	for _, tt := range validNetworkListMatchStringTests {
		list := []*Network{}
		for _, n := range tt.netlist {
			net, _ := NewNetworkFromString(n)
			list = append(list, net)
		}
		netlist, _ := NewNetworkListFromList(list)

		got, err := netlist.MatchString(tt.net)
		if err != nil {
			t.Errorf("ip: %s, got '%+v' want '%+v', err: %s", tt.netlist, got, tt.want, err)
		}
		if got != tt.want {
			t.Errorf("ip: %s, got '%+v' want '%+v'", tt.netlist, got, tt.want)
		}
	}
}

var validNetworkListAggregateTests = []struct {
	netlist []string
	net     string
	list    []string
}{
	{[]string{"192.168.1.0-192.168.1.253", "192.168.0.0/24"}, "192.168.0.0-192.168.1.253", nil},
	{[]string{"192.168.1.0-192.168.1.253", "10.1.0.0/16", "192.168.0.0/24"}, "", []string{"192.168.0.0-192.168.1.253", "10.1.0.0/16"}},
	{[]string{"192.168.1.0-192.168.3.255", "10.1.0.0/16", "192.168.0.0/24"}, "", []string{"192.168.0.0/22", "10.1.0.0/16"}},
}

func TestNetworkListAggregate(t *testing.T) {
	for _, tt := range validNetworkListAggregateTests {
		list := []*Network{}
		for _, n := range tt.netlist {
			net, _ := NewNetworkFromString(n)
			list = append(list, net)
		}
		netlist, _ := NewNetworkListFromList(list)
		agg_net, agg_list := netlist.Aggregate()
		var want_net *Network
		if tt.net != "" {
			want_net, _ = NewNetworkFromString(tt.net)
		}

		var want_list *NetworkList
		if tt.list != nil {
			ls := []*Network{}
			for _, n := range tt.list {
				net, _ := NewNetworkFromString(n)
				ls = append(ls, net)
			}
			want_list, _ = NewNetworkListFromList(ls)
		}

		if agg_net != nil {
			if !(agg_net.Same(want_net)) {
				t.Errorf("ip: %s, got '%+v' want '%+v'", tt.netlist, agg_net, want_net)
			}
		}

		if agg_list != nil {
			if !(agg_list.Same(want_list)) {
				t.Errorf("ip: %s, got '%+v' want '%+v'", tt.netlist, agg_list, want_list)
			}
		}

		//if got != tt.want {
		//t.Errorf("ip: %s, got '%+v' want '%+v'", tt.netlist, got, tt.want)
		//}
	}
}

var validNetworkListSuperNetTests = []struct {
	netlist []string
	net     string
}{
	{[]string{"192.168.1.0-192.168.1.253", "192.168.0.0/24"}, "192.168.0.0/23"},
	{[]string{"192.168.1.0-192.168.3.253", "192.168.0.0/24"}, "192.168.0.0/22"},
	{[]string{"192.168.1.0/32", "192.168.3.253/32", "192.168.0.0/24"}, "192.168.0.0/22"},
}

func TestNetworkListSuperNet(t *testing.T) {
	for _, tt := range validNetworkListSuperNetTests {
		list := []*Network{}
		for _, n := range tt.netlist {
			net, _ := NewNetworkFromString(n)
			list = append(list, net)
		}
		netlist, _ := NewNetworkListFromList(list)
		net, _ := netlist.SuperNet()

		want_net, _ := ParseIPNet(tt.net)
		if !(net.MatchIPNet(want_net) && want_net.MatchIPNet(net)) {
			t.Errorf("ip: %s, got '%+v' want '%+v'", tt.netlist, net, want_net)
		}
	}
}

var validIPMaskBitsLenTests = []struct {
	mask  string
	bit1s int
	bit0s int
}{
	{"255.255.0.255", 24, 8},
	{"127.255.0.255", 23, 9},
}

func TestIPMaskBitsLen(t *testing.T) {
	for _, ss := range validIPMaskBitsLenTests {
		mask, _ := NewIPMaskFromString(ss.mask)
		bit0s, bit1s := mask.BitsLen()
		if (bit1s != ss.bit1s) || (bit0s != ss.bit0s) {
			t.Errorf("mask: %s, got 1s '%d %d', want '%d %d'", ss.mask, bit1s, bit0s, ss.bit1s, ss.bit0s)
		}
	}

}

var validNetworkListCmpTests = []struct {
	t     IPFamily
	this  *[]string
	other *[]string
	left  *[]string
	mid   *[]string
	right *[]string
}{
	{
		IPv4,
		&[]string{"192.168.1.0/24", "192.168.3.0/24"},
		&[]string{"192.168.1.200-192.168.3.200"},
		&[]string{"192.168.1.0-192.168.1.199", "192.168.3.201-192.168.3.255"},
		&[]string{"192.168.1.200-192.168.1.255", "192.168.3.0-192.168.3.200"},
		&[]string{"192.168.2.0-192.168.2.255"},
	},
	{
		IPv4,
		&[]string{"10.0.0.0/8"},
		//&[]string{"10.148.254.0/23", "10.233.252.0/27"},
		&[]string{"10.148.254.0/23"},
		&[]string{"10.0.0.0-10.148.253.255", "10.149.0.0-10.255.255.255"},
		&[]string{"10.148.254.0/23"},
		//&[]string{"10.148.254.0/23", "10.233.252.0/27"},
		//&[]string{"192.168.1.200-192.168.1.255", "192.168.3.0-192.168.3.200"},
		nil,
		//&[]string{},
	},
	{
		IPv4,
		&[]string{"10.0.0.0/8"},
		&[]string{"10.148.254.0/23", "10.233.252.0/27"},
		&[]string{"10.0.0.0-10.148.253.255", "10.149.0.0-10.233.251.255", "10.233.252.32-10.255.255.255"},
		&[]string{"10.148.254.0/23", "10.233.252.0/27"},
		//&[]string{"10.148.254.0/23", "10.233.252.0/27"},
		//&[]string{"192.168.1.200-192.168.1.255", "192.168.3.0-192.168.3.200"},
		nil,
		//&[]string{},
	},
	{
		IPv4,
		&[]string{"10.0.0.0/8", "100.65.248.96/32"}, //"100.65.248.97/32"},
		&[]string{"10.148.254.0/23", "10.233.252.0/27"},
		&[]string{"10.0.0.0-10.148.253.255", "10.149.0.0-10.233.251.255",
			"10.233.252.32-10.255.255.255", "100.65.248.96-100.65.248.96"},
		&[]string{"10.148.254.0/23", "10.233.252.0/27"},
		//&[]string{"10.148.254.0/23", "10.233.252.0/27"},
		//&[]string{"192.168.1.200-192.168.1.255", "192.168.3.0-192.168.3.200"},
		nil,
		//&[]string{},
	},

	{
		IPv4,
		&[]string{"10.0.0.0/8"},
		//&[]string{"10.148.254.0/23", "10.233.252.0/27"},
		&[]string{"10.148.254.0/23"},
		&[]string{"10.0.0.0/9", "10.128.0.0/12", "10.144.0.0/14", "10.148.0.0/17",
			"10.148.128.0/18", "10.148.192.0/19", "10.148.224.0/20", "10.148.240.0/21",
			"10.148.248.0/22", "10.148.252.0/23", "10.149.0.0/16", "10.150.0.0/15",
			"10.152.0.0/13", "10.160.0.0/11", "10.192.0.0/10"},
		&[]string{"10.148.254.0/23"},
		//&[]string{"10.148.254.0/23", "10.233.252.0/27"},
		//&[]string{"192.168.1.200-192.168.1.255", "192.168.3.0-192.168.3.200"},
		nil,
		//&[]string{},
	},

	{
		IPv6,
		&[]string{"::192.168.1.0/120", "::192.168.3.0/120"},
		&[]string{"::192.168.1.200-::192.168.3.200"},
		&[]string{"::192.168.1.0-::192.168.1.199", "::192.168.3.201-::192.168.3.255"},
		&[]string{"::192.168.1.200-::192.168.1.255", "::192.168.3.0-::192.168.3.200"},
		&[]string{"::192.168.2.0-::192.168.2.255"},
	},
	{
		IPv4,
		&[]string{"192.168.1.0/24", "192.168.3.0/24", "192.168.5.0/24"},
		&[]string{"192.168.1.200-192.168.3.200", "192.168.3.254-192.168.5.1"},
		&[]string{"192.168.1.0-192.168.1.199", "192.168.3.201-192.168.3.253", "192.168.5.2-192.168.5.255"},
		&[]string{"192.168.1.200-192.168.1.255", "192.168.3.0-192.168.3.200", "192.168.3.254-192.168.3.255", "192.168.5.0-192.168.5.1"},
		&[]string{"192.168.2.0-192.168.2.255", "192.168.4.0-192.168.4.255"},
	},
}

func TestNetworkListCmp(t *testing.T) {
	for _, ss := range validNetworkListCmpTests {
		this_net := []*Network{}
		for _, s := range *ss.this {
			net, _ := NewNetworkFromString(s)
			this_net = append(this_net, net)
		}
		this, _ := NewNetworkListFromList(this_net)

		other_net := []*Network{}
		for _, s := range *ss.other {
			net, _ := NewNetworkFromString(s)
			other_net = append(other_net, net)
		}
		other, _ := NewNetworkListFromList(other_net)

		var left *NetworkList
		if ss.left == nil {
			left = nil
		} else {
			left_net := []*Network{}
			for _, s := range *ss.left {
				net, _ := NewNetworkFromString(s)
				left_net = append(left_net, net)
			}
			left, _ = NewNetworkListFromList(left_net)
		}

		var mid *NetworkList
		if ss.mid == nil {
			mid = nil
		} else {
			mid_net := []*Network{}
			for _, s := range *ss.mid {
				net, _ := NewNetworkFromString(s)
				mid_net = append(mid_net, net)
			}
			mid, _ = NewNetworkListFromList(mid_net)
		}

		var right *NetworkList
		if ss.right == nil {
			right = nil
		} else {
			right_net := []*Network{}
			for _, s := range *ss.right {
				net, _ := NewNetworkFromString(s)
				right_net = append(right_net, net)
			}
			right, _ = NewNetworkListFromList(right_net)
		}

		got_left, got_mid, got_right := NetworkListCmp(*this, *other)
		if left == nil {
			if got_left != nil {
				t.Errorf("s:'%+v', got_left:'%+v', left:'%+v'", ss, got_left, left)
			}
		} else {
			if got_left == nil {
				t.Errorf("s:'%+v', got_left:'%+v', left:'%+v'", ss, got_left, left)
			} else {
				if !got_left.Same(left) {
					t.Errorf("s:'%+v', got_left:'%+v', left:'%+v'", ss, got_left, left)
				}
			}
		}
		if mid == nil {
			if got_mid != nil {
				t.Errorf("s:'%+v', got_mid:'%+v', mid:'%+v'", ss, got_mid, mid)
			}
		} else {
			if got_mid == nil {
				t.Errorf("s:'%+v', got_mid:'%+v', mid:'%+v'", ss, got_mid, mid)
			} else {
				if !got_mid.Same(mid) {
					t.Errorf("s:'%+v', got_mid:'%+v', mid:'%+v'", ss, got_mid, mid)
				}

			}
		}
		if right == nil {
			if got_right != nil {
				t.Errorf("s:'%+v', got_right:'%+v', right:'%+v'", ss, got_right, right)
			}
		} else {
			if got_right == nil {
				t.Errorf("s:'%+v', got_right:'%+v', right:'%+v'", ss, got_right, right)
			} else {
				if !got_right.Same(right) {
					t.Errorf("s:'%+v', got_right:'%+v', right:'%+v'", ss, got_right, right)
				}
			}
		}

	}
}

type gs struct {
	ipv4 []string
	ipv6 []string
}

var validNetworkGroupCmpTests = []struct {
	this  *gs
	other *gs
	left  *gs
	mid   *gs
	right *gs
}{
	{
		&gs{
			[]string{"192.168.1.0/24", "192.168.3.0/24"},
			[]string{"::192.168.1.0/120", "::192.168.3.0/120"},
		},
		&gs{
			[]string{"192.168.1.1-192.168.3.254"},
			[]string{},
		},
		&gs{
			[]string{"192.168.1.0", "192.168.3.255"},
			[]string{"::192.168.1.0/120", "::192.168.3.0/120"},
		},
		&gs{
			[]string{"192.168.1.1-192.168.1.255", "192.168.3.0-192.168.3.254"},
			[]string{},
		},
		&gs{
			[]string{"192.168.2.0-192.168.2.255"},
			[]string{},
		},
	},
	{
		&gs{
			[]string{"192.168.1.0/24", "192.168.3.0/24"},
			[]string{"::192.168.1.0/120", "::192.168.3.0/120"},
		},
		&gs{
			[]string{"192.168.1.1-192.168.3.254"},
			[]string{"::192.168.1.1-::192.168.1.50", "::192.168.2.50-::192.168.3.100"},
		},
		&gs{
			[]string{"192.168.1.0", "192.168.3.255"},
			[]string{"::192.168.1.0", "::192.168.1.51-::192.168.1.255", "::192.168.3.101-::192.168.3.255"},
		},
		&gs{
			[]string{"192.168.1.1-192.168.1.255", "192.168.3.0-192.168.3.254"},
			[]string{"::192.168.1.1-::192.168.1.50", "::192.168.3.0-::192.168.3.100"},
		},
		&gs{
			[]string{"192.168.2.0-192.168.2.255"},
			[]string{"::192.168.2.50-::192.168.2.255"},
		},
	},
	{
		&gs{

			[]string{"1.1.1.1", "2.2.2.2", "6.6.6.6", "100.100.100.100", "33.33.33.33"},
			[]string{},
		},
		&gs{
			[]string{"1.1.1.0-3.255.255.255"},
			[]string{},
		},
		&gs{
			[]string{"6.6.6.6", "100.100.100.100", "33.33.33.33"},
			[]string{},
		},
		&gs{
			[]string{"1.1.1.1", "2.2.2.2"},
			[]string{},
		},
		&gs{
			[]string{"1.1.1.0", "1.1.1.2-2.2.2.1", "2.2.2.3-3.255.255.255"},
			[]string{},
		},
	},
}

func TestNetworkGroupCmp(t *testing.T) {
	for _, ss := range validNetworkGroupCmpTests {
		var this *NetworkGroup
		{
			this_ipv4_net := []*Network{}
			for _, s := range ss.this.ipv4 {
				net, _ := NewNetworkFromString(s)
				this_ipv4_net = append(this_ipv4_net, net)
			}
			this_ipv4, _ := NewNetworkListFromList(this_ipv4_net)

			this_ipv6_net := []*Network{}
			for _, s := range ss.this.ipv6 {
				net, _ := NewNetworkFromString(s)
				this_ipv6_net = append(this_ipv6_net, net)
			}
			this_ipv6, _ := NewNetworkListFromList(this_ipv6_net)

			this = &NetworkGroup{
				*this_ipv4,
				*this_ipv6,
			}
		}
		var other *NetworkGroup
		{
			other_ipv4_net := []*Network{}
			for _, s := range ss.other.ipv4 {
				net, _ := NewNetworkFromString(s)
				other_ipv4_net = append(other_ipv4_net, net)
			}
			other_ipv4, _ := NewNetworkListFromList(other_ipv4_net)

			other_ipv6_net := []*Network{}
			for _, s := range ss.other.ipv6 {
				net, _ := NewNetworkFromString(s)
				other_ipv6_net = append(other_ipv6_net, net)
			}
			other_ipv6, _ := NewNetworkListFromList(other_ipv6_net)

			other = &NetworkGroup{
				*other_ipv4,
				*other_ipv6,
			}
		}

		var left *NetworkGroup
		if ss.left != nil {
			left_ipv4_net := []*Network{}
			for _, s := range ss.left.ipv4 {
				net, _ := NewNetworkFromString(s)
				left_ipv4_net = append(left_ipv4_net, net)
			}
			left_ipv4, _ := NewNetworkListFromList(left_ipv4_net)

			left_ipv6_net := []*Network{}
			for _, s := range ss.left.ipv6 {
				net, _ := NewNetworkFromString(s)
				left_ipv6_net = append(left_ipv6_net, net)
			}
			left_ipv6, _ := NewNetworkListFromList(left_ipv6_net)

			left = &NetworkGroup{
				*left_ipv4,
				*left_ipv6,
			}
		} else {
			left = nil
		}

		var mid *NetworkGroup
		if ss.mid != nil {
			mid_ipv4_net := []*Network{}
			for _, s := range ss.mid.ipv4 {
				net, _ := NewNetworkFromString(s)
				mid_ipv4_net = append(mid_ipv4_net, net)
			}
			mid_ipv4, _ := NewNetworkListFromList(mid_ipv4_net)

			mid_ipv6_net := []*Network{}
			for _, s := range ss.mid.ipv6 {
				net, _ := NewNetworkFromString(s)
				mid_ipv6_net = append(mid_ipv6_net, net)
			}
			mid_ipv6, _ := NewNetworkListFromList(mid_ipv6_net)

			mid = &NetworkGroup{
				*mid_ipv4,
				*mid_ipv6,
			}
		} else {
			mid = nil
		}

		var right *NetworkGroup
		if ss.right != nil {
			right_ipv4_net := []*Network{}
			for _, s := range ss.right.ipv4 {
				net, _ := NewNetworkFromString(s)
				right_ipv4_net = append(right_ipv4_net, net)
			}
			right_ipv4, _ := NewNetworkListFromList(right_ipv4_net)

			right_ipv6_net := []*Network{}
			for _, s := range ss.right.ipv6 {
				net, _ := NewNetworkFromString(s)
				right_ipv6_net = append(right_ipv6_net, net)
			}
			right_ipv6, _ := NewNetworkListFromList(right_ipv6_net)

			right = &NetworkGroup{
				*right_ipv4,
				*right_ipv6,
			}
		} else {
			right = nil
		}

		got_left, got_mid, got_right := NetworkGroupCmp(*this, *other)
		if left == nil {
			if got_left != nil {
				t.Errorf("s:'%+v', got_left:'%+v', left:'%+v'", ss, got_left, left)
			}
		} else {
			if !got_left.Same(left) {
				t.Errorf("s:'%+v', got_left:'%+v', left:'%+v'", ss, got_left, left)
			}
		}

		if mid == nil {
			if got_mid != nil {
				t.Errorf("s:'%+v', got_mid:'%+v', mid:'%+v'", ss, got_mid, mid)
			}
		} else {
			if !got_mid.Same(mid) {
				t.Errorf("s:'%+v', got_mid:'%+v', mid:'%+v'", ss, got_mid, mid)
			}
		}

		if right == nil {
			if got_right != nil {
				t.Errorf("s:'%+v', got_right:'%+v', right:'%+v'", ss, got_right, right)
			}
		} else {
			if !got_right.Same(right) {
				t.Errorf("s:'%+v', got_right:'%+v', right:'%+v'", ss, got_right, right)
			}
		}

	}
}

var validNetworkGroupFromStringCmpTests = []struct {
	this string
	want *gs
}{
	{
		this: "1.1.1.0/24,1.1.1.2-2.2.2.1,2.2.2.3-3.255.255.255",
		want: &gs{
			[]string{"1.1.1.0/24", "1.1.2.0-2.2.2.1", "2.2.2.3-3.255.255.255"},
			[]string{},
		},
	},
	{
		this: "2001::1.1.1.0/120,2001::1.1.1.2-2001::2.2.2.1,2001::2.2.2.3-2001::3.255.255.255",
		want: &gs{
			[]string{},
			[]string{"2001::1.1.1.0/120", "2001::1.1.2.0-2001::2.2.2.1", "2001::2.2.2.3-2001::3.255.255.255"},
		},
	},
	{
		this: "1.1.1.0/24,1.1.1.2-2.2.2.1,2.2.2.3-3.255.255.255,2001::1.1.1.1/120",
		want: &gs{
			[]string{"1.1.1.0/24", "1.1.2.0-2.2.2.1", "2.2.2.3-3.255.255.255"},
			[]string{"2001::1.1.1.0/120"},
		},
	},
	// {
	// this: "",
	// want: &gs{
	// []string{},
	// []string{},
	// },
	// },
}

func TestNetworkGroupFromStringCmp(t *testing.T) {
	for _, ss := range validNetworkGroupFromStringCmpTests {
		got, err := NewNetworkGroupFromString(ss.this)
		if err != nil {
			fmt.Println("err")
		}

		want_ipv4_net := []*Network{}
		for _, s := range ss.want.ipv4 {
			net, _ := NewNetworkFromString(s)
			want_ipv4_net = append(want_ipv4_net, net)
		}
		want_ipv4, _ := NewNetworkListFromList(want_ipv4_net)

		want_ipv6_net := []*Network{}
		for _, s := range ss.want.ipv6 {
			net, _ := NewNetworkFromString(s)
			want_ipv6_net = append(want_ipv6_net, net)
		}
		want_ipv6, _ := NewNetworkListFromList(want_ipv6_net)

		want := &NetworkGroup{
			*want_ipv4,
			*want_ipv6,
		}
		fmt.Printf("want: %+v\n", want)
		fmt.Printf("got: %+v\n", got)

		if !want.Same(got) {
			t.Errorf("got: %+v, want: %+v\n", got, want)
		}
	}
}

func TestJSON(t *testing.T) {
	ip, _ := ParseIP("192.168.1.1")
	b, err := json.Marshal(ip)
	if err != nil {
		t.Errorf("%s", err)
	}
	var ii IP
	err = json.Unmarshal(b, &ii)
	if err != nil {
		t.Errorf("%s", err)
	}

	net, err := ParseIPNet("172.16.0.5/24")
	if err != nil {
		t.Errorf("%s", err)
	}
	b, err = json.Marshal(net)
	var nn IPNet
	err = json.Unmarshal(b, &nn)
	if err != nil {
		t.Errorf("%s", err)
	}

	rg, _ := NewIPRange("172.16.5.1-172.16.5.10")
	b, err = json.Marshal(rg)
	if err != nil {
		t.Errorf("%s", err)
	}
	var rr IPRange
	err = json.Unmarshal(b, &rr)
	if err != nil {
		t.Errorf("%s", err)
	}

	net2, _ := NewNetworkFromString("172.1.1.5/24")
	b, err = json.Marshal(net2)
	if err != nil {
		t.Errorf("%s", err)
	}
	var n2 Network
	err = json.Unmarshal(b, &n2)
	if err != nil {
		t.Errorf("%s", err)
	}

	net3, _ := NewNetworkFromString("10.1.1.0/24")
	nl, _ := NewNetworkListFromList([]*Network{net2, net3})
	b, err = json.Marshal(nl)
	if err != nil {
		t.Errorf("%s", err)
	}

	var n3 NetworkList
	err = json.Unmarshal(b, &n3)
	if err != nil {
		t.Errorf("%s", err)
	}

	ng := NewNetworkGroup()
	ng.Add(net3)
	ng.Add(net2)

	b, err = json.Marshal(ng)
	if err != nil {
		t.Errorf("%s", err)
	}
	var n4 NetworkGroup
	err = json.Unmarshal(b, &n4)
	if err != nil {
		t.Errorf("%s", err)
	}
	//fmt.Printf("n4 = %+v\n", n4)
}

var validNetworkGroupTranslate = []struct {
	input      *gs
	translator *gs
	want       *gs
}{
	{
		input: &gs{
			// ipv4: []string{"1.1.1.1"},
		},
		translator: &gs{
			// ipv4: []string{"2.2.2.2"},
		},
		want: &gs{
			// ipv4: []string{"2.2.2.2"},
		},
	},
	{
		input: &gs{
			// ipv4: []string{"1.1.1.1"},
		},
		translator: &gs{
			ipv4: []string{"0.0.0.0/0"},
			ipv6: []string{"::/0"},
		},
		want: &gs{
			// ipv4: []string{"2.2.2.2"},
		},
	},
	{
		input: &gs{
			ipv4: []string{"1.1.1.1"},
		},
		translator: &gs{
			ipv4: []string{"0.0.0.0/0"},
			ipv6: []string{"::/0"},
		},
		want: &gs{
			ipv4: []string{"1.1.1.1"},
		},
	},
	{
		input: &gs{
			ipv4: []string{"1.1.1.0/24"},
		},
		translator: &gs{
			ipv4: []string{"0.0.0.0/0"},
			ipv6: []string{"::/0"},
		},
		want: &gs{
			ipv4: []string{"1.1.1.0/24"},
		},
	},
	{
		input: &gs{
			ipv4: []string{"0.0.0.0/0"},
		},
		translator: &gs{
			ipv4: []string{"0.0.0.0/0"},
			ipv6: []string{"::/0"},
		},
		want: &gs{
			ipv4: []string{"0.0.0.0/0"},
		},
	},

	{
		input: &gs{
			ipv4: []string{"1.1.1.1"},
		},
		translator: &gs{
			ipv4: []string{"2.2.2.2"},
		},
		want: &gs{
			ipv4: []string{"2.2.2.2"},
		},
	},
	{
		input: &gs{
			ipv4: []string{"1.1.1.1"},
		},
		translator: &gs{
			// ipv4: []string{"2.2.2.2"},
		},
		want: &gs{
			ipv4: []string{"1.1.1.1"},
		},
	},
	{
		input: &gs{
			ipv4: []string{"1.1.1.1"},
			ipv6: []string{"::1.1.1.1"},
		},
		translator: &gs{
			ipv4: []string{"2.2.2.2"},
			ipv6: []string{"::2.2.2.2"},
		},
		want: &gs{
			ipv4: []string{"2.2.2.2"},
			ipv6: []string{"::2.2.2.2"},
		},
	},
	{
		input: &gs{
			ipv6: []string{"::1.1.1.1"},
		},
		translator: &gs{
			ipv4: []string{"2.2.2.2"},
			ipv6: []string{"::2.2.2.2"},
		},
		want: &gs{
			ipv6: []string{"::2.2.2.2"},
		},
	},
	{
		input: &gs{
			ipv4: []string{"1.1.1.0/24"},
		},
		translator: &gs{
			ipv4: []string{"0.0.0.0/0"},
			ipv6: []string{"::2.2.2.2"},
		},
		want: &gs{
			ipv4: []string{"1.1.1.0/24"},
		},
	},
	{
		input: &gs{
			ipv4: []string{"1.1.1.0/24"},
		},
		translator: &gs{
			ipv4: []string{"1.1.1.0-1.1.1.255"},
			ipv6: []string{"::2.2.2.2"},
		},
		want: &gs{
			ipv4: []string{"1.1.1.0/24"},
		},
	},
	{
		input: &gs{
			ipv4: []string{"1.1.1.0/24"},
		},
		translator: &gs{
			ipv4: []string{"1.1.1.1-1.1.1.255"},
			ipv6: []string{"::2.2.2.2"},
		},
		want: &gs{
			ipv4: []string{"1.1.1.1-1.1.1.255"},
		},
	},
	{
		input: &gs{
			ipv6: []string{"::1.1.1.0/120"},
		},
		translator: &gs{
			ipv4: []string{"2.2.2.2"},
			ipv6: []string{"::1.1.1.0-::1.1.1.255"},
		},
		want: &gs{
			ipv6: []string{"::1.1.1.0/120"},
		},
	},
	{
		input: &gs{
			ipv6: []string{"::1.1.1.0/120"},
		},
		translator: &gs{
			ipv4: []string{"2.2.2.2"},
			ipv6: []string{"::1.1.1.1-::1.1.1.255"},
		},
		want: &gs{
			ipv6: []string{"::1.1.1.1-::1.1.1.255"},
		},
	},
}

func TestNetworkGroupTranslate(t *testing.T) {
	for _, ss := range validNetworkGroupTranslate {
		input_ipv4_net := []*Network{}
		for _, s := range ss.input.ipv4 {
			net, _ := NewNetworkFromString(s)
			input_ipv4_net = append(input_ipv4_net, net)
		}
		input_ipv4, _ := NewNetworkListFromList(input_ipv4_net)

		input_ipv6_net := []*Network{}
		for _, s := range ss.input.ipv6 {
			net, _ := NewNetworkFromString(s)
			input_ipv6_net = append(input_ipv6_net, net)
		}
		input_ipv6, _ := NewNetworkListFromList(input_ipv6_net)

		input := &NetworkGroup{
			*input_ipv4,
			*input_ipv6,
		}

		translator_ipv4_net := []*Network{}
		for _, s := range ss.translator.ipv4 {
			net, _ := NewNetworkFromString(s)
			translator_ipv4_net = append(translator_ipv4_net, net)
		}
		translator_ipv4, _ := NewNetworkListFromList(translator_ipv4_net)

		translator_ipv6_net := []*Network{}
		for _, s := range ss.translator.ipv6 {
			net, _ := NewNetworkFromString(s)
			translator_ipv6_net = append(translator_ipv6_net, net)
		}
		translator_ipv6, _ := NewNetworkListFromList(translator_ipv6_net)

		translator := &NetworkGroup{
			*translator_ipv4,
			*translator_ipv6,
		}

		want_ipv4_net := []*Network{}
		for _, s := range ss.want.ipv4 {
			net, _ := NewNetworkFromString(s)
			want_ipv4_net = append(want_ipv4_net, net)
		}
		want_ipv4, _ := NewNetworkListFromList(want_ipv4_net)

		want_ipv6_net := []*Network{}
		for _, s := range ss.want.ipv6 {
			net, _ := NewNetworkFromString(s)
			want_ipv6_net = append(want_ipv6_net, net)
		}
		want_ipv6, _ := NewNetworkListFromList(want_ipv6_net)

		want := &NetworkGroup{
			*want_ipv4,
			*want_ipv6,
		}

		got, err := Translate(input, translator)
		if err != nil {
			t.Error(err)
		}
		if !want.Same(got) {
			t.Errorf("input:%+v, translator:%+v, want:%+v, got:%+v", ss.input, ss.translator, ss.want, got)
		}

	}
}

var validRangeCidrs = []struct {
	ips   string
	wants []string
}{
	{
		ips: "1.1.1.1-1.1.1.1",
		wants: []string{
			"1.1.1.1/32",
		},
	},
	{
		ips: "1.1.1.1-1.1.1.2",
		wants: []string{
			"1.1.1.1/32",
			"1.1.1.2/32",
		},
	},
	{
		ips: "1.1.1.1-1.1.1.3",
		wants: []string{
			"1.1.1.1/32",
			"1.1.1.2/31",
		},
	},
	{
		ips: "1.1.1.1-1.1.1.4",
		wants: []string{
			"1.1.1.1/32",
			"1.1.1.2/31",
			"1.1.1.4/32",
		},
	},
	{
		ips: "1.1.1.1-1.1.1.5",
		wants: []string{
			"1.1.1.1/32",
			"1.1.1.2/31",
			"1.1.1.4/31",
		},
	},
	{
		ips: "1.1.1.1-1.1.1.6",
		wants: []string{
			"1.1.1.1/32",
			"1.1.1.2/31",
			"1.1.1.4/31",
			"1.1.1.6/32",
		},
	},
	{
		ips: "1.1.1.1-1.1.1.7",
		wants: []string{
			"1.1.1.1/32",
			"1.1.1.2/31",
			"1.1.1.4/30",
		},
	},
	{
		ips: "1.1.1.1-1.1.1.16",
		wants: []string{
			"1.1.1.1/32",
			"1.1.1.2/31",
			"1.1.1.4/30",
			"1.1.1.8/29",
			"1.1.1.16/32",
		},
	},
	{
		ips: "1.1.1.1-1.1.1.255",
		wants: []string{
			"1.1.1.1/32",
			"1.1.1.2/31",
			"1.1.1.4/30",
			"1.1.1.8/29",
			"1.1.1.16/28",
			"1.1.1.32/27",
			"1.1.1.64/26",
			"1.1.1.128/25",
		},
	},
	{
		ips: "0.0.0.0-255.255.255.255",
		wants: []string{
			"0.0.0.0/0",
		},
	},
	{
		ips: "1.1.1.1-1.1.255.255",
		wants: []string{
			"1.1.1.1/32",
			"1.1.1.2/31",
			"1.1.1.4/30",
			"1.1.1.8/29",
			"1.1.1.16/28",
			"1.1.1.32/27",
			"1.1.1.64/26",
			"1.1.1.128/25",
			"1.1.2.0/23",
			"1.1.4.0/22",
			"1.1.8.0/21",
			"1.1.16.0/20",
			"1.1.32.0/19",
			"1.1.64.0/18",
			"1.1.128.0/17",
		},
	},
	{
		ips: "1.1.1.1-1.255.255.255",
		wants: []string{
			"1.1.1.1/32",
			"1.1.1.2/31",
			"1.1.1.4/30",
			"1.1.1.8/29",
			"1.1.1.16/28",
			"1.1.1.32/27",
			"1.1.1.64/26",
			"1.1.1.128/25",
			"1.1.2.0/23",
			"1.1.4.0/22",
			"1.1.8.0/21",
			"1.1.16.0/20",
			"1.1.32.0/19",
			"1.1.64.0/18",
			"1.1.128.0/17",
			"1.2.0.0/15",
			"1.4.0.0/14",
			"1.8.0.0/13",
			"1.16.0.0/12",
			"1.32.0.0/11",
			"1.64.0.0/10",
			"1.128.0.0/9",
		},
	},
	{
		ips: "1.1.1.0-1.1.2.0",
		wants: []string{
			"1.1.1.0/24",
			"1.1.2.0/32",
		},
	},
	{
		ips: "::1.1.1.0-::1.1.2.0",
		wants: []string{
			"::1.1.1.0/120",
			"::1.1.2.0/128",
		},
	},
	{
		ips: "::1.1.1.1-::1.1.255.255",
		wants: []string{
			"::1.1.1.1/128",
			"::1.1.1.2/127",
			"::1.1.1.4/126",
			"::1.1.1.8/125",
			"::1.1.1.16/124",
			"::1.1.1.32/123",
			"::1.1.1.64/122",
			"::1.1.1.128/121",
			"::1.1.2.0/119",
			"::1.1.4.0/118",
			"::1.1.8.0/117",
			"::1.1.16.0/116",
			"::1.1.32.0/115",
			"::1.1.64.0/114",
			"::1.1.128.0/113",
		},
	},
	{
		ips: "::1.1.1.0-::1.1.2.0",
		wants: []string{
			"::1.1.1.0/120",
			"::1.1.2.0/128",
		},
	},
}

func TestRangeCIDRs(t *testing.T) {
	for _, ss := range validRangeCidrs {
		ipRange, _ := NewIPRange(ss.ips)

		wants := []*IPNet{}
		for _, w := range ss.wants {
			net, _ := ParseIPNet(w)
			wants = append(wants, net)
		}
		gots := ipRange.CIDRs()
		if len(gots) != len(wants) {
			t.Errorf("ips:%s, wants:%s, gots:%+v", ss.ips, ss.wants, gots)
		}

		for index, got := range gots {
			if !got.MatchIPNet(wants[index]) || !wants[index].MatchIPNet(got) {
				t.Errorf("ips:%s, wants:%s, gots:%+v", ss.ips, ss.wants, gots)
			}

		}
	}
}

func TestIPRangeToIPNetList2(t *testing.T) {
	for _, tt := range validRangeCidrs {
		r, _ := NewIPRange(tt.ips)

		ns := []*IPNet{}
		for _, s := range tt.wants {
			net, _ := ParseIPNet(s)
			ns = append(ns, net)
		}

		if out, _ := r.IPNetList(); !cmp.Equal(ns, out) {
			t.Errorf("ip: %s, got '%+v' want '%+v'", tt.ips, out, tt.wants)
		}
	}
}

var validNetworkGroupMatch = []struct {
	one  string
	two  string
	want bool
}{
	{
		one:  "172.28.62.0/24",
		two:  "172.28.62.114/32",
		want: true,
	},
}

func TestNetworkGroupMatch(t *testing.T) {
	for _, ss := range validNetworkGroupMatch {
		one, _ := NewNetworkGroupFromString(ss.one)
		two, _ := NewNetworkGroupFromString(ss.two)

		got := one.MatchNetworkGroup(two)
		if got != ss.want {
			t.Errorf("one: %s, two: %s, want:%v, got:%v", ss.one, ss.two, ss.want, got)
		}
	}
}
