package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Haya372/ai-trial/backend/domain"
	"github.com/Haya372/ai-trial/backend/domain/session"
	"github.com/Haya372/ai-trial/backend/domain/user"
	"github.com/Haya372/ai-trial/backend/usecase"
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
	txManager   usecase.TransactionManager
}

func NewSignupCommand(ur user.Repository, sr session.Repository, tx usecase.TransactionManager) *SignupCommand {
	return &SignupCommand{userRepo: ur, sessionRepo: sr, txManager: tx}
}

func (c *SignupCommand) Execute(ctx context.Context, in SignupInput) (*AuthOutput, error) {
	var details []domain.ValidationDetail

	email, emailErr := user.NewEmail(in.Email)
	if emailErr != nil {
		details = append(details, toValidationDetail("email", emailErr))
	}

	password, pwErr := user.NewPassword(in.Password)
	if pwErr != nil {
		details = append(details, toValidationDetail(fieldPassword, pwErr))
	}

	if len([]rune(in.DisplayName)) > maxDisplayName {
		details = append(details, domain.ValidationDetail{
			Field:   "displayName",
			Code:    "TOO_LONG",
			Message: fmt.Sprintf("Display name must be at most %d characters", maxDisplayName),
		})
	}

	if len(details) > 0 {
		return nil, &domain.ValidationError{Details: details}
	}

	displayName := in.DisplayName
	if displayName == "" {
		displayName = strings.SplitN(in.Email, "@", 2)[0]
	}

	var out *AuthOutput
	if err := c.txManager.RunInTx(ctx, func(ctx context.Context) error {
		u, err := c.userRepo.Create(ctx, email, displayName, password)
		if err != nil {
			return err
		}
		sess, err := c.sessionRepo.Create(ctx, u.ID(), time.Now().Add(sessionExpiry))
		if err != nil {
			return fmt.Errorf("create session: %w", err)
		}
		out = &AuthOutput{User: u, SessionID: sess.ID()}
		return nil
	}); err != nil {
		return nil, err
	}

	return out, nil
}

func toValidationDetail(field string, err error) domain.ValidationDetail {
	var domErr *domain.DomainError
	if errors.As(err, &domErr) {
		return domain.ValidationDetail{Field: field, Code: domErr.Code, Message: domErr.Message}
	}
	return domain.ValidationDetail{Field: field, Code: "INVALID", Message: err.Error()}
}
