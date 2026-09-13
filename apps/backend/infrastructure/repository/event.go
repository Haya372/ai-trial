package repository

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/google/uuid"

	query "github.com/Haya372/ai-trial/backend/infrastructure/db/generated"
	eventuc "github.com/Haya372/ai-trial/backend/usecase/event"
)

type eventQueryRepository struct {
	baseRepository

	logger *slog.Logger
}

func NewEventQueryRepository(pool *pgxpool.Pool, logger *slog.Logger) eventuc.QueryService {
	return &eventQueryRepository{baseRepository{pool: pool}, logger}
}

func (r *eventQueryRepository) List(ctx context.Context, filter eventuc.ListFilter) ([]eventuc.EventReadModel, error) {
	rows, err := r.querier(ctx).ListEventsByUserAndDateRange(ctx, query.ListEventsByUserAndDateRangeParams{
		UserID:  pgtype.UUID{Bytes: filter.UserID, Valid: true},
		EndAt:   pgtype.Timestamptz{Time: filter.StartDate, Valid: true},
		StartAt: pgtype.Timestamptz{Time: filter.EndDate, Valid: true},
	})
	if err != nil {
		r.logger.Error("list events query failed", "error", err)
		return nil, fmt.Errorf("list events: %w", err)
	}

	result := make([]eventuc.EventReadModel, 0, len(rows))
	for _, row := range rows {
		var desc string
		if row.Description.Valid {
			desc = row.Description.String
		}
		result = append(result, eventuc.EventReadModel{
			ID:          uuid.UUID(row.ID.Bytes),
			Title:       row.Title,
			Description: desc,
			StartAt:     row.StartAt.Time,
			EndAt:       row.EndAt.Time,
		})
	}
	return result, nil
}
