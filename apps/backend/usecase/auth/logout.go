package auth

import (
	"context"
	"fmt"

	"github.com/Haya372/ai-trial/backend/domain/session"
)

type LogoutCommand struct {
	sessionRepo session.Repository
}

func NewLogoutCommand(sr session.Repository) *LogoutCommand {
	return &LogoutCommand{sessionRepo: sr}
}

func (c *LogoutCommand) Execute(ctx context.Context, token session.Token) error {
	if err := c.sessionRepo.DeleteByTokenHash(ctx, token.Hash()); err != nil {
		return fmt.Errorf("delete session: %w", err)
	}
	return nil
}
