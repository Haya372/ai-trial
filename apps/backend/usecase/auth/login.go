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
	// hash, so only one bcrypt operation (inside authenticate below) runs per
	// login. This runs even when the user was not found above, so that a
	// malformed password is rejected the same way regardless of whether the
	// email exists.
	password, pwErr := user.NewLoginPassword(in.Password)
	if pwErr != nil {
		// A malformed password can never match a real, policy-compliant hash,
		// so treat it the same as a wrong password rather than exposing
		// password-policy details to an unauthenticated caller.
		return nil, fmt.Errorf("compare password: %w", user.ErrPasswordMismatch)
	}

	authenticated, err := authenticate(notFound, u, password, findErr)
	if err != nil {
		return nil, err
	}

	sess, err := c.sessionRepo.Create(ctx, authenticated.ID(), time.Now().Add(sessionExpiry))
	if err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}

	return &AuthOutput{User: authenticated, SessionID: sess.ID()}, nil
}

// authenticate is the single call site that runs the bcrypt comparison: the
// real hash when the user was found, or a fixed dummy hash otherwise, so the
// found and not-found paths always pay the same bcrypt cost and a
// nonexistent email cannot be distinguished from a wrong password by timing.
//
// It also re-checks notFound after a successful comparison before returning
// u. A successful dummy-hash match is cryptographically infeasible for any
// attacker-supplied input, but u is nil on the not-found path, so this guards
// explicitly against touching it instead of relying on that infeasibility alone.
func authenticate(notFound bool, u user.User, password user.LoginPassword, findErr error) (user.User, error) {
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
	if notFound {
		return nil, fmt.Errorf("find user: %w", findErr)
	}
	return u, nil
}
