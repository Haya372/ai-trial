package auth

import (
	"errors"
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
