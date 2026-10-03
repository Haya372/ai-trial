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

	domainevent "github.com/Haya372/ai-trial/backend/domain/event"
	domaineventshare "github.com/Haya372/ai-trial/backend/domain/eventshare"
	"github.com/Haya372/ai-trial/backend/interface/ctxkey"
	"github.com/Haya372/ai-trial/backend/interface/handler"
	eventshareuc "github.com/Haya372/ai-trial/backend/usecase/eventshare"
)

type stubGetShareByTokenExec struct {
	fn func(context.Context, eventshareuc.GetShareByTokenInput) (eventshareuc.GetShareByTokenOutput, error)
}

func (s *stubGetShareByTokenExec) Execute(
	ctx context.Context, in eventshareuc.GetShareByTokenInput,
) (eventshareuc.GetShareByTokenOutput, error) {
	if s.fn == nil {
		return eventshareuc.GetShareByTokenOutput{}, nil
	}
	return s.fn(ctx, in)
}

func shareTokenRequest(t *testing.T, token string) *http.Request {
	t.Helper()
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "/shares/"+token, nil)
	if err != nil {
		t.Fatalf("create request: %v", err)
	}
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("token", token)
	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
}

func shareTokenRequestAsUser(t *testing.T, token string, userID uuid.UUID) *http.Request {
	t.Helper()
	req := shareTokenRequest(t, token)
	u := newStubUser(userID, "user@example.com", "User")
	return req.WithContext(context.WithValue(req.Context(), ctxkey.User, u))
}

func validShareOutput() eventshareuc.GetShareByTokenOutput {
	now := time.Now().UTC()
	return eventshareuc.GetShareByTokenOutput{
		Title:   "Test Event",
		StartAt: now,
		EndAt:   now.Add(time.Hour),
	}
}

func TestShareHandler_GetShareByToken_success_200(t *testing.T) {
	out := validShareOutput()
	stub := &stubGetShareByTokenExec{
		fn: func(_ context.Context, _ eventshareuc.GetShareByTokenInput) (eventshareuc.GetShareByTokenOutput, error) {
			return out, nil
		},
	}
	h := handler.NewShareHandler(stub, testLogger)

	req := shareTokenRequest(t, "sometoken")
	w := httptest.NewRecorder()
	h.GetShareByToken(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp map[string]any
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp["title"] != out.Title {
		t.Errorf("title mismatch: got %v, want %q", resp["title"], out.Title)
	}
	if _, ok := resp["isOwnEvent"]; !ok {
		t.Error("expected isOwnEvent in response")
	}
	if _, ok := resp["isSubscribed"]; !ok {
		t.Error("expected isSubscribed in response")
	}
}

func TestShareHandler_GetShareByToken_notFound_404(t *testing.T) {
	stub := &stubGetShareByTokenExec{
		fn: func(_ context.Context, _ eventshareuc.GetShareByTokenInput) (eventshareuc.GetShareByTokenOutput, error) {
			return eventshareuc.GetShareByTokenOutput{}, domaineventshare.ErrEventShareNotFound
		},
	}
	h := handler.NewShareHandler(stub, testLogger)

	req := shareTokenRequest(t, "unknown")
	w := httptest.NewRecorder()
	h.GetShareByToken(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
	}
}

func TestShareHandler_GetShareByToken_expired_410(t *testing.T) {
	stub := &stubGetShareByTokenExec{
		fn: func(_ context.Context, _ eventshareuc.GetShareByTokenInput) (eventshareuc.GetShareByTokenOutput, error) {
			return eventshareuc.GetShareByTokenOutput{}, domaineventshare.ErrEventShareExpired
		},
	}
	h := handler.NewShareHandler(stub, testLogger)

	req := shareTokenRequest(t, "expiredtoken")
	w := httptest.NewRecorder()
	h.GetShareByToken(w, req)

	if w.Code != http.StatusGone {
		t.Fatalf("expected 410, got %d: %s", w.Code, w.Body.String())
	}
	var resp map[string]any
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp["code"] != "GONE" {
		t.Errorf("expected code GONE, got %v", resp["code"])
	}
}

func TestShareHandler_GetShareByToken_eventDeletedRace_returns404(t *testing.T) {
	stub := &stubGetShareByTokenExec{
		fn: func(_ context.Context, _ eventshareuc.GetShareByTokenInput) (eventshareuc.GetShareByTokenOutput, error) {
			return eventshareuc.GetShareByTokenOutput{}, domainevent.ErrEventNotFound
		},
	}
	h := handler.NewShareHandler(stub, testLogger)

	req := shareTokenRequest(t, "racytoken")
	w := httptest.NewRecorder()
	h.GetShareByToken(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
	}
}

func TestShareHandler_GetShareByToken_internalError_500(t *testing.T) {
	stub := &stubGetShareByTokenExec{
		fn: func(_ context.Context, _ eventshareuc.GetShareByTokenInput) (eventshareuc.GetShareByTokenOutput, error) {
			return eventshareuc.GetShareByTokenOutput{}, errInternal
		},
	}
	h := handler.NewShareHandler(stub, testLogger)

	req := shareTokenRequest(t, "tok")
	w := httptest.NewRecorder()
	h.GetShareByToken(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d: %s", w.Code, w.Body.String())
	}
}

func TestShareHandler_GetShareByToken_unauthenticated_isOwnAndSubscribedFalse(t *testing.T) {
	out := validShareOutput()
	out.IsOwnEvent = false
	out.IsSubscribed = false

	stub := &stubGetShareByTokenExec{
		fn: func(_ context.Context, in eventshareuc.GetShareByTokenInput) (eventshareuc.GetShareByTokenOutput, error) {
			if in.Viewer != nil {
				t.Error("expected nil viewer for unauthenticated request")
			}
			return out, nil
		},
	}
	h := handler.NewShareHandler(stub, testLogger)

	req := shareTokenRequest(t, "tok")
	w := httptest.NewRecorder()
	h.GetShareByToken(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp map[string]any
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp["isOwnEvent"] != false {
		t.Errorf("expected isOwnEvent=false for unauthenticated, got %v", resp["isOwnEvent"])
	}
	if resp["isSubscribed"] != false {
		t.Errorf("expected isSubscribed=false for unauthenticated, got %v", resp["isSubscribed"])
	}
}

func TestShareHandler_GetShareByToken_authenticatedOwner(t *testing.T) {
	userID := uuid.New()
	out := validShareOutput()
	out.IsOwnEvent = true
	out.IsSubscribed = false

	stub := &stubGetShareByTokenExec{
		fn: func(_ context.Context, in eventshareuc.GetShareByTokenInput) (eventshareuc.GetShareByTokenOutput, error) {
			if in.Viewer == nil {
				t.Error("expected non-nil viewer for authenticated request")
			}
			return out, nil
		},
	}
	h := handler.NewShareHandler(stub, testLogger)

	req := shareTokenRequestAsUser(t, "tok", userID)
	w := httptest.NewRecorder()
	h.GetShareByToken(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp map[string]any
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp["isOwnEvent"] != true {
		t.Errorf("expected isOwnEvent=true for owner, got %v", resp["isOwnEvent"])
	}
}

func TestShareHandler_GetShareByToken_authenticatedSubscribed(t *testing.T) {
	userID := uuid.New()
	out := validShareOutput()
	out.IsOwnEvent = false
	out.IsSubscribed = true

	stub := &stubGetShareByTokenExec{
		fn: func(_ context.Context, _ eventshareuc.GetShareByTokenInput) (eventshareuc.GetShareByTokenOutput, error) {
			return out, nil
		},
	}
	h := handler.NewShareHandler(stub, testLogger)

	req := shareTokenRequestAsUser(t, "tok", userID)
	w := httptest.NewRecorder()
	h.GetShareByToken(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp map[string]any
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp["isSubscribed"] != true {
		t.Errorf("expected isSubscribed=true for subscribed user, got %v", resp["isSubscribed"])
	}
}

func TestShareHandler_GetShareByToken_authenticatedNotSubscribed(t *testing.T) {
	userID := uuid.New()
	out := validShareOutput()
	out.IsOwnEvent = false
	out.IsSubscribed = false

	stub := &stubGetShareByTokenExec{
		fn: func(_ context.Context, _ eventshareuc.GetShareByTokenInput) (eventshareuc.GetShareByTokenOutput, error) {
			return out, nil
		},
	}
	h := handler.NewShareHandler(stub, testLogger)

	req := shareTokenRequestAsUser(t, "tok", userID)
	w := httptest.NewRecorder()
	h.GetShareByToken(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp map[string]any
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp["isSubscribed"] != false {
		t.Errorf("expected isSubscribed=false for non-subscribed user, got %v", resp["isSubscribed"])
	}
	if resp["isOwnEvent"] != false {
		t.Errorf("expected isOwnEvent=false for non-owner, got %v", resp["isOwnEvent"])
	}
}
