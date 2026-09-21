package repository

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/google/uuid"

	domainevent "github.com/Haya372/ai-trial/backend/domain/event"
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

type eventRepository struct {
	baseRepository

	logger *slog.Logger
}

func NewEventRepository(pool *pgxpool.Pool, logger *slog.Logger) domainevent.Repository {
	return &eventRepository{baseRepository{pool: pool}, logger}
}

func (r *eventRepository) FindByID(ctx context.Context, id uuid.UUID) (domainevent.Event, error) {
	row, err := r.querier(ctx).FindEventByID(ctx, pgtype.UUID{Bytes: id, Valid: true})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domainevent.ErrEventNotFound
	}
	if err != nil {
		r.logger.Error("find event by id query failed", "error", err)
		return nil, fmt.Errorf("find event by id: %w", err)
	}

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

	return domainevent.New(
		uuid.UUID(row.ID.Bytes), uuid.UUID(row.UserID.Bytes),
		row.Title, desc, row.StartAt.Time, row.EndAt.Time, location, url,
	)
}

func (r *eventRepository) Update(ctx context.Context, e domainevent.Event) error {
	_, err := r.querier(ctx).UpdateEvent(ctx, query.UpdateEventParams{
		ID:          pgtype.UUID{Bytes: e.ID(), Valid: true},
		Title:       e.Title(),
		Description: textOrNull(e.Description()),
		StartAt:     pgtype.Timestamptz{Time: e.StartAt(), Valid: true},
		EndAt:       pgtype.Timestamptz{Time: e.EndAt(), Valid: true},
		Location:    textOrNull(e.Location()),
		Url:         textOrNull(e.URL()),
	})
	if err != nil {
		r.logger.Error("update event query failed", "error", err)
		return fmt.Errorf("update event: %w", err)
	}
	return nil
}

func textOrNull(s string) pgtype.Text {
	if s == "" {
		return pgtype.Text{}
	}
	return pgtype.Text{String: s, Valid: true}
}
