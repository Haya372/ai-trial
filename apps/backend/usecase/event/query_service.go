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
}

type QueryService interface {
	List(ctx context.Context, filter ListFilter) ([]EventReadModel, error)
	FindByID(ctx context.Context, id uuid.UUID) (EventReadModel, error)
	// ListByIDs resolves multiple event IDs within a date range in one call,
	// used to fetch EventSubscription targets without an N+1 query per
	// subscription. Events outside the range, or whose ID doesn't exist
	// (e.g. the source event was deleted), are simply omitted rather than
	// erroring.
	ListByIDs(ctx context.Context, ids []uuid.UUID, startDate, endDate time.Time) ([]EventReadModel, error)
}
