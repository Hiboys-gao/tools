package network

import (
	"encoding/json"
	"gin-vue-admin/core"
	"gin-vue-admin/global"
	"gin-vue-admin/initialize"
	"gin-vue-admin/utils"
	"gin-vue-admin/utils/network"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func init() {
	global.GVA_VP = core.Viper("../../../config.yaml") // 初始化Viper
	global.GVA_LOG = core.Zap()                        // 初始化zap日志库
	test := true
	initialize.SQLite(test)
	err := global.GVA_SQLite.AutoMigrate(&networkTestStruct{})
	if err != nil {
		panic(err)
	}
}

type networkTestStruct struct {
	global.GVA_MODEL `mapstructure:",squash" json:"-"`
	IP               *network.IP           `gorm:"type:ip"`
	Mask             *network.IPMask       `gorm:"type:mask"`
	IPNet            *network.IPNet        `gorm:"type:ipnet"`
	IPRange          *network.IPRange      `gorm:"type:iprange"`
	Network          *network.Network      `gorm:"type:network"`
	NetworkList      *network.NetworkList  `gorm:"type:network_list"`
	NetworkGroup     *network.NetworkGroup `gorm:"type:network_group"`
}

type networkTestInput struct {
	ip           string
	mask         string
	ipnet        string
	iprange      string
	network      string
	networkList  string
	networkGroup string
}

func (n networkTestInput) newTestStruct() *networkTestStruct {
	var ip *network.IP
	var mask *network.IPMask
	var ipnet *network.IPNet
	var iprange *network.IPRange
	var net *network.Network
	var nl *network.NetworkList
	var ng *network.NetworkGroup

	if n.ip != "" {
		ip, _ = network.ParseIP(n.ip)
	}
	if n.mask != "" {
		maskIp, _ := network.ParseIP(n.mask)
		mask, _ = network.IPtoMask(maskIp)
	}
	if n.ipnet != "" {
		ipnet, _ = network.ParseIPNet(n.ipnet)
	}
	if n.iprange != "" {
		iprange, _ = network.NewIPRange(n.iprange)
	}

	if n.network != "" {
		net, _ = network.NewNetworkFromString(n.network)
	}

	if n.networkList != "" {
		nlNet, _ := network.NewNetworkFromString(n.networkList)
		nl, _ = network.NewNetworkListFromList([]*network.Network{nlNet})
	}
	if n.networkGroup != "" {
		ng, _ = network.NewNetworkGroupFromString(n.networkGroup)
	}

	netStruct := networkTestStruct{}

	if ip != nil {
		netStruct.IP = ip
	}

	if mask != nil {
		netStruct.Mask = mask
	}

	if ipnet != nil {
		netStruct.IPNet = ipnet
	}

	if iprange != nil {
		netStruct.IPRange = iprange
	}

	if net != nil {
		netStruct.Network = net
	}

	if nl != nil {
		netStruct.NetworkList = nl
	}

	if ng != nil {
		netStruct.NetworkGroup = ng
	}

	return &netStruct
}

var validNetworkGormTest = []struct {
	got  networkTestInput
	want networkTestInput
}{
	{
		got: networkTestInput{
			ip:    "1.1.1.1",
			mask:  "255.255.255.0",
			ipnet: "1.1.1.0/24",
		},
		want: networkTestInput{
			ip:    "1.1.1.1",
			mask:  "255.255.255.0",
			ipnet: "1.1.1.0/24",
		},
	},
	{
		got: networkTestInput{
			networkList: "1.1.1.1",
		},
		want: networkTestInput{
			networkList: "1.1.1.1",
		},
	},
	{
		got: networkTestInput{
			networkGroup: "1.1.1.0/24",
		},
		want: networkTestInput{
			networkGroup: "1.1.1.0/24",
		},
	},
}

func TestNetworkGorm(t *testing.T) {
	for _, ss := range validNetworkGormTest {
		netStruct := ss.got.newTestStruct()
		global.GVA_SQLite.Create(netStruct)
		got := networkTestStruct{}
		got.ID = netStruct.ID
		global.GVA_SQLite.Find(&got)
		want := ss.want.newTestStruct()
		byteWant, _ := json.Marshal(&want)
		byteGot, _ := json.Marshal(&got)

		if !cmp.Equal(byteWant, byteGot) {
			t.Errorf("%+v, got:%+v, want:%+v", ss.got, got, want)
		}
	}
}

var validNetworkGroupFlexMatchData = []struct {
	one   string
	two   string
	any   bool
	flags int
	ok    bool
}{
	{
		one:   "1.1.1.0/24",
		two:   "1.1.1.1",
		any:   false,
		flags: utils.OBJECT_OPTION_ONE_MATCH_TWO,
		ok:    true,
	},
	{
		one:   "1.1.1.1",
		two:   "1.1.1.0/24",
		any:   false,
		flags: utils.OBJECT_OPTION_TWO_MATCH_ONE,
		ok:    true,
	},
	{
		one:   "1.1.1.0/24",
		two:   "1.1.1.0/24",
		any:   false,
		flags: utils.OBJECT_OPTION_SAME,
		ok:    true,
	},
	{
		one:   "1.1.1.1,::1.1.1.1",
		two:   "::1.1.1.1",
		any:   false,
		flags: utils.OBJECT_OPTION_ONE_MATCH_TWO,
		ok:    true,
	},
	{
		one:   "0.0.0.0/0",
		two:   "1.1.1.1",
		any:   false,
		flags: utils.OBJECT_OPTION_ONE_MATCH_TWO,
		ok:    false,
	},
	{
		one:   "0.0.0.0/0",
		two:   "1.1.1.1",
		any:   true,
		flags: utils.OBJECT_OPTION_ONE_MATCH_TWO,
		ok:    true,
	},
	{
		one:   "0.0.0.0/0",
		two:   "0.0.0.0/0",
		any:   false,
		flags: utils.OBJECT_OPTION_ONE_MATCH_TWO,
		ok:    false,
	},
	{
		one:   "0.0.0.0/0",
		two:   "0.0.0.0/0",
		any:   true,
		flags: utils.OBJECT_OPTION_ONE_MATCH_TWO,
		ok:    true,
	},
	{
		one:   "0.0.0.0/0,::/0",
		two:   "0.0.0.0/0",
		any:   false,
		flags: utils.OBJECT_OPTION_SAME,
		ok:    false,
	},
	{
		one:   "0.0.0.0/0,::1.1.1.1",
		two:   "0.0.0.0/0",
		any:   true,
		flags: utils.OBJECT_OPTION_ONE_MATCH_TWO,
		ok:    true,
	},
	{
		one:   "0.0.0.0/0,::1.1.1.1",
		two:   "0.0.0.0/0",
		any:   false,
		flags: utils.OBJECT_OPTION_ONE_MATCH_TWO,
		ok:    true,
	},
}

func TestNetworkGroupFlexMatch(t *testing.T) {
	for _, ss := range validNetworkGroupFlexMatchData {
		one, _ := network.NewNetworkGroupFromString(ss.one)
		two, _ := network.NewNetworkGroupFromString(ss.two)

		got := one.MatchWithOption(two, ss.flags, ss.any)
		if got != ss.ok {
			t.Errorf("one: %s, two: %s, flags: %d, got: %v, want: %v", ss.one, ss.two, ss.flags, got, ss.ok)
		}
	}
}
