package event_test

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/Haya372/ai-trial/backend/domain/event"
)

func newTestEvent(t *testing.T) event.Event {
	t.Helper()
	start := time.Now()
	e, err := event.New(uuid.New(), uuid.New(), "Meeting", "", start, start.Add(time.Hour), "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	return e
}

func mustEventTag(t *testing.T, tagID uuid.UUID, position int) event.EventTag {
	t.Helper()
	et, err := event.NewEventTag(tagID, position)
	if err != nil {
		t.Fatalf("NewEventTag(%v, %d): unexpected error: %v", tagID, position, err)
	}
	return et
}

func TestNewEventTag_validPositions(t *testing.T) {
	tagID := uuid.New()
	for _, pos := range []int{1, 5} {
		et, err := event.NewEventTag(tagID, pos)
		if err != nil {
			t.Errorf("NewEventTag(position=%d): unexpected error: %v", pos, err)
			continue
		}
		if et.TagID() != tagID {
			t.Errorf("TagID mismatch: got %v, want %v", et.TagID(), tagID)
		}
		if et.Position() != pos {
			t.Errorf("Position mismatch: got %d, want %d", et.Position(), pos)
		}
	}
}

func TestNewEventTag_positionOutOfRange(t *testing.T) {
	for _, pos := range []int{0, 6, -1} {
		_, err := event.NewEventTag(uuid.New(), pos)
		if !errors.Is(err, event.ErrEventTagPositionInvalid) {
			t.Errorf("NewEventTag(position=%d): got %v, want ErrEventTagPositionInvalid", pos, err)
		}
	}
}

func TestNewEventTags_assignsPositionsInOrder(t *testing.T) {
	ids := []uuid.UUID{uuid.New(), uuid.New(), uuid.New()}

	tags, err := event.NewEventTags(ids)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(tags) != len(ids) {
		t.Fatalf("len mismatch: got %d, want %d", len(tags), len(ids))
	}
	for i, et := range tags {
		if et.TagID() != ids[i] {
			t.Errorf("tags[%d].TagID: got %v, want %v", i, et.TagID(), ids[i])
		}
		if et.Position() != i+1 {
			t.Errorf("tags[%d].Position: got %d, want %d", i, et.Position(), i+1)
		}
	}
}

func TestNewEventTags_empty(t *testing.T) {
	tags, err := event.NewEventTags(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tags) != 0 {
		t.Errorf("got %d tags, want 0", len(tags))
	}
}

func TestNewEventTags_maxCount(t *testing.T) {
	ids := []uuid.UUID{uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New()}
	if _, err := event.NewEventTags(ids); err != nil {
		t.Errorf("5 tags should be valid, got: %v", err)
	}
}

func TestNewEventTags_tooMany(t *testing.T) {
	ids := []uuid.UUID{uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New()}
	_, err := event.NewEventTags(ids)
	if !errors.Is(err, event.ErrEventTagLimitExceeded) {
		t.Errorf("got %v, want ErrEventTagLimitExceeded", err)
	}
}

func TestNewEventTags_duplicated(t *testing.T) {
	id := uuid.New()
	_, err := event.NewEventTags([]uuid.UUID{id, uuid.New(), id})
	if !errors.Is(err, event.ErrEventTagDuplicated) {
		t.Errorf("got %v, want ErrEventTagDuplicated", err)
	}
}

func TestNew_HasNoTags(t *testing.T) {
	e := newTestEvent(t)
	if len(e.Tags()) != 0 {
		t.Errorf("got %d tags, want 0", len(e.Tags()))
	}
}

func TestWithTags_sortsByPositionAndKeepsGaps(t *testing.T) {
	e := newTestEvent(t)
	a, b, c := uuid.New(), uuid.New(), uuid.New()

	// Positions 2 and 4 are missing, as after their tags were deleted.
	got, err := e.WithTags([]event.EventTag{
		mustEventTag(t, c, 5),
		mustEventTag(t, a, 1),
		mustEventTag(t, b, 3),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []struct {
		tagID    uuid.UUID
		position int
	}{{a, 1}, {b, 3}, {c, 5}}
	tags := got.Tags()
	if len(tags) != len(want) {
		t.Fatalf("len mismatch: got %d, want %d", len(tags), len(want))
	}
	for i, w := range want {
		if tags[i].TagID() != w.tagID || tags[i].Position() != w.position {
			t.Errorf("tags[%d]: got (%v, %d), want (%v, %d)",
				i, tags[i].TagID(), tags[i].Position(), w.tagID, w.position)
		}
	}
}

func TestWithTags_keepsOtherFieldsAndOriginalUnchanged(t *testing.T) {
	e := newTestEvent(t)

	got, err := e.WithTags([]event.EventTag{mustEventTag(t, uuid.New(), 1)})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got.ID() != e.ID() || got.UserID() != e.UserID() || got.Title() != e.Title() ||
		!got.StartAt().Equal(e.StartAt()) || !got.EndAt().Equal(e.EndAt()) {
		t.Errorf("WithTags changed event fields: got %+v, want %+v", got, e)
	}
	if len(e.Tags()) != 0 {
		t.Errorf("original event was mutated: got %d tags, want 0", len(e.Tags()))
	}
}

func TestWithTags_tooMany(t *testing.T) {
	e := newTestEvent(t)
	tags := make([]event.EventTag, 0, 6)
	for i := range 5 {
		tags = append(tags, mustEventTag(t, uuid.New(), i+1))
	}
	// A sixth tag can only collide on position, but the count limit is checked first.
	tags = append(tags, mustEventTag(t, uuid.New(), 1))

	_, err := e.WithTags(tags)
	if !errors.Is(err, event.ErrEventTagLimitExceeded) {
		t.Errorf("got %v, want ErrEventTagLimitExceeded", err)
	}
}

func TestWithTags_duplicatedTag(t *testing.T) {
	e := newTestEvent(t)
	id := uuid.New()

	_, err := e.WithTags([]event.EventTag{mustEventTag(t, id, 1), mustEventTag(t, id, 2)})
	if !errors.Is(err, event.ErrEventTagDuplicated) {
		t.Errorf("got %v, want ErrEventTagDuplicated", err)
	}
}

func TestWithTags_duplicatedPosition(t *testing.T) {
	e := newTestEvent(t)

	_, err := e.WithTags([]event.EventTag{mustEventTag(t, uuid.New(), 2), mustEventTag(t, uuid.New(), 2)})
	if !errors.Is(err, event.ErrEventTagPositionInvalid) {
		t.Errorf("got %v, want ErrEventTagPositionInvalid", err)
	}
}

func TestWithTags_zeroValueEventTagRejected(t *testing.T) {
	e := newTestEvent(t)

	_, err := e.WithTags([]event.EventTag{{}})
	if !errors.Is(err, event.ErrEventTagPositionInvalid) {
		t.Errorf("got %v, want ErrEventTagPositionInvalid", err)
	}
}

func TestTags_returnsCopy(t *testing.T) {
	e, err := newTestEvent(t).WithTags([]event.EventTag{mustEventTag(t, uuid.New(), 1)})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tags := e.Tags()
	tags[0] = mustEventTag(t, uuid.New(), 2)

	if e.Tags()[0].Position() != 1 {
		t.Errorf("mutating the returned slice changed the event")
	}
}
