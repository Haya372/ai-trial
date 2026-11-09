package event

//go:generate go tool mockgen -destination=generated/repository.go -package=mock github.com/Haya372/ai-trial/backend/domain/event Repository

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	// Create saves the whole aggregate, including e.Tags(). It writes several
	// rows, so call it inside a transaction (ADR-024).
	Create(ctx context.Context, e Event) (Event, error)
	// FindByID returns the event with the given ID, or ErrEventNotFound if it does not exist.
	FindByID(ctx context.Context, id uuid.UUID) (Event, error)
	// FindByIDForUpdate behaves like FindByID but locks the event row
	// (SELECT ... FOR UPDATE), so concurrent read-modify-writes of the same
	// event serialize and each reads the tags the previous one committed.
	// It must be called inside a transaction (ADR-024).
	FindByIDForUpdate(ctx context.Context, id uuid.UUID) (Event, error)
	// Update saves the whole aggregate, replacing the stored tags with
	// e.Tags(). It writes several rows, so call it inside a transaction (ADR-024).
	Update(ctx context.Context, e Event) error
	Delete(ctx context.Context, id uuid.UUID) error
}
