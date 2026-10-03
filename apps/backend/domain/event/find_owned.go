package event

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
)

// FindOwned finds the event with the given id and verifies that userID is
// its owner. A mismatch is reported as ErrEventNotFound, the same error
// returned when the event does not exist at all, so callers never leak an
// event's existence to a non-owner.
func FindOwned(ctx context.Context, repo Repository, logger *slog.Logger, id, userID uuid.UUID) (Event, error) {
	ev, err := repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("find event: %w", err)
	}
	if ev.UserID() != userID {
		logger.Warn("attempted to access event owned by another user",
			"event_id", id, "user_id", userID, "owner_id", ev.UserID())
		return nil, ErrEventNotFound
	}
	return ev, nil
}
