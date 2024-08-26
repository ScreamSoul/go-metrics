// correctly compresses a valid byte slice body
package client_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/screamsoul/go-metrics-tpl/internal/client"
	"github.com/screamsoul/go-metrics-tpl/internal/models/metrics"
	"github.com/stretchr/testify/assert"
)

// Successfully sends a list of metrics to the specified upload URL
func TestSendMetric_Success(t *testing.T) {
	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Create a MetricsClient instance
	client := client.NewMetricsClient(
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
