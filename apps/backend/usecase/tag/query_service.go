package tag

import (
	"context"

	"github.com/google/uuid"
)

type QueryService interface {
	// ListByUserID returns the user's tags in creation order.
	ListByUserID(ctx context.Context, userID uuid.UUID) ([]TagReadModel, error)
}
