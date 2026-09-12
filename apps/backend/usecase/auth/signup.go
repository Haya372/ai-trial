package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode"

	"github.com/Haya372/ai-trial/backend/domain/session"
	"github.com/Haya372/ai-trial/backend/domain/user"
)

const (
	maxDisplayName = 50
	fieldPassword  = "password"
)

type SignupInput struct {
	Email       string
	Password    string
	DisplayName string
}

type SignupCommand struct {
	userRepo    user.Repository
	sessionRepo session.Repository
}

func NewSignupCommand(ur user.Repository, sr session.Repository) *SignupCommand {
	return &SignupCommand{userRepo: ur, sessionRepo: sr}
}

func (c *SignupCommand) Execute(ctx context.Context, in SignupInput) (*AuthOutput, error) {
	var details []ValidationDetail

	email, emailErr := user.NewEmail(in.Email)
	if emailErr != nil {
		details = append(details, ValidationDetail{
			Field:   "email",
			Code:    "INVALID_FORMAT",
			Message: "Invalid email format",
		})
	}

	password, pwErr := user.NewPassword(in.Password)
	if pwErr != nil {
		details = append(details, passwordErrorDetail(pwErr))
	} else if d := checkPasswordComplexity(in.Password); d != nil {
		details = append(details, *d)
	}

	if len([]rune(in.DisplayName)) > maxDisplayName {
		details = append(details, ValidationDetail{
			Field:   "displayName",
			Code:    "TOO_LONG",
			Message: fmt.Sprintf("Display name must be at most %d characters", maxDisplayName),
		})
	}

	if len(details) > 0 {
		return nil, &ValidationError{Details: details}
	}

	displayName := in.DisplayName
	if displayName == "" {
		displayName = strings.SplitN(in.Email, "@", 2)[0]
	}

	u, err := c.userRepo.Create(ctx, email, displayName, password)
	if err != nil {
		return nil, err
	}

	sess, err := c.sessionRepo.Create(ctx, u.ID(), time.Now().Add(sessionExpiry))
	if err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}

	return &AuthOutput{User: u, SessionID: sess.ID()}, nil
}

func passwordErrorDetail(err error) ValidationDetail {
	switch {
	case errors.Is(err, user.ErrPasswordTooShort):
		return ValidationDetail{Field: fieldPassword, Code: "TOO_SHORT", Message: "Password must be at least 8 characters"}
	case errors.Is(err, user.ErrPasswordTooLong):
		return ValidationDetail{Field: fieldPassword, Code: "TOO_LONG", Message: "Password must be at most 72 characters"}
	case errors.Is(err, user.ErrPasswordNotASCII):
		return ValidationDetail{
			Field:   fieldPassword,
			Code:    "INVALID_CHARACTER",
			Message: "Password must contain only ASCII printable characters",
		}
	default:
		return ValidationDetail{Field: fieldPassword, Code: "INVALID", Message: err.Error()}
	}
}

func checkPasswordComplexity(p string) *ValidationDetail {
	var hasUpper, hasLower, hasDigit, hasSymbol bool
	for _, r := range p {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsDigit(r):
			hasDigit = true
		default:
			hasSymbol = true
		}
	}
	if !hasUpper || !hasLower || !hasDigit || !hasSymbol {
		return &ValidationDetail{
			Field:   fieldPassword,
			Code:    "INSUFFICIENT_COMPLEXITY",
			Message: "Password must contain uppercase, lowercase, digit, and symbol",
		}
	}
	return nil
}
