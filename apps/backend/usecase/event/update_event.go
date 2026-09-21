package event

import (
	"context"
	"fmt"
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
	repo domainevent.Repository
}

func NewUpdateEventCommand(r domainevent.Repository) *UpdateEventCommand {
	return &UpdateEventCommand{repo: r}
}

func (c *UpdateEventCommand) Execute(
	ctx context.Context, userID uuid.UUID, in UpdateEventInput,
) (domainevent.Event, error) {
	existing, err := c.repo.FindByID(ctx, in.ID)
	if err != nil {
		return nil, fmt.Errorf("find event: %w", err)
	}
	if existing.UserID() != userID {
		// Reported the same as a missing event to avoid leaking event
		// existence to users who don't own it.
		return nil, domainevent.ErrEventNotFound
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
