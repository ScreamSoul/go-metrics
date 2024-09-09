// correctly compresses a valid byte slice body
package restymetric_test

import (
	"context"
	"crypto/rsa"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/screamsoul/go-metrics-tpl/internal/client/restymetric"
	"github.com/screamsoul/go-metrics-tpl/internal/models/metrics"
	"github.com/stretchr/testify/assert"
)

func TestNewRestyMetricsClient(t *testing.T) {
	hashKey := "secretKey"
	uploadURL := "https://example.com/upload"
	localIP := "192.168.1.1"
	pubKey := &rsa.PublicKey{}

	client := restymetric.NewRestyMetricsClient(true, hashKey, uploadURL, localIP, pubKey)

	assert.NotNil(t, client, "Client should not be nil")
}

// Successfully sends a list of metrics to the specified upload URL
func TestSendMetric_Success(t *testing.T) {
	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Create a MetricsClient instance
	client := restymetric.NewRestyMetricsClient(
		false, "", server.URL, "127.0.0.1", nil,
	)

	// Create a context
	ctx := context.Background()

	// Create a sample metrics list
	metricsList := []metrics.Metrics{
		{ID: "metric1", MType: "gauge", Value: new(float64)},
		{ID: "metric2", MType: "counter", Delta: new(int64)},
	}

	// Call the SendMetric method
	err := client.SendMetric(ctx, metricsList)

	// Assert no error occurred
	assert.NoError(t, err)
}

// Handles JSON marshalling errors gracefully
func TestSendMetricFail(t *testing.T) {
	// Create a MetricsClient instance
	client := restymetric.NewRestyMetricsClient(
		false, "", "fakeurl", "127.0.0.1", nil,
	)

	// Create a sample metrics list
	metricsList := []metrics.Metrics{
		{ID: "metric1", MType: "gauge", Value: new(float64)},
		{ID: "metric2", MType: "counter", Delta: new(int64)},
	}

	err := client.SendMetric(context.Background(), metricsList)

	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}
