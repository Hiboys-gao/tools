package network

import (
	"fmt"
)

type PairFormatter string

func (ng *NetworkGroup) HostList() (nets []string) {
	nets = []string{}

	for _, netList := range []NetworkList{ng.ipv4, ng.ipv6} {
		n, nl := netList.Aggregate()
		if nl == nil {
			if n == nil {
				continue
			}
			switch n.AbbrNet.(type) {
			case *IPNet:
				if n.AddressType() == HOST {
					nets = append(nets, n.AbbrNet.(*IPNet).IP.String())
				}
			case *IPRange:
				if n.AddressType() == HOST {
					nets = append(nets, n.AbbrNet.(*IPRange).Start.String())
				}
			}
		}
	}

	return
}

func (ng *NetworkGroup) SubnetList(mask bool, maskFormat, prefixFormat PairFormatter) (nets []string) {
	nets = []string{}

	for _, netList := range []NetworkList{ng.ipv4, ng.ipv6} {
		n, nl := netList.Aggregate()
		if nl == nil {
			if n == nil {
				continue
			}
			switch n.AbbrNet.(type) {
			case *IPNet:
				if n.AddressType() == SUBNET {
					if n.Type() == IPv4 {
						if mask {
							nets = append(nets, fmt.Sprintf(string(maskFormat), n.AbbrNet.(*IPNet).IP.String(), MasktoIP(n.AbbrNet.(*IPNet).Mask).String()))
						} else {
							nets = append(nets, fmt.Sprintf(string(prefixFormat), n.AbbrNet.(*IPNet).IP.String(), n.AbbrNet.(*IPNet).Mask.Prefix()))
						}
					} else {
						nets = append(nets, fmt.Sprintf(string(prefixFormat), n.AbbrNet.(*IPNet).IP.String(), n.AbbrNet.(*IPNet).Mask.Prefix()))
					}
				}
			case *IPRange:
				if n.AddressType() == SUBNET {
					ips := n.AbbrNet.(*IPRange).CIDRs()
					ipnet := ips[0]
					if ipnet.Type() == IPv4 {
						if mask {
							nets = append(nets, fmt.Sprintf(string(maskFormat), ipnet.IP.String(), MasktoIP(ipnet.Mask).String()))
						} else {
							nets = append(nets, fmt.Sprintf(string(prefixFormat), ipnet.IP.String(), ipnet.Mask.Prefix()))
						}
					} else {
						nets = append(nets, fmt.Sprintf(string(prefixFormat), ipnet.IP.String(), ipnet.Mask.Prefix()))
					}
				}
			}
		}

	}

	return
}

func (ng *NetworkGroup) RangeList(format PairFormatter) (nets []string) {
	nets = []string{}

	for _, netList := range []NetworkList{ng.ipv4, ng.ipv6} {
		n, nl := netList.Aggregate()
		if nl == nil {
			if n == nil {
				continue
			}
			switch n.AbbrNet.(type) {
			case *IPRange:
				if n.AddressType() == RANGE {
					nets = append(nets, fmt.Sprintf(string(format), n.AbbrNet.(*IPRange).Start.String(), n.AbbrNet.(*IPRange).End.String()))
				}
			}
		}
	}

	return
}
