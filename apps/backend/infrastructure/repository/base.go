package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Haya372/ai-trial/backend/infrastructure/db"
	query "github.com/Haya372/ai-trial/backend/infrastructure/db/generated"
)

type baseRepository struct {
	pool *pgxpool.Pool
}

func (r *baseRepository) querier(ctx context.Context) *query.Queries {
	if tx, ok := db.GetTx(ctx); ok {
		return query.New(tx)
	}
	return query.New(r.pool)
}
