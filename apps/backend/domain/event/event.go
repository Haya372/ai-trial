package event

import (
	"slices"
	"time"

	"github.com/google/uuid"

	"github.com/Haya372/ai-trial/backend/domain"
)

// Event is the aggregate root, so its interface grows with the aggregate's
// fields; splitting it would not let any caller depend on less.
//
//nolint:interfacebloat // one accessor per field of the aggregate
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
	// WithContent returns a copy with the user-editable content replaced and
	// the same identity, tags and series. Rebuilding an updated event with
	// New would silently drop the tags and the series.
	WithContent(title, description string, startAt, endAt time.Time, location, url string) (Event, error)
	// SeriesID returns the recurring series the event is an occurrence of;
	// ok is false for a one-off event.
	SeriesID() (id uuid.UUID, ok bool)
	// WithSeriesID returns a copy of the event as an occurrence of the series.
	WithSeriesID(seriesID uuid.UUID) Event
	// DetachFromSeries returns a copy of the event as a one-off event, so later
	// changes to the series no longer reach it.
	DetachFromSeries() Event
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
	seriesID    uuid.UUID
}

func New(
	id, userID uuid.UUID,
	title, description string,
	startAt, endAt time.Time,
	location, url string,
) (Event, error) {
	if err := validateContent(title, startAt, endAt); err != nil {
		return nil, err
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

func (e *eventEntity) WithContent(
	title, description string, startAt, endAt time.Time, location, url string,
) (Event, error) {
	if err := validateContent(title, startAt, endAt); err != nil {
		return nil, err
	}
	copied := *e
	copied.title = title
	copied.description = description
	copied.startAt = startAt
	copied.endAt = endAt
	copied.location = location
	copied.url = url
	return &copied, nil
}

func validateContent(title string, startAt, endAt time.Time) error {
	if title == "" {
		return &domain.ValidationError{Details: []domain.ValidationDetail{
			{Field: "title", Code: "REQUIRED", Message: "title is required"},
		}}
	}
	if !endAt.After(startAt) {
		return &domain.ValidationError{Details: []domain.ValidationDetail{
			{Field: "endAt", Code: "INVALID_DATE_RANGE", Message: "endAt must be strictly after startAt"},
		}}
	}
	return nil
}

func (e *eventEntity) Tags() []EventTag { return slices.Clone(e.tags) }

func (e *eventEntity) WithTags(tags []EventTag) (Event, error) {
	if err := validateTags(tags); err != nil {
		return nil, err
	}
	copied := *e
	copied.tags = sortedByPosition(tags)
	return &copied, nil
}

// A series ID is a random UUID, so uuid.Nil never names a real series and
// can stand for "not in a series".
func (e *eventEntity) SeriesID() (uuid.UUID, bool) {
	return e.seriesID, e.seriesID != uuid.Nil
}

func (e *eventEntity) WithSeriesID(seriesID uuid.UUID) Event {
	copied := *e
	copied.seriesID = seriesID
	return &copied
}

func (e *eventEntity) DetachFromSeries() Event {
	return e.WithSeriesID(uuid.Nil)
}
