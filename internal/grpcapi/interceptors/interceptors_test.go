package interceptors_test

import (
	"context"
	"net"
	"testing"

	"github.com/screamsoul/go-metrics-tpl/internal/grpcapi/interceptors"
	"github.com/screamsoul/go-metrics-tpl/pkg/ipmask"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/peer"
)

func TestMiddlewareAllowsRequestWithinCIDR(t *testing.T) {
	cidrip := ipmask.CIDRIP{
		Network: &net.IPNet{
			IP:   net.ParseIP("192.168.1.0"),
			Mask: net.CIDRMask(24, 32),
		},
	}
	middleware := interceptors.NewTrustedIPMiddleware(cidrip)
	ctx := peer.NewContext(context.Background(), &peer.Peer{
		Addr: &net.IPAddr{IP: net.ParseIP("192.168.1.10")},
	})
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "success", nil
	}
	resp, err := middleware(ctx, nil, nil, handler)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp != "success" {
		t.Fatalf("expected success, got %v", resp)
	}
}

func TestMiddlewareAllowsRequestWithinCIDRFailedIAddr(t *testing.T) {
	cidrip := ipmask.CIDRIP{
		Network: &net.IPNet{
			IP:   net.ParseIP("192.168.1.0"),
			Mask: net.CIDRMask(24, 32),
		},
	}
	middleware := interceptors.NewTrustedIPMiddleware(cidrip)
	ctx := peer.NewContext(context.Background(), &peer.Peer{
		Addr: &net.IPAddr{IP: net.ParseIP("192.168.2.10")},
	})
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "success", nil
	}
	_, err := middleware(ctx, nil, nil, handler)
	assert.Error(t, err)
}

func TestMiddlewareAllowsRequestWithinCIDRWithoutPeer(t *testing.T) {
	cidrip := ipmask.CIDRIP{
		Network: &net.IPNet{
			IP:   net.ParseIP("192.168.1.0"),
			Mask: net.CIDRMask(24, 32),
		},
	}
	middleware := interceptors.NewTrustedIPMiddleware(cidrip)

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "success", nil
	}
	_, err := middleware(context.Background(), nil, nil, handler)
	assert.Error(t, err)
}

func TestMiddlewareHandlesNilNetworkGracefully(t *testing.T) {
	cidrip := ipmask.CIDRIP{
		Network: nil,
	}
	middleware := interceptors.NewTrustedIPMiddleware(cidrip)
	ctx := peer.NewContext(context.Background(), &peer.Peer{
		Addr: &net.IPAddr{IP: net.ParseIP("192.168.1.10")},
	})
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "success", nil
	}
	resp, err := middleware(ctx, nil, nil, handler)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp != "success" {
		t.Fatalf("expected success, got %v", resp)
	}
}
