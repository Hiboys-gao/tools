package network

import (
	"encoding/json"
	"fmt"
	"math/big"
	"strings"
	"tools/flexrange"
	"tools/utils"
	"tools/validator"

	gojsonq "github.com/thedevsaddam/gojsonq/v2"
)

type addressFilter struct {
	at    *AddressTable
	jsonq *gojsonq.JSONQ
}

type VsInt interface {
	String() string
	Copy() utils.CopyAble
}

type HopInt interface {
	String() string
	Copy() utils.CopyAble
}

type Hop struct {
	Interface string `json:"interface"`
	Ip        string `json:"ip"`
	Connected bool   `json:"connected"`
	DefaultGw bool   `json:"default_gw"`
	Vs        interface{}
}

func NewHop(it string, ip string, connect, defaultGw bool, vs interface{}) (*Hop, error) {
	data := map[string]interface{}{
		"interface":  it,
		"ip":         ip,
		"connect":    connect,
		"default_gw": defaultGw,
		"vs":         vs,
	}
	result := NextHopFormatValidator{}.Validate(data)
	if result.Status() == false {
		panic(result.Msg())
	}
	return &Hop{
		it,
		ip,
		connect,
		defaultGw,
		vs,
	}, nil
}

func FlattenEntryList(el *flexrange.EntryList, ip IPFamily) []map[string]interface{} {
	ms := []map[string]interface{}{}
	for it := el.Iterator(); it.HasNext(); {

		_, re := it.Next()
		next := re.Data().Data.(*NextHop)

		for _, nh := range next.next {
			m := map[string]interface{}{}
			m["net"] = NewIPRangeFromEtnry(re, ip).String()
			m["interface"] = nh.(*Hop).Interface
			m["connected"] = nh.(*Hop).Connected
			m["ip"] = nh.(*Hop).Ip
			m["default_gw"] = nh.(*Hop).DefaultGw
			ms = append(ms, m)
		}
	}

	return ms
}

func (h Hop) Copy() utils.CopyAble {
	return &Hop{
		h.Interface,
		h.Ip,
		h.Connected,
		h.DefaultGw,
		h.CopyVs(),
	}
}

func (h Hop) CopyVs() interface{} {
	if h.Vs == nil {
		return nil
	} else {
		return h.Vs.(utils.CopyAble).Copy()
	}
}

func (h Hop) String() string {
	return h.Json()
}

func (h Hop) Json() string {
	s, _ := json.Marshal(h)
	return string(s)
}

type NextHop struct {
	next []HopInt
}

func (nh *NextHop) Count() int {
	return len(nh.next)
}

func (nh *NextHop) MakeExtendData() (*flexrange.ExtendData, error) {
	if _, ok := flexrange.MAPPER["NextHop"]; !ok {
		flexrange.MAPPER["NextHop"] = &NextHop{}
	}

	raw, err := json.Marshal(&nh)
	if err != nil {
		return nil, err
	}
	//fmt.Printf("raw = %+v\n", raw)
	ext := flexrange.ExtendData{
		Type: "NextHop",
		Data: nh,
	}
	ext.Property = make([]byte, len(raw))

	copy(ext.Property, raw)
	//fmt.Printf("ext.Property = %+v\n", ext.Property)
	return &ext, nil
}

func (nh NextHop) MarshalJSON() (b []byte, err error) {
	type nexthop struct {
		Next []*Hop
	}
	_nh := nexthop{}
	for _, i := range nh.next {
		_nh.Next = append(_nh.Next, i.(*Hop))
	}

	return json.Marshal(&_nh)
}

func (nh *NextHop) UnmarshalJSON(b []byte) error {
	type nexthop struct {
		Next []*Hop
	}

	_nh := nexthop{}
	err := json.Unmarshal(b, &_nh)
	if err != nil {
		return err
	}

	for _, i := range _nh.Next {
		nh.next = append(nh.next, i)
	}

	return nil
}

func NewNextHop() *NextHop {
	nx := []HopInt{}
	return &NextHop{
		nx,
	}
}

func (nh NextHop) OutInterfaces() []string {
	il := []string{}
	for it := nh.Iterator(); it.HasNext(); {
		_, h := it.Next()
		il = append(il, h.(*Hop).Interface)
	}

	return il
}

func (nh NextHop) NextIp(intName string) []string {
	ips := []string{}
	for it := nh.Iterator(); it.HasNext(); {
		_, h := it.Next()
		if h.(*Hop).Interface == intName {
			if h.(*Hop).Ip != "" {
				ips = append(ips, h.(*Hop).Ip)
			}
		}
	}
	return ips
}

func (nh NextHop) IsConnected() bool {
	c := false
	for index, e := range nh.next {
		if index == 0 {
			c = e.(*Hop).Connected
		} else {
			if c != e.(*Hop).Connected {
				return false
			}
		}
	}
	return c
}
func (nh NextHop) IsDefaultGw() bool {
	defaultGw := false
	for index, e := range nh.next {
		if index == 0 {
			defaultGw = e.(*Hop).DefaultGw
		} else {
			if defaultGw != e.(*Hop).DefaultGw {
				return false
			}
		}
	}
	return defaultGw
}

func (nh NextHop) IsSameInterface() (bool, string) {
	inf := ""
	for index, e := range nh.next {
		if index == 0 {
			inf = e.(*Hop).Interface
		} else {
			if inf != e.(*Hop).Interface {
				return false, ""
			}
		}
	}
	if inf != "" {
		return true, inf
	} else {
		return false, ""
	}
}

func (nh NextHop) IsSameIp() (bool, string) {
	ip := ""
	for index, e := range nh.next {
		if index == 0 {
			ip = e.(*Hop).Ip
		} else {
			if ip != e.(*Hop).Ip {
				return false, ""
			}
		}
	}
	if ip != "" {
		return true, ip
	} else {
		return false, ""
	}
}

func (nh *NextHop) AddHop(it string, ip string, connect, defaultGw bool, vs interface{}) (*Hop, error) {
	h, err := NewHop(it, ip, connect, defaultGw, vs)
	if err == nil {
		nh.next = append(nh.next, h)
		return h, err
	} else {
		return h, err
	}
}

func (n NextHop) String() string {
	l := []string{}
	for it := n.Iterator(); it.HasNext(); {
		_, h := it.Next()
		l = append(l, h.String())
	}

	return strings.Join(l, ",")
	//return strings.Join(l, ",")
}

func (n NextHop) Copy() utils.CopyAble {
	nx := []HopInt{}
	for it := n.Iterator(); it.HasNext(); {
		_, h := it.Next()

		nx = append(nx, h.(utils.CopyAble).Copy().(HopInt))
	}

	return &NextHop{
		nx,
	}
}

type NextHopIterator struct {
	nh    *NextHop
	index int
}

func (nh *NextHop) Iterator() *NextHopIterator {
	return &NextHopIterator{
		nh,
		0,
	}
}

func (ni *NextHopIterator) HasNext() bool {
	return ni.index < len(ni.nh.next)
}

func (ni *NextHopIterator) Next() (int, HopInt) {
	n := ni.nh.next[ni.index]
	index := ni.index
	ni.index++
	return index, n
}

func NewIPv4EntryList() *flexrange.EntryList {
	return flexrange.NewEntryList(32, big.NewInt(0))
}

func NewIPv6EntryList() *flexrange.EntryList {
	return flexrange.NewEntryList(128, big.NewInt(0))
}

func NewIPEntryList(t IPFamily) *flexrange.EntryList {
	var r *flexrange.EntryList
	if t == IPv4 {
		r = NewIPv4EntryList()
	} else if t == IPv6 {
		r = NewIPv6EntryList()
	}

	r.WithStrFunc(func() string {
		ls := []string{}
		for it := r.Iterator(); it.HasNext(); {
			_, e := it.Next()

			//if ls == "" {
			//ipr := NewIPRange()
			ipr := NewIPRangeFromInt(e.Low(), e.High(), t)
			ip, _ := ipr.SuperNet()

			//ls = fmt.Sprintf("Low:%s, High:%s, NextHop:%s",

			ls = append(ls, fmt.Sprintf("{\"net\":\"%s\", \"next\":%s}", ip, e.Data().Data))
			//NewIPFromInt(e.Low(), t), NewIPFromInt(e.High(), t), e.Data()))
			//} else {
			//ls = append(ls, fmt.Sprintf("%s; Low:%s, High:%s, NextHop:%s",
			//ls, NewIPFromInt(e.Low(), t), NewIPFromInt(e.High(), t), e.Data()))
			//}
		}
		return "[" + strings.Join(ls, ",\n") + "]"
	})

	return r
}

func NewIPEntryListFromList(t IPFamily, list []flexrange.EntryInt) *flexrange.EntryList {
	if t == IPv4 {
		return flexrange.NewEntryListFromList(32, big.NewInt(0), list)
	} else if t == IPv6 {
		return flexrange.NewEntryListFromList(128, big.NewInt(0), list)
	}

	return nil
}

//targetList := flexrange.NewEntryListFromList(uint32(t.Size()), big.NewInt(0), []flexrange.EntryInt{other})

type AddressTable struct {
	ip IPFamily
	//table [][]*flexrange.Entry
	table []*flexrange.EntryList
	dgw   *flexrange.Entry
	//jsonq *jsonq.JSONQ
}

func (at AddressTable) Type() IPFamily {
	return at.ip
}

func (at AddressTable) MarshalJSON() (b []byte, err error) {
	type addresstable struct {
		IP    IPFamily
		Table []*flexrange.EntryList
		Dgw   *flexrange.Entry
	}

	a := addresstable{
		IP:    at.ip,
		Table: at.table,
		Dgw:   at.dgw,
	}

	return json.Marshal(&a)
}

func (at *AddressTable) UnmarshalJSON(b []byte) error {
	type addresstable struct {
		IP    IPFamily
		Table []*flexrange.EntryList
		Dgw   *flexrange.Entry
	}
	var a addresstable
	err := json.Unmarshal(b, &a)
	//fmt.Printf("err = %+v\n", err)
	if err != nil {
		return err
	}
	at.ip = a.IP
	for _, i := range a.Table {
		at.table = append(at.table, i)

	}
	if a.Dgw != nil {
		at.dgw = a.Dgw.Copy().(*flexrange.Entry)
	}

	return nil
}

type NextHopFormatValidator struct{}

func (nv NextHopFormatValidator) Validate(data map[string]interface{}) validator.Result {
	// routeTable := data["table"].(*AddressTable)
	it := data["interface"].(string)
	ip := data["ip"].(string)
	connect := data["connect"].(bool)
	vs := data["vs"]

	if vs != nil {
		_, ok := vs.(VsInt)
		if !ok {
			return validator.NewValidateResult(false, fmt.Sprintf("error 1: interface:%s, ip:%s, connect:%t, vs:%T", it, ip, connect, vs))
		}
	}

	if ip != "" {
		if !(validator.IsIPv4Address(ip) || validator.IsIPv6Address(ip)) {
			return validator.NewValidateResult(false, fmt.Sprintf("error 2: interface:%s, ip:%s, connect:%t, vs:%T", it, ip, connect, vs))
		}
		// if it == "" {
		// return nil, errors.New(fmt.Sprintf("error 3: interface:%s, ip:%s, connect:%t, vs:%T", it, ip, connect, vs))
		// }
	}

	if it != "" {
		if !connect {
			if ip == "" {
				return validator.NewValidateResult(false, fmt.Sprintf("error 3: interface:%s, ip:%s, connect:%t, vs:%T", it, ip, connect, vs))
			}
		}
	}

	if !connect {
		if ip == "" {
			return validator.NewValidateResult(false, fmt.Sprintf("error 4: interface:%s, ip:%s, connect:%t, vs:%T", it, ip, connect, vs))
		}
	} else {
		if it == "" {
			return validator.NewValidateResult(false, fmt.Sprintf("error 5: interface:%s, ip:%s, connect:%t, vs:%T", it, ip, connect, vs))
		}
	}

	return validator.NewValidateResult(true, "")
}

type NextHopRecursionValidator struct{}

func (nv NextHopRecursionValidator) Validate(data map[string]interface{}) validator.Result {
	routeTable := data["table"].(*AddressTable)
	// for i := routeTable.Size() - 1; i >= 0; i-- {
	for i := routeTable.Size(); i >= 0; i-- {
		el := routeTable.table[i]
		for it := el.Iterator(); it.HasNext(); {
			_, re := it.Next()
			next := re.Data().Data.(*NextHop)

			for _, nh := range next.next {
				hop := nh.(*Hop)
				if hop.Interface == "" && hop.Ip == "" {
					return validator.NewValidateResult(false, fmt.Sprintf("%s ip and interface is empty", re))
				}

				if hop.Interface == "" {
					net, err := ParseIPNet(hop.Ip)
					if err != nil {
						return validator.NewValidateResult(false, fmt.Sprintf("%v", err))
					}
					// 递归检查时，应该允许默认路由进行递归查询
					rmr := routeTable.Match(net, true, false)
					if rmr.IsMatch() == false {
						return validator.NewValidateResult(false, fmt.Sprintf("next hop ip %s match route failed", hop.Ip))
					}
				}
			}
		}

	}

	return validator.NewValidateResult(true, "")

}

// 返回net_list和gw_list
// 从
func (at AddressTable) Flatten() (*flexrange.EntryList, *flexrange.EntryList) {
	netList := NewIPEntryList(at.ip)
	gwList := NewIPEntryList(at.ip)

	// for i := at.Size() - 1; i >= 0; i-- {
	for i := at.Size(); i >= 0; i-- {
		el := at.table[i]
		for it := el.Iterator(); it.HasNext(); {
			_, re := it.Next()
			next := re.Data().Data.(*NextHop)

			for _, nh := range next.next {
				newRe := re.Copy().(flexrange.EntryInt)
				dnh := &NextHop{
					next: []HopInt{nh.Copy().(HopInt)},
				}
				ext, err := dnh.MakeExtendData()

				if err != nil {
					panic(err)
				}
				newRe.SetData(ext)
				//newRe.SetData(&NextHop{
				//next: []HopInt{nh.Copy().(HopInt)},
				//})

				netList.PushEntry(newRe)
			}
		}
	}

	if at.dgw != nil {
		next := at.dgw.Data().Data.(*NextHop)

		for _, nh := range next.next {
			newRe := at.dgw.Copy().(flexrange.EntryInt)
			dnh := &NextHop{
				next: []HopInt{nh.Copy().(HopInt)},
			}

			ext, err := dnh.MakeExtendData()
			if err != nil {
				panic(err)
			}
			newRe.SetData(ext)

			//newRe.SetData(&NextHop{
			//next: []HopInt{nh.Copy().(HopInt)},
			//})

			gwList.PushEntry(newRe)
		}
	}

	return netList, gwList
}

func (at AddressTable) AddressFilter(gw bool) *addressFilter {
	a, b := at.Flatten()

	if gw {
		for it := b.Iterator(); it.HasNext(); {
			_, e := it.Next()
			a.PushEntry(e)
		}
	}

	text := a.String()
	return &addressFilter{
		at:    &at,
		jsonq: gojsonq.New().FromString(text),
	}
}

func (af *addressFilter) Select(name ...string) *addressFilter {
	af.jsonq.Select(name...)
	return af
}

func (af *addressFilter) Interface(name string) *addressFilter {
	af.jsonq.Where("next.interface", "=", name)
	return af
}

func (af *addressFilter) NotInterface(name string) *addressFilter {
	af.jsonq.Where("next.interface", "!=", name)
	return af
}

func (af *addressFilter) Ip(name string) *addressFilter {
	af.jsonq.Where("next.ip", "=", name)
	return af
}

func (af *addressFilter) NotIp(name string) *addressFilter {
	af.jsonq.Where("next.ip", "!=", name)
	return af
}

func (af *addressFilter) Connected(connect bool) *addressFilter {
	af.jsonq.Where("next.connected", "=", connect)
	return af
}

func (af *addressFilter) NotConnected(connect bool) *addressFilter {
	af.jsonq.Where("next.connected", "!=", connect)
	return af
}

func (af *addressFilter) Get() []interface{} {
	return af.jsonq.Get().([]interface{})
}

func (at AddressTable) Jsonq() *gojsonq.JSONQ {
	text := at.Json()

	json := gojsonq.New().FromString(text)
	return json
}

//func (at AddressTable) Filter(t string, fd string) (*flexrange.EntryList, *flexrange.EntryList) {
////match := flexrange.NewEntryList(uint32(at.Size()), big.NewInt(0))
////dgw := flexrange.NewEntryList(uint32(at.Size()), big.NewInt(0))

//match := NewIPEntryList(at.ip)
//dgw := NewIPEntryList(at.ip)
//for i := at.Size() - 1; i >= 0; i-- {
//el := at.table[i]
//for it := el.Iterator(); it.HasNext(); {
//_, e := it.Next()
//nexthop := e.Data().(*NextHop)
//for hopit := nexthop.Iterator(); hopit.HasNext(); {
//_, hop := hopit.Next()
//switch t {
//case "interface":
//if hop.(*Hop).Interface == fd {
//match.PushEntry(e)
//}
//break
//case "ip":
//if hop.(*Hop).Ip == fd {
//match.PushEntry(e)
//}
//break
//}

//}

//}

//}

//if at.dgw != nil {
//nh := at.dgw.Data().(*NextHop)
//for it := nh.Iterator(); it.HasNext(); {
//_, hop := it.Next()
//switch t {
//case "interface":
//if hop.(*Hop).Interface == fd {
////match.PushEntry(at.dgw)
//dgw.PushEntry(at.dgw)
//}
//break
//case "ip":
//if hop.(*Hop).Ip == fd {
////match.PushEntry(at.dgw)
//dgw.PushEntry(at.dgw)
//}
//break
//}

//}

//}
//return match, dgw
//}

func NewAddressTable(ip IPFamily) *AddressTable {
	var l int
	if ip == IPv4 {
		l = 32
	} else {
		l = 128
	}

	var m []*flexrange.EntryList = make([]*flexrange.EntryList, l+1)
	for i := 0; i <= l; i++ {
		m[i] = NewIPEntryList(ip)
	}

	return &AddressTable{
		ip,
		m,
		nil,
	}
}
func (t *AddressTable) DefaultGw() *flexrange.Entry {
	return t.dgw
}

// func (t *AddressTable) DefaultGw() bool {
// if t.dgw != nil {
// return true
// }
// return false
// }

func (t *AddressTable) Size() int {
	if t.ip == IPv4 {
		return 32
	}
	if t.ip == IPv6 {
		return 128
	}
	panic(fmt.Sprintf("%+v", t))
}

func (t *AddressTable) Push(net AbbrNet, data *flexrange.ExtendData) error {
	n, ok := net.(*IPNet)
	if !ok {
		return fmt.Errorf("net: %T, data: %T", net, data)
		//return fmt.Errorf("net: %T", net)
	}

	switch net.(type) {
	case *IPNet:
		if net.Copy().(*IPNet).Mask.Prefix() == -1 {
			return fmt.Errorf("Prefix() == -1, net: %+v, data: %+v", net, data)
		}
	}

	first := net.First().Int()
	last := net.Last().Int()

	et, err := flexrange.NewEntry(first, last, data)
	if err != nil {
		panic(err)
	}

	et.SetData(data)
	l := n.Mask.Prefix()

	if l == 0 {
		// 如果路由项的Prefix为0，其实就是默认路由，单独保存为t.dgw中
		t.dgw = et
	} else {
		t.table[l].Push(first, last, data)
	}

	return nil
}

func (t *AddressTable) PushRoute(net AbbrNet, nx *NextHop) error {
	n, ok := net.(*IPNet)
	if !ok {
		return fmt.Errorf("net: %T, nexthop: %T", net, nx)
	}

	l := n.Mask.Prefix()
	if l == 0 {
		for _, hop := range nx.next {
			hop.(*Hop).DefaultGw = true
		}
	}
	ext, err := nx.MakeExtendData()
	if err != nil {
		panic(err)
	}
	return t.Push(net, ext)
}

func (t AddressTable) Remove(net AbbrNet) (ok bool) {
	entry, err := flexrange.NewEntry(net.First().Int(), net.Last().Int(), nil)
	if err != nil {
		panic(err)
	}

	for _, nl := range t.table {
		ok = nl.Remove(entry)
		if ok {
			return
		}
	}

	return
}

func (t *AddressTable) Equal(net AbbrNet) *NextHop {
	entry, err := flexrange.NewEntry(net.First().Int(), net.Last().Int(), nil)
	if err != nil {
		panic(err)
	}

	for _, nl := range t.table {
		for it := nl.Iterator(); it.HasNext(); {
			_, e := it.Next()
			//fmt.Printf("Compare(entry) = %+v\n", e.Compare(entry))
			if e.Compare(entry) == flexrange.Equal {
				nh := e.Data().Data.Copy()
				return nh.(*NextHop)
			}
		}
	}

	return nil
}

func (t *AddressTable) OutputInterface(route flexrange.EntryInt) []string {
	nextHop := route.Data().Data.(*NextHop)

	portMap := map[string]int{}
	for it := nextHop.Iterator(); it.HasNext(); {
		_, hop := it.Next()
		portMap[hop.(*Hop).Interface] = 1
	}

	portList := []string{}
	for k, _ := range portMap {
		portList = append(portList, k)
	}

	return portList

}

func (t *AddressTable) MatchNetList(nl NetworkList, dgw, ignoreGateway bool) *MatchResult {
	if nl.Type() != t.ip {
		panic(fmt.Sprintf("AddressTable type is: %s, NetworkList type is: %s", t.ip, nl.Type()))
	}

	//match := flexrange.NewEntryList(uint32(t.Size()), big.NewInt(0))
	match := NewIPEntryList(t.ip)
	//unmatch := flexrange.NewEntryList(uint32(t.Size()), big.NewInt(0))
	unmatch := NewIPEntryList(t.ip)
	for _, n := range nl.list {
		res := t.Match(n, dgw, ignoreGateway)
		for it := res.Match.Iterator(); it.HasNext(); {
			_, e := it.Next()
			match.PushEntry(e)
		}
		for it := res.Unmatch.Iterator(); it.HasNext(); {
			_, e := it.Next()
			unmatch.PushEntry(e)
		}
	}
	return &MatchResult{
		Ip:      t.ip,
		Match:   match,
		Unmatch: unmatch,
	}
}

func (t *AddressTable) Match(net AbbrNet, dgw, ignoreGateway bool) *MatchResult {
	//unmatch表示net匹配路由以后的，还剩余的部分
	other, err := flexrange.NewEntry(net.First().Int(), net.Last().Int(), nil)
	if err != nil {
		panic(err)
	}

	targetList := NewIPEntryListFromList(t.ip, []flexrange.EntryInt{other})
	match := NewIPEntryList(t.ip)

	low := 0
	if ignoreGateway {
		low = 1
	}

	// for i := t.Size() - 1; i >= low; i-- {
	for i := t.Size(); i >= low; i-- {
		el := t.table[i]
		unmatch := NewIPEntryList(t.ip)
		var _match []flexrange.EntryInt
		var _unmatch []flexrange.EntryInt
		for it := targetList.Iterator(); it.HasNext(); {
			_, e := it.Next()
			_match, _unmatch = e.MatchResult(el)

			for l := 0; l < len(_match); l++ {
				match.PushEntry(_match[l])
			}

			for l := 0; l < len(_unmatch); l++ {
				unmatch.PushEntry(_unmatch[l])
			}

		}

		targetList = unmatch
	}

	if ignoreGateway == false {
		if dgw && t.dgw != nil {
			netDataRange := net.DataRange()
			gwDataRange := t.dgw.WarpperWithDataRange(uint32(t.Size()), big.NewInt(0))
			if netDataRange.Same(gwDataRange) {
				match = NewIPEntryList(t.ip)
				match.PushEntry(t.dgw)
				targetList = NewIPEntryList(t.ip)
			} else {
				if targetList.Len() > 0 {
					for it := targetList.Iterator(); it.HasNext(); {
						_, e := it.Next()

						m := e.Copy().(*flexrange.Entry)

						m.SetData(t.dgw.Data().Copy().(*flexrange.ExtendData))
						match.PushEntry(m)
					}
					targetList = NewIPEntryList(t.ip)
				}

			}

		}
	}

	return &MatchResult{
		Ip:      t.ip,
		Match:   match,
		Unmatch: targetList,
	}
}

// func (t AddressTable) Verify() bool {
// result := NextHopRecursionValidator{}.Validate(data)
// return result.Status()
// }

func (t *AddressTable) RecursionRouteProcess() {
	data := map[string]interface{}{
		"table": t,
	}
	result := NextHopRecursionValidator{}.Validate(data)

	if result.Status() == false {
		panic(result.Msg())
	}

	// for i := t.Size() - 1; i >= 0; i-- {
	for i := t.Size(); i >= 0; i-- {
		el := t.table[i]
		if i == 0 && t.dgw != nil {
			el = NewIPEntryList(t.ip)
			el.PushEntry(t.dgw)
		}
		// fmt.Println(el)
		for it := el.Iterator(); it.HasNext(); {
			_, re := it.Next()
			next := re.Data().Data.(*NextHop)

			for _, nh := range next.next {
				hop := nh.(*Hop)
				// fmt.Println(re, hop.Interface, hop.Ip)
				if hop.Interface == "" && hop.Ip != "" {
					fmt.Printf("进入递归路由查询: route: %+v, hop: %+v\n", re, hop)
					// 如果接口为空，IP地址不为空，符合递归路由查询条件
					net, _ := ParseIPNet(hop.Ip)

					// 进行递归路由检查时，也应该对默认路由进行匹配
					rmr := t.Match(net, true, false)
					if rmr.IsMatch() == false {
						panic(fmt.Sprintf("next hop ip %s match route failed", hop.Ip))
					}
					// 轮询Match结果，如果有多个下一跳，则针对每个下一跳生成Nexthop
					for it2 := rmr.Match.Iterator(); it2.HasNext(); {
						_, re2 := it2.Next()
						next2 := re2.Data().Data.(*NextHop)

						for index, nh2 := range next2.next {
							if nh2.(*Hop).Interface == "" {
								panic(fmt.Sprintf("next hop ip %s's recursion route is invalid, %v", hop.Ip, nh2))
							}
							if index == 0 {
								// 递归路由的第一个下一跳，直接修改原有HOP信息
								hop.Interface = nh2.(*Hop).Interface
								if nh2.(*Hop).Connected {
									// fmt.Println(nh2)
									fmt.Printf("递归路由查询成功: 原hop信息: %+v, 修改信息, {interface: %s, ip: %s}\n", hop, nh2.(*Hop).Interface, hop.Ip)
									// hop.Ip = h
								} else {
									fmt.Printf("递归路由查询成功: 原hop信息: %+v, 修改信息, {interface: %s, ip: %s}\n", hop, nh2.(*Hop).Interface, nh2.(*Hop).Ip)
									hop.Ip = nh2.(*Hop).Ip
								}
							} else {
								// 递归路由的其他下一跳，需要添加到原有NextHop的next列表中
								n := nh2.Copy()
								fmt.Printf("递归路由查询成功: 添加信息, {interface: %s, ip: %s}\n", nh2.(*Hop).Interface, nh2.(*Hop).Ip)
								next.next = append(next.next, n.(*Hop))
							}

						}
					}

				}
			}
		}
	}
}

func (t AddressTable) String() string {
	return t.Json()
}

func (t AddressTable) Json() string {
	var l int
	if t.ip == IPv4 {
		l = 32
	} else {
		l = 128
	}
	var rs []string
	// for i := l - 1; i >= 0; i-- {
	for i := l; i >= 0; i-- {
		dr := t.table[i]

		if dr.Len() > 0 {
			var ls string
			for it := dr.Iterator(); it.HasNext(); {
				_, e := it.Next()
				//func NewIPRangeFromInt(low *big.Int, high *big.Int, fa IPFamily) *IPRange {
				ipr := NewIPRangeFromInt(e.Low(), e.High(), t.ip)
				ip, err := ipr.SuperNet()
				if err != nil {
					panic(err)
				}

				if ls == "" {
					ls = fmt.Sprintf("{\"net\":\"%s\", \"next\":[%s]}", ip, e.Data().Data)
				} else {
					ls = fmt.Sprintf("{\"net\":\"%s\", \"next\":[%s]}", ip, e.Data().Data)
				}
				rs = append(rs, ls)
			}
		}

	}

	if t.dgw != nil {

		ipr := NewIPRangeFromInt(t.dgw.Low(), t.dgw.High(), t.ip)
		ip, err := ipr.SuperNet()
		if err != nil {
			panic(err)
		}
		ls := fmt.Sprintf("{\"net\":\"%s\", \"next\":%s}", ip, t.dgw.Data().Data)
		rs = append(rs, ls)
	}
	text := fmt.Sprintf("{\"type:\":\"%s\", \"list\":[%s]}", t.ip, strings.Join(rs, ",\n"))
	return text
}

type AddressTableIterator struct {
	ip IPFamily
	el *flexrange.EntryList
	it *flexrange.EntryListIterator
}

func (t *AddressTable) Iterator() *AddressTableIterator {
	el := flexrange.EntryList{}

	for _, tel := range t.table {
		for it := tel.Iterator(); it.HasNext(); {
			_, e := it.Next()
			el.PushEntry(e.Copy().(flexrange.EntryInt))
		}
	}
	if t.dgw != nil {
		el.PushEntry(t.dgw.Copy().(flexrange.EntryInt))
	}

	ati := AddressTableIterator{
		ip: t.ip,
		el: &el,
		it: el.Iterator(),
	}

	return &ati
}

func (ati *AddressTableIterator) HasNext() bool {
	return ati.it.HasNext()
}

func (ati *AddressTableIterator) Next() (*IPNet, *NextHop) {
	_, e := ati.it.Next()
	ipr := NewIPRangeFromInt(e.Low(), e.High(), ati.ip)
	ip, err := ipr.SuperNet()
	if err != nil {
		panic(err)
	}

	return ip, e.Data().Data.(*NextHop)
}
