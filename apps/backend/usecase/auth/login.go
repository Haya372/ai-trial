package auth

import (
	"context"
	"errors"
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
		return nil, ErrInvalidCredentials
	}

	u, err := c.userRepo.FindByEmail(ctx, email)
	if errors.Is(err, user.ErrUserNotFound) {
		return nil, ErrInvalidCredentials
	}
	if err != nil {
		return nil, fmt.Errorf("find user: %w", err)
	}

	password, err := user.NewPassword(in.Password)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	if err := u.ComparePassword(password); err != nil {
		return nil, ErrInvalidCredentials
	}

	sess, err := c.sessionRepo.Create(ctx, u.ID(), time.Now().Add(sessionExpiry))
	if err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}

	return &AuthOutput{User: u, SessionID: sess.ID()}, nil
}
