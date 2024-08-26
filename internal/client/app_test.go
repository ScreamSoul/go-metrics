package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"syscall"
	"testing"
	"time"

	"github.com/screamsoul/go-metrics-tpl/internal/models/metrics"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type MockMetricStorage struct {
	mock.Mock
}

// List mocks the List method
func (m *MockMetricStorage) List(ctx context.Context) ([]metrics.Metrics, error) {
	args := m.Called(ctx)
	return args.Get(0).([]metrics.Metrics), args.Error(1)
}

func (m *MockMetricStorage) Update() {
	m.Called()
}

func (m *MockMetricStorage) UpdateRuntime() {
	m.Called()
}

func (m *MockMetricStorage) UpdateGopsutil() {
	m.Called()
}

func TestUpdaterUpdatesMetricsAtRegularIntervals(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mockRepo := new(MockMetricStorage)
	pollInterval := 100 * time.Millisecond

	mockRepo.On("Update").Return()
	mockRepo.On("UpdateRuntime").Return()
	mockRepo.On("UpdateGopsutil").Return()

	go updater(ctx, mockRepo, pollInterval)

	time.Sleep(300 * time.Millisecond)
	cancel()

	mockRepo.AssertNumberOfCalls(t, "Update", 3)
	mockRepo.AssertNumberOfCalls(t, "UpdateRuntime", 3)
	mockRepo.AssertNumberOfCalls(t, "UpdateGopsutil", 3)
}

// Successfully retrieves metrics from metricRepo and sends them using metricClient
func TestSender_SuccessfullySendsMetrics(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Create a MetricsClient instance
	metricClient := NewMetricsClient(
		false, "", server.URL, "127.0.0.1", nil,
	)
	mockMetricStorage := new(MockMetricStorage)

	metricsList := []metrics.Metrics{{ID: "test_metric", MType: metrics.Gauge, Value: new(float64)}}
	mockMetricStorage.On("List", ctx).Return(metricsList, nil)

	backoffIntervals := []time.Duration{time.Millisecond}
	reportInterval := time.Millisecond

	go sender(ctx, mockMetricStorage, backoffIntervals, metricClient, reportInterval)

	time.Sleep(2 * reportInterval)
}

func TestStartAgent(t *testing.T) {
	cfg := &Config{}
	logger := zap.NewNop()

	time.AfterFunc(3*time.Second, func() {
		require.NoError(t, syscall.Kill(syscall.Getpid(), syscall.SIGINT))
	})

	Start(cfg, logger)
}
