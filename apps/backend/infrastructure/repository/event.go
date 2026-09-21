package repository

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/google/uuid"

	domainevent "github.com/Haya372/ai-trial/backend/domain/event"
	query "github.com/Haya372/ai-trial/backend/infrastructure/db/generated"
	eventuc "github.com/Haya372/ai-trial/backend/usecase/event"
)

// --- read side ---

type eventQueryRepository struct {
	baseRepository

	logger *slog.Logger
}

func NewEventQueryRepository(pool *pgxpool.Pool, logger *slog.Logger) eventuc.QueryService {
	return &eventQueryRepository{baseRepository{pool: pool}, logger}
}

func (r *eventQueryRepository) List(ctx context.Context, filter eventuc.ListFilter) ([]eventuc.EventReadModel, error) {
	rows, err := r.querier(ctx).ListEventsByUserAndDateRange(ctx, query.ListEventsByUserAndDateRangeParams{
		UserID:    pgtype.UUID{Bytes: filter.UserID, Valid: true},
		StartDate: pgtype.Timestamptz{Time: filter.StartDate, Valid: true},
		EndDate:   pgtype.Timestamptz{Time: filter.EndDate, Valid: true},
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

// --- write side ---

type eventRepository struct {
	baseRepository
}

// NewEventRepository returns an event.Repository backed by PostgreSQL.
func NewEventRepository(pool *pgxpool.Pool) domainevent.Repository {
	return &eventRepository{baseRepository{pool: pool}}
}

func (r *eventRepository) Create(ctx context.Context, e domainevent.Event) (domainevent.Event, error) {
	row, err := r.querier(ctx).InsertEvent(ctx, query.InsertEventParams{
		ID:          pgtype.UUID{Bytes: e.ID(), Valid: true},
		UserID:      pgtype.UUID{Bytes: e.UserID(), Valid: true},
		Title:       e.Title(),
		Description: nullableText(e.Description()),
		StartAt:     pgtype.Timestamptz{Time: e.StartAt(), Valid: true},
		EndAt:       pgtype.Timestamptz{Time: e.EndAt(), Valid: true},
		Location:    nullableText(e.Location()),
		Url:         nullableText(e.URL()),
	})
	if err != nil {
		return nil, err
	}

	// Reconstruct from the persisted row so the caller gets the exact stored values.
	var desc, location, url string
	if row.Description.Valid {
		desc = row.Description.String
	}
	if row.Location.Valid {
		location = row.Location.String
	}
	if row.Url.Valid {
		url = row.Url.String
	}

	saved, err := domainevent.New(
		uuid.UUID(row.ID.Bytes),
		uuid.UUID(row.UserID.Bytes),
		row.Title,
		desc,
		row.StartAt.Time,
		row.EndAt.Time,
		location,
		url,
	)
	if err != nil {
		return nil, fmt.Errorf("reconstruct saved event: %w", err)
	}
	return saved, nil
}

// nullableText converts an empty Go string to a NULL pgtype.Text;
// non-empty values are stored as-is.
func nullableText(s string) pgtype.Text {
	if s == "" {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: s, Valid: true}
}
