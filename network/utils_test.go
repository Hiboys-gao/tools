package network

import (
	"testing"

	"gin-vue-admin/utils"

	"github.com/google/go-cmp/cmp"
)

var validHostListTests = []struct {
	this   string
	format string
	want   []string
}{
	{
		this:   "1.1.1.1/32",
		format: "range %s %s",
		want: []string{
			"1.1.1.1",
		},
	},
	{
		this:   "::1:1:1:1/128",
		format: "range %s %s",
		want: []string{
			"::1:1:1:1",
		},
	},
	{
		this:   "1.1.1.1-1.1.1.1",
		format: "range %s %s",
		want: []string{
			"1.1.1.1",
		},
	},
	{
		this:   "::1:1:1:1-::1:1:1:1",
		format: "range %s %s",
		want: []string{
			"::1:1:1:1",
		},
	},
}

func TestHostList(t *testing.T) {
	for _, ss := range validHostListTests {
		ng, _ := NewNetworkGroupFromString(ss.this)

		got := ng.HostList()

		if !cmp.Equal(got, ss.want) {
			t.Errorf("this:%s, got:%+v, want:%+v", ss.this, got, ss.want)
		}
	}
}

var validSubnetListTests = []struct {
	this         string
	maskFormat   string
	prefixFormat string
	mask         bool
	want         []string
}{
	{
		this:         "1.1.1.1/24",
		maskFormat:   "%s %s",
		prefixFormat: "%s/%d",
		mask:         true,
		want: []string{
			"1.1.1.0 255.255.255.0",
		},
	},
	{
		this:         "::1:1:1:1/120",
		maskFormat:   "%s %s",
		prefixFormat: "%s/%d",
		mask:         true,
		want: []string{
			"::1:1:1:0/120",
		},
	},
	{
		this:         "1.1.1.1/24",
		maskFormat:   "%s/%s",
		prefixFormat: "%s/%d",
		mask:         true,
		want: []string{
			"1.1.1.0/255.255.255.0",
		},
	},
	{
		this:         "1.1.1.1-1.1.1.1",
		maskFormat:   "%s/%s",
		prefixFormat: "%s/%d",
		mask:         true,
		want:         []string{},
	},
	{
		this:         "1.1.1.1-1.1.1.2",
		maskFormat:   "%s/%s",
		prefixFormat: "%s/%d",
		mask:         true,
		want:         []string{},
	},
}

func TestSubnetList(t *testing.T) {
	for _, ss := range validSubnetListTests {
		ng, _ := NewNetworkGroupFromString(ss.this)
		maskFormat, _ := utils.NewPairFormatter(ss.maskFormat)
		prefixFormat, _ := utils.NewPairFormatter(ss.prefixFormat)

		got := ng.SubnetList(ss.mask, maskFormat, prefixFormat)

		if !cmp.Equal(got, ss.want) {
			t.Errorf("this:%s, got:%+v, want:%+v", ss.this, got, ss.want)
		}
	}
}

var validRangeListTests = []struct {
	this   string
	format string
	mask   bool
	want   []string
}{
	{
		this:   "1.1.1.1/24",
		format: "%s %s",
		mask:   true,
		want:   []string{},
	},
	{
		this:   "1.1.1.1-1.1.1.2",
		format: "%s %s",
		mask:   true,
		want: []string{
			"1.1.1.1 1.1.1.2",
		},
	},
	{
		this:   "::1:1:1:1-::1:1:1:2",
		format: "%s-%s",
		mask:   true,
		want: []string{
			"::1:1:1:1-::1:1:1:2",
		},
	},
	{
		this:   "::1:1:1:1-::1:1:1:1",
		format: "%s-%s",
		mask:   true,
		want:   []string{},
	},
}

func TestRangeList(t *testing.T) {
	for _, ss := range validRangeListTests {
		ng, _ := NewNetworkGroupFromString(ss.this)
		format, _ := utils.NewPairFormatter(ss.format)

		got := ng.RangeList(format)

		if !cmp.Equal(got, ss.want) {
			t.Errorf("this:%s, got:%+v, want:%+v", ss.this, got, ss.want)
		}
	}
}
