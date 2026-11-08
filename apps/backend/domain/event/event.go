package event

import (
	"slices"
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
	Location() string
	URL() string
	// Tags returns the attached tags sorted by ascending position.
	Tags() []EventTag
	// WithTags returns a copy of the event with its tags replaced by tags.
	WithTags(tags []EventTag) (Event, error)
}

type eventEntity struct {
	id          uuid.UUID
	userID      uuid.UUID
	title       string
	description string
	startAt     time.Time
	endAt       time.Time
	location    string
	url         string
	tags        []EventTag
}

func New(
	id, userID uuid.UUID,
	title, description string,
	startAt, endAt time.Time,
	location, url string,
) (Event, error) {
	if title == "" {
		return nil, &domain.ValidationError{Details: []domain.ValidationDetail{
			{Field: "title", Code: "REQUIRED", Message: "title is required"},
		}}
	}
	if !endAt.After(startAt) {
		return nil, &domain.ValidationError{Details: []domain.ValidationDetail{
			{Field: "endAt", Code: "INVALID_DATE_RANGE", Message: "endAt must be strictly after startAt"},
		}}
	}
	return &eventEntity{
		id:          id,
		userID:      userID,
		title:       title,
		description: description,
		startAt:     startAt,
		endAt:       endAt,
		location:    location,
		url:         url,
	}, nil
}

func (e *eventEntity) ID() uuid.UUID       { return e.id }
func (e *eventEntity) UserID() uuid.UUID   { return e.userID }
func (e *eventEntity) Title() string       { return e.title }
func (e *eventEntity) Description() string { return e.description }
func (e *eventEntity) StartAt() time.Time  { return e.startAt }
func (e *eventEntity) EndAt() time.Time    { return e.endAt }
func (e *eventEntity) Location() string    { return e.location }
func (e *eventEntity) URL() string         { return e.url }

func (e *eventEntity) Tags() []EventTag { return slices.Clone(e.tags) }

func (e *eventEntity) WithTags(tags []EventTag) (Event, error) {
	if err := validateTags(tags); err != nil {
		return nil, err
	}
	copied := *e
	copied.tags = sortedByPosition(tags)
	return &copied, nil
}
