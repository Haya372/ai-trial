package eventsubscription

import (
	"time"

	"github.com/google/uuid"
)

type EventSubscription interface {
	ID() uuid.UUID
	EventID() uuid.UUID
	UserID() uuid.UUID
	CreatedAt() time.Time
}

type eventSubscriptionEntity struct {
	id        uuid.UUID
	eventID   uuid.UUID
	userID    uuid.UUID
	createdAt time.Time
}

func New(id, eventID, userID uuid.UUID, createdAt time.Time) EventSubscription {
	return &eventSubscriptionEntity{
		id:        id,
		eventID:   eventID,
		userID:    userID,
		createdAt: createdAt,
	}
}

func (s *eventSubscriptionEntity) ID() uuid.UUID        { return s.id }
func (s *eventSubscriptionEntity) EventID() uuid.UUID   { return s.eventID }
func (s *eventSubscriptionEntity) UserID() uuid.UUID    { return s.userID }
func (s *eventSubscriptionEntity) CreatedAt() time.Time { return s.createdAt }
