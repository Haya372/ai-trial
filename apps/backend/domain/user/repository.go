package user

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, email Email, displayName string, password Password) (User, error)
	FindByEmail(ctx context.Context, email Email) (User, error)
	FindByID(ctx context.Context, id uuid.UUID) (User, error)
}
