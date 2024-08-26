package server

import (
	"context"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"

	"github.com/screamsoul/go-metrics-tpl/internal/handlers"
	"github.com/screamsoul/go-metrics-tpl/internal/middlewares"
	"github.com/screamsoul/go-metrics-tpl/internal/repositories"
	"github.com/screamsoul/go-metrics-tpl/internal/repositories/file"
	"github.com/screamsoul/go-metrics-tpl/internal/repositories/memory"
	"github.com/screamsoul/go-metrics-tpl/internal/repositories/postgres"
	"github.com/screamsoul/go-metrics-tpl/internal/routers"
	"go.uber.org/zap"
)

func GracefulShutdownListenAndServe(
	ctx context.Context,
	cancel context.CancelFunc,
	server *http.Server,
	logger *zap.Logger,
) {

	idleConnsClosed := make(chan struct{})
	sigint := make(chan os.Signal, 1)

	signal.Notify(sigint, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	go func() {

		<-sigint
		if err := server.Shutdown(ctx); err != nil {
			// Error close Listener
			logger.Error("HTTP server Shutdown", zap.Error(err))
		}
		cancel()
		close(idleConnsClosed)
	}()

	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		logger.Error("HTTP server ListenAndServe", zap.Error(err))
	}

	<-idleConnsClosed

	logger.Info("Server Shutdown gracefully")
}

// Start starts the server
func Start(cfg *Config, logger *zap.Logger) {
	ctx, cansel := context.WithCancel(context.Background())
	defer cansel()
	// Create MetricStorage.
	var mStorage repositories.MetricStorage

	if cfg.DatabaseDSN == "" {
		// if no connection to the database is specified, the in-memory storage will be used.

		memS := memory.NewMemStorage()
		mStorage = memS
	} else {
		postgresS := postgres.NewPostgresStorage(cfg.DatabaseDSN, cfg.BackoffIntervals)
		defer postgresS.Close()

		if err := postgresS.Bootstrap(ctx); err != nil {
			panic(err)
		}

		mStorage = postgresS
	}

	// Create restore wrapper.
	mStorageRestore := file.NewFileRestoreMetricWrapper(
		ctx,
		mStorage,
		cfg.FileStoragePath,
		cfg.StoreInterval,
		cfg.Restore,
	)

	if mStorageRestore.IsActiveRestore {
		defer mStorageRestore.Save(ctx)
	}

	var metricServer = handlers.NewMetricServer(
		mStorageRestore,
	)

	var router = routers.NewMetricRouter(
		metricServer,
		middlewares.LoggingMiddleware,
		middlewares.NewTrustedIpMiddleware(cfg.TrustedSubnetCIDR),
		middlewares.NewDecryptMiddleware(cfg.CryptoKey.Key),
		middlewares.NewHashSumHeaderMiddleware(cfg.HashBodyKey),
		middlewares.GzipDecompressMiddleware,
		middlewares.GzipCompressMiddleware,
	)

	if cfg.Debug {
		router.Mount("/debug", http.DefaultServeMux)
		logger.Info("mount debug pprof")
	}

	logger.Info("starting server", zap.String("ListenAddress", cfg.ListenAddress))

	server := http.Server{Addr: cfg.ListenAddress, Handler: router}

	GracefulShutdownListenAndServe(
		ctx,
		cansel,
		&server,
		logger,
	)
}
