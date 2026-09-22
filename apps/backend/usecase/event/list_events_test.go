package event_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	eventuc "github.com/Haya372/ai-trial/backend/usecase/event"
)

type stubQueryService struct {
	fn func(context.Context, eventuc.ListFilter) ([]eventuc.EventReadModel, error)
}

func (s *stubQueryService) List(ctx context.Context, filter eventuc.ListFilter) ([]eventuc.EventReadModel, error) {
	if s.fn == nil {
		return nil, nil
	}
	return s.fn(ctx, filter)
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
