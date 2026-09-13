package event

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/Haya372/ai-trial/backend/domain/event"
)

type ListEventsInput struct {
	StartDate time.Time
	EndDate   time.Time
}

type EventReadModel struct {
	ID          uuid.UUID
	Title       string
	Description string
	StartAt     time.Time
	EndAt       time.Time
}

type ListEventsQuery struct {
	eventQueryRepo event.QueryRepository
}

func NewListEventsQuery(r event.QueryRepository) *ListEventsQuery {
	return &ListEventsQuery{eventQueryRepo: r}
}

func (q *ListEventsQuery) Execute(ctx context.Context, userID uuid.UUID, in ListEventsInput) ([]EventReadModel, error) {
	events, err := q.eventQueryRepo.List(ctx, event.ListFilter{
		UserID:    userID,
		StartDate: in.StartDate,
		EndDate:   in.EndDate,
	})
	if err != nil {
		return nil, fmt.Errorf("list events: %w", err)
	}
	result := make([]EventReadModel, len(events))
	for i, e := range events {
		result[i] = EventReadModel{
			ID:          e.ID(),
			Title:       e.Title(),
			Description: e.Description(),
			StartAt:     e.StartAt(),
			EndAt:       e.EndAt(),
		}
	}
	return result, nil
}
