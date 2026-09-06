//go:build integration

package db_test

import (
	"context"
	"testing"

	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"

	"github.com/Haya372/ai-trial/backend/infrastructure/db"
)

func setupPostgres(t *testing.T) string {
	t.Helper()
	ctx := context.Background()
	c, err := tcpostgres.Run(ctx, "postgres:17-alpine",
		tcpostgres.WithDatabase("testdb"),
		tcpostgres.WithUsername("test"),
		tcpostgres.WithPassword("test"),
		tcpostgres.BasicWaitStrategies(),
	)
	if err != nil {
		t.Fatalf("start postgres container: %v", err)
	}
	t.Cleanup(func() { _ = c.Terminate(ctx) })

	dsn, err := c.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("get connection string: %v", err)
	}
	return dsn
}

func TestNewPool_emptyDSN_returnsError(t *testing.T) {
	_, err := db.NewPool(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for empty DSN, got nil")
	}
}

func TestNewPool_invalidDSN_returnsError(t *testing.T) {
	_, err := db.NewPool(context.Background(), "not-a-valid-dsn")
	if err == nil {
		t.Fatal("expected error for invalid DSN, got nil")
	}
}

func TestNewPool_validDSN_pingSucceeds(t *testing.T) {
	dsn := setupPostgres(t)

	pool, err := db.NewPool(context.Background(), dsn)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	defer pool.Close()
}
