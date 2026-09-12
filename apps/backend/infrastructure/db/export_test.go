package db

import (
	"context"

	"github.com/jackc/pgx/v5"
)

func NewPgxTxManagerForTest(pool interface {
	Begin(ctx context.Context) (pgx.Tx, error)
}) *PgxTxManager {
	return &PgxTxManager{pool: pool}
}
