package tag

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	domaintag "github.com/Haya372/ai-trial/backend/domain/tag"
	"github.com/Haya372/ai-trial/backend/usecase"
)

type DeleteTagCommand struct {
	tagRepo   domaintag.Repository
	txManager usecase.TransactionManager
}

func NewDeleteTagCommand(r domaintag.Repository, tx usecase.TransactionManager) *DeleteTagCommand {
	return &DeleteTagCommand{tagRepo: r, txManager: tx}
}

// Execute deletes the tag; its assignments to events are removed by the
// event_tags ON DELETE CASCADE, while the events themselves are kept.
func (c *DeleteTagCommand) Execute(ctx context.Context, userID uuid.UUID, id uuid.UUID) error {
	if err := c.txManager.RunInTx(ctx, func(ctx context.Context) error {
		existing, err := c.tagRepo.FindByID(ctx, id)
		if err != nil {
			return err
		}
		// Another user's tag is reported as not found so that its existence
		// does not leak (SPEC-008).
		if existing.UserID() != userID {
			return domaintag.ErrTagNotFound
		}
		return c.tagRepo.Delete(ctx, id)
	}); err != nil {
		return fmt.Errorf("delete tag: %w", err)
	}
	return nil
}
