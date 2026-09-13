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
	Title       string
	Description string
	StartAt     time.Time
	EndAt       time.Time
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
