package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel/trace"

	"github.com/google/uuid"

	"github.com/Haya372/ai-trial/backend/domain/eventshare"
	query "github.com/Haya372/ai-trial/backend/infrastructure/db/generated"
)

const eventSharesTable = "event_shares"

type eventShareRepository struct {
	baseRepository
}

// NewEventShareRepository returns an eventshare.Repository backed by PostgreSQL.
func NewEventShareRepository(pool *pgxpool.Pool, tp trace.TracerProvider) eventshare.Repository {
	return &eventShareRepository{baseRepository{pool: pool, tracer: tp.Tracer(tracerName)}}
}

func (r *eventShareRepository) Create(ctx context.Context, s eventshare.EventShare) (eventshare.EventShare, error) {
	spanCtx, span := r.startDBSpan(ctx, "INSERT", eventSharesTable)
	row, err := r.querier(spanCtx).InsertEventShare(spanCtx, query.InsertEventShareParams{
		ID:        pgtype.UUID{Bytes: s.ID(), Valid: true},
		EventID:   pgtype.UUID{Bytes: s.EventID(), Valid: true},
		TokenHash: s.TokenHash(),
		ExpiresAt: pgtype.Timestamptz{Time: s.ExpiresAt(), Valid: true},
	})
	endDBSpan(span, err)
	if err != nil {
		return nil, fmt.Errorf("insert event share: %w", err)
	}

	saved, err := eventshare.New(
		uuid.UUID(row.ID.Bytes), uuid.UUID(row.EventID.Bytes), row.TokenHash, row.ExpiresAt.Time,
	)
	if err != nil {
		return nil, fmt.Errorf("reconstruct saved event share: %w", err)
	}
	return saved, nil
}

// FindByToken hashes the given plain token and looks up the share whose
// stored hash matches (eventshare.Repository's contract).
func (r *eventShareRepository) FindByToken(ctx context.Context, token string) (eventshare.EventShare, error) {
	spanCtx, span := r.startDBSpan(ctx, "SELECT", eventSharesTable)
	row, err := r.querier(spanCtx).FindEventShareByTokenHash(spanCtx, eventshare.HashToken(token))
	endDBSpanNotFound(span, err)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, eventshare.ErrEventShareNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find event share by token: %w", err)
	}

	return eventshare.New(
		uuid.UUID(row.ID.Bytes), uuid.UUID(row.EventID.Bytes), row.TokenHash, row.ExpiresAt.Time,
	)
}
