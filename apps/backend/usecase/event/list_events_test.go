package event_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	domaineventsubscription "github.com/Haya372/ai-trial/backend/domain/eventsubscription"
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

// stubSubscriptionRepo implements domaineventsubscription.Repository. Only
// ListByUserID is exercised by ListEventsQuery; the remaining methods are
// unused no-ops required to satisfy the interface.
type stubSubscriptionRepo struct {
	listByUserIDFn func(context.Context, uuid.UUID) ([]domaineventsubscription.EventSubscription, error)
}

func (s *stubSubscriptionRepo) Create(
	_ context.Context, sub domaineventsubscription.EventSubscription,
) (domaineventsubscription.EventSubscription, error) {
	return sub, nil
}

func (s *stubSubscriptionRepo) FindByID(
	_ context.Context, _ uuid.UUID,
) (domaineventsubscription.EventSubscription, error) {
	return nil, domaineventsubscription.ErrEventSubscriptionNotFound
}

func (s *stubSubscriptionRepo) FindByEventAndUserID(
	_ context.Context, _, _ uuid.UUID,
) (domaineventsubscription.EventSubscription, error) {
	return nil, domaineventsubscription.ErrEventSubscriptionNotFound
}

func (s *stubSubscriptionRepo) ListByUserID(
	ctx context.Context, userID uuid.UUID,
) ([]domaineventsubscription.EventSubscription, error) {
	if s.listByUserIDFn == nil {
		return nil, nil
	}
	return s.listByUserIDFn(ctx, userID)
}

func (s *stubSubscriptionRepo) Delete(_ context.Context, _ uuid.UUID) error {
	return nil
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

	q := eventuc.NewListEventsQuery(stub, &stubSubscriptionRepo{})
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

func TestListEventsQuery_Execute_RepoError(t *testing.T) {
	userID := uuid.New()
	now := time.Now()

	stub := &stubQueryService{
		fn: func(_ context.Context, _ eventuc.ListFilter) ([]eventuc.EventReadModel, error) {
			return nil, errDBFailure
		},
	}

	q := eventuc.NewListEventsQuery(stub, &stubSubscriptionRepo{})
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

	q := eventuc.NewListEventsQuery(stub, &stubSubscriptionRepo{})
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

func TestListEventsQuery_Execute_IncludesSubscribedEvents(t *testing.T) {
	userID := uuid.New()
	ownerID := uuid.New()
	now := time.Now()
	start := now
	end := now.Add(24 * time.Hour)

	ownEventID := uuid.New()
	subscribedEventID := uuid.New()

	queryStub := &stubQueryService{
		fn: func(_ context.Context, _ eventuc.ListFilter) ([]eventuc.EventReadModel, error) {
			return []eventuc.EventReadModel{
				{ID: ownEventID, UserID: userID, Title: "My event", StartAt: start, EndAt: start.Add(time.Hour)},
			}, nil
		},
		findByIDFn: func(_ context.Context, id uuid.UUID) (eventuc.EventReadModel, error) {
			if id != subscribedEventID {
				t.Fatalf("unexpected FindByID id: %v", id)
			}
			return eventuc.EventReadModel{
				ID:      subscribedEventID,
				UserID:  ownerID,
				Title:   "Shared meeting",
				StartAt: start.Add(2 * time.Hour),
				EndAt:   start.Add(3 * time.Hour),
			}, nil
		},
	}

	subsStub := &stubSubscriptionRepo{
		listByUserIDFn: func(_ context.Context, uid uuid.UUID) ([]domaineventsubscription.EventSubscription, error) {
			if uid != userID {
				t.Errorf("userID mismatch: got %v, want %v", uid, userID)
			}
			return []domaineventsubscription.EventSubscription{
				domaineventsubscription.New(uuid.New(), subscribedEventID, userID, now),
			}, nil
		},
	}

	q := eventuc.NewListEventsQuery(queryStub, subsStub)
	result, err := q.Execute(context.Background(), userID, eventuc.ListEventsInput{StartDate: start, EndDate: end})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 events, got %d: %+v", len(result), result)
	}

	var own, subscribed *eventuc.EventReadModel
	for i := range result {
		switch result[i].ID {
		case ownEventID:
			own = &result[i]
		case subscribedEventID:
			subscribed = &result[i]
		}
	}
	if own == nil || subscribed == nil {
		t.Fatalf("missing expected events in result: %+v", result)
	}
	if own.IsSubscribed {
		t.Errorf("expected own event to not be marked IsSubscribed")
	}
	if !subscribed.IsSubscribed {
		t.Errorf("expected subscribed event to be marked IsSubscribed")
	}
}

func TestListEventsQuery_Execute_ExcludesSubscribedEventOutsideDateRange(t *testing.T) {
	userID := uuid.New()
	now := time.Now()
	start := now
	end := now.Add(24 * time.Hour)

	subscribedEventID := uuid.New()

	queryStub := &stubQueryService{
		fn: func(_ context.Context, _ eventuc.ListFilter) ([]eventuc.EventReadModel, error) {
			return []eventuc.EventReadModel{}, nil
		},
		findByIDFn: func(_ context.Context, _ uuid.UUID) (eventuc.EventReadModel, error) {
			// Entirely before the requested range: end_at (-23h) <= start (0h).
			return eventuc.EventReadModel{
				ID:      subscribedEventID,
				Title:   "Old shared meeting",
				StartAt: start.Add(-24 * time.Hour),
				EndAt:   start.Add(-23 * time.Hour),
			}, nil
		},
	}

	subsStub := &stubSubscriptionRepo{
		listByUserIDFn: func(_ context.Context, _ uuid.UUID) ([]domaineventsubscription.EventSubscription, error) {
			return []domaineventsubscription.EventSubscription{
				domaineventsubscription.New(uuid.New(), subscribedEventID, userID, now),
			}, nil
		},
	}

	q := eventuc.NewListEventsQuery(queryStub, subsStub)
	result, err := q.Execute(context.Background(), userID, eventuc.ListEventsInput{StartDate: start, EndDate: end})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Fatalf("expected subscribed event outside date range to be excluded, got %d: %+v", len(result), result)
	}
}

func TestListEventsQuery_Execute_SortsMergedEventsByStartAt(t *testing.T) {
	userID := uuid.New()
	now := time.Now()
	start := now
	end := now.Add(24 * time.Hour)

	laterOwnEventID := uuid.New()
	earlierSubscribedEventID := uuid.New()

	queryStub := &stubQueryService{
		fn: func(_ context.Context, _ eventuc.ListFilter) ([]eventuc.EventReadModel, error) {
			return []eventuc.EventReadModel{
				{ID: laterOwnEventID, Title: "Later own event", StartAt: start.Add(3 * time.Hour), EndAt: start.Add(4 * time.Hour)},
			}, nil
		},
		findByIDFn: func(_ context.Context, _ uuid.UUID) (eventuc.EventReadModel, error) {
			return eventuc.EventReadModel{
				ID:      earlierSubscribedEventID,
				Title:   "Earlier subscribed event",
				StartAt: start.Add(time.Hour),
				EndAt:   start.Add(2 * time.Hour),
			}, nil
		},
	}

	subsStub := &stubSubscriptionRepo{
		listByUserIDFn: func(_ context.Context, _ uuid.UUID) ([]domaineventsubscription.EventSubscription, error) {
			return []domaineventsubscription.EventSubscription{
				domaineventsubscription.New(uuid.New(), earlierSubscribedEventID, userID, now),
			}, nil
		},
	}

	q := eventuc.NewListEventsQuery(queryStub, subsStub)
	result, err := q.Execute(context.Background(), userID, eventuc.ListEventsInput{StartDate: start, EndDate: end})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 events, got %d", len(result))
	}
	if result[0].ID != earlierSubscribedEventID || result[1].ID != laterOwnEventID {
		t.Errorf("expected events sorted by StartAt, got order: %v, %v", result[0].ID, result[1].ID)
	}
}

func TestListEventsQuery_Execute_SubscriptionRepoError(t *testing.T) {
	userID := uuid.New()
	now := time.Now()

	queryStub := &stubQueryService{
		fn: func(_ context.Context, _ eventuc.ListFilter) ([]eventuc.EventReadModel, error) {
			return []eventuc.EventReadModel{}, nil
		},
	}
	subsStub := &stubSubscriptionRepo{
		listByUserIDFn: func(_ context.Context, _ uuid.UUID) ([]domaineventsubscription.EventSubscription, error) {
			return nil, errDBFailure
		},
	}

	q := eventuc.NewListEventsQuery(queryStub, subsStub)
	_, err := q.Execute(context.Background(), userID, eventuc.ListEventsInput{
		StartDate: now,
		EndDate:   now.Add(24 * time.Hour),
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
