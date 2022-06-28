package network

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"reflect"
	"strings"
	"tools/flexrange"
	"tools/utils"
	"tools/validator"

	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type AbbrNetType int

const (
	OBJECT_OPTION_ONE_MATCH_TWO = 0x1
	OBJECT_OPTION_TWO_MATCH_ONE = 0x2
	OBJECT_OPTION_SAME          = 0x4
)

const (
	_ AbbrNetType = iota
	IPNET
	IPRANGE
	NETWORK
	NETWORKLIST
)

func InSameNetwork(one string, two string) (bool, error) {
	net1, err := ParseIPNet(one)
	if err != nil {
		return false, err
	}

	net2, err := ParseIPNet(two)
	if err != nil {
		return false, err
	}

	if net1.MatchIPNet(net2) {
		return true, nil
	}

	return false, fmt.Errorf("one: %s not same network wtih two: %s", one, two)
}

type Network struct {
	AbbrNet
}

func (net Network) Copy() utils.CopyAble {
	var n AbbrNet
	n = net.AbbrNet.Copy().(AbbrNet)
	return &Network{
		n,
	}
}

func (net Network) IPNet() (*IPNet, bool) {
	return net.AbbrNet.IPNet()
}

func (net Network) MarshalJSON() ([]byte, error) {
	data := struct {
		Type AbbrNetType
		Data []byte
	}{}

	var d []byte
	var err error
	switch net.AbbrNet.(type) {
	case *IPNet:
		d, err = json.Marshal(net.AbbrNet.(*IPNet))
		if err != nil {
			return nil, err
		}
		data.Type = IPNET
	case *IPRange:
		d, err = json.Marshal(net.AbbrNet.(*IPRange))
		if err != nil {
			return nil, err
		}
		data.Type = IPRANGE
	case *Network:
		d, err = json.Marshal(net.AbbrNet.(*Network))
		if err != nil {
			return nil, err
		}
		data.Type = NETWORK
	case *NetworkList:
		d, err = json.Marshal(net.AbbrNet.(*NetworkList))
		if err != nil {
			return nil, err
		}
		data.Type = NETWORKLIST
	default:
		panic(fmt.Sprint("unknown error", reflect.TypeOf(net.AbbrNet)))
	}

	data.Data = append(data.Data, d...)

	return json.Marshal(&data)
}

//
func (net *Network) UnmarshalJSON(b []byte) error {
	data := struct {
		Type AbbrNetType
		Data []byte
	}{}

	err := json.Unmarshal(b, &data)
	if err != nil {
		return err
	}

	if data.Type == IPNET {
		var n IPNet
		err = json.Unmarshal(data.Data, &n)
		if err != nil {
			return err
		}
		net.AbbrNet = &n
	} else if data.Type == IPRANGE {
		var n IPRange
		err = json.Unmarshal(data.Data, &n)
		if err != nil {
			return err
		}
		net.AbbrNet = &n
	} else if data.Type == NETWORK {
		var n Network
		err = json.Unmarshal(data.Data, &n)
		if err != nil {
			return err
		}
		net.AbbrNet = &n
	} else if data.Type == NETWORKLIST {
		var n NetworkList
		err = json.Unmarshal(data.Data, &n)
		if err != nil {
			return err
		}
		net.AbbrNet = &n

	} else {
		fmt.Println(data.Type)
		panic("unknown error")
	}

	return nil
}

func (net *Network) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New(fmt.Sprint("Failed to unmarshal IPNet value:", value))
	}

	err := json.Unmarshal(bytes, net)
	return err
}

func (net Network) Value() (driver.Value, error) {
	return json.Marshal(&net)
}

func (Network) GormDataType() string {
	return "network"
}

func (net Network) Match(n AbbrNet) bool {
	dr := net.DataRange()
	other := n.DataRange()
	return dr.Match(other)
}

func (net Network) Same(n AbbrNet) bool {
	dr := net.DataRange()
	other := n.DataRange()

	if !dr.Match(other) {
		return false
	}

	return other.Match(dr)
}

type NetworkList struct {
	family IPFamily
	list   []*Network
}

func (nl *NetworkList) AddressType() AddressType {
	if len(nl.list) == 0 {
		return ADDRESS_NONE
	}
	net, netList := nl.Aggregate()
	if net == nil && netList == nil {
		return ADDRESS_NONE
	}
	if net != nil {
		return net.AddressType()
	}

	if netList != nil {
		return LIST
	}

	return ADDRESS_NONE

}

func (nl *NetworkList) List() []*Network {
	return nl.list
}

func (nl *NetworkList) Type() IPFamily {
	return nl.family
}

func (nl *NetworkList) MarshalJSON() (b []byte, err error) {
	type NL struct {
		F IPFamily
		L []*Network
	}

	return json.Marshal(&NL{
		F: nl.family,
		L: nl.list,
	})
}

func (nl *NetworkList) UnmarshalJSON(b []byte) error {
	type NL struct {
		F IPFamily
		L []*Network
	}

	var n NL
	err := json.Unmarshal(b, &n)
	if err != nil {
		return err
	}

	nl.family = n.F
	nl.list = n.L

	return nil
}

func (nl *NetworkList) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New(fmt.Sprint("Failed to unmarshal IPNet value:", value))
	}

	err := json.Unmarshal(bytes, nl)
	return err
}

func (nl NetworkList) Value() (driver.Value, error) {
	return json.Marshal(&nl)
}

func (NetworkList) GormDataType() string {
	return "network_list"
}

func (nl *NetworkList) Copy() utils.CopyAble {
	list := []*Network{}
	for _, e := range nl.list {
		list = append(list, &Network{e.Copy().(AbbrNet)})
	}

	return &NetworkList{
		nl.Type(),
		list,
	}
}

//Type() IPFamily
func (nl *NetworkList) First() *IP {
	var ip *IP
	for i, n := range nl.list {
		if i == 0 {
			ip = n.First()
		} else {
			if ip.Int().Cmp(n.First().Int()) > 0 {
				ip = n.First().Copy()
			}
		}
	}

	return ip
}

func (nl *NetworkList) Last() *IP {
	var ip *IP
	for i, n := range nl.list {
		if i == 0 {
			ip = n.Last()
		} else {
			if ip.Int().Cmp(n.Last().Int()) < 0 {
				ip = n.Last().Copy()
			}
		}
	}

	return ip
}

func (nl *NetworkList) Size() int {
	if nl.family == IPv4 {
		return 32
	} else if nl.family == IPv6 {
		return 128
	} else {
		return -1
	}
}

func (nl *NetworkList) IsEmpty() bool {
	if nl.Count().Cmp(big.NewInt(0)) > 0 {
		return false
	} else {
		return true
	}
}

func (nl *NetworkList) Count() *big.Int {
	dr := nl.DataRange()
	//fmt.Printf("dr = %+v\n", dr)
	if dr == nil {
		return big.NewInt(0)
	}
	return dr.Count()
}

//func (nl NetworkList) Count() *big.Int {
//count := big.NewInt(0)
//for _, n := range nl.list {
//count = count.Add(count, n.Count())
//}
//return count
//}

func (nl *NetworkList) MatchIPNet(n *IPNet) bool {
	dr := nl.DataRange()
	other := n.DataRange()
	return dr.Match(other)
}

func (nl *NetworkList) MatchIPRange(n *IPRange) bool {
	dr := nl.DataRange()
	other := n.DataRange()
	return dr.Match(other)
}

func (nl *NetworkList) MatchNetwork(net *Network) bool {
	dr := nl.DataRange()
	other := net.DataRange()
	return dr.Match(other)
}

func (nl *NetworkList) MatchString(s string) (bool, error) {
	net, err := NewNetworkFromString(s)
	if err != nil {
		return false, err
	}

	if net == nil {
		return false, errors.New(fmt.Sprintf("NewNetworkFromString failed: %s", s))
	}
	dr := nl.DataRange()
	other := net.DataRange()
	return dr.Match(other), nil
}

func (nl *NetworkList) String() string {
	list := []string{}
	for _, n := range nl.list {
		list = append(list, fmt.Sprintf("%s", n))
	}
	return strings.Join(list, ",")
}

func (nl *NetworkList) Match(n AbbrNet) bool {
	dr := nl.DataRange()
	if dr == nil {
		return false
	}
	other := n.DataRange()
	if other == nil {
		return false
	}
	return dr.Match(other)
}

func (nl *NetworkList) Same(n AbbrNet) bool {
	dr := nl.DataRange()
	other := n.DataRange()

	if !dr.Match(other) {
		return false
	}

	return other.Match(dr)
}

func (nl *NetworkList) IPNetList() ([]*IPNet, error) {
	net, netlist := nl.Aggregate()
	if net != nil {
		return net.IPNetList()
	}

	list := []*IPNet{}
	if netlist != nil {
		for _, n := range netlist.list {
			nets, err := n.IPNetList()
			if err != nil {
				return nil, err
			}
			list = append(list, nets...)
		}
	}
	return list, nil
}

func (nl *NetworkList) IPNet() (*IPNet, bool) {
	net, _ := nl.Aggregate()
	if net != nil {
		return net.IPNet()
	} else {
		return nil, false
	}
}

func (nl *NetworkList) SuperNet() (*IPNet, error) {
	ip1 := nl.First()
	ip2 := nl.Last()
	rg := NewIPRangeFromInt(ip1.Int(), ip2.Int(), nl.Type())
	return rg.SuperNet()
}

func (nl *NetworkList) DataRange() flexrange.DataRangeInf {
	var dr flexrange.DataRangeInf
	if len(nl.list) == 0 {
		return nil
	}
	for i := 0; i < len(nl.list); i++ {
		if i == 0 {
			dr = nl.list[i].DataRange()
		} else {
			d := nl.list[i].DataRange()
			dr = dr.Add(d)
			//if dr == nil {
			//return nil
			//}
		}
	}

	return dr
}

func NewNetworkListFromEntryList(el flexrange.EntryList, af IPFamily) (*NetworkList, error) {
	nl := []*Network{}
	for it := el.Iterator(); it.HasNext(); {
		_, entry := it.Next()
		net, err := NewNetworkFromEntry(entry, af)
		if err != nil {
			return nil, err
		}
		nl = append(nl, net)

	}
	return NewNetworkListFromList(nl)
}

func NewNetworkListFromDataRange(dr flexrange.DataRangeInf) (*NetworkList, error) {
	if !(dr.Size() == 128 || dr.Size() == 32) {
		return nil, fmt.Errorf("dr.Size = %d", dr.Size())
	}

	var t IPFamily
	if dr.Size() == 128 {
		t = IPv6
	}
	if dr.Size() == 32 {
		t = IPv4
	}

	//func NewNetworkListFromList(nl []*Network) *NetworkList {
	nl := []*Network{}
	for it := dr.Iterator(); it.HasNext(); {
		_, e := it.Next()
		ip := NewIPRangeFromInt(e.Low(), e.High(), t)
		//net := NewNetworkFromIPRange(ip)
		//nl = append(nl, net)
		ip_list, err := ip.IPNetList()
		if err != nil {
			return nil, err
		}
		//fmt.Printf("ip_list = %+v\n", ip_list)
		//fmt.Println("===============================================")
		//fmt.Printf("ip_list = %+v\n", ip_list)

		for _, n := range ip_list {
			net := NewNetworkFromIPNet(n)
			nl = append(nl, net)
		}
	}

	return NewNetworkListFromList(nl)
}

func NewNetworkListFromList(nl []*Network) (*NetworkList, error) {
	var ipf IPFamily
	for i := 0; i < len(nl); i++ {
		if i == 0 {
			ipf = nl[i].Type()
		} else {
			if ipf != nl[i].Type() {
				return nil, fmt.Errorf("NewNetworkListFromList: nl[i].Type() = %d, ipf = %d", nl[i].Type(), ipf)
			}
		}
	}

	l := []*Network{}
	for _, n := range nl {
		var net *Network
		net = &Network{n.Copy().(AbbrNet)}
		l = append(l, net)
	}
	return &NetworkList{
		ipf,
		l,
	}, nil
}

// func (nl NetworkList) Type() IPFamily {
// return nl.family
// }

func (nl NetworkList) Aggregate() (*Network, *NetworkList) {
	var dr flexrange.DataRangeInf
	if len(nl.list) == 0 {
		return nil, nil
	}
	for i := 0; i < len(nl.list); i++ {
		if i == 0 {
			dr = nl.list[i].DataRange()
		} else {
			d := nl.list[i].DataRange()
			dr = dr.Add(d)
			if dr == nil {
				return nil, nil
			}
		}
	}
	// fmt.Println(nl, dr)
	if dr == nil || len(dr.List()) == 0 {
		return nil, nil
	}

	list := []*Network{}

	if len(dr.List()) == 1 {
		l := dr.List()
		ip1 := NewIPFromInt(utils.CopyInt(l[0].(flexrange.EntryInt).Low()), nl.Type())
		ip2 := NewIPFromInt(utils.CopyInt(l[0].(flexrange.EntryInt).High()), nl.Type())
		rg := NewIPRangeFromInt(ip1.Int(), ip2.Int(), nl.Type())
		cidrs := rg.CIDRs()
		if len(cidrs) == 1 {
			return NewNetworkFromIPNet(cidrs[0]), nil
		} else {
			return NewNetworkFromIPRange(rg), nil
		}

	} else {
		for it := dr.Iterator(); it.HasNext(); {
			_, e := it.Next()
			ip1 := NewIPFromInt(utils.CopyInt(e.(flexrange.EntryInt).Low()), nl.Type())
			ip2 := NewIPFromInt(utils.CopyInt(e.(flexrange.EntryInt).High()), nl.Type())
			rg := NewIPRangeFromInt(ip1.Int(), ip2.Int(), nl.Type())
			list = append(list, NewNetworkFromIPRange(rg))
		}

		return nil, &NetworkList{
			nl.Type(),
			list,
		}
	}

}

func NetworkListCmp(this NetworkList, other NetworkList) (left *NetworkList, mid *NetworkList, right *NetworkList) {
	if this.Type() != other.Type() {
		return &this, nil, &other
	}

	this_dr := this.DataRange()
	other_dr := other.DataRange()
	this_dr_copy := this_dr.Copy().(*flexrange.DataRange)
	var err error
	//for it := this_dr.Iterator(); it.HasNext(); {
	//_, e := it.Next()
	//fmt.Printf("e = %+v\n", e)
	//}
	//fmt.Println("-----------------------------------------------")
	//other_dr_copy := other_dr.Copy().(*flexrange.DataRange)

	//tt := NewNetworkListFromDataRange(this_dr)
	//oo := NewNetworkListFromDataRange(other_dr)
	//fmt.Printf("tt = %+v\n", tt)
	//fmt.Printf("this_dr = %+v\n", this_dr)
	//fmt.Println("-------")
	//fmt.Printf("oo = %+v\n", oo)
	//fmt.Printf("other_dr = %+v\n", other_dr)
	rm_this, _ := this_dr.Sub(other_dr)
	//fmt.Println("===============================================================")
	_, _ = other_dr.Sub(this_dr_copy)

	if len(rm_this.List()) == 0 {
		mid = nil
	} else {
		mid, err = NewNetworkListFromDataRange(rm_this)
		if err != nil {
			panic(err)
		}
	}
	if len(this_dr.List()) == 0 {
		left = nil
	} else {
		left, err = NewNetworkListFromDataRange(this_dr)
		if err != nil {
			panic(err)
		}
		//fmt.Printf("left = %+v\n", left)
		//fmt.Println("===============================================================")
	}
	if len(other_dr.List()) == 0 {
		right = nil
	} else {
		right, err = NewNetworkListFromDataRange(other_dr)
		if err != nil {
			panic(err)
		}
	}
	return
}

//func (n Network) Type() IPFamily {
//return n.Net.Type()
//}

//func (n Network) First() IP {
//return n.Net.First()
//}

//func (n Network) Last() IP {
//return n.Net.Last()
//}

//func (n Network) Count() *big.Int {
//return n.Net.Count()
//}

//func (n Network) MatchIPNet(o *IPNet) bool {
//return n.Net.MatchIPNet(o)
//}

//func (n Network) MatchIPRange(o *IPRange) bool {
//return n.Net.MatchIPRange(o)
//}

//func (n Network) SuperNet() *IPNet {
//return n.Net.SuperNet()
//}

//支持IPRange和IPNet两种字符串格式
func NewNetworkFromString(s string) (*Network, error) {
	if validator.IsIPRange(s) {
		rg, err := NewIPRange(s)
		if err != nil {
			return nil, err
		}
		return &Network{
			rg,
			//NewIPRange(s),
		}, nil
	} else {
		n, err := ParseIPNet(s)
		if err != nil {
			return nil, err
		}
		return &Network{
			n,
		}, nil
	}
}

func NewNetworkFromIPNet(n *IPNet) *Network {
	return &Network{
		n,
	}
}

func NewNetworkFromIPRange(n *IPRange) *Network {
	return &Network{
		n,
	}
}

func NewNetworkFromEntry(e flexrange.EntryInt, t IPFamily) (*Network, error) {
	if !(t == IPv4 || t == IPv6) {
		panic("only support IPv4 and IPv6 Entry")
	}

	r := NewIPRangeFromInt(e.Low(), e.High(), t)

	net, err := r.SuperNet()
	if err != nil {
		return nil, err
	}
	if r.MatchIPNet(net) && net.MatchIPRange(r) {
		return NewNetworkFromIPNet(net), nil
	} else {
		return NewNetworkFromIPRange(r), nil
	}
}

//func NewNetworkFromAbbrNet(n AbbrNet) *Network {
//return &Net

//}

type NetworkIterator struct {
	net   *Network
	index *big.Int
}

func (net *Network) Iterator() *NetworkIterator {
	return &NetworkIterator{
		net,
		big.NewInt(0),
	}
}

func (it *NetworkIterator) HasNext() bool {
	s1 := it.net.First().Int()
	s2 := it.net.Last().Int()

	if s1.Add(s1, it.index); !(s1.Cmp(s2) > 0) {
		return true
	}
	return false
}

func (it *NetworkIterator) Next() *IP {
	s1 := it.net.First().Int()
	s2 := it.net.Last().Int()

	r := s1.Add(s1, it.index)
	if r.Cmp(s2) > 0 {
		return nil
	}
	it.index = utils.AddInt(it.index, 1)
	return NewIPFromInt(r, it.net.Type())

}

func (n Network) DataRange() flexrange.DataRangeInf {
	return n.AbbrNet.DataRange()
}

type NetworkGroup struct {
	ipv4 NetworkList
	ipv6 NetworkList
}

func (ng NetworkGroup) IsEmpty() bool {
	if len(ng.ipv6.list) == 0 && len(ng.ipv4.list) == 0 {
		return true
	}

	return false
}

func (ng *NetworkGroup) IsAny(emptyIsAny bool) bool {
	if ng == nil {
		return true
	}

	if ng.IsEmpty() {
		return true
	}

	if ng.IsIPv4() {
		return ng.Same(NewAny4Group())
	} else if ng.IsIPv6() {
		return ng.Same(NewAny6Group())
	} else {
		return ng.Same(NewAny46Group())
	}
}

func (ng NetworkGroup) MustOne() *NetworkList {
	if len(ng.ipv6.list) > 0 && len(ng.ipv4.list) > 0 {
		panic("len(ipv6) > 0 and len(ipv4) > 0")
	}
	if len(ng.ipv6.list) == 0 && len(ng.ipv4.list) == 0 {
		panic("len(ipv6) == 0 and len(ipv4) == 0")
	}

	if len(ng.ipv6.list) > 0 {
		return &ng.ipv6
	}

	return &ng.ipv4
}

func (ng NetworkGroup) MarshalJSON() (b []byte, err error) {
	type NG struct {
		IPv4 NetworkList
		IPv6 NetworkList
	}

	return json.Marshal(&NG{
		IPv4: ng.ipv4,
		IPv6: ng.ipv6,
	})
}

func (ng *NetworkGroup) UnmarshalJSON(b []byte) error {
	type NG struct {
		IPv4 NetworkList
		IPv6 NetworkList
	}

	var n NG
	err := json.Unmarshal(b, &n)
	if err != nil {
		return err
	}

	ng.ipv4 = n.IPv4
	ng.ipv6 = n.IPv6

	return nil
}

func (ng *NetworkGroup) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New(fmt.Sprint("Failed to unmarshal NetworkGroup value:", value))
	}

	err := json.Unmarshal(bytes, ng)
	return err
}

func (ng NetworkGroup) Value() (driver.Value, error) {
	return json.Marshal(&ng)
}

func (NetworkGroup) GormDataType() string {
	return "network_group"
}

func (NetworkGroup) GormDBDataType(db *gorm.DB, field *schema.Field) string {
	return "blob"
}

func (ng NetworkGroup) String() string {
	text := []string{}
	if len(ng.ipv4.list) > 0 {
		tmp := ng.ipv4.String()
		text = append(text, tmp)
	}
	if len(ng.ipv6.list) > 0 {
		tmp := ng.ipv6.String()
		text = append(text, tmp)
	}
	return strings.Join(text, "\n")
}

func (ng NetworkGroup) NetworkList(af IPFamily) *NetworkList {
	if af == IPv4 {
		return ng.IPv4()
	} else {
		return ng.IPv6()
	}
}

func (ng NetworkGroup) Split() (v4, v6 *NetworkGroup) {
	if len(ng.ipv4.list) == 0 {
		v4 = nil
	} else {
		v4 = ng.Copy().(*NetworkGroup)
		v4.ipv6 = NetworkList{
			family: IPv6,
			list:   []*Network{},
		}
	}

	if len(ng.ipv6.list) == 0 {
		v6 = nil
	} else {
		v6 = ng.Copy().(*NetworkGroup)

		v6.ipv4 = NetworkList{
			family: IPv4,
			list:   []*Network{},
		}
	}
	return
}

func (ng NetworkGroup) IPv4() *NetworkList {
	return &ng.ipv4
}

func (ng NetworkGroup) IPv6() *NetworkList {
	return &ng.ipv6
}

func (ng NetworkGroup) HasIPv4() bool {
	if len(ng.ipv4.list) > 0 {
		return true
	}

	return false
}

func (ng NetworkGroup) HasIPv6() bool {
	return len(ng.ipv6.list) > 0
}

func (ng NetworkGroup) IsIPv4() bool {
	if len(ng.ipv4.list) == 0 {
		return false
	}

	if len(ng.ipv6.list) > 0 {
		return false
	}

	return true
}

func (ng NetworkGroup) IsIPv6() bool {
	if len(ng.ipv6.list) == 0 {
		return false
	}

	if len(ng.ipv4.list) > 0 {
		return false
	}

	return true
}

func NewAny4Group() *NetworkGroup {
	n, _ := NewNetworkFromString("0.0.0.0/0")
	return &NetworkGroup{
		NetworkList{
			IPv4,
			[]*Network{n},
		},
		NetworkList{
			IPv6,
			[]*Network{},
		},
	}
}

func NewAny6Group() *NetworkGroup {
	n6, _ := NewNetworkFromString("::/0")
	return &NetworkGroup{
		NetworkList{
			IPv4,
			[]*Network{},
		},
		NetworkList{
			IPv6,
			[]*Network{n6},
		},
	}
}

func NewAny46Group() *NetworkGroup {
	n4, _ := NewNetworkFromString("0.0.0.0/0")
	n6, _ := NewNetworkFromString("::/0")
	return &NetworkGroup{
		NetworkList{
			IPv4,
			[]*Network{n4},
		},

		NetworkList{
			IPv6,
			[]*Network{n6},
		},
	}
}

func NewNetworkGroup() *NetworkGroup {
	return &NetworkGroup{
		NetworkList{
			IPv4,
			[]*Network{},
		},
		NetworkList{
			IPv6,
			[]*Network{},
		},
	}
}

func NewNetworkGroupFromString(s string) (*NetworkGroup, error) {
	if s == "" {
		return nil, fmt.Errorf("s is empty")
	}

	list := strings.Split(s, ",")

	ipv4_netlist := []*Network{}
	ipv6_netlist := []*Network{}
	for _, addr := range list {
		n, err := NewNetworkFromString(addr)
		if err != nil {
			return nil, err
		}
		if n.Type() == IPv4 {
			ipv4_netlist = append(ipv4_netlist, n)
		} else {
			ipv6_netlist = append(ipv6_netlist, n)
		}
	}

	ipv4NetworkList, err := NewNetworkListFromList(ipv4_netlist)
	if err != nil {
		return nil, err
	}
	ipv6NetworkList, err := NewNetworkListFromList(ipv6_netlist)
	if err != nil {
		return nil, err
	}

	return &NetworkGroup{
		ipv4: *ipv4NetworkList,
		ipv6: *ipv6NetworkList,
	}, nil
}

func (ng *NetworkGroup) Add(net AbbrNet) {
	var n []*Network
	switch net.(type) {
	case *IPNet:
		n = []*Network{NewNetworkFromIPNet(net.Copy().(*IPNet))}
	case *IPRange:
		n = []*Network{NewNetworkFromIPRange(net.Copy().(*IPRange))}
	case *NetworkList:
		n = net.Copy().(*NetworkList).list
	case *Network:
		// n = []*Network{}
		n = append(n, &Network{net.Copy().(AbbrNet)})
	}

	//func NewNetworkListFromDataRange(dr flexrange.DataRangeInf) *NetworkList {

	if net.Type() == IPv4 {
		ng.ipv4.list = append(ng.ipv4.list, n...)
	}
	if net.Type() == IPv6 {
		ng.ipv6.list = append(ng.ipv6.list, n...)
	}

}

func (ng NetworkGroup) Copy() utils.CopyAble {
	//ipv4 NetworkList
	//ipv6 NetworkList

	return &NetworkGroup{
		*ng.ipv4.Copy().(*NetworkList),
		*ng.ipv6.Copy().(*NetworkList),
	}
}

func (ng *NetworkGroup) AddGroup(n *NetworkGroup) {
	ng.Add(n.ipv4.Copy().(AbbrNet))
	ng.Add(n.ipv6.Copy().(AbbrNet))
}

func (ng NetworkGroup) Match(n AbbrNet) bool {
	if n.Type() == IPv4 {
		return ng.ipv4.Match(n)
	} else {
		return ng.ipv6.Match(n)
	}
}

func (ng NetworkGroup) Same(other *NetworkGroup) bool {
	if ng.MatchNetworkGroup(other) && other.MatchNetworkGroup(&ng) {
		return true
	} else {
		return false
	}
}

func (ng NetworkGroup) MatchNetworkGroup(n *NetworkGroup) bool {
	if len(n.ipv4.list) > 0 {
		for _, i := range n.ipv4.list {
			if !ng.Match(i) {
				return false
			}
		}
	}

	if len(n.ipv6.list) > 0 {
		for _, i := range n.ipv6.list {
			if !ng.Match(i) {
				return false
			}
		}
	}

	return true
}

func (ng NetworkGroup) Count() *big.Int {
	count1 := ng.ipv4.Count()
	count2 := ng.ipv6.Count()
	var z big.Int

	z.Add(count1, count2)
	return &z
}

func (ng NetworkGroup) Aggregate() (*NetworkGroup, error) {
	var err error
	ipv4_dr := ng.ipv4.DataRange()
	ipv6_dr := ng.ipv6.DataRange()
	ipv4_nl := &NetworkList{
		IPv4,
		[]*Network{},
	}

	//family IPFamily
	//list   []*Network
	//fmt.Printf("ipv4_dr = %+v\n", ipv4_dr)
	if ipv4_dr != nil {
		//ipv4_nl = NewNetworkListFromDataRange(ipv4_dr)
		ipv4_nl, err = NewNetworkListFromDataRange(ipv4_dr)
		if err != nil {
			return nil, err
		}
	}

	ipv6_nl := &NetworkList{
		IPv6,
		[]*Network{},
	}

	if ipv6_dr != nil {
		ipv6_nl, err = NewNetworkListFromDataRange(ipv6_dr)
		if err != nil {
			return nil, err
		}
	}

	return &NetworkGroup{
		*ipv4_nl,
		*ipv6_nl,
	}, nil

}

func (ng *NetworkGroup) AddressType() AddressType {
	if len(ng.ipv4.list) > 0 && len(ng.ipv6.list) > 0 {
		return MIXED
	}

	if len(ng.ipv4.list) > 0 {
		return ng.ipv4.AddressType()
	}

	if len(ng.ipv6.list) > 0 {
		return ng.ipv6.AddressType()
	}

	return ADDRESS_NONE
}

func (ng NetworkGroup) GenerateNetwork() *Network {
	var nl NetworkList
	if ng.IsIPv4() {
		nl = ng.ipv4
	} else if ng.IsIPv6() {
		nl = ng.ipv6
	} else {
		panic("GenerateNetwork failed")
	}
	net, _ := nl.Aggregate()
	if net == nil {
		panic("GenerateNetwork failed")
	}
	return net
}

func NetworkGroupCmp(this NetworkGroup, other NetworkGroup) (left *NetworkGroup, mid *NetworkGroup, right *NetworkGroup) {
	var v4_left, v4_mid, v4_right *NetworkList
	var v6_left, v6_mid, v6_right *NetworkList

	if len(this.ipv4.list) == 0 && len(other.ipv4.list) == 0 {
		v4_left = nil
		v4_mid = nil
		v4_right = nil
	} else if len(this.ipv4.list) == 0 {
		v4_left = nil
		v4_mid = nil
		v4_right = other.ipv4.Copy().(*NetworkList)
	} else if len(other.ipv4.list) == 0 {
		v4_left = this.ipv4.Copy().(*NetworkList)
		v4_mid = nil
		v4_right = nil
	} else {
		v4_left, v4_mid, v4_right = NetworkListCmp(this.ipv4, other.ipv4)
	}

	if len(this.ipv6.list) == 0 && len(other.ipv6.list) == 0 {
		v6_left = nil
		v6_mid = nil
		v6_right = nil
	} else if len(this.ipv6.list) == 0 {
		v6_left = nil
		v6_mid = nil
		v6_right = other.ipv6.Copy().(*NetworkList)
	} else if len(other.ipv6.list) == 0 {
		v6_left = this.ipv6.Copy().(*NetworkList)
		v6_mid = nil
		v6_right = nil
	} else {
		v6_left, v6_mid, v6_right = NetworkListCmp(this.ipv6, other.ipv6)
	}

	//list   []*Network
	if v4_left == nil && v6_left == nil {
		left = nil
	} else if v4_left == nil {
		nl := NetworkList{
			IPv4,
			[]*Network{},
		}
		left = &NetworkGroup{
			nl,
			*v6_left,
		}
	} else if v6_left == nil {
		nl := NetworkList{
			IPv6,
			[]*Network{},
		}
		left = &NetworkGroup{
			*v4_left,
			nl,
		}
	} else {
		left = &NetworkGroup{
			*v4_left,
			*v6_left,
		}
	}

	if v4_mid == nil && v6_mid == nil {
		mid = nil
	} else if v4_mid == nil {
		nl := NetworkList{
			IPv4,
			[]*Network{},
		}
		mid = &NetworkGroup{
			nl,
			*v6_mid,
		}
	} else if v6_mid == nil {
		nl := NetworkList{
			IPv6,
			[]*Network{},
		}
		mid = &NetworkGroup{
			*v4_mid,
			nl,
		}
	} else {
		mid = &NetworkGroup{
			*v4_mid,
			*v6_mid,
		}
	}

	if v4_right == nil && v6_right == nil {
		right = nil
	} else if v4_right == nil {
		nl := NetworkList{
			IPv4,
			[]*Network{},
		}
		right = &NetworkGroup{
			nl,
			*v6_right,
		}
	} else if v6_right == nil {
		nl := NetworkList{
			IPv6,
			[]*Network{},
		}
		right = &NetworkGroup{
			*v4_right,
			nl,
		}
	} else {
		right = &NetworkGroup{
			*v4_right,
			*v6_right,
		}
	}

	return
}

func Translate(input, translator *NetworkGroup) (*NetworkGroup, error) {
	// 原则：1、进行地址转换，如果translator为空，表示任何input地址不会做任何变化
	//       2、如果translator地址不为空，即为表示需要对input地址中的ipv4和ipv6都进行匹配
	//          如果translator的ipv4或ipv6为空，表示不运行进行对应的地址进行转换
	//            比如：translator的ipv6为空，但是巧合input的ipv6不为空，此时将进行报错，返回error
	inIpv4, inIpv6 := input.Split()
	translatorIpv4, translatorIpv6 := translator.Split()
	var resultIpv4, resultIpv6 *NetworkGroup

	if translator.IsEmpty() {
		return input.Copy().(*NetworkGroup), nil
	} else {
		if translatorIpv4 == nil && inIpv4 != nil {
			return nil, errors.New("translator ipv4 is empty")
		}

		if translatorIpv6 == nil && inIpv6 != nil {
			return nil, errors.New("translator ipv6 is empty")
		}
		// if inIpv4.Count().Cmp(big.NewInt(0)) > 0 && translatorIpv4.Count().Cmp(big.NewInt(0)) == 0 {
		// return nil, errors.New("translator ipv4 is empty")
		// }
		//
		// if inIpv6.Count().Cmp(big.NewInt(0)) > 0 && translatorIpv6.Count().Cmp(big.NewInt(0)) == 0 {
		// return nil, errors.New("translator ipv6 is empty")
		// }
	}

	if translatorIpv4 != nil && inIpv4 != nil {
		if translatorIpv4.MatchNetworkGroup(inIpv4) {
			resultIpv4 = inIpv4.Copy().(*NetworkGroup)
		} else {
			resultIpv4 = translatorIpv4.Copy().(*NetworkGroup)
		}
	}

	if translatorIpv6 != nil && inIpv6 != nil {
		if translatorIpv6.MatchNetworkGroup(inIpv6) {
			resultIpv6 = inIpv6.Copy().(*NetworkGroup)
		} else {
			resultIpv6 = translatorIpv6.Copy().(*NetworkGroup)
		}
	}

	result := NewNetworkGroup()
	if resultIpv4 != nil {
		result.AddGroup(resultIpv4)
	}
	if resultIpv6 != nil {
		result.AddGroup(resultIpv6)
	}

	return result, nil

}

func (ng *NetworkGroup) StringList() []string {
	netList := []string{}
	for _, ipv4 := range ng.ipv4.list {
		netList = append(netList, ipv4.String())
	}

	for _, ipv6 := range ng.ipv6.list {
		netList = append(netList, ipv6.String())
	}

	return netList
}

func (one *NetworkGroup) MatchWithOption(two *NetworkGroup, optionFlags int, includeAny bool) bool {
	if !includeAny {
		if one.IsIPv4() {
			if one.Same(NewAny4Group()) {
				return false
			}
		} else if one.IsIPv6() {
			if one.Same(NewAny6Group()) {
				return false
			}
		} else {
			if one.Same(NewAny46Group()) {
				return false
			}
		}
	}

	var ok bool
	if optionFlags&OBJECT_OPTION_ONE_MATCH_TWO != 0 && one.MatchNetworkGroup(two) {
		ok = true
	} else if optionFlags&OBJECT_OPTION_TWO_MATCH_ONE != 0 && two.MatchNetworkGroup(one) {
		ok = true
	} else if optionFlags&OBJECT_OPTION_SAME != 0 && one.Same(two) {
		ok = true
	}

	return ok
}
