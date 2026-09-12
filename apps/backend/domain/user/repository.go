package user

//go:generate go tool mockgen -destination=generated/repository.go -package=mock github.com/Haya372/ai-trial/backend/domain/user Repository

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, email Email, displayName string, password Password) (User, error)
	FindByEmail(ctx context.Context, email Email) (User, error)
	FindByID(ctx context.Context, id uuid.UUID) (User, error)
}
