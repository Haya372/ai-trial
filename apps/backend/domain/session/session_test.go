package session_test

import (
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/Haya372/ai-trial/backend/domain/session"
)

func TestNewSession_returnsAccessibleFields(t *testing.T) {
	id := uuid.New()
	userID := uuid.New()
	expiresAt := time.Now().Add(24 * time.Hour).Truncate(time.Second)

	s := session.New(id, userID, expiresAt)

	if s.ID() != id {
		t.Errorf("ID() = %v, want %v", s.ID(), id)
	}
	if s.UserID() != userID {
		t.Errorf("UserID() = %v, want %v", s.UserID(), userID)
	}
	if !s.ExpiresAt().Equal(expiresAt) {
		t.Errorf("ExpiresAt() = %v, want %v", s.ExpiresAt(), expiresAt)
	}
}
