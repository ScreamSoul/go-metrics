package interceptors

import (
	"context"

	"github.com/screamsoul/go-metrics-tpl/pkg/ipmask"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

func NewTrustedIPMiddleware(cidrip ipmask.CIDRIP) func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if cidrip.Network == nil {
			return handler(ctx, req)
		}
		p, ok := peer.FromContext(ctx)
		// extracts the client's IP address
		if !ok {
			return nil, status.Error(codes.Internal, "not check ip addr")
		}

		if !cidrip.CheckIPIncluded(p.Addr.String()) {
			return nil, status.Error(codes.PermissionDenied, "access dinied")
		}
		return handler(ctx, req)
	}
}
