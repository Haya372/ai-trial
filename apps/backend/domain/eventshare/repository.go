package eventshare

//go:generate go tool mockgen -destination=generated/repository.go -package=mock github.com/Haya372/ai-trial/backend/domain/eventshare Repository

import (
	"context"
)

type Repository interface {
	Create(ctx context.Context, s EventShare) (EventShare, error)
	// FindByToken hashes the given plain token (via HashToken) and looks up the
	// share whose stored hash matches. It returns the share even if expired so
	// callers can decide via IsExpired(); it returns ErrEventShareNotFound only
	// when no share matches the hash at all.
	FindByToken(ctx context.Context, token string) (EventShare, error)
}
