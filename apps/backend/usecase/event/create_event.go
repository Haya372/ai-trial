package event

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	domainevent "github.com/Haya372/ai-trial/backend/domain/event"
)

type CreateEventInput struct {
	Title       string
	Description string
	StartAt     time.Time
	EndAt       time.Time
	Location    string
	URL         string
}

type CreateEventCommand struct {
	eventRepo domainevent.Repository
}

func NewCreateEventCommand(r domainevent.Repository) *CreateEventCommand {
	return &CreateEventCommand{eventRepo: r}
}

func (c *CreateEventCommand) Execute(
	ctx context.Context, userID uuid.UUID, in CreateEventInput,
) (domainevent.Event, error) {
	id := uuid.New()
	e, err := domainevent.New(id, userID, in.Title, in.Description, in.StartAt, in.EndAt, in.Location, in.URL)
	if err != nil {
		return nil, err
	}

	saved, err := c.eventRepo.Create(ctx, e)
	if err != nil {
		return nil, fmt.Errorf("create event: %w", err)
	}
	return saved, nil
}
