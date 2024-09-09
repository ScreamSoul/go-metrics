package ipmask_test

import (
	"net"
	"testing"

	"github.com/screamsoul/go-metrics-tpl/pkg/ipmask"
	"github.com/stretchr/testify/assert"
)

func TestUnmarshalTextValidCIDR(t *testing.T) {
	cidr := &ipmask.CIDRIP{}
	err := cidr.UnmarshalText([]byte("192.168.1.0/24"))

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	expectedNetwork := &net.IPNet{
		IP:   net.IPv4(192, 168, 1, 0),
		Mask: net.CIDRMask(24, 32),
	}

	if !cidr.Network.IP.Equal(expectedNetwork.IP) || cidr.Network.Mask.String() != expectedNetwork.Mask.String() {
		t.Errorf("expected network %v, got %v", expectedNetwork, cidr.Network)
	}
}

func TestUnmarshalTextInvalidCIDR(t *testing.T) {
	cidr := &ipmask.CIDRIP{}
	err := cidr.UnmarshalText([]byte("invalid-cidr"))
	assert.Error(t, err)

	err = cidr.UnmarshalText([]byte(""))
	assert.NoError(t, err)
	assert.Nil(t, cidr.Network)
}

func TestValidIPv4AddressWithinCIDR(t *testing.T) {
	cidr := ipmask.CIDRIP{
		Network: &net.IPNet{
			IP:   net.ParseIP("192.168.1.0"),
			Mask: net.CIDRMask(24, 32),
		},
	}

	assert.Equal(t, true, cidr.CheckIPIncluded("192.168.1.10"))
	assert.Equal(t, false, cidr.CheckIPIncluded("127.0.0.1"))
	assert.Equal(t, false, cidr.CheckIPIncluded("sdasdas"))
}
