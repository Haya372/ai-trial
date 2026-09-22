package event

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"

	domainevent "github.com/Haya372/ai-trial/backend/domain/event"
)

type DeleteEventCommand struct {
	repo   domainevent.Repository
	logger *slog.Logger
}

func NewDeleteEventCommand(r domainevent.Repository, logger *slog.Logger) *DeleteEventCommand {
	return &DeleteEventCommand{repo: r, logger: logger}
}

func (c *DeleteEventCommand) Execute(ctx context.Context, userID uuid.UUID, id uuid.UUID) error {
	existing, err := c.repo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("find event: %w", err)
	}
	if existing.UserID() != userID {
		c.logger.Warn("attempted to delete event owned by another user",
			"event_id", id, "user_id", userID, "owner_id", existing.UserID())
		return domainevent.ErrEventNotFound
	}

	if err := c.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete event: %w", err)
	}
	return nil
}
