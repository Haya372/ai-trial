package eventsubscription

//go:generate go tool mockgen -destination=generated/repository.go -package=mock github.com/Haya372/ai-trial/backend/domain/eventsubscription Repository

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, s EventSubscription) (EventSubscription, error)
	FindByID(ctx context.Context, id uuid.UUID) (EventSubscription, error)
	// FindByEventAndUserID returns ErrEventSubscriptionNotFound if the user has
	// not subscribed to the event, so callers can treat "add to my calendar"
	// as idempotent (SPEC-004: repeated adds keep the existing subscription).
	FindByEventAndUserID(ctx context.Context, eventID, userID uuid.UUID) (EventSubscription, error)
	ListByUserID(ctx context.Context, userID uuid.UUID) ([]EventSubscription, error)
	Delete(ctx context.Context, id uuid.UUID) error
}
