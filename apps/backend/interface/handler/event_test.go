package handler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"

	api "github.com/Haya372/ai-trial/backend/interface/api/generated"
	"github.com/Haya372/ai-trial/backend/interface/ctxkey"
	"github.com/Haya372/ai-trial/backend/interface/handler"
	eventuc "github.com/Haya372/ai-trial/backend/usecase/event"
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
					Title:       "Meeting",
					Description: desc,
					StartAt:     now,
					EndAt:       now.Add(time.Hour),
				},
			}, nil
		},
	}

	h := handler.NewEventHandler(stub, testLogger)

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

	var resp api.EventsListResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(resp.Events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(resp.Events))
	}
	if resp.Events[0].Title != "Meeting" {
		t.Errorf("title mismatch: got %q", resp.Events[0].Title)
	}
}

func TestEventHandler_GetEvents_Unauthorized(t *testing.T) {
	stub := &stubListEventsExec{}
	h := handler.NewEventHandler(stub, testLogger)

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
	h := handler.NewEventHandler(stub, testLogger)

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
