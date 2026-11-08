package event

import (
	"fmt"
	"slices"

	"github.com/google/uuid"
)

// MaxTagsPerEvent is the number of tags one event may carry (SPEC-008).
const MaxTagsPerEvent = 5

// EventTag is one tag attached to an event. Position orders the tags within
// the event: ascending position is the order they were attached in, and gaps
// left by deleted tags are allowed.
type EventTag struct {
	tagID    uuid.UUID
	position int
}

// NewEventTag builds an EventTag with an explicit position, as when restoring
// stored tags whose positions may have gaps.
func NewEventTag(tagID uuid.UUID, position int) (EventTag, error) {
	if position < 1 || position > MaxTagsPerEvent {
		return EventTag{}, fmt.Errorf("%w: position %d", ErrEventTagPositionInvalid, position)
	}
	return EventTag{tagID: tagID, position: position}, nil
}

// NewEventTags builds EventTags from tag IDs in the order the user chose,
// numbering their positions from 1.
func NewEventTags(tagIDs []uuid.UUID) ([]EventTag, error) {
	tags := make([]EventTag, len(tagIDs))
	for i, id := range tagIDs {
		tags[i] = EventTag{tagID: id, position: i + 1}
	}
	if err := validateTags(tags); err != nil {
		return nil, err
	}
	return tags, nil
}

func (t EventTag) TagID() uuid.UUID { return t.tagID }
func (t EventTag) Position() int    { return t.position }

// validateTags checks the count first so that more than MaxTagsPerEvent tags
// is always reported as ErrEventTagLimitExceeded, whatever else is wrong.
func validateTags(tags []EventTag) error {
	if len(tags) > MaxTagsPerEvent {
		return fmt.Errorf("%w: got %d", ErrEventTagLimitExceeded, len(tags))
	}
	tagIDs := make(map[uuid.UUID]struct{}, len(tags))
	positions := make(map[int]struct{}, len(tags))
	for _, t := range tags {
		if _, ok := tagIDs[t.tagID]; ok {
			return fmt.Errorf("%w: %v", ErrEventTagDuplicated, t.tagID)
		}
		tagIDs[t.tagID] = struct{}{}
		// Re-check the range here too: the zero-value EventTag{} bypasses NewEventTag.
		if _, ok := positions[t.position]; ok || t.position < 1 || t.position > MaxTagsPerEvent {
			return fmt.Errorf("%w: position %d", ErrEventTagPositionInvalid, t.position)
		}
		positions[t.position] = struct{}{}
	}
	return nil
}

func sortedByPosition(tags []EventTag) []EventTag {
	sorted := slices.Clone(tags)
	slices.SortFunc(sorted, func(a, b EventTag) int { return a.position - b.position })
	return sorted
}
