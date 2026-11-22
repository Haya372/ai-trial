package recurringeventseries

//go:generate go tool mockgen -destination=generated/repository.go -package=mock github.com/Haya372/ai-trial/backend/domain/recurringeventseries Repository

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/Haya372/ai-trial/backend/domain/event"
)

// Repository stores series and their occurrences. The occurrences are Event
// aggregates, but their bulk operations live here because they are always
// selected by series.
type Repository interface {
	// Create saves the whole aggregate, including s.ExcludedDates(). It
	// writes several rows, so call it inside a transaction (ADR-024).
	Create(ctx context.Context, s RecurringEventSeries) (RecurringEventSeries, error)
	// FindByID returns the series with the given ID, or
	// ErrRecurringSeriesNotFound if it does not exist.
	FindByID(ctx context.Context, id uuid.UUID) (RecurringEventSeries, error)
	// FindByIDForUpdate behaves like FindByID but locks the series row
	// (SELECT ... FOR UPDATE), so concurrent changes to the same series, such
	// as two occurrences excluded at once, serialize. It must be called
	// inside a transaction (ADR-024).
	FindByIDForUpdate(ctx context.Context, id uuid.UUID) (RecurringEventSeries, error)
	// Update saves the whole aggregate, replacing the stored excluded dates
	// with s.ExcludedDates(). It writes several rows, so call it inside a
	// transaction (ADR-024).
	Update(ctx context.Context, s RecurringEventSeries) error
	// Delete deletes the series together with its occurrences and excluded
	// dates.
	Delete(ctx context.Context, id uuid.UUID) error

	// FindOccurrences returns the events of the series with their tags, in
	// ascending order of start.
	FindOccurrences(ctx context.Context, seriesID uuid.UUID) ([]event.Event, error)
	// CountOccurrences returns how many events the series has, so a command
	// can delete a series left with none.
	CountOccurrences(ctx context.Context, seriesID uuid.UUID) (int, error)
	// CreateOccurrences saves up to about 730 events (2 years of daily
	// occurrences) with their tags at once. Each event must already carry
	// its series ID. Call it inside a transaction (ADR-024).
	CreateOccurrences(ctx context.Context, occurrences []event.Event) error
	// DeleteOccurrencesFrom deletes the events of the series that start at or
	// after from, for "this and following" changes and deletes.
	DeleteOccurrencesFrom(ctx context.Context, seriesID uuid.UUID, from time.Time) error
	// DeleteOccurrences deletes every event of the series but keeps the
	// series, for an "all" change that regenerates the occurrences.
	DeleteOccurrences(ctx context.Context, seriesID uuid.UUID) error
}
