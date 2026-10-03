package eventshare_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"

	domaineventshare "github.com/Haya372/ai-trial/backend/domain/eventshare"
	eventsharemock "github.com/Haya372/ai-trial/backend/domain/eventshare/generated"
	"github.com/Haya372/ai-trial/backend/domain/eventsubscription"
	subscriptionmock "github.com/Haya372/ai-trial/backend/domain/eventsubscription/generated"
	eventuc "github.com/Haya372/ai-trial/backend/usecase/event"
	eventshareuc "github.com/Haya372/ai-trial/backend/usecase/eventshare"
)

const testToken = "tok"

func buildQuery(
	ctrl *gomock.Controller,
	shareRepoFn func(*eventsharemock.MockRepository),
	eventQueryFn func(*stubEventQueryService),
	subsRepoFn func(*subscriptionmock.MockRepository),
) *eventshareuc.GetShareByTokenQuery {
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
	return eventshareuc.NewGetShareByTokenQuery(loader, subsRepo)
}

func findByIDReturning(ev eventuc.EventReadModel) func(*stubEventQueryService) {
	return func(q *stubEventQueryService) {
		q.findByIDFn = func(_ context.Context, _ uuid.UUID) (eventuc.EventReadModel, error) {
			return ev, nil
		}
	}
}

func TestGetShareByTokenQuery_Execute_tokenNotFound_returns404error(t *testing.T) {
	ctrl := gomock.NewController(t)
	q := buildQuery(ctrl,
		func(r *eventsharemock.MockRepository) {
			r.EXPECT().FindByToken(gomock.Any(), testToken).Return(nil, domaineventshare.ErrEventShareNotFound)
		},
		nil, nil,
	)

	_, err := q.Execute(context.Background(), eventshareuc.GetShareByTokenInput{Token: testToken})

	if !errors.Is(err, domaineventshare.ErrEventShareNotFound) {
		t.Errorf("expected ErrEventShareNotFound, got %v", err)
	}
}

func TestGetShareByTokenQuery_Execute_tokenExpired_returns410error(t *testing.T) {
	ctrl := gomock.NewController(t)
	ownerID := uuid.New()
	ev := newTestEvent(ownerID)
	expiredShare := newTestShare(ev.ID, true)

	q := buildQuery(ctrl,
		func(r *eventsharemock.MockRepository) {
			r.EXPECT().FindByToken(gomock.Any(), testToken).Return(expiredShare, nil)
		},
		nil, nil,
	)

	_, err := q.Execute(context.Background(), eventshareuc.GetShareByTokenInput{Token: testToken})

	if !errors.Is(err, domaineventshare.ErrEventShareExpired) {
		t.Errorf("expected ErrEventShareExpired, got %v", err)
	}
}

func TestGetShareByTokenQuery_Execute_viewerNil_isOwnAndSubscribedFalse(t *testing.T) {
	ctrl := gomock.NewController(t)
	ownerID := uuid.New()
	ev := newTestEvent(ownerID)
	share := newTestShare(ev.ID, false)

	q := buildQuery(ctrl,
		func(r *eventsharemock.MockRepository) {
			r.EXPECT().FindByToken(gomock.Any(), testToken).Return(share, nil)
		},
		findByIDReturning(ev),
		nil,
	)

	out, err := q.Execute(context.Background(), eventshareuc.GetShareByTokenInput{Token: testToken, Viewer: nil})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.IsOwnEvent {
		t.Error("expected IsOwnEvent=false for nil viewer")
	}
	if out.IsSubscribed {
		t.Error("expected IsSubscribed=false for nil viewer")
	}
	if out.Title != ev.Title {
		t.Errorf("Title mismatch: got %q, want %q", out.Title, ev.Title)
	}
}

func TestGetShareByTokenQuery_Execute_viewerIsOwner(t *testing.T) {
	ctrl := gomock.NewController(t)
	ownerID := uuid.New()
	ev := newTestEvent(ownerID)
	share := newTestShare(ev.ID, false)
	viewer := newTestUser(ownerID)

	q := buildQuery(ctrl,
		func(r *eventsharemock.MockRepository) {
			r.EXPECT().FindByToken(gomock.Any(), testToken).Return(share, nil)
		},
		findByIDReturning(ev),
		func(r *subscriptionmock.MockRepository) {
			r.EXPECT().FindByEventAndUserID(gomock.Any(), ev.ID, ownerID).
				Return(nil, eventsubscription.ErrEventSubscriptionNotFound)
		},
	)

	out, err := q.Execute(context.Background(), eventshareuc.GetShareByTokenInput{Token: testToken, Viewer: viewer})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !out.IsOwnEvent {
		t.Error("expected IsOwnEvent=true for owner viewer")
	}
	if out.IsSubscribed {
		t.Error("expected IsSubscribed=false (owner has no subscription record)")
	}
}

func TestGetShareByTokenQuery_Execute_viewerIsSubscribed(t *testing.T) {
	ctrl := gomock.NewController(t)
	ownerID := uuid.New()
	viewerID := uuid.New()
	ev := newTestEvent(ownerID)
	share := newTestShare(ev.ID, false)
	viewer := newTestUser(viewerID)

	q := buildQuery(ctrl,
		func(r *eventsharemock.MockRepository) {
			r.EXPECT().FindByToken(gomock.Any(), testToken).Return(share, nil)
		},
		findByIDReturning(ev),
		func(r *subscriptionmock.MockRepository) {
			sub := eventsubscription.New(uuid.New(), ev.ID, viewerID, newTime())
			r.EXPECT().FindByEventAndUserID(gomock.Any(), ev.ID, viewerID).Return(sub, nil)
		},
	)

	out, err := q.Execute(context.Background(), eventshareuc.GetShareByTokenInput{Token: testToken, Viewer: viewer})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.IsOwnEvent {
		t.Error("expected IsOwnEvent=false for non-owner viewer")
	}
	if !out.IsSubscribed {
		t.Error("expected IsSubscribed=true for subscribed viewer")
	}
}

func TestGetShareByTokenQuery_Execute_viewerNotSubscribed(t *testing.T) {
	ctrl := gomock.NewController(t)
	ownerID := uuid.New()
	viewerID := uuid.New()
	ev := newTestEvent(ownerID)
	share := newTestShare(ev.ID, false)
	viewer := newTestUser(viewerID)

	q := buildQuery(ctrl,
		func(r *eventsharemock.MockRepository) {
			r.EXPECT().FindByToken(gomock.Any(), testToken).Return(share, nil)
		},
		findByIDReturning(ev),
		func(r *subscriptionmock.MockRepository) {
			r.EXPECT().FindByEventAndUserID(gomock.Any(), ev.ID, viewerID).
				Return(nil, eventsubscription.ErrEventSubscriptionNotFound)
		},
	)

	out, err := q.Execute(context.Background(), eventshareuc.GetShareByTokenInput{Token: testToken, Viewer: viewer})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.IsOwnEvent {
		t.Error("expected IsOwnEvent=false for non-owner viewer")
	}
	if out.IsSubscribed {
		t.Error("expected IsSubscribed=false for non-subscribed viewer")
	}
}

func TestGetShareByTokenQuery_Execute_subsRepoError_propagates(t *testing.T) {
	ctrl := gomock.NewController(t)
	ownerID := uuid.New()
	viewerID := uuid.New()
	ev := newTestEvent(ownerID)
	share := newTestShare(ev.ID, false)
	viewer := newTestUser(viewerID)

	q := buildQuery(ctrl,
		func(r *eventsharemock.MockRepository) {
			r.EXPECT().FindByToken(gomock.Any(), testToken).Return(share, nil)
		},
		findByIDReturning(ev),
		func(r *subscriptionmock.MockRepository) {
			r.EXPECT().FindByEventAndUserID(gomock.Any(), ev.ID, viewerID).Return(nil, errDBFailure)
		},
	)

	_, err := q.Execute(context.Background(), eventshareuc.GetShareByTokenInput{Token: testToken, Viewer: viewer})

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, errDBFailure) {
		t.Errorf("expected errDBFailure to propagate, got %v", err)
	}
}
