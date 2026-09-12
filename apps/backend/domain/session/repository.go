package session

//go:generate go tool mockgen -destination=generated/repository.go -package=mock github.com/Haya372/ai-trial/backend/domain/session Repository

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, userID uuid.UUID, expiresAt time.Time) (Session, error)
	FindByID(ctx context.Context, id uuid.UUID) (Session, error)
	Delete(ctx context.Context, id uuid.UUID) error
}
