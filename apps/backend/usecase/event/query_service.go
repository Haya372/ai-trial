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
}
