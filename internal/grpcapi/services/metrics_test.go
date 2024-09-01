package services_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/gojuno/minimock/v3"
	"github.com/screamsoul/go-metrics-tpl/internal/grpcapi/services"
	pb "github.com/screamsoul/go-metrics-tpl/internal/proto"

	"github.com/stretchr/testify/assert"
)

// Successfully processes and updates metrics in chunks of 100
func TestUpdateMetricsProcessesChunks__Success(t *testing.T) {
	ctx := context.Background()
	mc := minimock.NewController(t)

	mockStore := NewMetricStorageMock(mc)
	defer mockStore.MinimockFinish()

	server := services.NewMetricServer(mockStore)

	metricsList := make([]*pb.Metric, 200)
	for i := 0; i < 200; i++ {
		metricsList[i] = &pb.Metric{
			Name:  fmt.Sprintf("metric%d", i),
			MType: pb.Metric_COUNTER,
			Value: float64(i),
			Delta: int64(i),
		}
	}

	req := &pb.MetricsRequest{Metrics: metricsList}

	mockStore.BulkAddMock.Return(nil)

	_, err := server.UpdateMetrics(ctx, req)
	assert.NoError(t, err)
}

func TestUpdateMetricsProcessesChunks__Error(t *testing.T) {
	ctx := context.Background()
	mc := minimock.NewController(t)

	mockStore := NewMetricStorageMock(mc)
	defer mockStore.MinimockFinish()

	server := services.NewMetricServer(mockStore)

	metricsList := make([]*pb.Metric, 200)
	for i := 0; i < 150; i++ {
		metricsList[i] = &pb.Metric{
			Name:  fmt.Sprintf("metric%d", i),
			MType: pb.Metric_COUNTER,
			Value: float64(i),
			Delta: int64(i),
		}
	}

	req := &pb.MetricsRequest{Metrics: metricsList}

	mockStore.BulkAddMock.Return(errors.New("some err"))

	_, err := server.UpdateMetrics(ctx, req)
	assert.Error(t, err)
}
