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
	return findOwned(ctx, repo.FindByID, logger, id, userID)
}

// FindOwnedForUpdate is FindOwned with the event row locked, for a caller
// that writes back what it read (see Repository.FindByIDForUpdate).
func FindOwnedForUpdate(
	ctx context.Context, repo Repository, logger *slog.Logger, id, userID uuid.UUID,
) (Event, error) {
	return findOwned(ctx, repo.FindByIDForUpdate, logger, id, userID)
}

func findOwned(
	ctx context.Context,
	find func(ctx context.Context, id uuid.UUID) (Event, error),
	logger *slog.Logger,
	id, userID uuid.UUID,
) (Event, error) {
	ev, err := find(ctx, id)
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
