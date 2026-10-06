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
	})
	if err != nil {
		return nil, fmt.Errorf("list events: %w", err)
	}
	return events, nil
}
