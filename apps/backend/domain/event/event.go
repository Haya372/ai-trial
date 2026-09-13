package event

import (
	"time"

	"github.com/google/uuid"

	"github.com/Haya372/ai-trial/backend/domain"
)

type Event interface {
	ID() uuid.UUID
	UserID() uuid.UUID
	Title() string
	Description() string
	StartAt() time.Time
	EndAt() time.Time
}

type eventEntity struct {
	id          uuid.UUID
	userID      uuid.UUID
	title       string
	description string
	startAt     time.Time
	endAt       time.Time
}

func New(id, userID uuid.UUID, title, description string, startAt, endAt time.Time) (Event, error) {
	if title == "" {
		return nil, &domain.ValidationError{Details: []domain.ValidationDetail{
			{Field: "title", Code: "REQUIRED", Message: "title is required"},
		}}
	}
	if endAt.Before(startAt) {
		return nil, &domain.ValidationError{Details: []domain.ValidationDetail{
			{Field: "endAt", Code: "INVALID_DATE_RANGE", Message: "endAt must be after or equal to startAt"},
		}}
	}
	return &eventEntity{
		id:          id,
		userID:      userID,
		title:       title,
		description: description,
		startAt:     startAt,
		endAt:       endAt,
	}, nil
}

func (e *eventEntity) ID() uuid.UUID       { return e.id }
func (e *eventEntity) UserID() uuid.UUID   { return e.userID }
func (e *eventEntity) Title() string       { return e.title }
func (e *eventEntity) Description() string { return e.description }
func (e *eventEntity) StartAt() time.Time  { return e.startAt }
func (e *eventEntity) EndAt() time.Time    { return e.endAt }
