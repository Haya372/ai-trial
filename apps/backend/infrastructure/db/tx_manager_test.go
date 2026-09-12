//go:build integration

package db_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Haya372/ai-trial/backend/infrastructure/db"
)

func TestRunInTx_success_commits(t *testing.T) {
	dsn := setupPostgres(t)
	pool, err := db.NewPool(context.Background(), dsn)
	if err != nil {
		t.Fatalf("create pool: %v", err)
	}
	defer pool.Close()

	mgr := db.NewPgxTxManager(pool)

	called := false
	err = mgr.RunInTx(context.Background(), func(ctx context.Context) error {
		called = true
		_, ok := db.GetTx(ctx)
		if !ok {
			t.Error("expected tx in context")
		}
		return nil
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !called {
		t.Fatal("fn was not called")
	}
}

func TestRunInTx_error_rollsback(t *testing.T) {
	dsn := setupPostgres(t)
	pool, err := db.NewPool(context.Background(), dsn)
	if err != nil {
		t.Fatalf("create pool: %v", err)
	}
	defer pool.Close()

	mgr := db.NewPgxTxManager(pool)

	sentinel := errors.New("sentinel error")
	err = mgr.RunInTx(context.Background(), func(ctx context.Context) error {
		return sentinel
	})

	if !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel error, got %v", err)
	}
}
