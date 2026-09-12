package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/Haya372/ai-trial/backend/domain/user"
)

const sessionExpiry = 30 * 24 * time.Hour

var ErrInvalidCredentials = errors.New("invalid email or password")

type AuthOutput struct {
	User      user.User
	SessionID uuid.UUID
}

type ValidationDetail struct {
	Field   string
	Code    string
	Message string
}

type ValidationError struct {
	Details []ValidationDetail
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error: %d field(s) invalid", len(e.Details))
}
