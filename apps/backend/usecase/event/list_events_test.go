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

func TestListEventsQuery_Execute_PassesThroughSubscriptionID(t *testing.T) {
	userID := uuid.New()
	now := time.Now()
	subscriptionID := uuid.New()

	stub := &stubQueryService{
		fn: func(_ context.Context, _ eventuc.ListFilter) ([]eventuc.EventReadModel, error) {
			return []eventuc.EventReadModel{
				{ID: uuid.New(), Title: "My event", StartAt: now, EndAt: now.Add(time.Hour), IsSubscribed: false},
				{
					ID: uuid.New(), Title: "Shared event", StartAt: now, EndAt: now.Add(time.Hour),
					IsSubscribed: true, SubscriptionID: &subscriptionID,
				},
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
	if result[0].SubscriptionID != nil {
		t.Errorf("expected own event SubscriptionID to be nil, got %v", result[0].SubscriptionID)
	}
	if result[1].SubscriptionID == nil || *result[1].SubscriptionID != subscriptionID {
		t.Errorf("expected subscribed event SubscriptionID to be %v, got %v", subscriptionID, result[1].SubscriptionID)
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

func TestListEventsQuery_Execute_NilTagID_PassesNilToQueryService(t *testing.T) {
	userID := uuid.New()
	now := time.Now()

	var capturedFilter eventuc.ListFilter
	stub := &stubQueryService{
		fn: func(_ context.Context, filter eventuc.ListFilter) ([]eventuc.EventReadModel, error) {
			capturedFilter = filter
			return nil, nil
		},
	}

	q := eventuc.NewListEventsQuery(stub)
	_, err := q.Execute(context.Background(), userID, eventuc.ListEventsInput{
		StartDate: now,
		EndDate:   now.Add(24 * time.Hour),
		TagID:     nil,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedFilter.TagID != nil {
		t.Errorf("expected TagID nil in filter, got %v", capturedFilter.TagID)
	}
}

func TestListEventsQuery_Execute_WithTagID_PassesTagIDToQueryService(t *testing.T) {
	userID := uuid.New()
	now := time.Now()
	tagID := uuid.New()

	var capturedFilter eventuc.ListFilter
	stub := &stubQueryService{
		fn: func(_ context.Context, filter eventuc.ListFilter) ([]eventuc.EventReadModel, error) {
			capturedFilter = filter
			return nil, nil
		},
	}

	q := eventuc.NewListEventsQuery(stub)
	_, err := q.Execute(context.Background(), userID, eventuc.ListEventsInput{
		StartDate: now,
		EndDate:   now.Add(24 * time.Hour),
		TagID:     &tagID,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedFilter.TagID == nil || *capturedFilter.TagID != tagID {
		t.Errorf("expected TagID %v in filter, got %v", tagID, capturedFilter.TagID)
	}
}

func TestListEventsQuery_Execute_TagsArePassedThrough(t *testing.T) {
	userID := uuid.New()
	now := time.Now()
	tagID := uuid.New()

	stub := &stubQueryService{
		fn: func(_ context.Context, _ eventuc.ListFilter) ([]eventuc.EventReadModel, error) {
			return []eventuc.EventReadModel{
				{
					ID:    uuid.New(),
					Title: "Tagged event",
					Tags: []eventuc.EventTagReadModel{
						{ID: tagID, Name: "Work", Color: "blue"},
					},
				},
				{
					ID:    uuid.New(),
					Title: "Untagged event",
					Tags:  []eventuc.EventTagReadModel{},
				},
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
	if len(result[0].Tags) != 1 || result[0].Tags[0].ID != tagID {
		t.Errorf("expected first event to have tag %v, got %v", tagID, result[0].Tags)
	}
	if len(result[1].Tags) != 0 {
		t.Errorf("expected second event to have no tags, got %d", len(result[1].Tags))
	}
}

func TestListEventsQuery_Execute_SubscribedEventTagsPassedThrough(t *testing.T) {
	userID := uuid.New()
	now := time.Now()

	stub := &stubQueryService{
		fn: func(_ context.Context, _ eventuc.ListFilter) ([]eventuc.EventReadModel, error) {
			return []eventuc.EventReadModel{
				{
					ID:           uuid.New(),
					Title:        "Subscribed event",
					IsSubscribed: true,
					Tags:         []eventuc.EventTagReadModel{}, // empty per spec
				},
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
	if len(result) != 1 {
		t.Fatalf("expected 1 event, got %d", len(result))
	}
	if len(result[0].Tags) != 0 {
		t.Errorf("expected subscribed event to have empty tags, got %d", len(result[0].Tags))
	}
}
