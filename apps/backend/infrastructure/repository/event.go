package repository

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel/trace"

	"github.com/google/uuid"

	domainevent "github.com/Haya372/ai-trial/backend/domain/event"
	query "github.com/Haya372/ai-trial/backend/infrastructure/db/generated"
	eventuc "github.com/Haya372/ai-trial/backend/usecase/event"
)

const eventsTable = "events"

// --- read side ---

type eventQueryRepository struct {
	baseRepository

	logger *slog.Logger
}

func NewEventQueryRepository(pool *pgxpool.Pool, logger *slog.Logger, tp trace.TracerProvider) eventuc.QueryService {
	return &eventQueryRepository{baseRepository{pool: pool, tracer: tp.Tracer(tracerName)}, logger}
}

func (r *eventQueryRepository) List(ctx context.Context, filter eventuc.ListFilter) ([]eventuc.EventReadModel, error) {
	spanCtx, span := r.startDBSpan(ctx, "SELECT", eventsTable)
	rows, err := r.querier(spanCtx).ListEventsByUserAndDateRange(spanCtx, query.ListEventsByUserAndDateRangeParams{
		UserID:    pgtype.UUID{Bytes: filter.UserID, Valid: true},
		StartDate: pgtype.Timestamptz{Time: filter.StartDate, Valid: true},
		EndDate:   pgtype.Timestamptz{Time: filter.EndDate, Valid: true},
	})
	endDBSpan(span, err)
	if err != nil {
		r.logger.Error("list events query failed", "error", err)
		return nil, fmt.Errorf("list events: %w", err)
	}

	result := make([]eventuc.EventReadModel, 0, len(rows))
	for _, row := range rows {
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
		result = append(result, eventuc.EventReadModel{
			ID:          uuid.UUID(row.ID.Bytes),
			Title:       row.Title,
			Description: desc,
			StartAt:     row.StartAt.Time,
			EndAt:       row.EndAt.Time,
			Location:    location,
			URL:         url,
		})
	}
	return result, nil
}

// --- write side ---

type eventRepository struct {
	baseRepository

	logger *slog.Logger
}

// NewEventRepository returns an event.Repository backed by PostgreSQL.
func NewEventRepository(pool *pgxpool.Pool, logger *slog.Logger, tp trace.TracerProvider) domainevent.Repository {
	return &eventRepository{baseRepository{pool: pool, tracer: tp.Tracer(tracerName)}, logger}
}

func (r *eventRepository) Create(ctx context.Context, e domainevent.Event) (domainevent.Event, error) {
	spanCtx, span := r.startDBSpan(ctx, "INSERT", eventsTable)
	row, err := r.querier(spanCtx).InsertEvent(spanCtx, query.InsertEventParams{
		ID:          pgtype.UUID{Bytes: e.ID(), Valid: true},
		UserID:      pgtype.UUID{Bytes: e.UserID(), Valid: true},
		Title:       e.Title(),
		Description: textOrNull(e.Description()),
		StartAt:     pgtype.Timestamptz{Time: e.StartAt(), Valid: true},
		EndAt:       pgtype.Timestamptz{Time: e.EndAt(), Valid: true},
		Location:    textOrNull(e.Location()),
		Url:         textOrNull(e.URL()),
	})
	endDBSpan(span, err)
	if err != nil {
		r.logger.Error("insert event query failed", "error", err)
		return nil, fmt.Errorf("insert event: %w", err)
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

func (r *eventRepository) FindByID(ctx context.Context, id uuid.UUID) (domainevent.Event, error) {
	spanCtx, span := r.startDBSpan(ctx, "SELECT", eventsTable)
	row, err := r.querier(spanCtx).FindEventByID(spanCtx, pgtype.UUID{Bytes: id, Valid: true})
	endDBSpan(span, err)
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
	spanCtx, span := r.startDBSpan(ctx, "UPDATE", eventsTable)
	_, err := r.querier(spanCtx).UpdateEvent(spanCtx, query.UpdateEventParams{
		ID:          pgtype.UUID{Bytes: e.ID(), Valid: true},
		Title:       e.Title(),
		Description: textOrNull(e.Description()),
		StartAt:     pgtype.Timestamptz{Time: e.StartAt(), Valid: true},
		EndAt:       pgtype.Timestamptz{Time: e.EndAt(), Valid: true},
		Location:    textOrNull(e.Location()),
		Url:         textOrNull(e.URL()),
	})
	endDBSpan(span, err)
	if err != nil {
		r.logger.Error("update event query failed", "error", err)
		return fmt.Errorf("update event: %w", err)
	}
	return nil
}

func (r *eventRepository) Delete(ctx context.Context, id uuid.UUID) error {
	spanCtx, span := r.startDBSpan(ctx, "DELETE", eventsTable)
	err := r.querier(spanCtx).DeleteEvent(spanCtx, pgtype.UUID{Bytes: id, Valid: true})
	endDBSpan(span, err)
	if err != nil {
		r.logger.Error("delete event query failed", "error", err)
		return fmt.Errorf("delete event: %w", err)
	}
	return nil
}

func textOrNull(s string) pgtype.Text {
	if s == "" {
		return pgtype.Text{}
	}
	return pgtype.Text{String: s, Valid: true}
}
