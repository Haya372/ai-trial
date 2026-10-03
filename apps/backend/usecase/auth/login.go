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

	// NewLoginPassword validates the password format without generating a bcrypt
	// hash, so only one bcrypt operation (the comparison below) runs per login.
	// This runs even when the user was not found above so that the validation
	// cost paid here plus the comparison below makes the not-found path
	// indistinguishable from a wrong-password path by timing.
	password, pwErr := user.NewLoginPassword(in.Password)
	if pwErr != nil {
		// A malformed password can never match a real, policy-compliant hash,
		// so treat it the same as a wrong password rather than exposing
		// password-policy details to an unauthenticated caller.
		return nil, fmt.Errorf("compare password: %w", user.ErrPasswordMismatch)
	}

	// Compare against the real user's hash when found, or a fixed dummy hash
	// otherwise, through a single call site so both paths always pay the same
	// bcrypt cost.
	compare := user.CompareDummyPassword
	if !notFound {
		compare = u.ComparePassword
	}
	if err := compare(password); err != nil {
		if notFound {
			return nil, fmt.Errorf("find user: %w", findErr)
		}
		return nil, fmt.Errorf("compare password: %w", err)
	}

	sess, err := c.sessionRepo.Create(ctx, u.ID(), time.Now().Add(sessionExpiry))
	if err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}

	return &AuthOutput{User: u, SessionID: sess.ID()}, nil
}
