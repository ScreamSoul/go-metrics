package postgres

import (
	"database/sql"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/screamsoul/go-metrics-tpl/internal/models/metrics"
	"github.com/screamsoul/go-metrics-tpl/pkg/logging"
	"go.uber.org/zap"
)

type PostgresStorage struct {
	db      *sql.DB
	logging *zap.Logger
}

func NewPostgresStorage(dataSourceName string) *PostgresStorage {
	db, err := sql.Open("pgx", dataSourceName)
	if err != nil {
		panic(err)
	}
	return &PostgresStorage{db, logging.GetLogger()}
}

func (storage *PostgresStorage) Add(m metrics.Metrics) {
	panic("not implemented") // TODO: Implement
}

func (storage *PostgresStorage) Get(m *metrics.Metrics) error {
	panic("not implemented") // TODO: Implement
}

func (storage *PostgresStorage) List() []metrics.Metrics {
	panic("not implemented") // TODO: Implement
}

func (storage *PostgresStorage) Ping() bool {
	err := storage.db.Ping()
	if err != nil {
		storage.logging.Error("db connect error", zap.Error(err))
	}
	return err == nil
}

func (storage *PostgresStorage) Close() {
	err := storage.db.Close()
	if err != nil {
		storage.logging.Error("db close connection error", zap.Error(err))
	}
}
