package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/Haya372/ai-trial/backend/domain"
	domainevent "github.com/Haya372/ai-trial/backend/domain/event"
	domaineventshare "github.com/Haya372/ai-trial/backend/domain/eventshare"
	"github.com/Haya372/ai-trial/backend/interface/ctxkey"
	"github.com/Haya372/ai-trial/backend/interface/handler"
	eventshareuc "github.com/Haya372/ai-trial/backend/usecase/eventshare"
)

type stubCreateShareExec struct {
	fn func(context.Context, uuid.UUID, eventshareuc.CreateShareInput) (*eventshareuc.CreateShareResult, error)
}

func (s *stubCreateShareExec) Execute(
	ctx context.Context, userID uuid.UUID, in eventshareuc.CreateShareInput,
) (*eventshareuc.CreateShareResult, error) {
	if s.fn == nil {
		return nil, nil
	}
	return s.fn(ctx, userID, in)
}

func createShareRequest(t *testing.T, id, body string) *http.Request {
	t.Helper()
	var reader *bytes.Buffer
	if body == "" {
		reader = bytes.NewBufferString("")
	} else {
		reader = bytes.NewBufferString(body)
	}
	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodPost,
		"/events/"+id+"/shares",
		reader,
	)
	if err != nil {
		t.Fatalf("create request: %v", err)
	}
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", id)
	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
}

func createShareRequestAsUser(t *testing.T, id, body string) *http.Request {
	t.Helper()
	req := createShareRequest(t, id, body)
	u := newStubUser(uuid.New(), "sharer@example.com", "Sharer")
	return req.WithContext(context.WithValue(req.Context(), ctxkey.User, u))
}

func newTestShare(t *testing.T, eventID uuid.UUID, expiresAt time.Time) domaineventshare.EventShare {
	t.Helper()
	s, err := domaineventshare.New(uuid.New(), eventID, domaineventshare.HashToken("plaintext-token"), expiresAt)
	if err != nil {
		t.Fatalf("build test share: %v", err)
	}
	return s
}

func TestEventShareHandler_CreateShare_Unauthenticated_Returns401(t *testing.T) {
	h := handler.NewEventShareHandler(&stubCreateShareExec{}, slog.New(slog.DiscardHandler))

	req := createShareRequest(t, uuid.New().String(), "")
	w := httptest.NewRecorder()
	h.CreateShare(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", w.Code, w.Body.String())
	}
}

func TestEventShareHandler_CreateShare_InvalidEventID_Returns400(t *testing.T) {
	h := handler.NewEventShareHandler(&stubCreateShareExec{}, slog.New(slog.DiscardHandler))

	req := createShareRequestAsUser(t, "not-a-uuid", "")
	w := httptest.NewRecorder()
	h.CreateShare(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestEventShareHandler_CreateShare_NoBody_DefaultsExpiresAtToNilInput(t *testing.T) {
	eventID := uuid.New()
	expiresAt := time.Now().UTC().Add(time.Hour)
	var gotExpiresAt *time.Time
	stub := &stubCreateShareExec{
		fn: func(_ context.Context, _ uuid.UUID, in eventshareuc.CreateShareInput) (*eventshareuc.CreateShareResult, error) {
			gotExpiresAt = in.ExpiresAt
			return &eventshareuc.CreateShareResult{
				Share: newTestShare(t, eventID, expiresAt),
				Token: "plaintext-token",
			}, nil
		},
	}
	h := handler.NewEventShareHandler(stub, slog.New(slog.DiscardHandler))

	req := createShareRequestAsUser(t, eventID.String(), "")
	w := httptest.NewRecorder()
	h.CreateShare(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
	if gotExpiresAt != nil {
		t.Errorf("expected nil ExpiresAt override, got %v", gotExpiresAt)
	}

	var resp struct {
		URL       string    `json:"url"`
		ExpiresAt time.Time `json:"expiresAt"`
	}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.URL != "/share/plaintext-token" {
		t.Errorf("url mismatch: got %q", resp.URL)
	}
	if !resp.ExpiresAt.Equal(expiresAt) {
		t.Errorf("expiresAt mismatch: got %v, want %v", resp.ExpiresAt, expiresAt)
	}
}

func TestEventShareHandler_CreateShare_WithExpiresAtBody_PassesExpiresAtToExecutor(t *testing.T) {
	eventID := uuid.New()
	override := time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)
	var gotExpiresAt *time.Time
	stub := &stubCreateShareExec{
		fn: func(_ context.Context, _ uuid.UUID, in eventshareuc.CreateShareInput) (*eventshareuc.CreateShareResult, error) {
			gotExpiresAt = in.ExpiresAt
			return &eventshareuc.CreateShareResult{
				Share: newTestShare(t, eventID, override),
				Token: "plaintext-token",
			}, nil
		},
	}
	h := handler.NewEventShareHandler(stub, slog.New(slog.DiscardHandler))

	body := `{"expiresAt":"2026-12-31T00:00:00Z"}`
	req := createShareRequestAsUser(t, eventID.String(), body)
	w := httptest.NewRecorder()
	h.CreateShare(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
	if gotExpiresAt == nil || !gotExpiresAt.Equal(override) {
		t.Errorf("expiresAt mismatch: got %v, want %v", gotExpiresAt, override)
	}
}

func TestEventShareHandler_CreateShare_InvalidBody_Returns400(t *testing.T) {
	h := handler.NewEventShareHandler(&stubCreateShareExec{}, slog.New(slog.DiscardHandler))

	req := createShareRequestAsUser(t, uuid.New().String(), `{"expiresAt":`)
	w := httptest.NewRecorder()
	h.CreateShare(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestEventShareHandler_CreateShare_EventNotFound_Returns404(t *testing.T) {
	stub := &stubCreateShareExec{
		fn: func(context.Context, uuid.UUID, eventshareuc.CreateShareInput) (*eventshareuc.CreateShareResult, error) {
			return nil, domainevent.ErrEventNotFound
		},
	}
	h := handler.NewEventShareHandler(stub, slog.New(slog.DiscardHandler))

	req := createShareRequestAsUser(t, uuid.New().String(), "")
	w := httptest.NewRecorder()
	h.CreateShare(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
	}
}

func TestEventShareHandler_CreateShare_Forbidden_Returns403(t *testing.T) {
	stub := &stubCreateShareExec{
		fn: func(context.Context, uuid.UUID, eventshareuc.CreateShareInput) (*eventshareuc.CreateShareResult, error) {
			return nil, domainevent.ErrEventForbidden
		},
	}
	h := handler.NewEventShareHandler(stub, slog.New(slog.DiscardHandler))

	req := createShareRequestAsUser(t, uuid.New().String(), "")
	w := httptest.NewRecorder()
	h.CreateShare(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d: %s", w.Code, w.Body.String())
	}
}

func TestEventShareHandler_CreateShare_ValidationError_Returns400(t *testing.T) {
	stub := &stubCreateShareExec{
		fn: func(context.Context, uuid.UUID, eventshareuc.CreateShareInput) (*eventshareuc.CreateShareResult, error) {
			return nil, &domain.ValidationError{Details: []domain.ValidationDetail{
				{Field: "expiresAt", Code: "BEFORE_EVENT_START", Message: "too early"},
			}}
		},
	}
	h := handler.NewEventShareHandler(stub, slog.New(slog.DiscardHandler))

	req := createShareRequestAsUser(t, uuid.New().String(), "")
	w := httptest.NewRecorder()
	h.CreateShare(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestEventShareHandler_CreateShare_InternalError_Returns500(t *testing.T) {
	stub := &stubCreateShareExec{
		fn: func(context.Context, uuid.UUID, eventshareuc.CreateShareInput) (*eventshareuc.CreateShareResult, error) {
			return nil, context.DeadlineExceeded
		},
	}
	h := handler.NewEventShareHandler(stub, slog.New(slog.DiscardHandler))

	req := createShareRequestAsUser(t, uuid.New().String(), "")
	w := httptest.NewRecorder()
	h.CreateShare(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d: %s", w.Code, w.Body.String())
	}
}
