package session

import (
	"time"

	"github.com/google/uuid"
)

type Session struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	ExpiresAt time.Time
}

type WithUser struct {
	Email       string
	DisplayName string

	Session
}
