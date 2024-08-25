package client

import (
	"context"
	"crypto/rsa"
	"encoding/json"

	"github.com/go-resty/resty/v2"
	"github.com/screamsoul/go-metrics-tpl/internal/client/middlewares"
	"github.com/screamsoul/go-metrics-tpl/internal/models/metrics"
	"github.com/screamsoul/go-metrics-tpl/pkg/logging"
	"go.uber.org/zap"
)

type MetricsClient struct {
	resty.Client
	logger    *zap.Logger
	uploadURL string
}

func NewMetricsClient(
	compressRequest bool,
	hashKey string,
	uploadURL string,
	pubKey *rsa.PublicKey,
) *MetricsClient {

	client := &MetricsClient{
		*resty.New(),
		logging.GetLogger(),
		uploadURL,
	}

	if compressRequest {
		client.OnBeforeRequest(middlewares.NewGzipCompressBodyMiddleware())
	}

	if hashKey != "" {
		client.OnBeforeRequest(middlewares.NewHashSumHeaderMiddleware(hashKey))
	}

	if pubKey != nil {
		client.OnBeforeRequest(middlewares.NewEncryptMiddleware(pubKey))
	}

	return client
}

func (client *MetricsClient) SendMetric(ctx context.Context, metricsList []metrics.Metrics) error {
	jsonData, err := json.Marshal(metricsList)
	if err != nil {
		panic(err)
	}

	resp, err := resty.New().R().
		SetContext(ctx).
		SetHeader("Content-Type", "application/json").
		SetBody(jsonData).
		Post(client.uploadURL)

	if err != nil {
		client.logger.Error("send error", zap.Error(err))
		return err
	}

	client.logger.Info(
		"send metric", zap.Any("metric", resp.Request.Body), zap.String("url", client.uploadURL),
	)
	return nil
}
