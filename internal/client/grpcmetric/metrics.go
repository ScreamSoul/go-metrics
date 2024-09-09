package grpcmetric

import (
	"context"

	"github.com/screamsoul/go-metrics-tpl/internal/models/metrics"
	pb "github.com/screamsoul/go-metrics-tpl/internal/proto"
	"github.com/screamsoul/go-metrics-tpl/pkg/logging"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type GRPCMetricsClient struct {
	logger *zap.Logger
	mc     pb.MetricsServiceClient
	conn   *grpc.ClientConn
}

func NewGRPCMetricsClient(
	conn *grpc.ClientConn,
) *GRPCMetricsClient {
	logger := logging.GetLogger()

	client := &GRPCMetricsClient{
		logger,
		pb.NewMetricsServiceClient(conn),
		conn,
	}
	return client
}

func (client *GRPCMetricsClient) SendMetric(ctx context.Context, metricsList []metrics.Metrics) error {
	var mr pb.MetricsRequest
	for _, m := range metricsList {
		var mType pb.Metric_MType

		switch m.MType {
		case metrics.Gauge:
			mType = pb.Metric_GAUGE
		case metrics.Counter:
			mType = pb.Metric_COUNTER
		}

		mr.Metrics = append(mr.Metrics, &pb.Metric{
			Name:  m.ID,
			Delta: *m.Delta,
			Value: *m.Value,
			MType: mType,
		})
	}

	_, err := client.mc.UpdateMetrics(ctx, &mr)

	return err
}
