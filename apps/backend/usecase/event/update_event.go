package event

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"

	domainevent "github.com/Haya372/ai-trial/backend/domain/event"
)

type UpdateEventInput struct {
	ID          uuid.UUID
	Title       string
	Description string
	StartAt     time.Time
	EndAt       time.Time
	Location    string
	URL         string
}

type UpdateEventCommand struct {
	repo   domainevent.Repository
	logger *slog.Logger
}

func NewUpdateEventCommand(r domainevent.Repository, logger *slog.Logger) *UpdateEventCommand {
	return &UpdateEventCommand{repo: r, logger: logger}
}

func (c *UpdateEventCommand) Execute(
	ctx context.Context, userID uuid.UUID, in UpdateEventInput,
) (domainevent.Event, error) {
	if _, err := domainevent.FindOwned(ctx, c.repo, c.logger, in.ID, userID); err != nil {
		return nil, err
	}

	updated, err := domainevent.New(in.ID, userID, in.Title, in.Description, in.StartAt, in.EndAt, in.Location, in.URL)
	if err != nil {
		return nil, err
	}

	if err := c.repo.Update(ctx, updated); err != nil {
		return nil, fmt.Errorf("update event: %w", err)
	}
	return updated, nil
}
