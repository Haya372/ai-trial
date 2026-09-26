package eventshare_test

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/Haya372/ai-trial/backend/domain"
	"github.com/Haya372/ai-trial/backend/domain/eventshare"
)

func TestNew_ValidEventShare(t *testing.T) {
	id := uuid.New()
	eventID := uuid.New()
	tokenHash := "hash-value"
	expiresAt := time.Now().Add(24 * time.Hour)

	s, err := eventshare.New(id, eventID, tokenHash, expiresAt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if s.ID() != id {
		t.Errorf("ID() = %v, want %v", s.ID(), id)
	}
	if s.EventID() != eventID {
		t.Errorf("EventID() = %v, want %v", s.EventID(), eventID)
	}
	if s.TokenHash() != tokenHash {
		t.Errorf("TokenHash() = %v, want %v", s.TokenHash(), tokenHash)
	}
	if !s.ExpiresAt().Equal(expiresAt) {
		t.Errorf("ExpiresAt() = %v, want %v", s.ExpiresAt(), expiresAt)
	}
}

func TestNew_EmptyTokenHash(t *testing.T) {
	_, err := eventshare.New(uuid.New(), uuid.New(), "", time.Now().Add(time.Hour))

	var ve *domain.ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected *domain.ValidationError, got %T (%v)", err, err)
	}
}

func TestNew_ZeroExpiresAt(t *testing.T) {
	_, err := eventshare.New(uuid.New(), uuid.New(), "hash-value", time.Time{})

	var ve *domain.ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected *domain.ValidationError, got %T (%v)", err, err)
	}
}

func TestIsExpired_notExpired(t *testing.T) {
	s, err := eventshare.New(uuid.New(), uuid.New(), "hash-value", time.Now().Add(time.Hour))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if s.IsExpired() {
		t.Error("IsExpired() = true, want false for future expiry")
	}
}

func TestIsExpired_expired(t *testing.T) {
	s, err := eventshare.New(uuid.New(), uuid.New(), "hash-value", time.Now().Add(-time.Second))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !s.IsExpired() {
		t.Error("IsExpired() = false, want true for past expiry")
	}
}
