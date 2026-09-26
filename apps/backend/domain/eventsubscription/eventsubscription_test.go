package eventsubscription_test

import (
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/Haya372/ai-trial/backend/domain/eventsubscription"
)

func TestNew_ValidEventSubscription(t *testing.T) {
	id := uuid.New()
	eventID := uuid.New()
	userID := uuid.New()
	createdAt := time.Now()

	s := eventsubscription.New(id, eventID, userID, createdAt)

	if s.ID() != id {
		t.Errorf("ID() = %v, want %v", s.ID(), id)
	}
	if s.EventID() != eventID {
		t.Errorf("EventID() = %v, want %v", s.EventID(), eventID)
	}
	if s.UserID() != userID {
		t.Errorf("UserID() = %v, want %v", s.UserID(), userID)
	}
	if !s.CreatedAt().Equal(createdAt) {
		t.Errorf("CreatedAt() = %v, want %v", s.CreatedAt(), createdAt)
	}
}
