package event

//go:generate go tool mockgen -destination=generated/repository.go -package=mock github.com/Haya372/ai-trial/backend/domain/event Repository

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	// Create also stores e.Tags(). It writes several rows, so call it inside
	// a transaction to keep the event and its tags consistent (ADR-024).
	Create(ctx context.Context, e Event) (Event, error)
	// FindByID returns the event with the given ID, or ErrEventNotFound if it does not exist.
	FindByID(ctx context.Context, id uuid.UUID) (Event, error)
	// Update replaces the stored tags with e.Tags(). It writes several rows,
	// so call it inside a transaction to keep the event and its tags consistent (ADR-024).
	Update(ctx context.Context, e Event) error
	Delete(ctx context.Context, id uuid.UUID) error
}
