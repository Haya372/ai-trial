package eventsubscription

import (
	"context"
	"log/slog"

	"github.com/google/uuid"

	domaineventsubscription "github.com/Haya372/ai-trial/backend/domain/eventsubscription"
)

type DeleteSubscriptionCommand struct {
	repo   domaineventsubscription.Repository
	logger *slog.Logger
}

func NewDeleteSubscriptionCommand(
	r domaineventsubscription.Repository, logger *slog.Logger,
) *DeleteSubscriptionCommand {
	return &DeleteSubscriptionCommand{repo: r, logger: logger}
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
		c.logger.Warn("attempted to delete subscription owned by another user",
			"subscription_id", id, "user_id", userID, "owner_id", sub.UserID())
		return domaineventsubscription.ErrNotSubscriptionOwner
	}
	return c.repo.Delete(ctx, id)
}
