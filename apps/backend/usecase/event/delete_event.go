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
	if _, err := domainevent.FindOwned(ctx, c.repo, c.logger, id, userID); err != nil {
		return err
	}

	if err := c.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete event: %w", err)
	}
	return nil
}
