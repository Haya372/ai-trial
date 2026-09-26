package eventsubscription

import (
	"time"

	"github.com/google/uuid"

	"github.com/Haya372/ai-trial/backend/domain"
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

func New(id, eventID, userID uuid.UUID, createdAt time.Time) (EventSubscription, error) {
	if eventID == uuid.Nil {
		return nil, &domain.ValidationError{Details: []domain.ValidationDetail{
			{Field: "eventID", Code: "REQUIRED", Message: "eventID is required"},
		}}
	}
	if userID == uuid.Nil {
		return nil, &domain.ValidationError{Details: []domain.ValidationDetail{
			{Field: "userID", Code: "REQUIRED", Message: "userID is required"},
		}}
	}
	return &eventSubscriptionEntity{
		id:        id,
		eventID:   eventID,
		userID:    userID,
		createdAt: createdAt,
	}, nil
}

func (s *eventSubscriptionEntity) ID() uuid.UUID        { return s.id }
func (s *eventSubscriptionEntity) EventID() uuid.UUID   { return s.eventID }
func (s *eventSubscriptionEntity) UserID() uuid.UUID    { return s.userID }
func (s *eventSubscriptionEntity) CreatedAt() time.Time { return s.createdAt }
