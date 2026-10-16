package auth

import (
	"time"

	"github.com/Haya372/ai-trial/backend/domain/session"
	"github.com/Haya372/ai-trial/backend/domain/user"
)

const sessionExpiry = 30 * 24 * time.Hour

type AuthOutput struct {
	User         user.User
	SessionToken session.Token
}
