package network

import (
	"encoding/json"
	"fmt"
	"gin-vue-admin/utils"
	"testing"
)

var validIPv4RouteTests = []struct {
	net       []string
	it        string
	ip        string
	connect   bool
	defaut_gw bool
}{
	{[]string{"192.168.1.0/24", "192.168.3.0/24"}, "Gi0/0", "10.1.1.254", false, false},
	{[]string{"192.168.1.0/16", "192.169.0.0/16"}, "Gi0/1", "20.1.1.254", false, false},
	{[]string{"10.0.0.0/16", "10.1.0.0/16"}, "", "20.1.1.254", false, false},
	{[]string{"0.0.0.0/0"}, "Gi0/2", "1.1.1.1", false, false},
}

func TestRoute(t *testing.T) {
	rt1 := NewAddressTable(IPv4)
	for _, ss := range validIPv4RouteTests {
		for _, n := range ss.net {
			net, err := NewIPNet(n)
			if err != nil {
				t.Error(err)
			}
			nx := NewNextHop()
			_, err = nx.AddHop(ss.it, ss.ip, ss.connect, ss.defaut_gw, nil)
			rt1.PushRoute(net, nx)
			//rt.Push(net, nil)
		}
	}

	fmt.Println("111---------------------->", rt1.String())
	rt1.RecursionRouteProcess()
	fmt.Println("222---------------------->", rt1.String())

	// ok := rt1.Verify()
	// if ok == false {
	// fmt.Println("verify result: ", ok)
	// }

	b, err := json.Marshal(rt1)
	if err != nil {
		t.Error(err)
	}

	rt := &AddressTable{}
	err = json.Unmarshal(b, rt)
	if err != nil {
		t.Error(err)
	}

	net, err := NewNetworkFromString("192.167.1.1-192.168.1.255")
	if err != nil {
		t.Error(err)
	}
	//fmt.Printf("rt = %s\n", rt)
	mr := rt.Match(AbbrNet(*net), true, false)
	fmt.Printf("%s\n", mr)

	ok, ip := mr.IsSameIp()
	fmt.Printf("ok = %+v\n", ok)
	fmt.Printf("ip = %+v\n", ip)
	// fmt.Println("----------------------------------")
	net, err = NewNetworkFromString("192.168.1.1-192.168.1.255")
	mr = rt.Match(AbbrNet(*net), true, false)

	ok, ip = mr.IsSameIp()
	fmt.Printf("ok = %+v\n", ok)
	fmt.Printf("ip = %+v\n", ip)

	table, _ := mr.Table()
	fmt.Println(table.Column("interface"))
	fmt.Println(table.Column("net"))
	fmt.Println(table.Column("ip"))
	fmt.Println(table.Column("connected"))
	fmt.Println(table.Row(0))
	fmt.Println(table.Row(0).Cells("net", "interface", "ip"))

	fmt.Println(len(table.Column("ip").List()))

	// mr.Table()
	// fmt.Println("--------------------->", mr.Match.Jsonq())
}

type vs struct {
	Name string
}

func (v vs) String() string {
	return v.Name
}

func (v vs) Copy() utils.CopyAble {
	return &vs{
		Name: v.Name,
	}
}

var validIPv4RouteFilterTests = []struct {
	net        []string
	it         string
	ip         string
	connect    bool
	vs         vs
	default_gw bool
}{
	{[]string{"192.168.1.0/24", "192.168.3.0/24"}, "Gi0/0", "10.1.1.254", false, vs{"RED"}, false},
	{[]string{"172.16.1.0/24", "172.16.2.0/24"}, "Gi0/0", "10.1.1.100", false, vs{"RED"}, false},
	{[]string{"192.168.1.0/16", "192.169.0.0/16"}, "Gi0/1", "20.1.1.254", false, vs{"RED"}, false},
	{[]string{"192.168.1.0/16", "192.169.0.0/16"}, "Gi0/2", "30.1.1.254", false, vs{"RED"}, false},
	{[]string{"0.0.0.0/0"}, "Gi0/2", "1.1.1.1", false, vs{"BLUE"}, true},
	{[]string{"191.168.1.0/16", "191.169.0.0/16"}, "Gi0/1", "20.1.1.254", false, vs{"RED"}, false},
	{[]string{"10.1.1.0/24", "20.1.1.0/24"}, "Gi0/1", "", true, vs{"RED"}, false},
}

func TestIPv4RouteFilter(t *testing.T) {
	rt := NewAddressTable(IPv4)
	for _, ss := range validIPv4RouteFilterTests {
		for _, n := range ss.net {
			net, _ := NewIPNet(n)
			nx := NewNextHop()
			nx.AddHop(ss.it, ss.ip, ss.connect, ss.default_gw, ss.vs)
			rt.PushRoute(net, nx)
			//rt.Push(net, nil)
		}
	}

	addressFilter := rt.AddressFilter(false)
	fmt.Println(addressFilter.jsonq)

	a := addressFilter.Select("net", "next.ip", "next.connected", "next.interface").Ip("10.1.1.254").Get()
	fmt.Println(a)

	addressFilter = rt.AddressFilter(false)
	a = addressFilter.Select("net", "next.ip", "next.connected", "next.interface", "next.vs.Name").Connected(true).Get()
	fmt.Println(a)

	//
	// fmt.Println("-----------------------------------------")
	// m, d := rt.Filter("interface", "Gi0/1")
	// fmt.Printf("m = %+v\n", m)
	// fmt.Println("-----------------------------------------")
	// m, d = rt.Filter("interface", "Gi0/2")
	// fmt.Printf("m = %+v\n", m)
	// fmt.Println("-----------------------------------------")
	// m, d = rt.Filter("ip", "10.1.1.254")
	// fmt.Printf("m = %+v\n", m)
	// fmt.Printf("d = %+v\n", d)

	// fl, gl := rt.Flatten()
	// fmt.Printf("fl = %+v\n", fl)
	// fmt.Printf("gl = %+v\n", gl)
	//
}

//
// func TestIPv4RouteQuery(t *testing.T) {
// rt := NewAddressTable(IPv4)
// for _, ss := range validIPv4RouteFilterTests {
// for _, n := range ss.net {
// net, err := NewIPNet(n)
// if err != nil {
// t.Error(err)
// }
// nx := NewNextHop()
// nx.AddHop(ss.it, ss.ip, ss.connect, nil)
// rt.PushRoute(net, nx)
// }
// }

//js := rt.Jsonq()
// a := rt.AddressFilter(true).Interface("Gi0/0").Select("net").Connected(false).Get()
//fmt.Printf("a = %+v\n", a)
// for _, l := range a {
// fmt.Printf("l = %+v\n", l.(map[string]interface{})["net"])
// }

//j := rt.Json()
//buf := new(bytes.Buffer)
//json.Indent(buf, []byte(j), "", "  ")
//fmt.Println(buf)

//at := rt.AddressFilter(true)
//fmt.Printf("at.jsonq = %+v\n", at.jsonq)
//itmes := at.jsonq.Where("next.Data.Next", "!=", "").Select("next.Data.Next").Get()
//fmt.Printf("itmes = %+v\n", itmes)
//for it := rt.Iterator(); it.HasNext(); {
//net, e := it.Next()
//fmt.Printf("net = %+v\n", net)
//fmt.Printf("e = %+v\n", e)
//}
// }
