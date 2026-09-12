package session

//go:generate go tool mockgen -destination=generated/repository.go -package=mock github.com/Haya372/ai-trial/backend/domain/session Repository

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, userID uuid.UUID, expiresAt time.Time) (Session, error)
	// FindByID returns the session with the given ID, or nil if not found.
	// It returns expired sessions as well; callers that need only active sessions should use FindActiveByID.
	FindByID(ctx context.Context, id uuid.UUID) (Session, error)
	// FindActiveByID returns the session only if it exists and has not expired.
	// Returns nil, nil when the session does not exist or is expired.
	FindActiveByID(ctx context.Context, id uuid.UUID) (Session, error)
	Delete(ctx context.Context, id uuid.UUID) error
}
