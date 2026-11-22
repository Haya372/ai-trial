package repository

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel/trace"

	domainevent "github.com/Haya372/ai-trial/backend/domain/event"
	rs "github.com/Haya372/ai-trial/backend/domain/recurringeventseries"
	"github.com/Haya372/ai-trial/backend/infrastructure/db"
	query "github.com/Haya372/ai-trial/backend/infrastructure/db/generated"
)

const (
	recurringSeriesTable           = "recurring_event_series"
	recurringSeriesExceptionsTable = "recurring_series_exceptions"
)

var (
	errOccurrenceWithoutSeries = errors.New("occurrence has no series ID")
	errStoredSeriesInvalid     = errors.New("stored recurring series is invalid")
)

// seriesRow is the stored form of a series without its excluded dates. The
// generated row types of the series queries share this shape.
type seriesRow = query.FindRecurringSeriesByIDForUpdateRow

type seriesRepository struct {
	baseRepository

	logger *slog.Logger
}

// NewRecurringEventSeriesRepository returns a recurringeventseries.Repository backed by PostgreSQL.
func NewRecurringEventSeriesRepository(
	pool *pgxpool.Pool, logger *slog.Logger, tp trace.TracerProvider,
) rs.Repository {
	return &seriesRepository{baseRepository{pool: pool, tracer: tp.Tracer(tracerName)}, logger}
}

func (r *seriesRepository) Create(ctx context.Context, s rs.RecurringEventSeries) (rs.RecurringEventSeries, error) {
	row, excluded := seriesRowOf(s)
	spanCtx, span := r.startDBSpan(ctx, "INSERT", recurringSeriesTable)
	err := r.querier(spanCtx).InsertRecurringSeries(spanCtx, query.InsertRecurringSeriesParams(row))
	endDBSpan(span, err)
	if err != nil {
		r.logger.Error("insert recurring series query failed", "error", err)
		return nil, fmt.Errorf("insert recurring series: %w", err)
	}
	if err := r.insertExceptions(ctx, s.ID(), excluded); err != nil {
		return nil, err
	}
	return s, nil
}

func (r *seriesRepository) FindByID(ctx context.Context, id uuid.UUID) (rs.RecurringEventSeries, error) {
	spanCtx, span := r.startDBSpan(ctx, "SELECT", recurringSeriesTable)
	rows, err := r.querier(spanCtx).FindRecurringSeriesWithExceptionsByID(spanCtx, toPgUUID(id))
	endDBSpan(span, err)
	if err != nil {
		r.logger.Error("find recurring series by id query failed", "error", err)
		return nil, fmt.Errorf("find recurring series by id: %w", err)
	}
	if len(rows) == 0 {
		return nil, rs.ErrRecurringSeriesNotFound
	}

	excluded := make([]pgtype.Date, 0, len(rows))
	for _, row := range rows {
		// A NULL date is the LEFT JOIN's row for a series with no excluded dates.
		if row.ExcludedDate.Valid {
			excluded = append(excluded, row.ExcludedDate)
		}
	}
	first := rows[0]
	return seriesFromRow(seriesRow{
		ID:           first.ID,
		UserID:       first.UserID,
		PatternType:  first.PatternType,
		MonthlyNth:   first.MonthlyNth,
		EndType:      first.EndType,
		EndUntilDate: first.EndUntilDate,
		EndCount:     first.EndCount,
		Timezone:     first.Timezone,
	}, excluded)
}

func (r *seriesRepository) FindByIDForUpdate(ctx context.Context, id uuid.UUID) (rs.RecurringEventSeries, error) {
	if _, ok := db.GetTx(ctx); !ok {
		return nil, fmt.Errorf("find recurring series for update: %w", errNoTransaction)
	}
	spanCtx, span := r.startDBSpan(ctx, "SELECT", recurringSeriesTable)
	row, err := r.querier(spanCtx).FindRecurringSeriesByIDForUpdate(spanCtx, toPgUUID(id))
	endDBSpanNotFound(span, err)
	if notFound, wasNoRows, mappedErr := checkNotFound(err, rs.ErrRecurringSeriesNotFound); notFound {
		if wasNoRows {
			return nil, mappedErr
		}
		r.logger.Error("find recurring series for update query failed", "error", mappedErr)
		return nil, fmt.Errorf("find recurring series for update: %w", mappedErr)
	}

	// Read the excluded dates after locking the series row: under READ
	// COMMITTED this statement sees the dates committed by the previous holder.
	spanCtx, span = r.startDBSpan(ctx, "SELECT", recurringSeriesExceptionsTable)
	excluded, err := r.querier(spanCtx).ListRecurringSeriesExceptions(spanCtx, toPgUUID(id))
	endDBSpan(span, err)
	if err != nil {
		r.logger.Error("list recurring series exceptions query failed", "error", err)
		return nil, fmt.Errorf("list recurring series exceptions: %w", err)
	}
	return seriesFromRow(row, excluded)
}

// Update replaces the stored excluded dates with s.ExcludedDates(); the
// series keeps few of them, so rewriting all is simpler than diffing.
func (r *seriesRepository) Update(ctx context.Context, s rs.RecurringEventSeries) error {
	row, excluded := seriesRowOf(s)
	spanCtx, span := r.startDBSpan(ctx, "UPDATE", recurringSeriesTable)
	err := r.querier(spanCtx).UpdateRecurringSeries(spanCtx, query.UpdateRecurringSeriesParams{
		PatternType:  row.PatternType,
		MonthlyNth:   row.MonthlyNth,
		EndType:      row.EndType,
		EndUntilDate: row.EndUntilDate,
		EndCount:     row.EndCount,
		Timezone:     row.Timezone,
		ID:           row.ID,
	})
	endDBSpan(span, err)
	if err != nil {
		r.logger.Error("update recurring series query failed", "error", err)
		return fmt.Errorf("update recurring series: %w", err)
	}

	spanCtx, span = r.startDBSpan(ctx, "DELETE", recurringSeriesExceptionsTable)
	err = r.querier(spanCtx).DeleteRecurringSeriesExceptions(spanCtx, row.ID)
	endDBSpan(span, err)
	if err != nil {
		r.logger.Error("delete recurring series exceptions query failed", "error", err)
		return fmt.Errorf("delete recurring series exceptions: %w", err)
	}
	return r.insertExceptions(ctx, s.ID(), excluded)
}

func (r *seriesRepository) Delete(ctx context.Context, id uuid.UUID) error {
	spanCtx, span := r.startDBSpan(ctx, "DELETE", recurringSeriesTable)
	err := r.querier(spanCtx).DeleteRecurringSeries(spanCtx, toPgUUID(id))
	endDBSpan(span, err)
	if err != nil {
		r.logger.Error("delete recurring series query failed", "error", err)
		return fmt.Errorf("delete recurring series: %w", err)
	}
	return nil
}

func (r *seriesRepository) FindOccurrences(ctx context.Context, seriesID uuid.UUID) ([]domainevent.Event, error) {
	spanCtx, span := r.startDBSpan(ctx, "SELECT", eventsTable)
	rows, err := r.querier(spanCtx).ListOccurrencesWithTagsBySeriesID(spanCtx, toPgUUID(seriesID))
	endDBSpan(span, err)
	if err != nil {
		r.logger.Error("list occurrences query failed", "error", err)
		return nil, fmt.Errorf("list occurrences: %w", err)
	}
	return occurrencesFromRows(rows)
}

func (r *seriesRepository) CountOccurrences(ctx context.Context, seriesID uuid.UUID) (int, error) {
	spanCtx, span := r.startDBSpan(ctx, "SELECT", eventsTable)
	n, err := r.querier(spanCtx).CountOccurrencesBySeriesID(spanCtx, toPgUUID(seriesID))
	endDBSpan(span, err)
	if err != nil {
		r.logger.Error("count occurrences query failed", "error", err)
		return 0, fmt.Errorf("count occurrences: %w", err)
	}
	return int(n), nil
}

// CreateOccurrences uses COPY: a series has up to about 730 occurrences,
// and inserting them one by one would take over a thousand round trips.
func (r *seriesRepository) CreateOccurrences(ctx context.Context, occurrences []domainevent.Event) error {
	events := make([]query.InsertOccurrencesParams, 0, len(occurrences))
	var tags []query.InsertOccurrenceTagsParams
	for _, e := range occurrences {
		seriesID, ok := e.SeriesID()
		if !ok {
			return fmt.Errorf("create occurrences: event %v: %w", e.ID(), errOccurrenceWithoutSeries)
		}
		events = append(events, query.InsertOccurrencesParams{
			ID:          toPgUUID(e.ID()),
			UserID:      toPgUUID(e.UserID()),
			Title:       e.Title(),
			Description: textOrNull(e.Description()),
			StartAt:     pgtype.Timestamptz{Time: e.StartAt(), Valid: true},
			EndAt:       pgtype.Timestamptz{Time: e.EndAt(), Valid: true},
			Location:    textOrNull(e.Location()),
			Url:         textOrNull(e.URL()),
			SeriesID:    toPgUUID(seriesID),
		})
		for _, t := range e.Tags() {
			tags = append(tags, query.InsertOccurrenceTagsParams{
				EventID: toPgUUID(e.ID()),
				TagID:   toPgUUID(t.TagID()),
				//nolint:gosec // position is validated to be 1..MaxTagsPerEvent by the domain
				Position: int16(t.Position()),
			})
		}
	}
	if len(events) == 0 {
		return nil
	}

	spanCtx, span := r.startDBSpan(ctx, "INSERT", eventsTable)
	_, err := r.querier(spanCtx).InsertOccurrences(spanCtx, events)
	endDBSpan(span, err)
	if err != nil {
		r.logger.Error("insert occurrences query failed", "error", err)
		return fmt.Errorf("insert occurrences: %w", err)
	}
	if len(tags) == 0 {
		return nil
	}

	spanCtx, span = r.startDBSpan(ctx, "INSERT", eventTagsTable)
	_, err = r.querier(spanCtx).InsertOccurrenceTags(spanCtx, tags)
	endDBSpan(span, err)
	if err != nil {
		r.logger.Error("insert occurrence tags query failed", "error", err)
		return fmt.Errorf("insert occurrence tags: %w", err)
	}
	return nil
}

func (r *seriesRepository) DeleteOccurrencesFrom(
	ctx context.Context, seriesID uuid.UUID, from time.Time,
) error {
	spanCtx, span := r.startDBSpan(ctx, "DELETE", eventsTable)
	err := r.querier(spanCtx).DeleteOccurrencesFromBySeriesID(spanCtx, query.DeleteOccurrencesFromBySeriesIDParams{
		SeriesID: toPgUUID(seriesID),
		FromAt:   pgtype.Timestamptz{Time: from, Valid: true},
	})
	endDBSpan(span, err)
	if err != nil {
		r.logger.Error("delete occurrences from query failed", "error", err)
		return fmt.Errorf("delete occurrences from: %w", err)
	}
	return nil
}

func (r *seriesRepository) DeleteOccurrences(ctx context.Context, seriesID uuid.UUID) error {
	spanCtx, span := r.startDBSpan(ctx, "DELETE", eventsTable)
	err := r.querier(spanCtx).DeleteOccurrencesBySeriesID(spanCtx, toPgUUID(seriesID))
	endDBSpan(span, err)
	if err != nil {
		r.logger.Error("delete occurrences query failed", "error", err)
		return fmt.Errorf("delete occurrences: %w", err)
	}
	return nil
}

func (r *seriesRepository) insertExceptions(
	ctx context.Context, seriesID uuid.UUID, excluded []pgtype.Date,
) error {
	if len(excluded) == 0 {
		return nil
	}
	params := make([]query.InsertRecurringSeriesExceptionsParams, 0, len(excluded))
	for _, d := range excluded {
		params = append(params, query.InsertRecurringSeriesExceptionsParams{SeriesID: toPgUUID(seriesID), ExcludedDate: d})
	}
	spanCtx, span := r.startDBSpan(ctx, "INSERT", recurringSeriesExceptionsTable)
	_, err := r.querier(spanCtx).InsertRecurringSeriesExceptions(spanCtx, params)
	endDBSpan(span, err)
	if err != nil {
		r.logger.Error("insert recurring series exceptions query failed", "error", err)
		return fmt.Errorf("insert recurring series exceptions: %w", err)
	}
	return nil
}

// occurrencesFromRows collapses the one-row-per-tag result, ordered so each
// event's rows are adjacent, into events.
func occurrencesFromRows(rows []query.ListOccurrencesWithTagsBySeriesIDRow) ([]domainevent.Event, error) {
	var out []domainevent.Event
	for i := 0; i < len(rows); {
		first := rows[i]
		var tags []domainevent.EventTag
		for ; i < len(rows) && rows[i].ID == first.ID; i++ {
			if !rows[i].TagID.Valid {
				continue
			}
			t, err := domainevent.NewEventTag(uuid.UUID(rows[i].TagID.Bytes), int(rows[i].Position.Int16))
			if err != nil {
				return nil, fmt.Errorf("reconstruct event tag: %w", err)
			}
			tags = append(tags, t)
		}
		e, err := eventFromRow(query.InsertEventRow{
			ID:          first.ID,
			UserID:      first.UserID,
			Title:       first.Title,
			Description: first.Description,
			StartAt:     first.StartAt,
			EndAt:       first.EndAt,
			Location:    first.Location,
			Url:         first.Url,
			SeriesID:    first.SeriesID,
			CreatedAt:   first.CreatedAt,
			UpdatedAt:   first.UpdatedAt,
		}, tags)
		if err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, nil
}

func seriesRowOf(s rs.RecurringEventSeries) (seriesRow, []pgtype.Date) {
	rule := s.Rule()
	row := seriesRow{
		ID:          toPgUUID(s.ID()),
		UserID:      toPgUUID(s.UserID()),
		PatternType: string(rule.Pattern.Type()),
		EndType:     string(rule.End.Type()),
		Timezone:    rule.Timezone.Name(),
	}
	if nth, ok := rule.Pattern.MonthlyNth(); ok {
		row.MonthlyNth = pgtype.Text{String: string(nth), Valid: true}
	}
	if d, ok := rule.End.UntilDate(); ok {
		row.EndUntilDate = dateOf(d)
	}
	if n, ok := rule.End.Count(); ok {
		//nolint:gosec // a count fits 2 years of daily occurrences (about 730)
		row.EndCount = pgtype.Int2{Int16: int16(n), Valid: true}
	}

	excluded := make([]pgtype.Date, 0, len(s.ExcludedDates()))
	for _, e := range s.ExcludedDates() {
		excluded = append(excluded, dateOf(e.Date()))
	}
	return row, excluded
}

// seriesFromRow validates the stored values with the domain constructors,
// because the database does not constrain them (ADR-019).
func seriesFromRow(row seriesRow, excluded []pgtype.Date) (rs.RecurringEventSeries, error) {
	var nth *string
	if row.MonthlyNth.Valid {
		nth = &row.MonthlyNth.String
	}
	pattern, err := rs.NewRecurrencePattern(row.PatternType, nth)
	if err != nil {
		return nil, fmt.Errorf("reconstruct recurring series: %w", err)
	}
	end, err := endFromRow(row)
	if err != nil {
		return nil, fmt.Errorf("reconstruct recurring series: %w", err)
	}
	tz, err := rs.NewSeriesTimezone(row.Timezone)
	if err != nil {
		return nil, fmt.Errorf("reconstruct recurring series: %w", err)
	}

	dates := make([]rs.ExcludedDate, 0, len(excluded))
	for _, d := range excluded {
		ld, err := localDateOf(d)
		if err != nil {
			return nil, fmt.Errorf("reconstruct recurring series: %w", err)
		}
		dates = append(dates, rs.NewExcludedDate(ld))
	}
	s, err := rs.New(uuid.UUID(row.ID.Bytes), uuid.UUID(row.UserID.Bytes),
		rs.RecurrenceRule{Pattern: pattern, End: end, Timezone: tz}, dates)
	if err != nil {
		return nil, fmt.Errorf("reconstruct recurring series: %w", err)
	}
	return s, nil
}

func endFromRow(row seriesRow) (rs.RecurrenceEnd, error) {
	switch rs.EndType(row.EndType) {
	case rs.EndIndefinite:
		return rs.NewIndefiniteEnd(), nil
	case rs.EndUntilDate:
		ld, err := localDateOf(row.EndUntilDate)
		if err != nil {
			return rs.RecurrenceEnd{}, err
		}
		return rs.NewUntilDateEnd(ld), nil
	case rs.EndCount:
		if !row.EndCount.Valid {
			return rs.RecurrenceEnd{}, fmt.Errorf("%w: end_count is NULL for a count end", errStoredSeriesInvalid)
		}
		return rs.RestoreCountEnd(int(row.EndCount.Int16))
	}
	return rs.RecurrenceEnd{}, fmt.Errorf("%w: unknown end_type %q", errStoredSeriesInvalid, row.EndType)
}

func dateOf(d rs.LocalDate) pgtype.Date {
	return pgtype.Date{Time: time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, time.UTC), Valid: true}
}

func localDateOf(d pgtype.Date) (rs.LocalDate, error) {
	if !d.Valid {
		return rs.LocalDate{}, fmt.Errorf("%w: date is NULL", errStoredSeriesInvalid)
	}
	return rs.NewLocalDate(d.Time.Year(), d.Time.Month(), d.Time.Day())
}
