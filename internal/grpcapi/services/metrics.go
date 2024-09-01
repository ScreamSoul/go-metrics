package services

import (
	"context"
	"strings"

	"github.com/screamsoul/go-metrics-tpl/internal/models/metrics"
	pb "github.com/screamsoul/go-metrics-tpl/internal/proto"
	"github.com/screamsoul/go-metrics-tpl/internal/repositories"
	"github.com/screamsoul/go-metrics-tpl/pkg/logging"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type MetricServer struct {
	pb.MetricsServiceServer

	store  repositories.MetricStorage
	logger *zap.Logger
}

func NewMetricServer(metricRepo repositories.MetricStorage) *MetricServer {
	logger := logging.GetLogger()

	return &MetricServer{store: metricRepo, logger: logger}
}

func (s *MetricServer) UpdateMetrics(ctx context.Context, in *pb.MetricsRequest) (*emptypb.Empty, error) {

	chunkSize := 100

	for i := 0; i < len(in.Metrics); i += chunkSize {
		end := i + chunkSize

		// necessary check to avoid slicing beyond
		// slice capacity
		if end > len(in.Metrics) {
			end = len(in.Metrics)
		}
		metricChunk := make([]metrics.Metrics, end-i)

		// generate metrics chunk
		for _, m := range in.Metrics[i:end] {
			metric, err := metrics.NewMetric(
				strings.ToLower(m.GetMType().String()),
				m.GetName(),
				"",
			)
			if err != nil {
				return nil, status.Errorf(codes.DataLoss, err.Error())
			}

			metric.Value = &m.Value
			metric.Delta = &m.Delta

			err = metric.ValidateValue()
			if err != nil {
				return nil, status.Errorf(codes.DataLoss, err.Error())
			}

			metricChunk = append(metricChunk, *metric)
		}

		err := s.store.BulkAdd(ctx, metricChunk)
		if err != nil {
			s.logger.Error("internal error", zap.Error(err))
			return nil, status.Errorf(codes.Internal, `internal error`)
		}

	}

	return &emptypb.Empty{}, nil
}
