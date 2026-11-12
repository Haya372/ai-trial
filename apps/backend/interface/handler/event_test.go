package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
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
	testEventTitle    = "Meeting"
	testKeyTitle      = "title"
	testKeyStartAt    = "startAt"
	testKeyEndAt      = "endAt"
	testEventLocation = "Tokyo"
	testEventURL      = "https://example.com"
	testEventTagColor = "blue"
	testFieldTagIDs   = "tagIds"
)

type stubUpdateEventExec struct {
	fn func(context.Context, uuid.UUID, eventuc.UpdateEventInput) (eventuc.UpdateEventOutput, error)
}

func (s *stubUpdateEventExec) Execute(
	ctx context.Context, userID uuid.UUID, in eventuc.UpdateEventInput,
) (eventuc.UpdateEventOutput, error) {
	if s.fn == nil {
		return eventuc.UpdateEventOutput{}, nil
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

func getEventsRequest(t *testing.T, startAt, endAt string) *http.Request {
	t.Helper()
	q := url.Values{}
	q.Set("startAt", startAt)
	q.Set("endAt", endAt)
	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodGet,
		"/events?"+q.Encode(),
		nil,
	)
	if err != nil {
		t.Fatalf("create request: %v", err)
	}
	return req
}

// JST の 2026年9月（2026-09-01T00:00+09:00 〜 2026-10-01T00:00+09:00）を UTC で表した範囲。
const (
	testRangeStartAt = "2026-08-31T15:00:00Z"
	testRangeEndAt   = "2026-09-30T15:00:00Z"
)

func testGetEventsParams(t *testing.T) api.GetEventsParams {
	t.Helper()
	return api.GetEventsParams{
		StartAt: mustParseTime(t, testRangeStartAt),
		EndAt:   mustParseTime(t, testRangeEndAt),
	}
}

func mustParseTime(t *testing.T, s string) time.Time {
	t.Helper()
	parsed, err := time.Parse(time.RFC3339Nano, s)
	if err != nil {
		t.Fatalf("parse time %q: %v", s, err)
	}
	return parsed
}

func TestEventHandler_ServeHTTP_PassesUTCRangeToUsecase(t *testing.T) {
	userID := uuid.New()
	var got eventuc.ListEventsInput
	stub := &stubListEventsExec{
		fn: func(_ context.Context, _ uuid.UUID, in eventuc.ListEventsInput) ([]eventuc.EventReadModel, error) {
			got = in
			return nil, nil
		},
	}
	h := handler.NewEventHandler(stub, &stubCreateEventExec{}, &stubUpdateEventExec{}, &stubDeleteEventExec{}, testLogger)

	// クライアントは toISOString() でミリ秒付きの値を送る
	req := getEventsRequest(t, "2026-08-31T15:00:00.000Z", "2026-09-30T15:00:00.000Z")
	req = req.WithContext(context.WithValue(req.Context(), ctxkey.User, newStubUser(userID, "user@example.com", "User")))
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	if !got.StartDate.Equal(mustParseTime(t, testRangeStartAt)) {
		t.Errorf("StartDate mismatch: got %v, want %v", got.StartDate, testRangeStartAt)
	}
	if !got.EndDate.Equal(mustParseTime(t, testRangeEndAt)) {
		t.Errorf("EndDate mismatch: got %v, want %v", got.EndDate, testRangeEndAt)
	}
}

func TestEventHandler_ServeHTTP_InvalidRange_Returns400(t *testing.T) {
	tests := []struct {
		name  string
		query string
	}{
		{"missing startAt", "endAt=" + testRangeEndAt},
		{"missing endAt", "startAt=" + testRangeStartAt},
		{"date-only startAt", "startAt=2026-09-01&endAt=" + testRangeEndAt},
		{"date-only endAt", "startAt=" + testRangeStartAt + "&endAt=2026-09-30"},
		{"offset startAt", "startAt=" + url.QueryEscape("2026-09-01T00:00:00+09:00") + "&endAt=" + testRangeEndAt},
		{"offset endAt", "startAt=" + testRangeStartAt + "&endAt=" + url.QueryEscape("2026-10-01T00:00:00+09:00")},
		{"zero offset instead of Z", "startAt=" + url.QueryEscape("2026-08-31T15:00:00+00:00") + "&endAt=" + testRangeEndAt},
		{"endAt equals startAt", "startAt=" + testRangeStartAt + "&endAt=" + testRangeStartAt},
		{"endAt before startAt", "startAt=" + testRangeEndAt + "&endAt=" + testRangeStartAt},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			called := false
			stub := &stubListEventsExec{
				fn: func(_ context.Context, _ uuid.UUID, _ eventuc.ListEventsInput) ([]eventuc.EventReadModel, error) {
					called = true
					return nil, nil
				},
			}
			h := handler.NewEventHandler(
				stub, &stubCreateEventExec{}, &stubUpdateEventExec{}, &stubDeleteEventExec{}, testLogger,
			)

			req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "/events?"+tt.query, nil)
			if err != nil {
				t.Fatalf("create request: %v", err)
			}
			u := newStubUser(uuid.New(), "user@example.com", "User")
			req = req.WithContext(context.WithValue(req.Context(), ctxkey.User, u))
			w := httptest.NewRecorder()

			h.ServeHTTP(w, req)

			if w.Code != http.StatusBadRequest {
				t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
			}
			var body struct {
				Code string `json:"code"`
			}
			if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
				t.Fatalf("decode response: %v", err)
			}
			if body.Code != codeValidationError {
				t.Errorf("code mismatch: got %q, want %q", body.Code, codeValidationError)
			}
			if called {
				t.Error("usecase must not be called for invalid range")
			}
		})
	}
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
					ID:           uuid.New(),
					Title:        testEventTitle,
					Description:  desc,
					StartAt:      now,
					EndAt:        now.Add(time.Hour),
					Location:     testEventLocation,
					URL:          testEventURL,
					IsSubscribed: false,
				},
			}, nil
		},
	}

	h := handler.NewEventHandler(stub, &stubCreateEventExec{}, &stubUpdateEventExec{}, &stubDeleteEventExec{}, testLogger)

	req := getEventsRequest(t, testRangeStartAt, testRangeEndAt)
	req = req.WithContext(context.WithValue(req.Context(), ctxkey.User, newStubUser(userID, "user@example.com", "User")))
	w := httptest.NewRecorder()

	h.GetEvents(w, req, testGetEventsParams(t))

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Events []struct {
			Title        string  `json:"title"`
			Location     *string `json:"location"`
			URL          *string `json:"url"`
			IsSubscribed bool    `json:"isSubscribed"`
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
	if resp.Events[0].Location == nil || *resp.Events[0].Location != testEventLocation {
		t.Errorf("location mismatch: got %v", resp.Events[0].Location)
	}
	if resp.Events[0].URL == nil || *resp.Events[0].URL != testEventURL {
		t.Errorf("url mismatch: got %v", resp.Events[0].URL)
	}
	if resp.Events[0].IsSubscribed {
		t.Errorf("expected isSubscribed=false for own event, got true")
	}
}

func TestEventHandler_GetEvents_IncludesIsSubscribed(t *testing.T) {
	now := time.Now().UTC()
	userID := uuid.New()

	stub := &stubListEventsExec{
		fn: func(_ context.Context, _ uuid.UUID, _ eventuc.ListEventsInput) ([]eventuc.EventReadModel, error) {
			return []eventuc.EventReadModel{
				{ID: uuid.New(), Title: "My event", StartAt: now, EndAt: now.Add(time.Hour), IsSubscribed: false},
				{ID: uuid.New(), Title: "Shared event", StartAt: now, EndAt: now.Add(time.Hour), IsSubscribed: true},
			}, nil
		},
	}

	h := handler.NewEventHandler(stub, &stubCreateEventExec{}, &stubUpdateEventExec{}, &stubDeleteEventExec{}, testLogger)

	req := getEventsRequest(t, testRangeStartAt, testRangeEndAt)
	req = req.WithContext(context.WithValue(req.Context(), ctxkey.User, newStubUser(userID, "user@example.com", "User")))
	w := httptest.NewRecorder()

	h.GetEvents(w, req, testGetEventsParams(t))

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Events []struct {
			Title        string `json:"title"`
			IsSubscribed bool   `json:"isSubscribed"`
		} `json:"events"`
	}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(resp.Events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(resp.Events))
	}
	if resp.Events[0].IsSubscribed {
		t.Errorf("expected own event isSubscribed=false, got true")
	}
	if !resp.Events[1].IsSubscribed {
		t.Errorf("expected shared event isSubscribed=true, got false")
	}
}

func TestEventHandler_GetEvents_IncludesSubscriptionID(t *testing.T) {
	now := time.Now().UTC()
	userID := uuid.New()
	subscriptionID := uuid.New()

	stub := &stubListEventsExec{
		fn: func(_ context.Context, _ uuid.UUID, _ eventuc.ListEventsInput) ([]eventuc.EventReadModel, error) {
			return []eventuc.EventReadModel{
				{ID: uuid.New(), Title: "My event", StartAt: now, EndAt: now.Add(time.Hour), IsSubscribed: false},
				{
					ID: uuid.New(), Title: "Shared event", StartAt: now, EndAt: now.Add(time.Hour),
					IsSubscribed: true, SubscriptionID: &subscriptionID,
				},
			}, nil
		},
	}

	h := handler.NewEventHandler(stub, &stubCreateEventExec{}, &stubUpdateEventExec{}, &stubDeleteEventExec{}, testLogger)

	req := getEventsRequest(t, testRangeStartAt, testRangeEndAt)
	req = req.WithContext(context.WithValue(req.Context(), ctxkey.User, newStubUser(userID, "user@example.com", "User")))
	w := httptest.NewRecorder()

	h.GetEvents(w, req, testGetEventsParams(t))

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Events []struct {
			Title          string  `json:"title"`
			SubscriptionID *string `json:"subscriptionId"`
		} `json:"events"`
	}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(resp.Events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(resp.Events))
	}
	if resp.Events[0].SubscriptionID != nil {
		t.Errorf("expected own event subscriptionId=null, got %v", *resp.Events[0].SubscriptionID)
	}
	if resp.Events[1].SubscriptionID == nil || *resp.Events[1].SubscriptionID != subscriptionID.String() {
		t.Errorf("expected shared event subscriptionId=%v, got %v", subscriptionID, resp.Events[1].SubscriptionID)
	}
}

func TestEventHandler_GetEvents_Unauthorized(t *testing.T) {
	stub := &stubListEventsExec{}
	h := handler.NewEventHandler(stub, &stubCreateEventExec{}, &stubUpdateEventExec{}, &stubDeleteEventExec{}, testLogger)

	req := getEventsRequest(t, testRangeStartAt, testRangeEndAt)
	w := httptest.NewRecorder()

	h.GetEvents(w, req, testGetEventsParams(t))

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
	h := handler.NewEventHandler(stub, &stubCreateEventExec{}, &stubUpdateEventExec{}, &stubDeleteEventExec{}, testLogger)

	req := getEventsRequest(t, testRangeStartAt, testRangeEndAt)
	req = req.WithContext(context.WithValue(req.Context(), ctxkey.User, newStubUser(userID, "user@example.com", "User")))
	w := httptest.NewRecorder()

	h.GetEvents(w, req, testGetEventsParams(t))

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

// --- CreateEvent tests ---

type stubCreateEventExec struct {
	fn func(context.Context, uuid.UUID, eventuc.CreateEventInput) (eventuc.CreateEventOutput, error)
}

func (s *stubCreateEventExec) Execute(
	ctx context.Context,
	userID uuid.UUID,
	in eventuc.CreateEventInput,
) (eventuc.CreateEventOutput, error) {
	if s.fn == nil {
		return eventuc.CreateEventOutput{}, nil
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
		fn: func(_ context.Context, uid uuid.UUID, in eventuc.CreateEventInput) (eventuc.CreateEventOutput, error) {
			if uid != userID {
				t.Errorf("userID mismatch: got %v, want %v", uid, userID)
			}
			ev, err := domainevent.New(
				eventID, userID, in.Title, in.Description, in.StartAt, in.EndAt, in.Location, in.URL,
			)
			if err != nil {
				return eventuc.CreateEventOutput{}, err
			}
			return eventuc.CreateEventOutput{Event: ev, Tags: []eventuc.EventTagView{}}, nil
		},
	}

	h := handler.NewEventHandler(&stubListEventsExec{}, stub, &stubUpdateEventExec{}, &stubDeleteEventExec{}, testLogger)

	req := createEventRequest(t, map[string]any{
		testKeyTitle:   testEventTitle,
		"description":  "Team sync",
		testKeyStartAt: now.Format(time.RFC3339),
		testKeyEndAt:   now.Add(time.Hour).Format(time.RFC3339),
		"location":     testEventLocation,
		"url":          testEventURL,
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
	h := handler.NewEventHandler(
		&stubListEventsExec{}, &stubCreateEventExec{}, &stubUpdateEventExec{}, &stubDeleteEventExec{}, testLogger,
	)

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
	h := handler.NewEventHandler(
		&stubListEventsExec{}, &stubCreateEventExec{}, &stubUpdateEventExec{}, &stubDeleteEventExec{}, testLogger,
	)

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
		fn: func(_ context.Context, _ uuid.UUID, _ eventuc.CreateEventInput) (eventuc.CreateEventOutput, error) {
			return eventuc.CreateEventOutput{}, &domain.ValidationError{Details: []domain.ValidationDetail{
				{Field: "title", Code: "REQUIRED", Message: "title is required"},
			}}
		},
	}

	h := handler.NewEventHandler(&stubListEventsExec{}, stub, &stubUpdateEventExec{}, &stubDeleteEventExec{}, testLogger)

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
	if body["code"] != codeValidationError {
		t.Errorf("expected code VALIDATION_ERROR, got %v", body["code"])
	}
}

func TestEventHandler_CreateEvent_EndAtEqualStartAt_Returns400(t *testing.T) {
	userID := uuid.New()

	stub := &stubCreateEventExec{
		fn: func(_ context.Context, _ uuid.UUID, _ eventuc.CreateEventInput) (eventuc.CreateEventOutput, error) {
			return eventuc.CreateEventOutput{}, &domain.ValidationError{Details: []domain.ValidationDetail{
				{Field: "endAt", Code: codeInvalidDateRange, Message: "endAt must be strictly after startAt"},
			}}
		},
	}

	h := handler.NewEventHandler(&stubListEventsExec{}, stub, &stubUpdateEventExec{}, &stubDeleteEventExec{}, testLogger)

	now := time.Now().UTC().Format(time.RFC3339)
	req := createEventRequest(t, map[string]any{
		testKeyTitle:   testEventTitle,
		testKeyStartAt: now,
		testKeyEndAt:   now, // same instant: invalid per SPEC-003
	})
	req = req.WithContext(context.WithValue(req.Context(), ctxkey.User, newStubUser(userID, "u@ex.com", "U")))
	w := httptest.NewRecorder()

	h.CreateEvent(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
	var body map[string]any
	_ = json.NewDecoder(w.Body).Decode(&body)
	if body["code"] != codeValidationError {
		t.Errorf("expected code VALIDATION_ERROR, got %v", body["code"])
	}
	details, ok := body["details"].([]any)
	if !ok || len(details) == 0 {
		t.Fatalf("expected validation details in response, got: %v", body)
	}
	detail, ok := details[0].(map[string]any)
	if !ok || detail["code"] != codeInvalidDateRange {
		t.Errorf("expected detail code INVALID_DATE_RANGE, got: %v", details[0])
	}
}

func TestEventHandler_CreateEvent_InternalError_Returns500(t *testing.T) {
	userID := uuid.New()

	stub := &stubCreateEventExec{
		fn: func(_ context.Context, _ uuid.UUID, _ eventuc.CreateEventInput) (eventuc.CreateEventOutput, error) {
			return eventuc.CreateEventOutput{}, errInternal
		},
	}

	h := handler.NewEventHandler(&stubListEventsExec{}, stub, &stubUpdateEventExec{}, &stubDeleteEventExec{}, testLogger)

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

	updatedEv, _ := domainevent.New(
		eventID, userID, "Updated title", "Updated desc", now, now.Add(time.Hour), testEventLocation, testEventURL,
	)

	stub := &stubUpdateEventExec{
		fn: func(_ context.Context, uid uuid.UUID, in eventuc.UpdateEventInput) (eventuc.UpdateEventOutput, error) {
			if uid != userID {
				t.Errorf("userID mismatch: got %v, want %v", uid, userID)
			}
			if in.ID != eventID {
				t.Errorf("eventID mismatch: got %v, want %v", in.ID, eventID)
			}
			return eventuc.UpdateEventOutput{Event: updatedEv, Tags: []eventuc.EventTagView{}}, nil
		},
	}
	h := handler.NewEventHandler(&stubListEventsExec{}, &stubCreateEventExec{}, stub, &stubDeleteEventExec{}, testLogger)

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
	if resp.Location == nil || *resp.Location != testEventLocation {
		t.Errorf("location mismatch: got %v", resp.Location)
	}
}

func TestEventHandler_UpdateEvent_Unauthorized(t *testing.T) {
	h := handler.NewEventHandler(
		&stubListEventsExec{}, &stubCreateEventExec{}, &stubUpdateEventExec{}, &stubDeleteEventExec{}, testLogger,
	)

	req := putEventRequest(t, uuid.New().String(), validUpdateEventBody)
	w := httptest.NewRecorder()

	h.UpdateEvent(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestEventHandler_UpdateEvent_InvalidIDFormat(t *testing.T) {
	h := handler.NewEventHandler(
		&stubListEventsExec{}, &stubCreateEventExec{}, &stubUpdateEventExec{}, &stubDeleteEventExec{}, testLogger,
	)

	req := putEventRequestAsUser(t, "not-a-uuid", validUpdateEventBody)
	w := httptest.NewRecorder()

	h.UpdateEvent(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestEventHandler_UpdateEvent_InvalidBody(t *testing.T) {
	h := handler.NewEventHandler(
		&stubListEventsExec{}, &stubCreateEventExec{}, &stubUpdateEventExec{}, &stubDeleteEventExec{}, testLogger,
	)

	req := putEventRequestAsUser(t, uuid.New().String(), `not-json`)
	w := httptest.NewRecorder()

	h.UpdateEvent(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestEventHandler_UpdateEvent_ValidationError_Returns400(t *testing.T) {
	stub := &stubUpdateEventExec{
		fn: func(_ context.Context, _ uuid.UUID, _ eventuc.UpdateEventInput) (eventuc.UpdateEventOutput, error) {
			return eventuc.UpdateEventOutput{}, &domain.ValidationError{Details: []domain.ValidationDetail{
				{Field: "endAt", Code: codeInvalidDateRange, Message: "endAt must be after or equal to startAt"},
			}}
		},
	}
	h := handler.NewEventHandler(&stubListEventsExec{}, &stubCreateEventExec{}, stub, &stubDeleteEventExec{}, testLogger)

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
		fn: func(_ context.Context, _ uuid.UUID, _ eventuc.UpdateEventInput) (eventuc.UpdateEventOutput, error) {
			return eventuc.UpdateEventOutput{}, domainevent.ErrEventNotFound
		},
	}
	h := handler.NewEventHandler(&stubListEventsExec{}, &stubCreateEventExec{}, stub, &stubDeleteEventExec{}, testLogger)

	req := putEventRequestAsUser(t, uuid.New().String(), validUpdateEventBody)
	w := httptest.NewRecorder()

	h.UpdateEvent(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d: %s", w.Code, w.Body.String())
	}
}

func TestEventHandler_UpdateEvent_InternalError_Returns500(t *testing.T) {
	stub := &stubUpdateEventExec{
		fn: func(_ context.Context, _ uuid.UUID, _ eventuc.UpdateEventInput) (eventuc.UpdateEventOutput, error) {
			return eventuc.UpdateEventOutput{}, errInternal
		},
	}
	h := handler.NewEventHandler(&stubListEventsExec{}, &stubCreateEventExec{}, stub, &stubDeleteEventExec{}, testLogger)

	req := putEventRequestAsUser(t, uuid.New().String(), validUpdateEventBody)
	w := httptest.NewRecorder()

	h.UpdateEvent(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d: %s", w.Code, w.Body.String())
	}
}

// --- DeleteEvent tests ---

type stubDeleteEventExec struct {
	fn func(context.Context, uuid.UUID, uuid.UUID) error
}

func (s *stubDeleteEventExec) Execute(ctx context.Context, userID uuid.UUID, id uuid.UUID) error {
	if s.fn == nil {
		return nil
	}
	return s.fn(ctx, userID, id)
}

func deleteEventRequest(t *testing.T, id string) *http.Request {
	t.Helper()
	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodDelete,
		"/events/"+id,
		nil,
	)
	if err != nil {
		t.Fatalf("create request: %v", err)
	}
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", id)
	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
}

func deleteEventRequestAsUser(t *testing.T, id string) *http.Request {
	t.Helper()
	req := deleteEventRequest(t, id)
	u := newStubUser(uuid.New(), "user@example.com", "User")
	return req.WithContext(context.WithValue(req.Context(), ctxkey.User, u))
}

func TestEventHandler_DeleteEvent_Success_Returns204(t *testing.T) {
	userID := uuid.New()
	eventID := uuid.New()

	stub := &stubDeleteEventExec{
		fn: func(_ context.Context, uid uuid.UUID, id uuid.UUID) error {
			if uid != userID {
				t.Errorf("userID mismatch: got %v, want %v", uid, userID)
			}
			if id != eventID {
				t.Errorf("eventID mismatch: got %v, want %v", id, eventID)
			}
			return nil
		},
	}
	h := handler.NewEventHandler(&stubListEventsExec{}, &stubCreateEventExec{}, &stubUpdateEventExec{}, stub, testLogger)

	req := deleteEventRequest(t, eventID.String())
	req = req.WithContext(context.WithValue(req.Context(), ctxkey.User, newStubUser(userID, "user@example.com", "User")))
	w := httptest.NewRecorder()

	h.DeleteEvent(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", w.Code, w.Body.String())
	}
	if w.Body.Len() != 0 {
		t.Errorf("expected empty body, got %q", w.Body.String())
	}
}

func TestEventHandler_DeleteEvent_Unauthorized_Returns401(t *testing.T) {
	h := handler.NewEventHandler(
		&stubListEventsExec{}, &stubCreateEventExec{}, &stubUpdateEventExec{}, &stubDeleteEventExec{}, testLogger,
	)

	req := deleteEventRequest(t, uuid.New().String())
	w := httptest.NewRecorder()

	h.DeleteEvent(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestEventHandler_DeleteEvent_InvalidIDFormat_Returns400(t *testing.T) {
	h := handler.NewEventHandler(
		&stubListEventsExec{}, &stubCreateEventExec{}, &stubUpdateEventExec{}, &stubDeleteEventExec{}, testLogger,
	)

	req := deleteEventRequestAsUser(t, "not-a-uuid")
	w := httptest.NewRecorder()

	h.DeleteEvent(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestEventHandler_DeleteEvent_NotFound_Returns404(t *testing.T) {
	stub := &stubDeleteEventExec{
		fn: func(_ context.Context, _ uuid.UUID, _ uuid.UUID) error {
			return domainevent.ErrEventNotFound
		},
	}
	h := handler.NewEventHandler(&stubListEventsExec{}, &stubCreateEventExec{}, &stubUpdateEventExec{}, stub, testLogger)

	req := deleteEventRequestAsUser(t, uuid.New().String())
	w := httptest.NewRecorder()

	h.DeleteEvent(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d: %s", w.Code, w.Body.String())
	}
}

func TestEventHandler_DeleteEvent_OtherUsersEvent_Returns404(t *testing.T) {
	stub := &stubDeleteEventExec{
		fn: func(_ context.Context, _ uuid.UUID, _ uuid.UUID) error {
			// Ownership mismatches surface as ErrEventNotFound from the usecase,
			// consistent with UpdateEvent's handling.
			return domainevent.ErrEventNotFound
		},
	}
	h := handler.NewEventHandler(&stubListEventsExec{}, &stubCreateEventExec{}, &stubUpdateEventExec{}, stub, testLogger)

	req := deleteEventRequestAsUser(t, uuid.New().String())
	w := httptest.NewRecorder()

	h.DeleteEvent(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d: %s", w.Code, w.Body.String())
	}
}

func TestEventHandler_DeleteEvent_InternalError_Returns500(t *testing.T) {
	stub := &stubDeleteEventExec{
		fn: func(_ context.Context, _ uuid.UUID, _ uuid.UUID) error {
			return errInternal
		},
	}
	h := handler.NewEventHandler(&stubListEventsExec{}, &stubCreateEventExec{}, &stubUpdateEventExec{}, stub, testLogger)

	req := deleteEventRequestAsUser(t, uuid.New().String())
	w := httptest.NewRecorder()

	h.DeleteEvent(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d: %s", w.Code, w.Body.String())
	}
}

// --- Tag-related handler tests ---

func TestEventHandler_GetEvents_TagsIncludedInResponse(t *testing.T) {
	userID := uuid.New()
	tagID := uuid.New()

	stub := &stubListEventsExec{
		fn: func(_ context.Context, _ uuid.UUID, _ eventuc.ListEventsInput) ([]eventuc.EventReadModel, error) {
			return []eventuc.EventReadModel{
				{
					ID:    uuid.New(),
					Title: "Tagged event",
					Tags:  []eventuc.EventTagReadModel{{ID: tagID, Name: "Work", Color: testEventTagColor}},
				},
				{
					ID:    uuid.New(),
					Title: "Untagged event",
					Tags:  []eventuc.EventTagReadModel{},
				},
			}, nil
		},
	}

	h := handler.NewEventHandler(stub, &stubCreateEventExec{}, &stubUpdateEventExec{}, &stubDeleteEventExec{}, testLogger)

	req := getEventsRequest(t, testRangeStartAt, testRangeEndAt)
	req = req.WithContext(context.WithValue(req.Context(), ctxkey.User, newStubUser(userID, "u@ex.com", "U")))
	w := httptest.NewRecorder()

	h.GetEvents(w, req, testGetEventsParams(t))

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Events []struct {
			Title string `json:"title"`
			Tags  []struct {
				ID    string `json:"id"`
				Name  string `json:"name"`
				Color string `json:"color"`
			} `json:"tags"`
		} `json:"events"`
	}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(resp.Events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(resp.Events))
	}
	if len(resp.Events[0].Tags) != 1 {
		t.Errorf("expected first event to have 1 tag, got %d", len(resp.Events[0].Tags))
	}
	if resp.Events[0].Tags[0].ID != tagID.String() {
		t.Errorf("tag ID mismatch: got %q, want %q", resp.Events[0].Tags[0].ID, tagID.String())
	}
	if resp.Events[0].Tags[0].Color != testEventTagColor {
		t.Errorf("tag color mismatch: got %q", resp.Events[0].Tags[0].Color)
	}
	if len(resp.Events[1].Tags) != 0 {
		t.Errorf("expected second event to have empty tags array, got %d", len(resp.Events[1].Tags))
	}
}

func TestEventHandler_GetEvents_TagIDQueryPassedToUseCase(t *testing.T) {
	userID := uuid.New()
	tagID := uuid.New()

	var capturedInput eventuc.ListEventsInput
	stub := &stubListEventsExec{
		fn: func(_ context.Context, _ uuid.UUID, in eventuc.ListEventsInput) ([]eventuc.EventReadModel, error) {
			capturedInput = in
			return nil, nil
		},
	}

	h := handler.NewEventHandler(stub, &stubCreateEventExec{}, &stubUpdateEventExec{}, &stubDeleteEventExec{}, testLogger)
	req := getEventsRequest(t, testRangeStartAt, testRangeEndAt)
	// Add tagId query param
	q := req.URL.Query()
	q.Set("tagId", tagID.String())
	req.URL.RawQuery = q.Encode()
	req = req.WithContext(context.WithValue(req.Context(), ctxkey.User, newStubUser(userID, "u@ex.com", "U")))
	w := httptest.NewRecorder()

	params := testGetEventsParams(t)
	params.TagId = &tagID
	h.GetEvents(w, req, params)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	if capturedInput.TagID == nil || *capturedInput.TagID != tagID {
		t.Errorf("expected TagID %v to be passed to UseCase, got %v", tagID, capturedInput.TagID)
	}
}

func TestEventHandler_GetEvents_InvalidTagIDReturns400(t *testing.T) {
	userID := uuid.New()
	stub := &stubListEventsExec{
		fn: func(context.Context, uuid.UUID, eventuc.ListEventsInput) ([]eventuc.EventReadModel, error) {
			t.Error("usecase must not be called for an invalid tagId")
			return nil, nil
		},
	}
	h := handler.NewEventHandler(stub, &stubCreateEventExec{}, &stubUpdateEventExec{}, &stubDeleteEventExec{}, testLogger)

	// ServeHTTP parses tagId before delegating to GetEvents; an invalid UUID
	// must be rejected with 400 before the usecase is called.
	req2, _ := http.NewRequestWithContext(
		context.WithValue(context.Background(), ctxkey.User, newStubUser(userID, "u@ex.com", "U")),
		http.MethodGet,
		"/events?startAt="+testRangeStartAt+"&endAt="+testRangeEndAt+"&tagId=not-a-uuid",
		nil,
	)
	req2 = req2.WithContext(context.WithValue(req2.Context(), ctxkey.User, newStubUser(userID, "u@ex.com", "U")))
	w2 := httptest.NewRecorder()
	h.ServeHTTP(w2, req2)

	if w2.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid tagId, got %d: %s", w2.Code, w2.Body.String())
	}
}

func TestEventHandler_CreateEvent_WithTagIDs_PassedToUseCase(t *testing.T) {
	userID := uuid.New()
	eventID := uuid.New()
	tagID := uuid.New()
	now := time.Now().UTC()

	var capturedInput eventuc.CreateEventInput
	stub := &stubCreateEventExec{
		fn: func(_ context.Context, _ uuid.UUID, in eventuc.CreateEventInput) (eventuc.CreateEventOutput, error) {
			capturedInput = in
			ev, _ := domainevent.New(eventID, userID, in.Title, "", in.StartAt, in.EndAt, "", "")
			tags := []eventuc.EventTagView{{ID: tagID, Name: "Work", Color: testEventTagColor}}
			return eventuc.CreateEventOutput{Event: ev, Tags: tags}, nil
		},
	}

	h := handler.NewEventHandler(&stubListEventsExec{}, stub, &stubUpdateEventExec{}, &stubDeleteEventExec{}, testLogger)

	req := createEventRequest(t, map[string]any{
		testKeyTitle:    testEventTitle,
		testKeyStartAt:  now.Format(time.RFC3339),
		testKeyEndAt:    now.Add(time.Hour).Format(time.RFC3339),
		testFieldTagIDs: []string{tagID.String()},
	})
	req = req.WithContext(context.WithValue(req.Context(), ctxkey.User, newStubUser(userID, "u@ex.com", "U")))
	w := httptest.NewRecorder()

	h.CreateEvent(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
	if len(capturedInput.TagIDs) != 1 || capturedInput.TagIDs[0] != tagID {
		t.Errorf("expected tagID %v to be passed to UseCase, got %v", tagID, capturedInput.TagIDs)
	}

	var resp struct {
		Tags []struct {
			ID    string `json:"id"`
			Name  string `json:"name"`
			Color string `json:"color"`
		} `json:"tags"`
	}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(resp.Tags) != 1 {
		t.Fatalf("expected 1 tag in response, got %d", len(resp.Tags))
	}
	if resp.Tags[0].ID != tagID.String() {
		t.Errorf("tag ID mismatch: got %q, want %q", resp.Tags[0].ID, tagID.String())
	}
}

func TestEventHandler_CreateEvent_TagValidationError_Returns400WithDetails(t *testing.T) {
	userID := uuid.New()

	stub := &stubCreateEventExec{
		fn: func(_ context.Context, _ uuid.UUID, _ eventuc.CreateEventInput) (eventuc.CreateEventOutput, error) {
			return eventuc.CreateEventOutput{}, &domain.ValidationError{Details: []domain.ValidationDetail{
				{Field: testFieldTagIDs, Code: "TOO_MANY", Message: "An event can have at most 5 tags"},
			}}
		},
	}

	h := handler.NewEventHandler(&stubListEventsExec{}, stub, &stubUpdateEventExec{}, &stubDeleteEventExec{}, testLogger)

	now := time.Now().UTC()
	req := createEventRequest(t, map[string]any{
		testKeyTitle:   testEventTitle,
		testKeyStartAt: now.Format(time.RFC3339),
		testKeyEndAt:   now.Add(time.Hour).Format(time.RFC3339),
		testFieldTagIDs: []string{
			uuid.New().String(), uuid.New().String(), uuid.New().String(),
			uuid.New().String(), uuid.New().String(), uuid.New().String(),
		},
	})
	req = req.WithContext(context.WithValue(req.Context(), ctxkey.User, newStubUser(userID, "u@ex.com", "U")))
	w := httptest.NewRecorder()

	h.CreateEvent(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
	var body map[string]any
	_ = json.NewDecoder(w.Body).Decode(&body)
	if body["code"] != codeValidationError {
		t.Errorf("expected code VALIDATION_ERROR, got %v", body["code"])
	}
	details, ok := body["details"].([]any)
	if !ok || len(details) == 0 {
		t.Fatalf("expected validation details in response: %v", body)
	}
	detail, ok2 := details[0].(map[string]any)
	if !ok2 {
		t.Fatalf("expected detail to be map[string]any, got %T", details[0])
	}
	if detail["field"] != testFieldTagIDs {
		t.Errorf("expected detail field=%s, got %v", testFieldTagIDs, detail["field"])
	}
	if detail["code"] != "TOO_MANY" {
		t.Errorf("expected detail code=TOO_MANY, got %v", detail["code"])
	}
}

func TestEventHandler_UpdateEvent_OmittedTagIDs_PassesNilToUseCase(t *testing.T) {
	userID := uuid.New()
	eventID := uuid.New()
	now := time.Now().UTC()

	var capturedInput eventuc.UpdateEventInput
	updatedEv2, _ := domainevent.New(eventID, userID, "Title", "", now, now.Add(time.Hour), "", "")
	stub := &stubUpdateEventExec{
		fn: func(_ context.Context, _ uuid.UUID, in eventuc.UpdateEventInput) (eventuc.UpdateEventOutput, error) {
			capturedInput = in
			return eventuc.UpdateEventOutput{Event: updatedEv2, Tags: []eventuc.EventTagView{}}, nil
		},
	}

	h := handler.NewEventHandler(&stubListEventsExec{}, &stubCreateEventExec{}, stub, &stubDeleteEventExec{}, testLogger)

	// No tagIds in request body
	body := `{"title":"Title","startAt":"` + now.Format(time.RFC3339) +
		`","endAt":"` + now.Add(time.Hour).Format(time.RFC3339) + `"}`
	req := putEventRequestAsUser(t, eventID.String(), body)
	w := httptest.NewRecorder()

	h.UpdateEvent(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	if capturedInput.TagIDs != nil {
		t.Errorf("expected TagIDs nil when omitted, got %v", capturedInput.TagIDs)
	}
}

func TestEventHandler_UpdateEvent_EmptyTagIDs_PassesEmptySliceToUseCase(t *testing.T) {
	userID := uuid.New()
	eventID := uuid.New()
	now := time.Now().UTC()

	var capturedInput eventuc.UpdateEventInput
	updatedEv3, _ := domainevent.New(eventID, userID, "Title", "", now, now.Add(time.Hour), "", "")
	stub := &stubUpdateEventExec{
		fn: func(_ context.Context, _ uuid.UUID, in eventuc.UpdateEventInput) (eventuc.UpdateEventOutput, error) {
			capturedInput = in
			return eventuc.UpdateEventOutput{Event: updatedEv3, Tags: []eventuc.EventTagView{}}, nil
		},
	}

	h := handler.NewEventHandler(&stubListEventsExec{}, &stubCreateEventExec{}, stub, &stubDeleteEventExec{}, testLogger)

	body := `{"title":"Title","startAt":"` + now.Format(time.RFC3339) +
		`","endAt":"` + now.Add(time.Hour).Format(time.RFC3339) + `","tagIds":[]}`
	req := putEventRequestAsUser(t, eventID.String(), body)
	w := httptest.NewRecorder()

	h.UpdateEvent(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	if capturedInput.TagIDs == nil {
		t.Error("expected TagIDs to be non-nil (explicit empty slice) when tagIds:[] sent")
	} else if len(*capturedInput.TagIDs) != 0 {
		t.Errorf("expected empty TagIDs slice, got len=%d", len(*capturedInput.TagIDs))
	}
}
