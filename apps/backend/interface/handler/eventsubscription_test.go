package handler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	domaineventshare "github.com/Haya372/ai-trial/backend/domain/eventshare"
	domaineventsubscription "github.com/Haya372/ai-trial/backend/domain/eventsubscription"
	"github.com/Haya372/ai-trial/backend/interface/ctxkey"
	"github.com/Haya372/ai-trial/backend/interface/handler"
	eventsubscriptionuc "github.com/Haya372/ai-trial/backend/usecase/eventsubscription"
)

type stubSubscribeToShareExec struct {
	fn func(context.Context, eventsubscriptionuc.SubscribeToShareInput) (eventsubscriptionuc.SubscribeToShareOutput, error)
}

func (s *stubSubscribeToShareExec) Execute(
	ctx context.Context, in eventsubscriptionuc.SubscribeToShareInput,
) (eventsubscriptionuc.SubscribeToShareOutput, error) {
	if s.fn == nil {
		return eventsubscriptionuc.SubscribeToShareOutput{}, nil
	}
	return s.fn(ctx, in)
}

func subscribeRequest(t *testing.T, token string) *http.Request {
	t.Helper()
	req, err := http.NewRequestWithContext(
		context.Background(), http.MethodPost, "/shares/"+token+"/subscriptions", nil,
	)
	if err != nil {
		t.Fatalf("create request: %v", err)
	}
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("token", token)
	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
}

func subscribeRequestAsUser(t *testing.T, token string, userID uuid.UUID) *http.Request {
	t.Helper()
	req := subscribeRequest(t, token)
	u := newStubUser(userID, "user@example.com", "User")
	return req.WithContext(context.WithValue(req.Context(), ctxkey.User, u))
}

func TestEventSubscriptionHandler_SubscribeToShare_unauthenticated_401(t *testing.T) {
	stub := &stubSubscribeToShareExec{}
	h := handler.NewEventSubscriptionHandler(stub, testLogger)

	req := subscribeRequest(t, "tok")
	w := httptest.NewRecorder()
	h.SubscribeToShare(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", w.Code, w.Body.String())
	}
}

func TestEventSubscriptionHandler_SubscribeToShare_created_201(t *testing.T) {
	userID := uuid.New()
	now := time.Now().UTC().Truncate(time.Second)
	subID := uuid.New()
	eventID := uuid.New()

	stub := &stubSubscribeToShareExec{
		fn: func(_ context.Context, in eventsubscriptionuc.SubscribeToShareInput) (
			eventsubscriptionuc.SubscribeToShareOutput, error,
		) {
			if in.UserID != userID {
				t.Errorf("expected UserID %v, got %v", userID, in.UserID)
			}
			return eventsubscriptionuc.SubscribeToShareOutput{
				SubscriptionID: subID, EventID: eventID, CreatedAt: now, Created: true,
			}, nil
		},
	}
	h := handler.NewEventSubscriptionHandler(stub, testLogger)

	req := subscribeRequestAsUser(t, "tok", userID)
	w := httptest.NewRecorder()
	h.SubscribeToShare(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
	var resp map[string]any
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp["id"] != subID.String() {
		t.Errorf("id mismatch: got %v, want %q", resp["id"], subID.String())
	}
	if resp["eventId"] != eventID.String() {
		t.Errorf("eventId mismatch: got %v, want %q", resp["eventId"], eventID.String())
	}
}

func TestEventSubscriptionHandler_SubscribeToShare_alreadySubscribed_200(t *testing.T) {
	userID := uuid.New()
	stub := &stubSubscribeToShareExec{
		fn: func(_ context.Context, _ eventsubscriptionuc.SubscribeToShareInput) (
			eventsubscriptionuc.SubscribeToShareOutput, error,
		) {
			return eventsubscriptionuc.SubscribeToShareOutput{
				SubscriptionID: uuid.New(), EventID: uuid.New(), CreatedAt: time.Now(), Created: false,
			}, nil
		},
	}
	h := handler.NewEventSubscriptionHandler(stub, testLogger)

	req := subscribeRequestAsUser(t, "tok", userID)
	w := httptest.NewRecorder()
	h.SubscribeToShare(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestEventSubscriptionHandler_SubscribeToShare_tokenNotFound_404(t *testing.T) {
	stub := &stubSubscribeToShareExec{
		fn: func(_ context.Context, _ eventsubscriptionuc.SubscribeToShareInput) (
			eventsubscriptionuc.SubscribeToShareOutput, error,
		) {
			return eventsubscriptionuc.SubscribeToShareOutput{}, domaineventshare.ErrEventShareNotFound
		},
	}
	h := handler.NewEventSubscriptionHandler(stub, testLogger)

	req := subscribeRequestAsUser(t, "unknown", uuid.New())
	w := httptest.NewRecorder()
	h.SubscribeToShare(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
	}
}

func TestEventSubscriptionHandler_SubscribeToShare_tokenExpired_410(t *testing.T) {
	stub := &stubSubscribeToShareExec{
		fn: func(_ context.Context, _ eventsubscriptionuc.SubscribeToShareInput) (
			eventsubscriptionuc.SubscribeToShareOutput, error,
		) {
			return eventsubscriptionuc.SubscribeToShareOutput{}, domaineventshare.ErrEventShareExpired
		},
	}
	h := handler.NewEventSubscriptionHandler(stub, testLogger)

	req := subscribeRequestAsUser(t, "expired", uuid.New())
	w := httptest.NewRecorder()
	h.SubscribeToShare(w, req)

	if w.Code != http.StatusGone {
		t.Fatalf("expected 410, got %d: %s", w.Code, w.Body.String())
	}
}

func TestEventSubscriptionHandler_SubscribeToShare_ownEvent_403(t *testing.T) {
	stub := &stubSubscribeToShareExec{
		fn: func(_ context.Context, _ eventsubscriptionuc.SubscribeToShareInput) (
			eventsubscriptionuc.SubscribeToShareOutput, error,
		) {
			return eventsubscriptionuc.SubscribeToShareOutput{}, domaineventsubscription.ErrCannotSubscribeToOwnEvent
		},
	}
	h := handler.NewEventSubscriptionHandler(stub, testLogger)

	req := subscribeRequestAsUser(t, "tok", uuid.New())
	w := httptest.NewRecorder()
	h.SubscribeToShare(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d: %s", w.Code, w.Body.String())
	}
}

func TestEventSubscriptionHandler_SubscribeToShare_internalError_500(t *testing.T) {
	stub := &stubSubscribeToShareExec{
		fn: func(_ context.Context, _ eventsubscriptionuc.SubscribeToShareInput) (
			eventsubscriptionuc.SubscribeToShareOutput, error,
		) {
			return eventsubscriptionuc.SubscribeToShareOutput{}, errInternal
		},
	}
	h := handler.NewEventSubscriptionHandler(stub, testLogger)

	req := subscribeRequestAsUser(t, "tok", uuid.New())
	w := httptest.NewRecorder()
	h.SubscribeToShare(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d: %s", w.Code, w.Body.String())
	}
}
