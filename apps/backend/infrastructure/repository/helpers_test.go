//go:build integration

package repository_test

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"

	"github.com/Haya372/ai-trial/backend/infrastructure/db"
)

var testPool *pgxpool.Pool

func TestMain(m *testing.M) {
	ctx := context.Background()

	c, err := tcpostgres.Run(ctx, "postgres:17-alpine",
		tcpostgres.WithDatabase("testdb"),
		tcpostgres.WithUsername("test"),
		tcpostgres.WithPassword("test"),
		tcpostgres.BasicWaitStrategies(),
	)
	if err != nil {
		panic("start postgres container: " + err.Error())
	}
	defer func() { _ = c.Terminate(ctx) }()

	dsn, err := c.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		panic("get connection string: " + err.Error())
	}

	testPool, err = db.NewPool(ctx, dsn)
	if err != nil {
		panic("connect to postgres: " + err.Error())
	}
	defer testPool.Close()

	// golang-migrate の pgx5 ドライバは pgx5:// スキームを要求する
	migrateDSN := strings.Replace(dsn, "postgres://", "pgx5://", 1)
	mg, err := migrate.New("file://../../db/migrations", migrateDSN)
	if err != nil {
		panic("create migrate: " + err.Error())
	}
	if err := mg.Up(); err != nil && err != migrate.ErrNoChange {
		panic("migrate up: " + err.Error())
	}

	os.Exit(m.Run())
}

func setupTest(t *testing.T) {
	t.Helper()
	t.Cleanup(func() {
		_, err := testPool.Exec(context.Background(), "TRUNCATE TABLE sessions, users RESTART IDENTITY CASCADE")
		if err != nil {
			t.Errorf("truncate tables: %v", err)
		}
	})
}
