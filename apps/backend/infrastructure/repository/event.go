package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/google/uuid"

	"github.com/Haya372/ai-trial/backend/domain/event"
	query "github.com/Haya372/ai-trial/backend/infrastructure/db/generated"
)

type eventQueryRepository struct {
	baseRepository
}

func NewEventQueryRepository(pool *pgxpool.Pool) event.QueryRepository {
	return &eventQueryRepository{baseRepository{pool: pool}}
}

func (r *eventQueryRepository) List(ctx context.Context, filter event.ListFilter) ([]event.Event, error) {
	rows, err := r.querier(ctx).ListEventsByUserAndDateRange(ctx, query.ListEventsByUserAndDateRangeParams{
		UserID:  pgtype.UUID{Bytes: filter.UserID, Valid: true},
		EndAt:   pgtype.Timestamptz{Time: filter.StartDate, Valid: true},
		StartAt: pgtype.Timestamptz{Time: filter.EndDate, Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("list events: %w", err)
	}

	events := make([]event.Event, 0, len(rows))
	for _, row := range rows {
		var desc string
		if row.Description.Valid {
			desc = row.Description.String
		}
		e, err := event.New(
			uuid.UUID(row.ID.Bytes),
			uuid.UUID(row.UserID.Bytes),
			row.Title,
			desc,
			row.StartAt.Time,
			row.EndAt.Time,
		)
		if err != nil {
			return nil, fmt.Errorf("reconstruct event: %w", err)
		}
		events = append(events, e)
	}
	return events, nil
}
