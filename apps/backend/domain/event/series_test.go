package event_test

import (
	"testing"

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
