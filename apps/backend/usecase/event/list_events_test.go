package event_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	eventuc "github.com/Haya372/ai-trial/backend/usecase/event"
)

type stubQueryService struct {
	fn         func(context.Context, eventuc.ListFilter) ([]eventuc.EventReadModel, error)
	findByIDFn func(context.Context, uuid.UUID) (eventuc.EventReadModel, error)
}

func (s *stubQueryService) List(ctx context.Context, filter eventuc.ListFilter) ([]eventuc.EventReadModel, error) {
	if s.fn == nil {
		return nil, nil
	}
	return s.fn(ctx, filter)
}

func (s *stubQueryService) FindByID(ctx context.Context, id uuid.UUID) (eventuc.EventReadModel, error) {
	if s.findByIDFn == nil {
		return eventuc.EventReadModel{}, nil
	}
	return s.findByIDFn(ctx, id)
}

func TestListEventsQuery_Execute_ReturnsEvents(t *testing.T) {
	userID := uuid.New()
	now := time.Now()
	start := now
	end := now.Add(24 * time.Hour)

	eventID := uuid.New()
	stub := &stubQueryService{
		fn: func(_ context.Context, filter eventuc.ListFilter) ([]eventuc.EventReadModel, error) {
			if filter.UserID != userID {
				t.Errorf("userID mismatch: got %v, want %v", filter.UserID, userID)
			}
			return []eventuc.EventReadModel{
				{
					ID:       eventID,
					Title:    "Team meeting",
					StartAt:  start,
					EndAt:    start.Add(time.Hour),
					Location: "Tokyo",
					URL:      "https://example.com",
				},
			}, nil
		},
	}

	q := eventuc.NewListEventsQuery(stub)
	result, err := q.Execute(context.Background(), userID, eventuc.ListEventsInput{
		StartDate: start,
		EndDate:   end,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 event, got %d", len(result))
	}
	if result[0].Title != "Team meeting" {
		t.Errorf("title mismatch: got %q", result[0].Title)
	}
	if result[0].Location != "Tokyo" {
		t.Errorf("location mismatch: got %q, want %q", result[0].Location, "Tokyo")
	}
	if result[0].URL != "https://example.com" {
		t.Errorf("url mismatch: got %q, want %q", result[0].URL, "https://example.com")
	}
	if result[0].IsSubscribed {
		t.Errorf("expected IsSubscribed to be false for own event")
	}
}

// IsSubscribed is computed entirely by the query side's SQL (ADR-022 logical
// CQRS); this usecase must pass whatever QueryService.List returns straight
// through without altering it.
func TestListEventsQuery_Execute_PassesThroughIsSubscribed(t *testing.T) {
	userID := uuid.New()
	now := time.Now()

	stub := &stubQueryService{
		fn: func(_ context.Context, _ eventuc.ListFilter) ([]eventuc.EventReadModel, error) {
			return []eventuc.EventReadModel{
				{ID: uuid.New(), Title: "My event", StartAt: now, EndAt: now.Add(time.Hour), IsSubscribed: false},
				{ID: uuid.New(), Title: "Shared event", StartAt: now, EndAt: now.Add(time.Hour), IsSubscribed: true},
			}, nil
		},
	}

	q := eventuc.NewListEventsQuery(stub)
	result, err := q.Execute(context.Background(), userID, eventuc.ListEventsInput{
		StartDate: now,
		EndDate:   now.Add(24 * time.Hour),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 events, got %d", len(result))
	}
	if result[0].IsSubscribed {
		t.Errorf("expected first event IsSubscribed=false, got true")
	}
	if !result[1].IsSubscribed {
		t.Errorf("expected second event IsSubscribed=true, got false")
	}
}

func TestListEventsQuery_Execute_RepoError(t *testing.T) {
	userID := uuid.New()
	now := time.Now()

	stub := &stubQueryService{
		fn: func(_ context.Context, _ eventuc.ListFilter) ([]eventuc.EventReadModel, error) {
			return nil, errDBFailure
		},
	}

	q := eventuc.NewListEventsQuery(stub)
	_, err := q.Execute(context.Background(), userID, eventuc.ListEventsInput{
		StartDate: now,
		EndDate:   now.Add(24 * time.Hour),
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestListEventsQuery_Execute_EmptyResult(t *testing.T) {
	userID := uuid.New()
	now := time.Now()

	stub := &stubQueryService{
		fn: func(_ context.Context, _ eventuc.ListFilter) ([]eventuc.EventReadModel, error) {
			return []eventuc.EventReadModel{}, nil
		},
	}

	q := eventuc.NewListEventsQuery(stub)
	result, err := q.Execute(context.Background(), userID, eventuc.ListEventsInput{
		StartDate: now,
		EndDate:   now.Add(24 * time.Hour),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected empty result, got %d events", len(result))
	}
}
