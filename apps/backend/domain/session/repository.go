package session

//go:generate go tool mockgen -destination=generated/repository.go -package=mock github.com/Haya372/ai-trial/backend/domain/session Repository

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Repository stores sessions keyed by the hash of their plaintext token (ADR-033).
// Implementations never receive or persist the plaintext token.
type Repository interface {
	Create(ctx context.Context, userID uuid.UUID, tokenHash TokenHash, expiresAt time.Time) (Session, error)
	// FindActiveByTokenHash returns the session only if it exists and has not expired.
	// Returns nil, nil when the session does not exist or is expired.
	FindActiveByTokenHash(ctx context.Context, tokenHash TokenHash) (Session, error)
	DeleteByTokenHash(ctx context.Context, tokenHash TokenHash) error
}
