package client

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/screamsoul/go-metrics-tpl/internal/repositories"
	"github.com/screamsoul/go-metrics-tpl/internal/repositories/memory"
	"github.com/screamsoul/go-metrics-tpl/pkg/backoff"
	"github.com/screamsoul/go-metrics-tpl/pkg/logging"
	"go.uber.org/zap"
)

func sender(
	ctx context.Context,
	metricRepo repositories.CollectionMetric,
	backoffIntervals []time.Duration,
	metricClient *MetricsClient,
	reportInterval time.Duration,
) {
	logger := logging.GetLogger()
	for {
		select {
		case <-ctx.Done():
			return
		default:
			metricsList, err := metricRepo.List(ctx)
			if err != nil {
				panic(err)
			}

			sendMetric := func() error {
				return metricClient.SendMetric(ctx, metricsList)
			}

			if err := backoff.RetryWithBackoff(backoffIntervals, IsTemporaryNetworkError, sendMetric); err != nil {
				logger.Error("send metric error", zap.Error(err))
			}
		}

		time.Sleep(reportInterval)
	}
}

func updater(
	ctx context.Context,
	metricRepo repositories.CollectionMetric,
	pollInterval time.Duration,
) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			metricRepo.Update()
			metricRepo.UpdateRuntime()
			metricRepo.UpdateGopsutil()
			time.Sleep(pollInterval)
		}
	}
}

func Start(cfg *Config, logger *zap.Logger) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger.Info("start agent")
	logger.Info("use metric server", zap.String("server", cfg.GetServerURL()))

	metricRepo := memory.NewCollectionMetricStorage()

	pollInterval := time.Duration(cfg.PollInterval) * time.Second
	reportInterval := time.Duration(cfg.ReportInterval) * time.Second

	metricClient := NewMetricsClient(cfg.CompressRequest, cfg.HashBodyKey, cfg.GetUpdateMetricURL(), cfg.CryptoKey.Key)

	go updater(ctx, metricRepo, pollInterval)
	logger.Info("start senders", zap.Int("count_senders", cfg.RateLimit))
	for i := 0; i < cfg.RateLimit; i++ {
		go sender(ctx, metricRepo, cfg.BackoffIntervals, metricClient, reportInterval)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	go func() {
		<-sigChan
		cancel()
	}()

	<-ctx.Done()
	fmt.Println("Agent gracefully closed:", ctx.Err())
}
