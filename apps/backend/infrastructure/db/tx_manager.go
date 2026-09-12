package db

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Haya372/ai-trial/backend/usecase"
)

var _ usecase.TransactionManager = (*PgxTxManager)(nil)

type PgxTxManager struct {
	pool *pgxpool.Pool
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
		// TODO: ロールバック失敗のユニットテストを追加する（pool.Begin の抽象化が必要）
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
