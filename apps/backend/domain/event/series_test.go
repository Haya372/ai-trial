package event_test

import (
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/Haya372/ai-trial/backend/domain/event"
)

func TestNew_HasNoSeries(t *testing.T) {
	e := newTestEvent(t)

	if id, ok := e.SeriesID(); ok {
		t.Errorf("SeriesID: got (%v, true), want not in a series", id)
	}
}

func TestWithSeriesID_setsSeriesAndKeepsOriginalUnchanged(t *testing.T) {
	e := newTestEvent(t)
	seriesID := uuid.New()

	got := e.WithSeriesID(seriesID)

	if id, ok := got.SeriesID(); !ok || id != seriesID {
		t.Errorf("SeriesID: got (%v, %v), want (%v, true)", id, ok, seriesID)
	}
	if got.ID() != e.ID() || got.Title() != e.Title() || !got.StartAt().Equal(e.StartAt()) {
		t.Errorf("WithSeriesID changed event fields: got %+v, want %+v", got, e)
	}
	if _, ok := e.SeriesID(); ok {
		t.Errorf("original event was mutated")
	}
}

func TestDetachFromSeries_removesSeriesAndKeepsOriginalUnchanged(t *testing.T) {
	seriesID := uuid.New()
	e := newTestEvent(t).WithSeriesID(seriesID)

	got := e.DetachFromSeries()

	if id, ok := got.SeriesID(); ok {
		t.Errorf("SeriesID after detach: got (%v, true), want not in a series", id)
	}
	if id, ok := e.SeriesID(); !ok || id != seriesID {
		t.Errorf("original event was mutated: got (%v, %v)", id, ok)
	}
}

func TestWithTags_keepsSeriesID(t *testing.T) {
	seriesID := uuid.New()
	e := newTestEvent(t).WithSeriesID(seriesID)

	got, err := e.WithTags([]event.EventTag{mustEventTag(t, uuid.New(), 1)})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if id, ok := got.SeriesID(); !ok || id != seriesID {
		t.Errorf("SeriesID: got (%v, %v), want (%v, true)", id, ok, seriesID)
	}
}

func TestWithContent_replacesContentAndKeepsIdentityTagsAndSeries(t *testing.T) {
	seriesID, tagID := uuid.New(), uuid.New()
	e, err := newTestEvent(t).WithSeriesID(seriesID).WithTags([]event.EventTag{mustEventTag(t, tagID, 1)})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	start := e.StartAt().Add(24 * time.Hour)

	got, err := e.WithContent("Review", "notes", start, start.Add(2*time.Hour), "Room A", "https://example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got.Title() != "Review" || got.Description() != "notes" || !got.StartAt().Equal(start) ||
		!got.EndAt().Equal(start.Add(2*time.Hour)) || got.Location() != "Room A" || got.URL() != "https://example.com" {
		t.Errorf("content not replaced: got %+v", got)
	}
	if got.ID() != e.ID() || got.UserID() != e.UserID() {
		t.Errorf("identity changed: got (%v, %v), want (%v, %v)", got.ID(), got.UserID(), e.ID(), e.UserID())
	}
	if id, ok := got.SeriesID(); !ok || id != seriesID {
		t.Errorf("SeriesID: got (%v, %v), want (%v, true)", id, ok, seriesID)
	}
	if tags := got.Tags(); len(tags) != 1 || tags[0].TagID() != tagID {
		t.Errorf("Tags: got %+v, want one tag %v", tags, tagID)
	}
	if e.Title() != "Meeting" {
		t.Errorf("original event was mutated")
	}
}

func TestWithContent_validatesLikeNew(t *testing.T) {
	e := newTestEvent(t)

	if _, err := e.WithContent("", "", e.StartAt(), e.EndAt(), "", ""); err == nil {
		t.Error("empty title: expected an error, got nil")
	}
	if _, err := e.WithContent("Meeting", "", e.EndAt(), e.StartAt(), "", ""); err == nil {
		t.Error("end before start: expected an error, got nil")
	}
}
