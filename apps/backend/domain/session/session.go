package session

import (
	"time"

	"github.com/google/uuid"
)

type Session interface {
	ID() uuid.UUID
	UserID() uuid.UUID
	ExpiresAt() time.Time
	IsExpired() bool
}

type sessionEntity struct {
	id        uuid.UUID
	userID    uuid.UUID
	expiresAt time.Time
}

func New(id, userID uuid.UUID, expiresAt time.Time) Session {
	return &sessionEntity{id: id, userID: userID, expiresAt: expiresAt}
}

func (s *sessionEntity) ID() uuid.UUID        { return s.id }
func (s *sessionEntity) UserID() uuid.UUID    { return s.userID }
func (s *sessionEntity) ExpiresAt() time.Time { return s.expiresAt }
func (s *sessionEntity) IsExpired() bool      { return time.Now().After(s.expiresAt) }
