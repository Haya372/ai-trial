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
		return nil, fmt.Errorf("validate email: %w", err)
	}

	u, findErr := c.userRepo.FindByEmail(ctx, email)
	notFound := errors.Is(findErr, user.ErrUserNotFound)
	if findErr != nil && !notFound {
		return nil, fmt.Errorf("find user: %w", findErr)
	}

	// NOTE: NewPassword calls bcrypt.GenerateFromPassword, which is expensive.
	// The generated hash is discarded; only the plain field is used for ComparePassword.
	// Consider a lightweight validator when this becomes a bottleneck.
	//
	// This runs even when the user was not found above, so the bcrypt cost paid
	// here (plus the dummy comparison below) matches the found-user path and a
	// nonexistent email cannot be distinguished from a wrong password by timing.
	password, err := user.NewPassword(in.Password)
	if err != nil {
		return nil, fmt.Errorf("validate password: %w", err)
	}

	if notFound {
		_ = user.CompareDummyPassword(password)
		return nil, fmt.Errorf("find user: %w", findErr)
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
