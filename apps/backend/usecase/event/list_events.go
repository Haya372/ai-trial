package event

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"

	domaineventsubscription "github.com/Haya372/ai-trial/backend/domain/eventsubscription"
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
	// caller. Such events are read-only.
	IsSubscribed bool
}

type ListEventsQuery struct {
	queryService     QueryService
	subscriptionRepo domaineventsubscription.Repository
}

func NewListEventsQuery(s QueryService, subs domaineventsubscription.Repository) *ListEventsQuery {
	return &ListEventsQuery{queryService: s, subscriptionRepo: subs}
}

func (q *ListEventsQuery) Execute(ctx context.Context, userID uuid.UUID, in ListEventsInput) ([]EventReadModel, error) {
	owned, err := q.queryService.List(ctx, ListFilter{
		UserID:    userID,
		StartDate: in.StartDate,
		EndDate:   in.EndDate,
	})
	if err != nil {
		return nil, fmt.Errorf("list events: %w", err)
	}

	result := make([]EventReadModel, 0, len(owned))
	result = append(result, owned...)

	subs, err := q.subscriptionRepo.ListByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list event subscriptions: %w", err)
	}
	for _, s := range subs {
		ev, err := q.queryService.FindByID(ctx, s.EventID())
		if err != nil {
			return nil, fmt.Errorf("find subscribed event: %w", err)
		}
		if !overlapsRange(ev.StartAt, ev.EndAt, in.StartDate, in.EndDate) {
			continue
		}
		ev.IsSubscribed = true
		result = append(result, ev)
	}

	sort.Slice(result, func(i, j int) bool { return result[i].StartAt.Before(result[j].StartAt) })
	return result, nil
}

// overlapsRange mirrors the SQL date-range filter used for the caller's own
// events (end_at > start AND start_at < end), so subscribed events are
// filtered to the requested range the same way.
func overlapsRange(eventStart, eventEnd, rangeStart, rangeEnd time.Time) bool {
	return eventEnd.After(rangeStart) && eventStart.Before(rangeEnd)
}
