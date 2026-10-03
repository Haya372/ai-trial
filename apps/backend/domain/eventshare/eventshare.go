package eventshare

import (
	"time"

	"github.com/google/uuid"

	"github.com/Haya372/ai-trial/backend/domain"
)

type EventShare interface {
	ID() uuid.UUID
	EventID() uuid.UUID
	TokenHash() TokenHash
	ExpiresAt() time.Time
	IsExpired() bool
}

type eventShareEntity struct {
	id        uuid.UUID
	eventID   uuid.UUID
	tokenHash TokenHash
	expiresAt time.Time
}

func New(id, eventID uuid.UUID, tokenHash TokenHash, expiresAt time.Time) (EventShare, error) {
	if tokenHash.IsZero() {
		return nil, &domain.ValidationError{Details: []domain.ValidationDetail{
			{Field: "tokenHash", Code: "REQUIRED", Message: "tokenHash is required"},
		}}
	}
	if expiresAt.IsZero() {
		return nil, &domain.ValidationError{Details: []domain.ValidationDetail{
			{Field: "expiresAt", Code: "REQUIRED", Message: "expiresAt is required"},
		}}
	}
	return &eventShareEntity{
		id:        id,
		eventID:   eventID,
		tokenHash: tokenHash,
		expiresAt: expiresAt,
	}, nil
}

func (s *eventShareEntity) ID() uuid.UUID        { return s.id }
func (s *eventShareEntity) EventID() uuid.UUID   { return s.eventID }
func (s *eventShareEntity) TokenHash() TokenHash { return s.tokenHash }
func (s *eventShareEntity) ExpiresAt() time.Time { return s.expiresAt }

func (s *eventShareEntity) IsExpired() bool { return isExpired(time.Now(), s.expiresAt) }

// isExpired treats expiresAt itself as expired (SPEC-004: the boundary is inclusive).
func isExpired(now, expiresAt time.Time) bool { return !now.Before(expiresAt) }
