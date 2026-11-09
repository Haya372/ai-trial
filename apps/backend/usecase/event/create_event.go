package event

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	domainevent "github.com/Haya372/ai-trial/backend/domain/event"
	"github.com/Haya372/ai-trial/backend/usecase"
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
	txManager usecase.TransactionManager
}

func NewCreateEventCommand(r domainevent.Repository, tx usecase.TransactionManager) *CreateEventCommand {
	return &CreateEventCommand{eventRepo: r, txManager: tx}
}

func (c *CreateEventCommand) Execute(
	ctx context.Context, userID uuid.UUID, in CreateEventInput,
) (domainevent.Event, error) {
	id := uuid.New()
	e, err := domainevent.New(id, userID, in.Title, in.Description, in.StartAt, in.EndAt, in.Location, in.URL)
	if err != nil {
		return nil, err
	}

	// Create writes the event and its tags as separate rows (ADR-024).
	var saved domainevent.Event
	if err := c.txManager.RunInTx(ctx, func(ctx context.Context) error {
		saved, err = c.eventRepo.Create(ctx, e)
		return err
	}); err != nil {
		return nil, fmt.Errorf("create event: %w", err)
	}
	return saved, nil
}
