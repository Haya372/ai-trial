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
	fn          func(context.Context, eventuc.ListFilter) ([]eventuc.EventReadModel, error)
	findByIDFn  func(context.Context, uuid.UUID) (eventuc.EventReadModel, error)
	listByIDsFn func(context.Context, []uuid.UUID, time.Time, time.Time) ([]eventuc.EventReadModel, error)
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

func (s *stubQueryService) ListByIDs(
	ctx context.Context, ids []uuid.UUID, startDate, endDate time.Time,
) ([]eventuc.EventReadModel, error) {
	if s.listByIDsFn == nil {
		return nil, nil
	}
	return s.listByIDsFn(ctx, ids, startDate, endDate)
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

func TestListEventsQuery_Execute_DoesNotCallListByIDsWhenNoSubscriptions(t *testing.T) {
	userID := uuid.New()
	now := time.Now()

	called := false
	queryStub := &stubQueryService{
		fn: func(_ context.Context, _ eventuc.ListFilter) ([]eventuc.EventReadModel, error) {
			return []eventuc.EventReadModel{}, nil
		},
		listByIDsFn: func(
			_ context.Context, _ []uuid.UUID, _, _ time.Time,
		) ([]eventuc.EventReadModel, error) {
			called = true
			return nil, nil
		},
	}

	q := eventuc.NewListEventsQuery(queryStub, &stubSubscriptionRepo{})
	_, err := q.Execute(context.Background(), userID, eventuc.ListEventsInput{
		StartDate: now,
		EndDate:   now.Add(24 * time.Hour),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if called {
		t.Errorf("expected ListByIDs not to be called when the user has no subscriptions")
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
		listByIDsFn: func(
			_ context.Context, ids []uuid.UUID, gotStart, gotEnd time.Time,
		) ([]eventuc.EventReadModel, error) {
			if len(ids) != 1 || ids[0] != subscribedEventID {
				t.Fatalf("unexpected ids: %v", ids)
			}
			if !gotStart.Equal(start) || !gotEnd.Equal(end) {
				t.Errorf("date range mismatch: got (%v, %v), want (%v, %v)", gotStart, gotEnd, start, end)
			}
			return []eventuc.EventReadModel{
				{
					ID:      subscribedEventID,
					UserID:  ownerID,
					Title:   "Shared meeting",
					StartAt: start.Add(2 * time.Hour),
					EndAt:   start.Add(3 * time.Hour),
				},
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

// A source event deleted between ListByUserID and ListByIDs (or already
// outside the date range) must simply be absent from ListByIDs' result,
// not surfaced as an error that would fail the whole request.
func TestListEventsQuery_Execute_SubscribedEventMissingFromListByIDs_SilentlyOmitted(t *testing.T) {
	userID := uuid.New()
	now := time.Now()
	start := now
	end := now.Add(24 * time.Hour)

	ownEventID := uuid.New()
	staleSubscribedEventID := uuid.New()

	queryStub := &stubQueryService{
		fn: func(_ context.Context, _ eventuc.ListFilter) ([]eventuc.EventReadModel, error) {
			return []eventuc.EventReadModel{
				{ID: ownEventID, Title: "My event", StartAt: start, EndAt: start.Add(time.Hour)},
			}, nil
		},
		listByIDsFn: func(
			_ context.Context, _ []uuid.UUID, _, _ time.Time,
		) ([]eventuc.EventReadModel, error) {
			// The source event no longer exists (or fell outside the range),
			// so the bulk query simply returns nothing for it.
			return nil, nil
		},
	}

	subsStub := &stubSubscriptionRepo{
		listByUserIDFn: func(_ context.Context, _ uuid.UUID) ([]domaineventsubscription.EventSubscription, error) {
			return []domaineventsubscription.EventSubscription{
				domaineventsubscription.New(uuid.New(), staleSubscribedEventID, userID, now),
			}, nil
		},
	}

	q := eventuc.NewListEventsQuery(queryStub, subsStub)
	result, err := q.Execute(context.Background(), userID, eventuc.ListEventsInput{StartDate: start, EndDate: end})
	if err != nil {
		t.Fatalf("expected no error when a subscribed event is missing, got: %v", err)
	}
	if len(result) != 1 || result[0].ID != ownEventID {
		t.Fatalf("expected only the own event to remain, got: %+v", result)
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
		listByIDsFn: func(
			_ context.Context, _ []uuid.UUID, _, _ time.Time,
		) ([]eventuc.EventReadModel, error) {
			return []eventuc.EventReadModel{
				{
					ID:      earlierSubscribedEventID,
					Title:   "Earlier subscribed event",
					StartAt: start.Add(time.Hour),
					EndAt:   start.Add(2 * time.Hour),
				},
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

func TestListEventsQuery_Execute_ListByIDsError(t *testing.T) {
	userID := uuid.New()
	now := time.Now()

	queryStub := &stubQueryService{
		fn: func(_ context.Context, _ eventuc.ListFilter) ([]eventuc.EventReadModel, error) {
			return []eventuc.EventReadModel{}, nil
		},
		listByIDsFn: func(
			_ context.Context, _ []uuid.UUID, _, _ time.Time,
		) ([]eventuc.EventReadModel, error) {
			return nil, errDBFailure
		},
	}
	subsStub := &stubSubscriptionRepo{
		listByUserIDFn: func(_ context.Context, _ uuid.UUID) ([]domaineventsubscription.EventSubscription, error) {
			return []domaineventsubscription.EventSubscription{
				domaineventsubscription.New(uuid.New(), uuid.New(), userID, now),
			}, nil
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
