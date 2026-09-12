package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/Haya372/ai-trial/backend/domain/session"
	"github.com/Haya372/ai-trial/backend/domain/user"
)

type LoginInput struct {
	Email    string
	Password string
}

type LoginCommand struct {
	userRepo    user.Repository
	sessionRepo session.Repository
}

func NewLoginCommand(ur user.Repository, sr session.Repository) *LoginCommand {
	return &LoginCommand{userRepo: ur, sessionRepo: sr}
}

func (c *LoginCommand) Execute(ctx context.Context, in LoginInput) (*AuthOutput, error) {
	email, err := user.NewEmail(in.Email)
	if err != nil {
		return nil, fmt.Errorf("validate email: %w", err)
	}

	u, err := c.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("find user: %w", err)
	}

	// NOTE: NewPassword calls bcrypt.GenerateFromPassword, which is expensive.
	// The generated hash is discarded; only the plain field is used for ComparePassword.
	// Consider a lightweight validator when this becomes a bottleneck.
	password, err := user.NewPassword(in.Password)
	if err != nil {
		return nil, fmt.Errorf("validate password: %w", err)
	}

	if err := u.ComparePassword(password); err != nil {
		return nil, fmt.Errorf("compare password: %w", err)
	}

	sess, err := c.sessionRepo.Create(ctx, u.ID(), time.Now().Add(sessionExpiry))
	if err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}

	return &AuthOutput{User: u, SessionID: sess.ID()}, nil
}
