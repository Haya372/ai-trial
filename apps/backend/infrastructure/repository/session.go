package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/google/uuid"

	"github.com/Haya372/ai-trial/backend/domain/session"
	query "github.com/Haya372/ai-trial/backend/infrastructure/db/generated"
)

type sessionRepository struct {
	baseRepository
}

func NewSessionRepository(pool *pgxpool.Pool) session.Repository {
	return &sessionRepository{baseRepository{pool: pool}}
}

func (r *sessionRepository) Create(
	ctx context.Context, userID uuid.UUID, expiresAt time.Time,
) (session.Session, error) {
	row, err := r.querier(ctx).InsertSession(ctx, query.InsertSessionParams{
		UserID:    pgtype.UUID{Bytes: userID, Valid: true},
		ExpiresAt: pgtype.Timestamptz{Time: expiresAt, Valid: true},
	})
	if err != nil {
		return nil, err
	}
	return session.New(uuid.UUID(row.ID.Bytes), uuid.UUID(row.UserID.Bytes), row.ExpiresAt.Time), nil
}

func (r *sessionRepository) FindByID(ctx context.Context, id uuid.UUID) (session.Session, error) {
	row, err := r.querier(ctx).FindSessionByID(ctx, pgtype.UUID{Bytes: id, Valid: true})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return session.New(uuid.UUID(row.ID.Bytes), uuid.UUID(row.UserID.Bytes), row.ExpiresAt.Time), nil
}

func (r *sessionRepository) FindActiveByID(ctx context.Context, id uuid.UUID) (session.Session, error) {
	row, err := r.querier(ctx).FindActiveSessionByID(ctx, pgtype.UUID{Bytes: id, Valid: true})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return session.New(uuid.UUID(row.ID.Bytes), uuid.UUID(row.UserID.Bytes), row.ExpiresAt.Time), nil
}

func (r *sessionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.querier(ctx).DeleteSession(ctx, pgtype.UUID{Bytes: id, Valid: true})
}
