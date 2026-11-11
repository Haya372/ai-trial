package repository

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel/trace"

	"github.com/google/uuid"

	domainevent "github.com/Haya372/ai-trial/backend/domain/event"
	"github.com/Haya372/ai-trial/backend/infrastructure/db"
	query "github.com/Haya372/ai-trial/backend/infrastructure/db/generated"
	eventuc "github.com/Haya372/ai-trial/backend/usecase/event"
)

const (
	eventsTable    = "events"
	eventTagsTable = "event_tags"
)

// --- read side ---

type eventQueryRepository struct {
	baseRepository

	logger *slog.Logger
}

func NewEventQueryRepository(pool *pgxpool.Pool, logger *slog.Logger, tp trace.TracerProvider) eventuc.QueryService {
	return &eventQueryRepository{baseRepository{pool: pool, tracer: tp.Tracer(tracerName)}, logger}
}

// eventListRow is the common shape returned by both tag-aware list queries;
// it lets aggregateEventRows handle both without duplicating logic.
type eventListRow struct {
	ID             pgtype.UUID
	UserID         pgtype.UUID
	Title          string
	Description    pgtype.Text
	StartAt        pgtype.Timestamptz
	EndAt          pgtype.Timestamptz
	Location       pgtype.Text
	Url            pgtype.Text
	IsSubscribed   bool
	SubscriptionID pgtype.UUID
	TagID          pgtype.UUID
	TagName        pgtype.Text
	TagColor       pgtype.Text
}

// newEventReadModelFromListRow builds an EventReadModel (without tags) from a
// list-query row. Tags are appended separately after the fact.
func newEventReadModelFromListRow(row eventListRow) eventuc.EventReadModel {
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
	var subscriptionID *uuid.UUID
	if row.SubscriptionID.Valid {
		sid := uuid.UUID(row.SubscriptionID.Bytes)
		subscriptionID = &sid
	}
	return eventuc.EventReadModel{
		ID:             uuid.UUID(row.ID.Bytes),
		UserID:         uuid.UUID(row.UserID.Bytes),
		Title:          row.Title,
		Description:    desc,
		StartAt:        row.StartAt.Time,
		EndAt:          row.EndAt.Time,
		Location:       location,
		URL:            url,
		IsSubscribed:   row.IsSubscribed,
		SubscriptionID: subscriptionID,
		Tags:           []eventuc.EventTagReadModel{},
	}
}

// aggregateEventRows collapses a 1-row-per-tag result set (from both
// list queries) into one EventReadModel per event in first-seen order.
func aggregateEventRows(rows []eventListRow) []eventuc.EventReadModel {
	seen := make(map[uuid.UUID]*eventuc.EventReadModel)
	order := make([]uuid.UUID, 0, len(rows))

	for _, row := range rows {
		id := uuid.UUID(row.ID.Bytes)
		rm, exists := seen[id]
		if !exists {
			m := newEventReadModelFromListRow(row)
			seen[id] = &m
			order = append(order, id)
			rm = seen[id]
		}
		// A NULL TagID means the LEFT JOIN found no tags (untagged event)
		// or the UNION ALL subscribed side always sends NULL tags.
		if row.TagID.Valid && row.TagName.Valid && row.TagColor.Valid {
			rm.Tags = append(rm.Tags, eventuc.EventTagReadModel{
				ID:    uuid.UUID(row.TagID.Bytes),
				Name:  row.TagName.String,
				Color: row.TagColor.String,
			})
		}
	}

	result := make([]eventuc.EventReadModel, 0, len(order))
	for _, id := range order {
		result = append(result, *seen[id])
	}
	return result
}

func (r *eventQueryRepository) List(ctx context.Context, filter eventuc.ListFilter) ([]eventuc.EventReadModel, error) {
	spanCtx, span := r.startDBSpan(ctx, "SELECT", eventsTable)

	var rawRows []eventListRow
	var err error
	if filter.TagID != nil {
		var rows []query.ListEventsWithTagsByUserAndDateRangeFilteredByTagRow
		rows, err = r.querier(spanCtx).ListEventsWithTagsByUserAndDateRangeFilteredByTag(spanCtx,
			query.ListEventsWithTagsByUserAndDateRangeFilteredByTagParams{
				TagID:     toPgUUID(*filter.TagID),
				UserID:    toPgUUID(filter.UserID),
				StartDate: pgtype.Timestamptz{Time: filter.StartDate, Valid: true},
				EndDate:   pgtype.Timestamptz{Time: filter.EndDate, Valid: true},
			})
		if err == nil {
			rawRows = make([]eventListRow, len(rows))
			for i, r := range rows {
				rawRows[i] = eventListRow{
					ID: r.ID, UserID: r.UserID, Title: r.Title, Description: r.Description,
					StartAt: r.StartAt, EndAt: r.EndAt, Location: r.Location, Url: r.Url,
					IsSubscribed: r.IsSubscribed, SubscriptionID: r.SubscriptionID,
					TagID: r.TagID, TagName: r.TagName, TagColor: r.TagColor,
				}
			}
		}
	} else {
		var rows []query.ListEventsWithTagsByUserAndDateRangeRow
		rows, err = r.querier(spanCtx).ListEventsWithTagsByUserAndDateRange(spanCtx,
			query.ListEventsWithTagsByUserAndDateRangeParams{
				UserID:    toPgUUID(filter.UserID),
				StartDate: pgtype.Timestamptz{Time: filter.StartDate, Valid: true},
				EndDate:   pgtype.Timestamptz{Time: filter.EndDate, Valid: true},
			})
		if err == nil {
			rawRows = make([]eventListRow, len(rows))
			for i, r := range rows {
				rawRows[i] = eventListRow{
					ID: r.ID, UserID: r.UserID, Title: r.Title, Description: r.Description,
					StartAt: r.StartAt, EndAt: r.EndAt, Location: r.Location, Url: r.Url,
					IsSubscribed: r.IsSubscribed, SubscriptionID: r.SubscriptionID,
					TagID: r.TagID, TagName: r.TagName, TagColor: r.TagColor,
				}
			}
		}
	}

	endDBSpan(span, err)
	if err != nil {
		r.logger.Error("list events query failed", "error", err)
		return nil, fmt.Errorf("list events: %w", err)
	}
	return aggregateEventRows(rawRows), nil
}

func (r *eventQueryRepository) FindByID(ctx context.Context, id uuid.UUID) (eventuc.EventReadModel, error) {
	spanCtx, span := r.startDBSpan(ctx, "SELECT", eventsTable)
	row, err := r.querier(spanCtx).FindEventReadModelByID(spanCtx, toPgUUID(id))
	endDBSpanNotFound(span, err)
	if notFound, wasNoRows, mappedErr := checkNotFound(err, domainevent.ErrEventNotFound); notFound {
		if wasNoRows {
			return eventuc.EventReadModel{}, mappedErr
		}
		r.logger.Error("find event read model by id query failed", "error", mappedErr)
		return eventuc.EventReadModel{}, fmt.Errorf("find event read model by id: %w", mappedErr)
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
	return eventuc.EventReadModel{
		ID:          uuid.UUID(row.ID.Bytes),
		UserID:      uuid.UUID(row.UserID.Bytes),
		Title:       row.Title,
		Description: desc,
		StartAt:     row.StartAt.Time,
		EndAt:       row.EndAt.Time,
		Location:    location,
		URL:         url,
	}, nil
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

// Create saves the whole aggregate: the event row and its tags.
func (r *eventRepository) Create(ctx context.Context, e domainevent.Event) (domainevent.Event, error) {
	row, err := r.insertEvent(ctx, e)
	if err != nil {
		return nil, err
	}
	if err := r.saveTags(ctx, e); err != nil {
		return nil, err
	}
	return eventFromRow(row, e.Tags())
}

// eventFromRow reconstructs the event from the persisted row so the caller
// gets the exact stored values.
func eventFromRow(row query.InsertEventRow, tags []domainevent.EventTag) (domainevent.Event, error) {
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
	e, err := domainevent.New(
		uuid.UUID(row.ID.Bytes), uuid.UUID(row.UserID.Bytes),
		row.Title, desc, row.StartAt.Time, row.EndAt.Time, location, url,
	)
	if err != nil {
		return nil, fmt.Errorf("reconstruct event: %w", err)
	}
	return e.WithTags(tags)
}

func (r *eventRepository) FindByID(ctx context.Context, id uuid.UUID) (domainevent.Event, error) {
	spanCtx, span := r.startDBSpan(ctx, "SELECT", eventsTable)
	rows, err := r.querier(spanCtx).FindEventWithTagsByID(spanCtx, toPgUUID(id))
	endDBSpan(span, err)
	if err != nil {
		r.logger.Error("find event by id query failed", "error", err)
		return nil, fmt.Errorf("find event by id: %w", err)
	}
	if len(rows) == 0 {
		return nil, domainevent.ErrEventNotFound
	}

	tags := make([]domainevent.EventTag, 0, len(rows))
	for _, row := range rows {
		// A row with no tag columns is the LEFT JOIN's row for an untagged event.
		if !row.TagID.Valid {
			continue
		}
		t, err := domainevent.NewEventTag(uuid.UUID(row.TagID.Bytes), int(row.Position.Int16))
		if err != nil {
			return nil, fmt.Errorf("reconstruct event tag: %w", err)
		}
		tags = append(tags, t)
	}
	first := rows[0]
	return eventFromRow(query.InsertEventRow{
		ID:          first.ID,
		UserID:      first.UserID,
		Title:       first.Title,
		Description: first.Description,
		StartAt:     first.StartAt,
		EndAt:       first.EndAt,
		Location:    first.Location,
		Url:         first.Url,
		CreatedAt:   first.CreatedAt,
		UpdatedAt:   first.UpdatedAt,
	}, tags)
}

func (r *eventRepository) FindByIDForUpdate(ctx context.Context, id uuid.UUID) (domainevent.Event, error) {
	if _, ok := db.GetTx(ctx); !ok {
		return nil, fmt.Errorf("find event for update: %w", errNoTransaction)
	}
	spanCtx, span := r.startDBSpan(ctx, "SELECT", eventsTable)
	row, err := r.querier(spanCtx).FindEventByIDForUpdate(spanCtx, toPgUUID(id))
	endDBSpanNotFound(span, err)
	if notFound, wasNoRows, mappedErr := checkNotFound(err, domainevent.ErrEventNotFound); notFound {
		if wasNoRows {
			return nil, mappedErr
		}
		r.logger.Error("find event for update query failed", "error", mappedErr)
		return nil, fmt.Errorf("find event for update: %w", mappedErr)
	}

	// Read the tags after locking the event row: under READ COMMITTED this
	// statement sees the tags committed by the writer that held the lock.
	tags, err := r.findTags(ctx, id)
	if err != nil {
		return nil, err
	}
	// The generated row types are identical structs, one per query.
	return eventFromRow(query.InsertEventRow(row), tags)
}

// Update saves the whole aggregate: the event row and its tags, replacing
// the stored tags with e.Tags().
func (r *eventRepository) Update(ctx context.Context, e domainevent.Event) error {
	spanCtx, span := r.startDBSpan(ctx, "UPDATE", eventsTable)
	_, err := r.querier(spanCtx).UpdateEvent(spanCtx, query.UpdateEventParams{
		ID:          toPgUUID(e.ID()),
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
	return r.saveTags(ctx, e)
}

func (r *eventRepository) Delete(ctx context.Context, id uuid.UUID) error {
	spanCtx, span := r.startDBSpan(ctx, "DELETE", eventsTable)
	err := r.querier(spanCtx).DeleteEvent(spanCtx, toPgUUID(id))
	endDBSpan(span, err)
	if err != nil {
		r.logger.Error("delete event query failed", "error", err)
		return fmt.Errorf("delete event: %w", err)
	}
	return nil
}

func (r *eventRepository) insertEvent(ctx context.Context, e domainevent.Event) (query.InsertEventRow, error) {
	spanCtx, span := r.startDBSpan(ctx, "INSERT", eventsTable)
	row, err := r.querier(spanCtx).InsertEvent(spanCtx, query.InsertEventParams{
		ID:          toPgUUID(e.ID()),
		UserID:      toPgUUID(e.UserID()),
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
		return query.InsertEventRow{}, fmt.Errorf("insert event: %w", err)
	}
	return row, nil
}

// saveTags makes the stored tags of e match e.Tags(). It deletes before
// inserting because a kept tag may move to a position another tag still
// holds, which would trip UNIQUE (event_id, position).
func (r *eventRepository) saveTags(ctx context.Context, e domainevent.Event) error {
	spanCtx, span := r.startDBSpan(ctx, "DELETE", eventTagsTable)
	err := r.querier(spanCtx).DeleteEventTagsByEventID(spanCtx, toPgUUID(e.ID()))
	endDBSpan(span, err)
	if err != nil {
		r.logger.Error("delete event tags query failed", "error", err)
		return fmt.Errorf("delete event tags: %w", err)
	}

	for _, t := range e.Tags() {
		spanCtx, span := r.startDBSpan(ctx, "INSERT", eventTagsTable)
		err := r.querier(spanCtx).InsertEventTag(spanCtx, query.InsertEventTagParams{
			EventID: toPgUUID(e.ID()),
			TagID:   toPgUUID(t.TagID()),
			//nolint:gosec // position is validated to be 1..MaxTagsPerEvent by the domain
			Position: int16(t.Position()),
		})
		endDBSpan(span, err)
		if err != nil {
			r.logger.Error("insert event tag query failed", "error", err)
			return fmt.Errorf("insert event tag: %w", err)
		}
	}
	return nil
}

func (r *eventRepository) findTags(ctx context.Context, eventID uuid.UUID) ([]domainevent.EventTag, error) {
	spanCtx, span := r.startDBSpan(ctx, "SELECT", eventTagsTable)
	rows, err := r.querier(spanCtx).ListEventTagsByEventID(spanCtx, toPgUUID(eventID))
	endDBSpan(span, err)
	if err != nil {
		r.logger.Error("list event tags query failed", "error", err)
		return nil, fmt.Errorf("list event tags: %w", err)
	}

	tags := make([]domainevent.EventTag, 0, len(rows))
	for _, row := range rows {
		t, err := domainevent.NewEventTag(uuid.UUID(row.TagID.Bytes), int(row.Position))
		if err != nil {
			return nil, fmt.Errorf("reconstruct event tag: %w", err)
		}
		tags = append(tags, t)
	}
	return tags, nil
}

func textOrNull(s string) pgtype.Text {
	if s == "" {
		return pgtype.Text{}
	}
	return pgtype.Text{String: s, Valid: true}
}
