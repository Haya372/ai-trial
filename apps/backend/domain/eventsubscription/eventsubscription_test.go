package eventsubscription_test

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/Haya372/ai-trial/backend/domain"
	"github.com/Haya372/ai-trial/backend/domain/eventsubscription"
)

func TestNew_ValidEventSubscription(t *testing.T) {
	id := uuid.New()
	eventID := uuid.New()
	userID := uuid.New()
	createdAt := time.Now()

	s, err := eventsubscription.New(id, eventID, userID, createdAt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

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

func TestNew_NilEventID(t *testing.T) {
	_, err := eventsubscription.New(uuid.New(), uuid.Nil, uuid.New(), time.Now())

	var ve *domain.ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected *domain.ValidationError, got %T (%v)", err, err)
	}
}

func TestNew_NilUserID(t *testing.T) {
	_, err := eventsubscription.New(uuid.New(), uuid.New(), uuid.Nil, time.Now())

	var ve *domain.ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected *domain.ValidationError, got %T (%v)", err, err)
	}
}
