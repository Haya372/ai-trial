package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel/trace"

	"github.com/google/uuid"

	"github.com/Haya372/ai-trial/backend/domain/session"
	query "github.com/Haya372/ai-trial/backend/infrastructure/db/generated"
)

const sessionsTable = "sessions"

type sessionRepository struct {
	baseRepository
}

func NewSessionRepository(pool *pgxpool.Pool, tp trace.TracerProvider) session.Repository {
	return &sessionRepository{baseRepository{pool: pool, tracer: tp.Tracer(tracerName)}}
}

func (r *sessionRepository) Create(
	ctx context.Context, userID uuid.UUID, expiresAt time.Time,
) (session.Session, error) {
	spanCtx, span := r.startDBSpan(ctx, "INSERT", sessionsTable)
	row, err := r.querier(spanCtx).InsertSession(spanCtx, query.InsertSessionParams{
		UserID:    pgtype.UUID{Bytes: userID, Valid: true},
		ExpiresAt: pgtype.Timestamptz{Time: expiresAt, Valid: true},
	})
	endDBSpan(span, err)
	if err != nil {
		return nil, err
	}
	return session.New(uuid.UUID(row.ID.Bytes), uuid.UUID(row.UserID.Bytes), row.ExpiresAt.Time), nil
}

func (r *sessionRepository) FindByID(ctx context.Context, id uuid.UUID) (session.Session, error) {
	spanCtx, span := r.startDBSpan(ctx, "SELECT", sessionsTable)
	row, err := r.querier(spanCtx).FindSessionByID(spanCtx, pgtype.UUID{Bytes: id, Valid: true})
	endDBSpanNotFound(span, err)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return session.New(uuid.UUID(row.ID.Bytes), uuid.UUID(row.UserID.Bytes), row.ExpiresAt.Time), nil
}

func (r *sessionRepository) FindActiveByID(ctx context.Context, id uuid.UUID) (session.Session, error) {
	spanCtx, span := r.startDBSpan(ctx, "SELECT", sessionsTable)
	row, err := r.querier(spanCtx).FindActiveSessionByID(spanCtx, pgtype.UUID{Bytes: id, Valid: true})
	endDBSpanNotFound(span, err)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return session.New(uuid.UUID(row.ID.Bytes), uuid.UUID(row.UserID.Bytes), row.ExpiresAt.Time), nil
}

func (r *sessionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	spanCtx, span := r.startDBSpan(ctx, "DELETE", sessionsTable)
	err := r.querier(spanCtx).DeleteSession(spanCtx, pgtype.UUID{Bytes: id, Valid: true})
	endDBSpan(span, err)
	return err
}
