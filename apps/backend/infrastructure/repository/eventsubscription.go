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

	"github.com/Haya372/ai-trial/backend/domain/eventsubscription"
	query "github.com/Haya372/ai-trial/backend/infrastructure/db/generated"
)

const eventSubscriptionsTable = "event_subscriptions"

type eventSubscriptionRepository struct {
	baseRepository
}

// NewEventSubscriptionRepository returns an eventsubscription.Repository backed by PostgreSQL.
func NewEventSubscriptionRepository(pool *pgxpool.Pool, tp trace.TracerProvider) eventsubscription.Repository {
	return &eventSubscriptionRepository{baseRepository{pool: pool, tracer: tp.Tracer(tracerName)}}
}

func (r *eventSubscriptionRepository) Create(
	ctx context.Context, s eventsubscription.EventSubscription,
) (eventsubscription.EventSubscription, error) {
	spanCtx, span := r.startDBSpan(ctx, "INSERT", eventSubscriptionsTable)
	row, err := r.querier(spanCtx).InsertEventSubscription(spanCtx, query.InsertEventSubscriptionParams{
		ID:      pgtype.UUID{Bytes: s.ID(), Valid: true},
		EventID: pgtype.UUID{Bytes: s.EventID(), Valid: true},
		UserID:  pgtype.UUID{Bytes: s.UserID(), Valid: true},
	})
	endDBSpan(span, err)
	if err != nil {
		return nil, fmt.Errorf("insert event subscription: %w", err)
	}

	return eventsubscription.New(
		uuid.UUID(row.ID.Bytes), uuid.UUID(row.EventID.Bytes), uuid.UUID(row.UserID.Bytes), row.CreatedAt.Time,
	), nil
}

func (r *eventSubscriptionRepository) FindByID(
	ctx context.Context, id uuid.UUID,
) (eventsubscription.EventSubscription, error) {
	spanCtx, span := r.startDBSpan(ctx, "SELECT", eventSubscriptionsTable)
	row, err := r.querier(spanCtx).FindEventSubscriptionByID(spanCtx, pgtype.UUID{Bytes: id, Valid: true})
	endDBSpanNotFound(span, err)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, eventsubscription.ErrEventSubscriptionNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find event subscription by id: %w", err)
	}

	return eventsubscription.New(
		uuid.UUID(row.ID.Bytes), uuid.UUID(row.EventID.Bytes), uuid.UUID(row.UserID.Bytes), row.CreatedAt.Time,
	), nil
}

func (r *eventSubscriptionRepository) FindByEventAndUserID(
	ctx context.Context, eventID, userID uuid.UUID,
) (eventsubscription.EventSubscription, error) {
	spanCtx, span := r.startDBSpan(ctx, "SELECT", eventSubscriptionsTable)
	params := query.FindEventSubscriptionByEventAndUserIDParams{
		EventID: pgtype.UUID{Bytes: eventID, Valid: true},
		UserID:  pgtype.UUID{Bytes: userID, Valid: true},
	}
	row, err := r.querier(spanCtx).FindEventSubscriptionByEventAndUserID(spanCtx, params)
	endDBSpanNotFound(span, err)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, eventsubscription.ErrEventSubscriptionNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find event subscription by event and user id: %w", err)
	}

	return eventsubscription.New(
		uuid.UUID(row.ID.Bytes), uuid.UUID(row.EventID.Bytes), uuid.UUID(row.UserID.Bytes), row.CreatedAt.Time,
	), nil
}

func (r *eventSubscriptionRepository) ListByUserID(
	ctx context.Context, userID uuid.UUID,
) ([]eventsubscription.EventSubscription, error) {
	spanCtx, span := r.startDBSpan(ctx, "SELECT", eventSubscriptionsTable)
	rows, err := r.querier(spanCtx).ListEventSubscriptionsByUserID(spanCtx, pgtype.UUID{Bytes: userID, Valid: true})
	endDBSpan(span, err)
	if err != nil {
		return nil, fmt.Errorf("list event subscriptions by user id: %w", err)
	}

	result := make([]eventsubscription.EventSubscription, 0, len(rows))
	for _, row := range rows {
		result = append(result, eventsubscription.New(
			uuid.UUID(row.ID.Bytes), uuid.UUID(row.EventID.Bytes), uuid.UUID(row.UserID.Bytes), row.CreatedAt.Time,
		))
	}
	return result, nil
}

func (r *eventSubscriptionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	spanCtx, span := r.startDBSpan(ctx, "DELETE", eventSubscriptionsTable)
	err := r.querier(spanCtx).DeleteEventSubscription(spanCtx, pgtype.UUID{Bytes: id, Valid: true})
	endDBSpan(span, err)
	if err != nil {
		return fmt.Errorf("delete event subscription: %w", err)
	}
	return nil
}
