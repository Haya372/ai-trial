package db

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Haya372/ai-trial/backend/usecase"
)

//go:generate go tool mockgen -destination=mock/tx_mock.go -package=mock github.com/jackc/pgx/v5 Tx
//go:generate go tool mockgen -destination=mock/tx_beginner_mock.go -package=mock github.com/Haya372/ai-trial/backend/infrastructure/db txBeginner

var _ usecase.TransactionManager = (*PgxTxManager)(nil)

type txBeginner interface {
	Begin(ctx context.Context) (pgx.Tx, error)
}

type PgxTxManager struct {
	pool txBeginner
}

func NewPgxTxManager(pool *pgxpool.Pool) *PgxTxManager {
	return &PgxTxManager{pool: pool}
}

func (m *PgxTxManager) RunInTx(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := m.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	txCtx := setTx(ctx, tx)

	if fnErr := fn(txCtx); fnErr != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return errors.Join(fnErr, fmt.Errorf("rollback: %w", rbErr))
		}
		return fnErr
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}
