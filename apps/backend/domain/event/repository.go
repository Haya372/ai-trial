package event

//go:generate go tool mockgen -destination=generated/repository.go -package=mock \
//   github.com/Haya372/ai-trial/backend/domain/event Repository

import "context"

type Repository interface {
	Create(ctx context.Context, e Event) (Event, error)
}
