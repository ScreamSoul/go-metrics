package server_test

import (
	"syscall"
	"testing"
	"time"

	"github.com/screamsoul/go-metrics-tpl/internal/server"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestStart_InitializesInMemoryStorage(t *testing.T) {
	cfg := &server.Config{
		Postgres: server.Postgres{
			DatabaseDSN: "",
		},
		FileStoragePath: "",
		StoreInterval:   0,
		Restore:         false,
		HashBodyKey:     "",
		ListenAddress:   ":8080",
		Debug:           false,
	}
	logger := zap.NewNop()

	time.AfterFunc(3*time.Second, func() {
		require.NoError(t, syscall.Kill(syscall.Getpid(), syscall.SIGINT))
	})

	server.Start(cfg, logger)

}
