package repository

import (
	"context"
	"time"

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
	ctx context.Context, userID uuid.UUID, tokenHash session.TokenHash, expiresAt time.Time,
) (session.Session, error) {
	spanCtx, span := r.startDBSpan(ctx, "INSERT", sessionsTable)
	row, err := r.querier(spanCtx).InsertSession(spanCtx, query.InsertSessionParams{
		UserID:    toPgUUID(userID),
		TokenHash: tokenHash.String(),
		ExpiresAt: pgtype.Timestamptz{Time: expiresAt, Valid: true},
	})
	endDBSpan(span, err)
	if err != nil {
		return nil, err
	}
	return session.New(uuid.UUID(row.ID.Bytes), uuid.UUID(row.UserID.Bytes), row.ExpiresAt.Time), nil
}

func (r *sessionRepository) FindActiveByTokenHash(
	ctx context.Context, tokenHash session.TokenHash,
) (session.Session, error) {
	spanCtx, span := r.startDBSpan(ctx, "SELECT", sessionsTable)
	row, err := r.querier(spanCtx).FindActiveSessionByTokenHash(spanCtx, tokenHash.String())
	endDBSpanNotFound(span, err)
	if notFound, _, err := checkNotFound(err, nil); notFound {
		return nil, err
	}
	return session.New(uuid.UUID(row.ID.Bytes), uuid.UUID(row.UserID.Bytes), row.ExpiresAt.Time), nil
}

func (r *sessionRepository) DeleteByTokenHash(ctx context.Context, tokenHash session.TokenHash) error {
	spanCtx, span := r.startDBSpan(ctx, "DELETE", sessionsTable)
	err := r.querier(spanCtx).DeleteSessionByTokenHash(spanCtx, tokenHash.String())
	endDBSpan(span, err)
	return err
}
