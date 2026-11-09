package event

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"

	domainevent "github.com/Haya372/ai-trial/backend/domain/event"
	"github.com/Haya372/ai-trial/backend/usecase"
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
	repo      domainevent.Repository
	txManager usecase.TransactionManager
	logger    *slog.Logger
}

func NewUpdateEventCommand(
	r domainevent.Repository, tx usecase.TransactionManager, logger *slog.Logger,
) *UpdateEventCommand {
	return &UpdateEventCommand{repo: r, txManager: tx, logger: logger}
}

func (c *UpdateEventCommand) Execute(
	ctx context.Context, userID uuid.UUID, in UpdateEventInput,
) (domainevent.Event, error) {
	// Read and write in one transaction so the carried-over tags are the ones
	// still stored when Update writes the event and its tags (ADR-024).
	var updated domainevent.Event
	err := c.txManager.RunInTx(ctx, func(ctx context.Context) error {
		existing, err := domainevent.FindOwned(ctx, c.repo, c.logger, in.ID, userID)
		if err != nil {
			return err
		}

		updated, err = domainevent.New(
			in.ID, userID, in.Title, in.Description, in.StartAt, in.EndAt, in.Location, in.URL,
		)
		if err != nil {
			return err
		}
		// Update replaces the stored tags with updated.Tags(), so carry them
		// over or editing the event would drop its tags.
		updated, err = updated.WithTags(existing.Tags())
		if err != nil {
			return err
		}

		if err := c.repo.Update(ctx, updated); err != nil {
			return fmt.Errorf("update event: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return updated, nil
}
