package tag

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/Haya372/ai-trial/backend/domain"
	domaintag "github.com/Haya372/ai-trial/backend/domain/tag"
	"github.com/Haya372/ai-trial/backend/usecase"
)

type UpdateTagInput struct {
	ID    uuid.UUID
	Name  string
	Color string
}

type UpdateTagCommand struct {
	tagRepo   domaintag.Repository
	txManager usecase.TransactionManager
}

func NewUpdateTagCommand(r domaintag.Repository, tx usecase.TransactionManager) *UpdateTagCommand {
	return &UpdateTagCommand{tagRepo: r, txManager: tx}
}

// Execute returns the updated tag together with its current event count, so
// that the response's eventCount is read in the same transaction as the update.
func (c *UpdateTagCommand) Execute(
	ctx context.Context, userID uuid.UUID, in UpdateTagInput,
) (domaintag.Tag, int, error) {
	var details []domain.ValidationDetail
	name, nameErr := domaintag.NewTagName(in.Name)
	if nameErr != nil {
		details = append(details, toValidationDetail("name", nameErr))
	}
	color, colorErr := domaintag.NewTagColor(in.Color)
	if colorErr != nil {
		details = append(details, toValidationDetail("color", colorErr))
	}
	if len(details) > 0 {
		return nil, 0, &domain.ValidationError{Details: details}
	}

	updated := domaintag.New(in.ID, userID, name, color)
	var eventCount int
	if err := c.txManager.RunInTx(ctx, func(ctx context.Context) error {
		existing, err := c.tagRepo.FindByID(ctx, in.ID)
		if err != nil {
			return err
		}
		// Another user's tag is reported as not found so that its existence
		// does not leak (SPEC-008).
		if existing.UserID() != userID {
			return domaintag.ErrTagNotFound
		}
		if err := c.tagRepo.Update(ctx, updated); err != nil {
			return err
		}
		eventCount, err = c.tagRepo.CountEventsByTagID(ctx, in.ID)
		return err
	}); err != nil {
		return nil, 0, fmt.Errorf("update tag: %w", err)
	}
	return updated, eventCount, nil
}
