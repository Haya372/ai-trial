package event

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type ListFilter struct {
	UserID    uuid.UUID
	StartDate time.Time
	EndDate   time.Time
	// TagID, when non-nil, restricts the result to own events that have this
	// tag. Subscribed events are excluded when TagID is set.
	TagID *uuid.UUID
}

type QueryService interface {
	List(ctx context.Context, filter ListFilter) ([]EventReadModel, error)
	FindByID(ctx context.Context, id uuid.UUID) (EventReadModel, error)
}
