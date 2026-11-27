package event

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type ListEventsInput struct {
	StartDate time.Time
	EndDate   time.Time
	// TagID, when non-nil, restricts the result to own events that have this
	// tag. Subscribed events are excluded when TagID is set.
	TagID *uuid.UUID
}

// EventTagReadModel is the tag data embedded in a list-events response.
// Tags are sorted by ascending position (SQL ORDER BY).
type EventTagReadModel struct {
	ID    uuid.UUID
	Name  string
	Color string
}

type EventReadModel struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	Title       string
	Description string
	StartAt     time.Time
	EndAt       time.Time
	Location    string
	URL         string
	// IsSubscribed is true when the event was added via EventSubscription
	// (another user's event the caller subscribed to), not created by the
	// caller. Such events are read-only. The query side's SQL (ADR-022
	// logical CQRS) sets this directly, so this layer just passes it through.
	IsSubscribed bool
	// SubscriptionID is the EventSubscription's own ID, set only when
	// IsSubscribed is true. Callers need it to remove the event from their
	// calendar via DELETE /subscriptions/{id}.
	SubscriptionID *uuid.UUID
	// Tags are the event's attached tags in position order. Untagged events
	// and subscribed events always have an empty (non-nil) slice.
	Tags []EventTagReadModel
	// RecurringSeries is nil for one-off events and for subscribed events:
	// the owner's series is never exposed to subscribers (SPEC-009).
	RecurringSeries *RecurringSeriesReadModel
}

// RecurringSeriesReadModel holds the stored values as they are, without domain
// validation (ADR-022). Each optional field is set whenever its column is not
// NULL, without looking at PatternType or EndType; that only monthly_nth_weekday
// has MonthlyNth, until_date has UntilDate and count has Count is an invariant
// the write side guarantees when it saves the series.
type RecurringSeriesReadModel struct {
	ID          uuid.UUID
	PatternType string
	MonthlyNth  *string
	EndType     string
	// UntilDate is a YYYY-MM-DD local date in Timezone.
	UntilDate *string
	Count     *int
	Timezone  string
}

type ListEventsQuery struct {
	queryService QueryService
}

func NewListEventsQuery(s QueryService) *ListEventsQuery {
	return &ListEventsQuery{queryService: s}
}

func (q *ListEventsQuery) Execute(ctx context.Context, userID uuid.UUID, in ListEventsInput) ([]EventReadModel, error) {
	events, err := q.queryService.List(ctx, ListFilter{
		UserID:    userID,
		StartDate: in.StartDate,
		EndDate:   in.EndDate,
		TagID:     in.TagID,
	})
	if err != nil {
		return nil, fmt.Errorf("list events: %w", err)
	}
	return events, nil
}
