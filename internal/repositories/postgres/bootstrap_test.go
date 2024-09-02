package postgres

import (
	"context"
	"testing"

	embeddedpostgres "github.com/fergusstrange/embedded-postgres"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

func connect() (*sqlx.DB, error) {
	db, err := sqlx.Connect("postgres", "host=localhost port=5432 user=postgres password=postgres dbname=postgres sslmode=disable")
	return db, err
}

func TestPostgresStorage_Bootstrap(t *testing.T) {
	// Create a temporary EmbeddedPostgres instance
	database := embeddedpostgres.NewDatabase()
	if err := database.Start(); err != nil {
		t.Fatal(err)
	}

	defer func() {
		if err := database.Stop(); err != nil {
			t.Fatal(err)
		}
	}()
	// Create a logger for testing
	logger := zaptest.NewLogger(t)

	db, err := connect()
	if err != nil {
		t.Fatal(err)
	}
	// Create a new PostgresStorage instance
	storage := &PostgresStorage{
		db:      db,
		logging: logger,
	}

	// Call the Bootstrap method
	err = storage.Bootstrap(context.Background())
	assert.NoError(t, err)
}
