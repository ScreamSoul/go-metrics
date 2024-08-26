package ipmask

import (
	"net"
)

type CIDRIP struct {
	Network *net.IPNet
}

func (cidr *CIDRIP) UnmarshalText(b []byte) error {
	ipMask := string(b)

	if ipMask == "" {
		return nil
	}

	_, network, err := net.ParseCIDR(ipMask)
	if err != nil {
		return err
	}

	cidr.Network = network

	return nil
}

func (cidr CIDRIP) CheckIPIncluded(ip string) bool {
	ipv4 := net.ParseIP(ip)
	if ipv4 == nil {
		return false
	}
	return cidr.Network.Contains(ipv4)
}
