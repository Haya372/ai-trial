package eventsubscription

import (
	"context"

	"github.com/google/uuid"

	domaineventsubscription "github.com/Haya372/ai-trial/backend/domain/eventsubscription"
)

type DeleteSubscriptionCommand struct {
	repo domaineventsubscription.Repository
}

func NewDeleteSubscriptionCommand(r domaineventsubscription.Repository) *DeleteSubscriptionCommand {
	return &DeleteSubscriptionCommand{repo: r}
}

// Execute removes the user's own addition of a shared event from their
// calendar (SPEC-004 "追加の取り消し"). It only ever touches the subscription
// row; the source event and its share links are never affected.
func (c *DeleteSubscriptionCommand) Execute(ctx context.Context, userID, id uuid.UUID) error {
	sub, err := c.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if sub.UserID() != userID {
		return domaineventsubscription.ErrNotSubscriptionOwner
	}
	return c.repo.Delete(ctx, id)
}
