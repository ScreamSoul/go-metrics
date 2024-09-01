package server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"

	pb "github.com/screamsoul/go-metrics-tpl/internal/proto"

	"github.com/screamsoul/go-metrics-tpl/internal/grpcapi/services"
	"github.com/screamsoul/go-metrics-tpl/internal/repositories"
	"github.com/screamsoul/go-metrics-tpl/internal/repositories/file"
	"github.com/screamsoul/go-metrics-tpl/internal/repositories/memory"
	"github.com/screamsoul/go-metrics-tpl/internal/repositories/postgres"
	"github.com/screamsoul/go-metrics-tpl/internal/restapi/handlers"
	"github.com/screamsoul/go-metrics-tpl/internal/restapi/middlewares"
	"github.com/screamsoul/go-metrics-tpl/internal/restapi/routers"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

var ErrRegularShutdown = errors.New("regular shutdown")
var ErrUnexpectedShutdown = errors.New("unexpected shutdown")

func StartHttpServer(
	ctx context.Context,
	errorResult chan error,
	cfg *Config,
	logger *zap.Logger,
	metricRepo repositories.MetricStorage,
) {
	var metricServer = handlers.NewMetricServer(
		metricRepo,
	)

	var router = routers.NewMetricRouter(
		metricServer,
		middlewares.LoggingMiddleware,
		middlewares.NewTrustedIPMiddleware(cfg.TrustedSubnetCIDR),
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

	// Graceful shutdown server
	idleConnsClosed := make(chan any)
	sigint := make(chan os.Signal, 1)

	signal.Notify(sigint, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	go func() {
		select {
		case <-sigint:
			logger.Info("receive signal")
			errorResult <- ErrRegularShutdown
		case <-ctx.Done():
			logger.Info("context is done")
		}
		if err := server.Shutdown(ctx); err != nil {
			// Error close Listener
			logger.Error("HTTP server Shutdown", zap.Error(err))
		}
		close(idleConnsClosed)
	}()

	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		logger.Error("HTTP server ListenAndServe", zap.Error(err))
		errorResult <- ErrUnexpectedShutdown

	}

	<-idleConnsClosed
}

func StartGRPCServer(
	ctx context.Context,
	errorResult chan error,
	cfg *Config,
	logger *zap.Logger,
	metricRepo repositories.MetricStorage,
) {
	listen, err := net.Listen("tcp", cfg.ListenGRPCAddress)
	if err != nil {
		logger.Error("listen tcp err", zap.Error(err))
		errorResult <- ErrUnexpectedShutdown
		return
	}

	server := grpc.NewServer()

	idleConnsClosed := make(chan any)
	sigint := make(chan os.Signal, 1)

	signal.Notify(sigint, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	go func() {
		select {
		case <-sigint:
			logger.Info("receive signal")
			errorResult <- ErrRegularShutdown
		case <-ctx.Done():
			logger.Info("context is done")
		}

		server.GracefulStop()
		close(idleConnsClosed)
	}()

	// register server
	pb.RegisterMetricsServiceServer(server, services.NewMetricServer(metricRepo))

	fmt.Println("Сервер gRPC начал работу")
	// start server
	if err := server.Serve(listen); err != nil {
		logger.Error("grpc server err", zap.Error(err))
		errorResult <- ErrUnexpectedShutdown
		return
	}

	<-idleConnsClosed
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

	errorResult := make(chan error)

	go StartHttpServer(ctx, errorResult, cfg, logger, mStorageRestore)
	go StartGRPCServer(ctx, errorResult, cfg, logger, mStorageRestore)

	if err := <-errorResult; err != nil {
		logger.Info(err.Error())
	}
}
