package event

//go:generate go tool mockgen -destination=generated/repository.go -package=mock github.com/Haya372/ai-trial/backend/domain/event Repository

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, e Event) (Event, error)
	// FindByID returns the event with the given ID, or ErrEventNotFound if it does not exist.
	FindByID(ctx context.Context, id uuid.UUID) (Event, error)
	Update(ctx context.Context, e Event) error
}
