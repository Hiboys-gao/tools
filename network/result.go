package network

import (
	"fmt"
	"strings"
	"tools/flexrange"
)

type MatchResult struct {
	//match   []flexrange.EntryInt
	//unmatch []flexrange.EntryInt
	Ip      IPFamily
	Match   *flexrange.EntryList
	Unmatch *flexrange.EntryList
}

func NewMatchResult(ip IPFamily, match *flexrange.EntryList, unmatch *flexrange.EntryList) *MatchResult {
	return &MatchResult{
		Ip:      ip,
		Match:   match,
		Unmatch: unmatch,
	}
}

func (m MatchResult) IsMatch() bool {
	if m.Match.Len() > 0 && m.Unmatch.Len() == 0 {
		return true
	} else {
		return false
	}
}

func (m MatchResult) IsSameIp() (bool, string) {
	if !m.IsMatch() {
		return false, ""
	}

	ip := ""
	for it := m.Match.Iterator(); it.HasNext(); {
		index, e := it.Next()
		nh := e.Data().Data.(*NextHop)
		b, i := nh.IsSameIp()
		if !b {
			return false, ""
		}

		if index == 0 {
			ip = i
		} else {
			if ip != i {
				return false, ""
			}
		}
	}
	if ip == "" {
		return false, ""
	} else {
		return true, ip
	}
}

func (m MatchResult) IsSameInterface() (bool, string) {
	if !m.IsMatch() {
		return false, ""
	}

	inf := ""
	for it := m.Match.Iterator(); it.HasNext(); {
		index, e := it.Next()
		nh := e.Data().Data.(*NextHop)
		b, i := nh.IsSameInterface()
		if !b {
			return false, ""
		}

		if index == 0 {
			inf = i
		} else {
			if inf != i {
				return false, ""
			}
		}
	}
	if inf == "" {
		return false, ""
	} else {
		return true, inf
	}

}

//

func (m MatchResult) String() string {
	match := "[]"
	unmatch := "[]"
	if m.Match != nil {
		l := []string{}
		for it := m.Match.Iterator(); it.HasNext(); {
			_, e := it.Next()
			ip1 := NewIPFromInt(e.Low(), m.Ip)
			ip2 := NewIPFromInt(e.High(), m.Ip)
			l = append(l, fmt.Sprintf("%s-%s:%+v", ip1, ip2, e.Data()))
		}
		match = "[" + strings.Join(l, ",") + "]"
	}
	if m.Unmatch != nil {
		l := []string{}
		for it := m.Unmatch.Iterator(); it.HasNext(); {
			_, e := it.Next()
			ip1 := NewIPFromInt(e.Low(), m.Ip)
			ip2 := NewIPFromInt(e.High(), m.Ip)
			l = append(l, fmt.Sprintf("%s-%s", ip1, ip2))
		}
		unmatch = "[" + strings.Join(l, ",") + "]"
	}

	return fmt.Sprintf("match: %s\nunmatch: %s", match, unmatch)
}

func (m MatchResult) IsFullMatch() bool {
	return false
}
