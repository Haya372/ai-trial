package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/Haya372/ai-trial/backend/domain"
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

	h := handler.NewEventHandler(stub, &stubCreateEventExec{}, testLogger)

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
	h := handler.NewEventHandler(stub, &stubCreateEventExec{}, testLogger)

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
	h := handler.NewEventHandler(stub, &stubCreateEventExec{}, testLogger)

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

	h := handler.NewEventHandler(&stubListEventsExec{}, stub, testLogger)

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
	h := handler.NewEventHandler(&stubListEventsExec{}, &stubCreateEventExec{}, testLogger)

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
	h := handler.NewEventHandler(&stubListEventsExec{}, &stubCreateEventExec{}, testLogger)

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

	h := handler.NewEventHandler(&stubListEventsExec{}, stub, testLogger)

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

	h := handler.NewEventHandler(&stubListEventsExec{}, stub, testLogger)

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
