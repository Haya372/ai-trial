package event_test

import (
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/Haya372/ai-trial/backend/domain/event"
)

func TestNew_ValidEvent(t *testing.T) {
	id := uuid.New()
	userID := uuid.New()
	title := "Meeting"
	desc := "Team sync"
	start := time.Now()
	end := start.Add(time.Hour)

	e, err := event.New(id, userID, title, desc, start, end, "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if e.ID() != id {
		t.Errorf("ID mismatch: got %v, want %v", e.ID(), id)
	}
	if e.UserID() != userID {
		t.Errorf("UserID mismatch: got %v, want %v", e.UserID(), userID)
	}
	if e.Title() != title {
		t.Errorf("Title mismatch: got %v, want %v", e.Title(), title)
	}
	if e.Description() != desc {
		t.Errorf("Description mismatch: got %v, want %v", e.Description(), desc)
	}
	if !e.StartAt().Equal(start) {
		t.Errorf("StartAt mismatch: got %v, want %v", e.StartAt(), start)
	}
	if !e.EndAt().Equal(end) {
		t.Errorf("EndAt mismatch: got %v, want %v", e.EndAt(), end)
	}
	if e.Location() != "" {
		t.Errorf("Location mismatch: got %v, want empty string", e.Location())
	}
	if e.URL() != "" {
		t.Errorf("URL mismatch: got %v, want empty string", e.URL())
	}
}

func TestNew_EmptyTitle(t *testing.T) {
	id := uuid.New()
	userID := uuid.New()
	start := time.Now()
	end := start.Add(time.Hour)

	_, err := event.New(id, userID, "", "", start, end, "", "")
	if err == nil {
		t.Fatal("expected error for empty title, got nil")
	}
}

func TestNew_InvalidDateRange(t *testing.T) {
	id := uuid.New()
	userID := uuid.New()
	now := time.Now()
	start := now.Add(time.Hour)
	end := now // end < start

	_, err := event.New(id, userID, "Meeting", "", start, end, "", "")
	if err == nil {
		t.Fatal("expected error for invalid date range, got nil")
	}
}

func TestNew_EqualStartAndEnd(t *testing.T) {
	id := uuid.New()
	userID := uuid.New()
	now := time.Now()

	// end == start is valid (zero-duration event)
	_, err := event.New(id, userID, "Meeting", "", now, now, "", "")
	if err != nil {
		t.Fatalf("unexpected error for equal start/end: %v", err)
	}
}

func TestNew_WithLocationAndURL(t *testing.T) {
	id := uuid.New()
	userID := uuid.New()
	start := time.Now()
	end := start.Add(time.Hour)
	location := "Tokyo"
	url := "https://example.com/event"

	e, err := event.New(id, userID, "Meeting", "desc", start, end, location, url)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if e.Location() != location {
		t.Errorf("Location mismatch: got %v, want %v", e.Location(), location)
	}
	if e.URL() != url {
		t.Errorf("URL mismatch: got %v, want %v", e.URL(), url)
	}
}

func TestNew_EmptyLocationAndURL(t *testing.T) {
	id := uuid.New()
	userID := uuid.New()
	start := time.Now()
	end := start.Add(time.Hour)

	e, err := event.New(id, userID, "Meeting", "desc", start, end, "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if e.Location() != "" {
		t.Errorf("Location mismatch: got %v, want empty string", e.Location())
	}
	if e.URL() != "" {
		t.Errorf("URL mismatch: got %v, want empty string", e.URL())
	}
}
