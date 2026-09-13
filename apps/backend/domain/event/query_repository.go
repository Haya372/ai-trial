package event

//go:generate go tool mockgen -destination=generated/query_repository.go -package=mock github.com/Haya372/ai-trial/backend/domain/event QueryRepository

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

type QueryRepository interface {
	List(ctx context.Context, filter ListFilter) ([]Event, error)
}
