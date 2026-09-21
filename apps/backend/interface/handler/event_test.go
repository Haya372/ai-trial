package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/Haya372/ai-trial/backend/domain"
	domainevent "github.com/Haya372/ai-trial/backend/domain/event"
	api "github.com/Haya372/ai-trial/backend/interface/api/generated"
	"github.com/Haya372/ai-trial/backend/interface/ctxkey"
	"github.com/Haya372/ai-trial/backend/interface/handler"
	eventuc "github.com/Haya372/ai-trial/backend/usecase/event"
)

const (
	testEventTitle = "Meeting"
	testKeyTitle   = "title"
	testKeyStartAt = "startAt"
	testKeyEndAt   = "endAt"
)

type stubUpdateEventExec struct {
	fn func(context.Context, uuid.UUID, eventuc.UpdateEventInput) (domainevent.Event, error)
}

func (s *stubUpdateEventExec) Execute(
	ctx context.Context, userID uuid.UUID, in eventuc.UpdateEventInput,
) (domainevent.Event, error) {
	if s.fn == nil {
		return nil, nil
	}
	return s.fn(ctx, userID, in)
}

const validUpdateEventBody = `{"title":"x","startAt":"2026-09-01T00:00:00Z","endAt":"2026-09-01T01:00:00Z"}`

func putEventRequest(t *testing.T, id, body string) *http.Request {
	t.Helper()
	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodPut,
		"/events/"+id,
		bytes.NewBufferString(body),
	)
	if err != nil {
		t.Fatalf("create request: %v", err)
	}
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", id)
	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
}

func putEventRequestAsUser(t *testing.T, id, body string) *http.Request {
	t.Helper()
	req := putEventRequest(t, id, body)
	u := newStubUser(uuid.New(), "user@example.com", "User")
	return req.WithContext(context.WithValue(req.Context(), ctxkey.User, u))
}

type stubListEventsExec struct {
	fn func(context.Context, uuid.UUID, eventuc.ListEventsInput) ([]eventuc.EventReadModel, error)
}

func (s *stubListEventsExec) Execute(
	ctx context.Context,
	userID uuid.UUID,
	in eventuc.ListEventsInput,
) ([]eventuc.EventReadModel, error) {
	if s.fn == nil {
		return nil, nil
	}
	return s.fn(ctx, userID, in)
}

func getEventsRequest(t *testing.T, startDate, endDate string) *http.Request {
	t.Helper()
	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodGet,
		"/events?startDate="+startDate+"&endDate="+endDate,
		nil,
	)
	if err != nil {
		t.Fatalf("create request: %v", err)
	}
	return req
}

func TestEventHandler_GetEvents_Success(t *testing.T) {
	now := time.Now().UTC()
	userID := uuid.New()

	stub := &stubListEventsExec{
		fn: func(_ context.Context, uid uuid.UUID, _ eventuc.ListEventsInput) ([]eventuc.EventReadModel, error) {
			if uid != userID {
				t.Errorf("userID mismatch: got %v, want %v", uid, userID)
			}
			desc := "Team sync"
			return []eventuc.EventReadModel{
				{
					ID:          uuid.New(),
					Title:       testEventTitle,
					Description: desc,
					StartAt:     now,
					EndAt:       now.Add(time.Hour),
				},
			}, nil
		},
	}

	h := handler.NewEventHandler(stub, &stubCreateEventExec{}, &stubUpdateEventExec{}, testLogger)

	req := getEventsRequest(t, "2026-09-01", "2026-09-30")
	req = req.WithContext(context.WithValue(req.Context(), ctxkey.User, newStubUser(userID, "user@example.com", "User")))
	w := httptest.NewRecorder()

	h.GetEvents(w, req, api.GetEventsParams{
		StartDate: mustParseDate(t, "2026-09-01"),
		EndDate:   mustParseDate(t, "2026-09-30"),
	})

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Events []struct {
			Title string `json:"title"`
		} `json:"events"`
	}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(resp.Events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(resp.Events))
	}
	if resp.Events[0].Title != testEventTitle {
		t.Errorf("title mismatch: got %q", resp.Events[0].Title)
	}
}

func TestEventHandler_GetEvents_Unauthorized(t *testing.T) {
	stub := &stubListEventsExec{}
	h := handler.NewEventHandler(stub, &stubCreateEventExec{}, &stubUpdateEventExec{}, testLogger)

	req := getEventsRequest(t, "2026-09-01", "2026-09-30")
	w := httptest.NewRecorder()

	h.GetEvents(w, req, api.GetEventsParams{
		StartDate: mustParseDate(t, "2026-09-01"),
		EndDate:   mustParseDate(t, "2026-09-30"),
	})

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestEventHandler_GetEvents_InternalError(t *testing.T) {
	userID := uuid.New()
	stub := &stubListEventsExec{
		fn: func(_ context.Context, _ uuid.UUID, _ eventuc.ListEventsInput) ([]eventuc.EventReadModel, error) {
			return nil, errInternal
		},
	}
	h := handler.NewEventHandler(stub, &stubCreateEventExec{}, &stubUpdateEventExec{}, testLogger)

	req := getEventsRequest(t, "2026-09-01", "2026-09-30")
	req = req.WithContext(context.WithValue(req.Context(), ctxkey.User, newStubUser(userID, "user@example.com", "User")))
	w := httptest.NewRecorder()

	h.GetEvents(w, req, api.GetEventsParams{
		StartDate: mustParseDate(t, "2026-09-01"),
		EndDate:   mustParseDate(t, "2026-09-30"),
	})

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

// --- CreateEvent tests ---

type stubCreateEventExec struct {
	fn func(context.Context, uuid.UUID, eventuc.CreateEventInput) (eventuc.EventDetail, error)
}

func (s *stubCreateEventExec) Execute(
	ctx context.Context,
	userID uuid.UUID,
	in eventuc.CreateEventInput,
) (eventuc.EventDetail, error) {
	if s.fn == nil {
		return eventuc.EventDetail{}, nil
	}
	return s.fn(ctx, userID, in)
}

func createEventRequest(t *testing.T, body any) *http.Request {
	t.Helper()
	b, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}
	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodPost,
		"/events",
		bytes.NewReader(b),
	)
	if err != nil {
		t.Fatalf("create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	return req
}

func TestEventHandler_CreateEvent_Success_Returns201(t *testing.T) {
	userID := uuid.New()
	eventID := uuid.New()
	now := time.Now().UTC().Truncate(time.Second)

	stub := &stubCreateEventExec{
		fn: func(_ context.Context, uid uuid.UUID, in eventuc.CreateEventInput) (eventuc.EventDetail, error) {
			if uid != userID {
				t.Errorf("userID mismatch: got %v, want %v", uid, userID)
			}
			return eventuc.EventDetail{
				ID:          eventID,
				UserID:      userID,
				Title:       in.Title,
				Description: in.Description,
				StartAt:     in.StartAt,
				EndAt:       in.EndAt,
				Location:    in.Location,
				URL:         in.URL,
			}, nil
		},
	}

	h := handler.NewEventHandler(&stubListEventsExec{}, stub, &stubUpdateEventExec{}, testLogger)

	req := createEventRequest(t, map[string]any{
		testKeyTitle:   testEventTitle,
		"description":  "Team sync",
		testKeyStartAt: now.Format(time.RFC3339),
		testKeyEndAt:   now.Add(time.Hour).Format(time.RFC3339),
		"location":     "Tokyo",
		"url":          "https://example.com",
	})
	req = req.WithContext(context.WithValue(req.Context(), ctxkey.User, newStubUser(userID, "u@ex.com", "U")))
	w := httptest.NewRecorder()

	h.CreateEvent(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
	var resp struct {
		ID    string `json:"id"`
		Title string `json:"title"`
	}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Title != testEventTitle {
		t.Errorf("title mismatch: got %q", resp.Title)
	}
	if resp.ID != eventID.String() {
		t.Errorf("id mismatch: got %q, want %q", resp.ID, eventID.String())
	}
}

func TestEventHandler_CreateEvent_Unauthorized_Returns401(t *testing.T) {
	h := handler.NewEventHandler(&stubListEventsExec{}, &stubCreateEventExec{}, &stubUpdateEventExec{}, testLogger)

	req := createEventRequest(t, map[string]any{
		testKeyTitle:   testEventTitle,
		testKeyStartAt: time.Now().Format(time.RFC3339),
		testKeyEndAt:   time.Now().Add(time.Hour).Format(time.RFC3339),
	})
	w := httptest.NewRecorder()

	h.CreateEvent(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestEventHandler_CreateEvent_InvalidJSON_Returns400(t *testing.T) {
	userID := uuid.New()
	h := handler.NewEventHandler(&stubListEventsExec{}, &stubCreateEventExec{}, &stubUpdateEventExec{}, testLogger)

	req, _ := http.NewRequestWithContext(
		context.Background(),
		http.MethodPost,
		"/events",
		bytes.NewReader([]byte("not json")),
	)
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(context.WithValue(req.Context(), ctxkey.User, newStubUser(userID, "u@ex.com", "U")))
	w := httptest.NewRecorder()

	h.CreateEvent(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestEventHandler_CreateEvent_ValidationError_Returns400(t *testing.T) {
	userID := uuid.New()

	stub := &stubCreateEventExec{
		fn: func(_ context.Context, _ uuid.UUID, _ eventuc.CreateEventInput) (eventuc.EventDetail, error) {
			return eventuc.EventDetail{}, &domain.ValidationError{Details: []domain.ValidationDetail{
				{Field: "title", Code: "REQUIRED", Message: "title is required"},
			}}
		},
	}

	h := handler.NewEventHandler(&stubListEventsExec{}, stub, &stubUpdateEventExec{}, testLogger)

	req := createEventRequest(t, map[string]any{
		testKeyTitle:   "",
		testKeyStartAt: time.Now().Format(time.RFC3339),
		testKeyEndAt:   time.Now().Add(time.Hour).Format(time.RFC3339),
	})
	req = req.WithContext(context.WithValue(req.Context(), ctxkey.User, newStubUser(userID, "u@ex.com", "U")))
	w := httptest.NewRecorder()

	h.CreateEvent(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
	var body map[string]any
	_ = json.NewDecoder(w.Body).Decode(&body)
	if body["code"] != "VALIDATION_ERROR" {
		t.Errorf("expected code VALIDATION_ERROR, got %v", body["code"])
	}
}

func TestEventHandler_CreateEvent_InternalError_Returns500(t *testing.T) {
	userID := uuid.New()

	stub := &stubCreateEventExec{
		fn: func(_ context.Context, _ uuid.UUID, _ eventuc.CreateEventInput) (eventuc.EventDetail, error) {
			return eventuc.EventDetail{}, errInternal
		},
	}

	h := handler.NewEventHandler(&stubListEventsExec{}, stub, &stubUpdateEventExec{}, testLogger)

	req := createEventRequest(t, map[string]any{
		testKeyTitle:   testEventTitle,
		testKeyStartAt: time.Now().Format(time.RFC3339),
		testKeyEndAt:   time.Now().Add(time.Hour).Format(time.RFC3339),
	})
	req = req.WithContext(context.WithValue(req.Context(), ctxkey.User, newStubUser(userID, "u@ex.com", "U")))
	w := httptest.NewRecorder()

	h.CreateEvent(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

// --- UpdateEvent tests ---

func TestEventHandler_UpdateEvent_Success(t *testing.T) {
	userID := uuid.New()
	eventID := uuid.New()
	now := time.Now().UTC()

	updated, _ := domainevent.New(
		eventID, userID, "Updated title", "Updated desc", now, now.Add(time.Hour), "Tokyo", "https://example.com",
	)

	stub := &stubUpdateEventExec{
		fn: func(_ context.Context, uid uuid.UUID, in eventuc.UpdateEventInput) (domainevent.Event, error) {
			if uid != userID {
				t.Errorf("userID mismatch: got %v, want %v", uid, userID)
			}
			if in.ID != eventID {
				t.Errorf("eventID mismatch: got %v, want %v", in.ID, eventID)
			}
			return updated, nil
		},
	}
	h := handler.NewEventHandler(&stubListEventsExec{}, &stubCreateEventExec{}, stub, testLogger)

	body := `{"title":"Updated title","description":"Updated desc","startAt":"` +
		now.Format(time.RFC3339) + `","endAt":"` + now.Add(time.Hour).Format(time.RFC3339) +
		`","location":"Tokyo","url":"https://example.com"}`
	req := putEventRequest(t, eventID.String(), body)
	req = req.WithContext(context.WithValue(req.Context(), ctxkey.User, newStubUser(userID, "user@example.com", "User")))
	w := httptest.NewRecorder()

	h.UpdateEvent(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp api.EventResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Title != "Updated title" {
		t.Errorf("title mismatch: got %q", resp.Title)
	}
	if resp.Location == nil || *resp.Location != "Tokyo" {
		t.Errorf("location mismatch: got %v", resp.Location)
	}
}

func TestEventHandler_UpdateEvent_Unauthorized(t *testing.T) {
	h := handler.NewEventHandler(&stubListEventsExec{}, &stubCreateEventExec{}, &stubUpdateEventExec{}, testLogger)

	req := putEventRequest(t, uuid.New().String(), validUpdateEventBody)
	w := httptest.NewRecorder()

	h.UpdateEvent(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestEventHandler_UpdateEvent_InvalidIDFormat(t *testing.T) {
	h := handler.NewEventHandler(&stubListEventsExec{}, &stubCreateEventExec{}, &stubUpdateEventExec{}, testLogger)

	req := putEventRequestAsUser(t, "not-a-uuid", validUpdateEventBody)
	w := httptest.NewRecorder()

	h.UpdateEvent(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestEventHandler_UpdateEvent_InvalidBody(t *testing.T) {
	h := handler.NewEventHandler(&stubListEventsExec{}, &stubCreateEventExec{}, &stubUpdateEventExec{}, testLogger)

	req := putEventRequestAsUser(t, uuid.New().String(), `not-json`)
	w := httptest.NewRecorder()

	h.UpdateEvent(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestEventHandler_UpdateEvent_ValidationError_Returns400(t *testing.T) {
	stub := &stubUpdateEventExec{
		fn: func(_ context.Context, _ uuid.UUID, _ eventuc.UpdateEventInput) (domainevent.Event, error) {
			return nil, &domain.ValidationError{Details: []domain.ValidationDetail{
				{Field: "endAt", Code: "INVALID_DATE_RANGE", Message: "endAt must be after or equal to startAt"},
			}}
		},
	}
	h := handler.NewEventHandler(&stubListEventsExec{}, &stubCreateEventExec{}, stub, testLogger)

	body := `{"title":"x","startAt":"2026-09-01T01:00:00Z","endAt":"2026-09-01T00:00:00Z"}`
	req := putEventRequestAsUser(t, uuid.New().String(), body)
	w := httptest.NewRecorder()

	h.UpdateEvent(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestEventHandler_UpdateEvent_NotFound_Returns404(t *testing.T) {
	stub := &stubUpdateEventExec{
		fn: func(_ context.Context, _ uuid.UUID, _ eventuc.UpdateEventInput) (domainevent.Event, error) {
			return nil, domainevent.ErrEventNotFound
		},
	}
	h := handler.NewEventHandler(&stubListEventsExec{}, &stubCreateEventExec{}, stub, testLogger)

	req := putEventRequestAsUser(t, uuid.New().String(), validUpdateEventBody)
	w := httptest.NewRecorder()

	h.UpdateEvent(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d: %s", w.Code, w.Body.String())
	}
}

func TestEventHandler_UpdateEvent_InternalError_Returns500(t *testing.T) {
	stub := &stubUpdateEventExec{
		fn: func(_ context.Context, _ uuid.UUID, _ eventuc.UpdateEventInput) (domainevent.Event, error) {
			return nil, errInternal
		},
	}
	h := handler.NewEventHandler(&stubListEventsExec{}, &stubCreateEventExec{}, stub, testLogger)

	req := putEventRequestAsUser(t, uuid.New().String(), validUpdateEventBody)
	w := httptest.NewRecorder()

	h.UpdateEvent(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d: %s", w.Code, w.Body.String())
	}
}
