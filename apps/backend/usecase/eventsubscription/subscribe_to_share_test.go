package eventsubscription_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"

	domaineventshare "github.com/Haya372/ai-trial/backend/domain/eventshare"
	eventsharemock "github.com/Haya372/ai-trial/backend/domain/eventshare/generated"
	"github.com/Haya372/ai-trial/backend/domain/eventsubscription"
	subscriptionmock "github.com/Haya372/ai-trial/backend/domain/eventsubscription/generated"
	eventuc "github.com/Haya372/ai-trial/backend/usecase/event"
	eventshareuc "github.com/Haya372/ai-trial/backend/usecase/eventshare"
	eventsubscriptionuc "github.com/Haya372/ai-trial/backend/usecase/eventsubscription"
)

var testToken = domaineventshare.NewToken("tok")

var errDBFailure = errors.New("db error")

func newTestEvent(ownerID uuid.UUID) eventuc.EventReadModel {
	now := time.Now()
	return eventuc.EventReadModel{
		ID:      uuid.New(),
		UserID:  ownerID,
		Title:   "Test Event",
		StartAt: now,
		EndAt:   now.Add(time.Hour),
	}
}

func newTestShare(eventID uuid.UUID, expired bool) domaineventshare.EventShare {
	expiresAt := time.Now().Add(time.Hour)
	if expired {
		expiresAt = time.Now().Add(-time.Hour)
	}
	s, err := domaineventshare.New(uuid.New(), eventID, domaineventshare.NewTokenHash("hash"), expiresAt)
	if err != nil {
		panic(err)
	}
	return s
}

// stubEventQueryService implements eventuc.QueryService with only the method
// ShareTokenLoader needs (FindByID); List is unused by this package's tests.
type stubEventQueryService struct {
	findByIDFn func(context.Context, uuid.UUID) (eventuc.EventReadModel, error)
}

func (s *stubEventQueryService) List(context.Context, eventuc.ListFilter) ([]eventuc.EventReadModel, error) {
	return nil, nil
}

func (s *stubEventQueryService) ListByIDs(
	context.Context, []uuid.UUID, time.Time, time.Time,
) ([]eventuc.EventReadModel, error) {
	return nil, nil
}

func (s *stubEventQueryService) FindByID(ctx context.Context, id uuid.UUID) (eventuc.EventReadModel, error) {
	if s.findByIDFn == nil {
		return eventuc.EventReadModel{}, nil
	}
	return s.findByIDFn(ctx, id)
}

func findByIDReturning(ev eventuc.EventReadModel) func(*stubEventQueryService) {
	return func(q *stubEventQueryService) {
		q.findByIDFn = func(_ context.Context, _ uuid.UUID) (eventuc.EventReadModel, error) {
			return ev, nil
		}
	}
}

// subMatcher matches an eventsubscription.EventSubscription by its
// event/user pair, ignoring the generated ID and CreatedAt.
type subMatcher struct{ eventID, userID uuid.UUID }

func (m subMatcher) Matches(x any) bool {
	s, ok := x.(eventsubscription.EventSubscription)
	return ok && s.EventID() == m.eventID && s.UserID() == m.userID
}

func (m subMatcher) String() string { return "subscription matching event/user" }

func buildCommand(
	ctrl *gomock.Controller,
	shareRepoFn func(*eventsharemock.MockRepository),
	eventQueryFn func(*stubEventQueryService),
	subsRepoFn func(*subscriptionmock.MockRepository),
) *eventsubscriptionuc.SubscribeToShareCommand {
	shareRepo := eventsharemock.NewMockRepository(ctrl)
	eventQuery := &stubEventQueryService{}
	subsRepo := subscriptionmock.NewMockRepository(ctrl)

	if shareRepoFn != nil {
		shareRepoFn(shareRepo)
	}
	if eventQueryFn != nil {
		eventQueryFn(eventQuery)
	}
	if subsRepoFn != nil {
		subsRepoFn(subsRepo)
	}

	loader := eventshareuc.NewShareTokenLoader(shareRepo, eventQuery)
	return eventsubscriptionuc.NewSubscribeToShareCommand(loader, subsRepo)
}

func TestSubscribeToShareCommand_Execute_tokenNotFound_returns404error(t *testing.T) {
	ctrl := gomock.NewController(t)
	c := buildCommand(ctrl,
		func(r *eventsharemock.MockRepository) {
			r.EXPECT().FindByToken(gomock.Any(), testToken).Return(nil, domaineventshare.ErrEventShareNotFound)
		},
		nil, nil,
	)

	_, err := c.Execute(context.Background(), eventsubscriptionuc.SubscribeToShareInput{
		Token: testToken, UserID: uuid.New(),
	})

	if !errors.Is(err, domaineventshare.ErrEventShareNotFound) {
		t.Errorf("expected ErrEventShareNotFound, got %v", err)
	}
}

func TestSubscribeToShareCommand_Execute_tokenExpired_returns410error(t *testing.T) {
	ctrl := gomock.NewController(t)
	ownerID := uuid.New()
	ev := newTestEvent(ownerID)
	expiredShare := newTestShare(ev.ID, true)

	c := buildCommand(ctrl,
		func(r *eventsharemock.MockRepository) {
			r.EXPECT().FindByToken(gomock.Any(), testToken).Return(expiredShare, nil)
		},
		nil, nil,
	)

	_, err := c.Execute(context.Background(), eventsubscriptionuc.SubscribeToShareInput{
		Token: testToken, UserID: uuid.New(),
	})

	if !errors.Is(err, domaineventshare.ErrEventShareExpired) {
		t.Errorf("expected ErrEventShareExpired, got %v", err)
	}
}

func TestSubscribeToShareCommand_Execute_ownEvent_returns403error(t *testing.T) {
	ctrl := gomock.NewController(t)
	ownerID := uuid.New()
	ev := newTestEvent(ownerID)
	share := newTestShare(ev.ID, false)

	c := buildCommand(ctrl,
		func(r *eventsharemock.MockRepository) {
			r.EXPECT().FindByToken(gomock.Any(), testToken).Return(share, nil)
		},
		findByIDReturning(ev),
		nil,
	)

	_, err := c.Execute(context.Background(), eventsubscriptionuc.SubscribeToShareInput{Token: testToken, UserID: ownerID})

	if !errors.Is(err, eventsubscription.ErrCannotSubscribeToOwnEvent) {
		t.Errorf("expected ErrCannotSubscribeToOwnEvent, got %v", err)
	}
}

func TestSubscribeToShareCommand_Execute_alreadySubscribed_returnsExistingNotCreated(t *testing.T) {
	ctrl := gomock.NewController(t)
	ownerID := uuid.New()
	viewerID := uuid.New()
	ev := newTestEvent(ownerID)
	share := newTestShare(ev.ID, false)
	existing := eventsubscription.New(uuid.New(), ev.ID, viewerID, time.Now())

	c := buildCommand(ctrl,
		func(r *eventsharemock.MockRepository) {
			r.EXPECT().FindByToken(gomock.Any(), testToken).Return(share, nil)
		},
		findByIDReturning(ev),
		func(r *subscriptionmock.MockRepository) {
			r.EXPECT().FindByEventAndUserID(gomock.Any(), ev.ID, viewerID).Return(existing, nil)
		},
	)

	out, err := c.Execute(context.Background(), eventsubscriptionuc.SubscribeToShareInput{
		Token: testToken, UserID: viewerID,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Created {
		t.Error("expected Created=false for an already-subscribed pair")
	}
	if out.SubscriptionID != existing.ID() {
		t.Errorf("SubscriptionID mismatch: got %v, want %v", out.SubscriptionID, existing.ID())
	}
	if out.EventID != ev.ID {
		t.Errorf("EventID mismatch: got %v, want %v", out.EventID, ev.ID)
	}
}

func TestSubscribeToShareCommand_Execute_notSubscribed_createsAndReturnsCreated(t *testing.T) {
	ctrl := gomock.NewController(t)
	ownerID := uuid.New()
	viewerID := uuid.New()
	ev := newTestEvent(ownerID)
	share := newTestShare(ev.ID, false)
	created := eventsubscription.New(uuid.New(), ev.ID, viewerID, time.Now())

	c := buildCommand(ctrl,
		func(r *eventsharemock.MockRepository) {
			r.EXPECT().FindByToken(gomock.Any(), testToken).Return(share, nil)
		},
		findByIDReturning(ev),
		func(r *subscriptionmock.MockRepository) {
			r.EXPECT().FindByEventAndUserID(gomock.Any(), ev.ID, viewerID).
				Return(nil, eventsubscription.ErrEventSubscriptionNotFound)
			r.EXPECT().Create(gomock.Any(), subMatcher{eventID: ev.ID, userID: viewerID}).Return(created, nil)
		},
	)

	out, err := c.Execute(context.Background(), eventsubscriptionuc.SubscribeToShareInput{
		Token: testToken, UserID: viewerID,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !out.Created {
		t.Error("expected Created=true for a new subscription")
	}
	if out.SubscriptionID != created.ID() {
		t.Errorf("SubscriptionID mismatch: got %v, want %v", out.SubscriptionID, created.ID())
	}
}

func TestSubscribeToShareCommand_Execute_createRaceAlreadyExists_returnsExistingNotCreated(t *testing.T) {
	ctrl := gomock.NewController(t)
	ownerID := uuid.New()
	viewerID := uuid.New()
	ev := newTestEvent(ownerID)
	share := newTestShare(ev.ID, false)
	existing := eventsubscription.New(uuid.New(), ev.ID, viewerID, time.Now())

	calls := 0
	c := buildCommand(ctrl,
		func(r *eventsharemock.MockRepository) {
			r.EXPECT().FindByToken(gomock.Any(), testToken).Return(share, nil)
		},
		findByIDReturning(ev),
		func(r *subscriptionmock.MockRepository) {
			r.EXPECT().FindByEventAndUserID(gomock.Any(), ev.ID, viewerID).DoAndReturn(
				func(context.Context, uuid.UUID, uuid.UUID) (eventsubscription.EventSubscription, error) {
					calls++
					if calls == 1 {
						return nil, eventsubscription.ErrEventSubscriptionNotFound
					}
					return existing, nil
				},
			).Times(2)
			r.EXPECT().Create(gomock.Any(), subMatcher{eventID: ev.ID, userID: viewerID}).
				Return(nil, eventsubscription.ErrAlreadySubscribed)
		},
	)

	out, err := c.Execute(context.Background(), eventsubscriptionuc.SubscribeToShareInput{
		Token: testToken, UserID: viewerID,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Created {
		t.Error("expected Created=false when Create loses a concurrent race")
	}
	if out.SubscriptionID != existing.ID() {
		t.Errorf("SubscriptionID mismatch: got %v, want %v", out.SubscriptionID, existing.ID())
	}
}

func TestSubscribeToShareCommand_Execute_findByEventAndUserIDError_propagates(t *testing.T) {
	ctrl := gomock.NewController(t)
	ownerID := uuid.New()
	viewerID := uuid.New()
	ev := newTestEvent(ownerID)
	share := newTestShare(ev.ID, false)

	c := buildCommand(ctrl,
		func(r *eventsharemock.MockRepository) {
			r.EXPECT().FindByToken(gomock.Any(), testToken).Return(share, nil)
		},
		findByIDReturning(ev),
		func(r *subscriptionmock.MockRepository) {
			r.EXPECT().FindByEventAndUserID(gomock.Any(), ev.ID, viewerID).Return(nil, errDBFailure)
		},
	)

	_, err := c.Execute(context.Background(), eventsubscriptionuc.SubscribeToShareInput{
		Token: testToken, UserID: viewerID,
	})

	if !errors.Is(err, errDBFailure) {
		t.Errorf("expected errDBFailure to propagate, got %v", err)
	}
}

func TestSubscribeToShareCommand_Execute_createError_propagates(t *testing.T) {
	ctrl := gomock.NewController(t)
	ownerID := uuid.New()
	viewerID := uuid.New()
	ev := newTestEvent(ownerID)
	share := newTestShare(ev.ID, false)

	c := buildCommand(ctrl,
		func(r *eventsharemock.MockRepository) {
			r.EXPECT().FindByToken(gomock.Any(), testToken).Return(share, nil)
		},
		findByIDReturning(ev),
		func(r *subscriptionmock.MockRepository) {
			r.EXPECT().FindByEventAndUserID(gomock.Any(), ev.ID, viewerID).
				Return(nil, eventsubscription.ErrEventSubscriptionNotFound)
			r.EXPECT().Create(gomock.Any(), subMatcher{eventID: ev.ID, userID: viewerID}).Return(nil, errDBFailure)
		},
	)

	_, err := c.Execute(context.Background(), eventsubscriptionuc.SubscribeToShareInput{
		Token: testToken, UserID: viewerID,
	})

	if !errors.Is(err, errDBFailure) {
		t.Errorf("expected errDBFailure to propagate, got %v", err)
	}
}
