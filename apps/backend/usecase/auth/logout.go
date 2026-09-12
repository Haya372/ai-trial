package auth

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/Haya372/ai-trial/backend/domain/session"
)

type LogoutCommand struct {
	sessionRepo session.Repository
}

func NewLogoutCommand(sr session.Repository) *LogoutCommand {
	return &LogoutCommand{sessionRepo: sr}
}

func (c *LogoutCommand) Execute(ctx context.Context, sessionID uuid.UUID) error {
	if err := c.sessionRepo.Delete(ctx, sessionID); err != nil {
		return fmt.Errorf("delete session: %w", err)
	}
	return nil
}
