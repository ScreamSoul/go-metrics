package client

import (
	"context"

	"github.com/screamsoul/go-metrics-tpl/internal/models/metrics"
)

type MetricsClient interface {
	SendMetric(ctx context.Context, metricsList []metrics.Metrics) error
}
