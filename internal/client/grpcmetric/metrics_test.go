package grpcmetric_test

import (
	"context"
	"log"
	"net"
	"testing"

	"github.com/screamsoul/go-metrics-tpl/internal/client/grpcmetric"
	"github.com/screamsoul/go-metrics-tpl/internal/models/metrics"
	pb "github.com/screamsoul/go-metrics-tpl/internal/proto"
	"github.com/screamsoul/go-metrics-tpl/pkg/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/types/known/emptypb"
)

type mockMetricsServer struct {
	pb.UnimplementedMetricsServiceServer
}

func (m *mockMetricsServer) UpdateMetrics(ctx context.Context, req *pb.MetricsRequest) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

func dialer() func(context.Context, string) (net.Conn, error) {
	listener := bufconn.Listen(1024 * 1024)
	server := grpc.NewServer()
	pb.RegisterMetricsServiceServer(server, &mockMetricsServer{})
	go func() {
		if err := server.Serve(listener); err != nil {
			log.Fatal(err)
		}
	}()

	return func(context.Context, string) (net.Conn, error) {
		return listener.Dial()
	}
}

func TestGRPCMetricsClient_SendMetric(t *testing.T) {
	// Setup
	conn, err := grpc.NewClient(
		"localhost",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(dialer()),
	)

	require.NoError(t, err)
	defer utils.CloseForse(conn)

	mc := grpcmetric.NewGRPCMetricsClient(conn)

	// Test data
	metricsList := []metrics.Metrics{
		{
			ID:    "metric1",
			Delta: new(int64),
			Value: new(float64),
			MType: metrics.Gauge,
		},
		{
			ID:    "metric2",
			Delta: new(int64),
			Value: new(float64),
			MType: metrics.Counter,
		},
	}

	err = mc.SendMetric(context.Background(), metricsList)

	assert.NoError(t, err)
}
