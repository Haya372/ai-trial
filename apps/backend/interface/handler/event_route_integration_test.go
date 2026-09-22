//go:build integration

package handler_test

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/Haya372/ai-trial/backend/infrastructure/db"
	"github.com/Haya372/ai-trial/backend/infrastructure/repository"
	"github.com/Haya372/ai-trial/backend/interface/handler"
	mw "github.com/Haya372/ai-trial/backend/interface/middleware"
	authuc "github.com/Haya372/ai-trial/backend/usecase/auth"
	eventuc "github.com/Haya372/ai-trial/backend/usecase/event"
)

func buildEventTestRouter() *chi.Mux {
	logger := slog.New(slog.DiscardHandler)
	userRepo := repository.NewUserRepository(routeTestPool)
	sessRepo := repository.NewSessionRepository(routeTestPool)
	eventQueryRepo := repository.NewEventQueryRepository(routeTestPool, logger)
	eventRepo := repository.NewEventRepository(routeTestPool, logger)
	txMgr := db.NewPgxTxManager(routeTestPool)

	signup := authuc.NewSignupCommand(userRepo, sessRepo, txMgr)
	login := authuc.NewLoginCommand(userRepo, sessRepo)
	logout := authuc.NewLogoutCommand(sessRepo)
	listEvents := eventuc.NewListEventsQuery(eventQueryRepo)
	createEvent := eventuc.NewCreateEventCommand(eventRepo)
	updateEvent := eventuc.NewUpdateEventCommand(eventRepo, logger)
	deleteEvent := eventuc.NewDeleteEventCommand(eventRepo, logger)

	auth := handler.NewAuthHandler(signup, login, logout, logger)
	ev := handler.NewEventHandler(listEvents, createEvent, updateEvent, deleteEvent, logger)

	r := chi.NewRouter()
	r.Post("/auth/signup", auth.Signup)
	r.Post("/auth/login", auth.Login)
	r.With(mw.RequireAuth(sessRepo, userRepo, logger)).Get("/auth/me", auth.GetMe)
	r.With(mw.RequireAuth(sessRepo, userRepo, logger)).Get("/events", ev.ServeHTTP)
	r.With(mw.RequireAuth(sessRepo, userRepo, logger)).Post("/events", ev.CreateEvent)
	r.With(mw.RequireAuth(sessRepo, userRepo, logger)).Put("/events/{id}", ev.UpdateEvent)
	r.With(mw.RequireAuth(sessRepo, userRepo, logger)).Delete("/events/{id}", ev.DeleteEvent)
	return r
}

func insertRouteTestEvent(t *testing.T, userID, title string, start, end time.Time) string {
	t.Helper()
	var id string
	err := routeTestPool.QueryRow(context.Background(),
		"INSERT INTO events (user_id, title, start_at, end_at) VALUES ($1, $2, $3, $4) RETURNING id",
		userID, title, start, end,
	).Scan(&id)
	if err != nil {
		t.Fatalf("insert event: %v", err)
	}
	return id
}

func TestRoute_GetEvents_withoutSession_returns401(t *testing.T) {
	setupRouteTest(t)
	router := buildEventTestRouter()

	req := httptest.NewRequest(http.MethodGet, "/events?startDate=2026-09-01&endDate=2026-09-30", nil)
	if code := responseCode(t, router, req); code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", code)
	}
}

func TestRoute_GetEvents_missingQueryParams_returns400(t *testing.T) {
	setupRouteTest(t)
	router := buildEventTestRouter()

	cookie := signupAndGetCookie(t, router, "events-noparam@ex.com")

	req := httptest.NewRequest(http.MethodGet, "/events", nil)
	req.AddCookie(cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestRoute_GetEvents_returnsOnlyAuthenticatedUsersEvents(t *testing.T) {
	setupRouteTest(t)
	router := buildEventTestRouter()

	cookie1 := signupAndGetCookie(t, router, "user1-events@ex.com")
	cookie2 := signupAndGetCookie(t, router, "user2-events@ex.com")

	userID1 := getUserIDFromCookie(t, router, cookie1)

	now := time.Now().UTC().Truncate(time.Second)
	insertRouteTestEvent(t, userID1, "User1 meeting", now, now.Add(time.Hour))
	insertRouteTestEvent(t, getUserIDFromCookie(t, router, cookie2), "User2 meeting", now, now.Add(time.Hour))

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/events?startDate=%s&endDate=%s",
		now.Format("2006-01-02"), now.Add(24*time.Hour).Format("2006-01-02")), nil)
	req.AddCookie(cookie1)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var body struct {
		Events []struct {
			Title string `json:"title"`
		} `json:"events"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(body.Events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(body.Events))
	}
	if body.Events[0].Title != "User1 meeting" {
		t.Errorf("title mismatch: got %q", body.Events[0].Title)
	}
}

func TestRoute_GetEvents_includesLocationAndUrl(t *testing.T) {
	setupRouteTest(t)
	router := buildEventTestRouter()

	cookie := signupAndGetCookie(t, router, "events-location-url@ex.com")
	userID := getUserIDFromCookie(t, router, cookie)

	now := time.Now().UTC().Truncate(time.Second)
	_, err := routeTestPool.Exec(context.Background(),
		`INSERT INTO events (user_id, title, start_at, end_at, location, url)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		userID, "Event with location and url", now, now.Add(time.Hour), "Tokyo Office", "https://example.com/meeting",
	)
	if err != nil {
		t.Fatalf("insert event: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/events?startDate=%s&endDate=%s",
		now.Format("2006-01-02"), now.Add(24*time.Hour).Format("2006-01-02")), nil)
	req.AddCookie(cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var body struct {
		Events []struct {
			Title    string `json:"title"`
			Location string `json:"location"`
			URL      string `json:"url"`
		} `json:"events"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(body.Events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(body.Events))
	}
	if body.Events[0].Location != "Tokyo Office" {
		t.Errorf("location mismatch: got %q", body.Events[0].Location)
	}
	if body.Events[0].URL != "https://example.com/meeting" {
		t.Errorf("url mismatch: got %q", body.Events[0].URL)
	}
}

func TestRoute_GetEvents_returnsEmptyListWhenNoEvents(t *testing.T) {
	setupRouteTest(t)
	router := buildEventTestRouter()

	cookie := signupAndGetCookie(t, router, "empty-events@ex.com")

	req := httptest.NewRequest(http.MethodGet, "/events?startDate=2026-09-01&endDate=2026-09-30", nil)
	req.AddCookie(cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var body struct {
		Events []any `json:"events"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(body.Events) != 0 {
		t.Errorf("expected empty events, got %d", len(body.Events))
	}
}

func TestRoute_PutEvent_withoutSession_returns401(t *testing.T) {
	setupRouteTest(t)
	router := buildEventTestRouter()

	req := httptest.NewRequest(http.MethodPut, "/events/"+uuid.New().String(),
		strings.NewReader(`{"title":"x","startAt":"2026-09-01T00:00:00Z","endAt":"2026-09-01T01:00:00Z"}`))
	if code := responseCode(t, router, req); code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", code)
	}
}

func TestRoute_PutEvent_updatesOwnEvent_returns200(t *testing.T) {
	setupRouteTest(t)
	router := buildEventTestRouter()

	cookie := signupAndGetCookie(t, router, "put-event-owner@ex.com")
	userID := getUserIDFromCookie(t, router, cookie)

	now := time.Now().UTC().Truncate(time.Second)
	eventID := insertRouteTestEvent(t, userID, "Original title", now, now.Add(time.Hour))

	newStart := now.Add(2 * time.Hour)
	newEnd := now.Add(3 * time.Hour)
	bodyFmt := `{"title":"Updated title","description":"Updated desc",` +
		`"startAt":"%s","endAt":"%s","location":"Tokyo","url":"https://example.com"}`
	body := fmt.Sprintf(bodyFmt, newStart.Format(time.RFC3339), newEnd.Format(time.RFC3339))
	req := httptest.NewRequest(http.MethodPut, "/events/"+eventID, strings.NewReader(body))
	req.AddCookie(cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp struct {
		Title    string `json:"title"`
		Location string `json:"location"`
		URL      string `json:"url"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Title != "Updated title" {
		t.Errorf("title mismatch: got %q", resp.Title)
	}
	if resp.Location != "Tokyo" {
		t.Errorf("location mismatch: got %q", resp.Location)
	}
}

func TestRoute_PutEvent_invalidDateRange_returns400(t *testing.T) {
	setupRouteTest(t)
	router := buildEventTestRouter()

	cookie := signupAndGetCookie(t, router, "put-event-invalid@ex.com")
	userID := getUserIDFromCookie(t, router, cookie)

	now := time.Now().UTC().Truncate(time.Second)
	eventID := insertRouteTestEvent(t, userID, "Original title", now, now.Add(time.Hour))

	body := fmt.Sprintf(
		`{"title":"Updated title","startAt":"%s","endAt":"%s"}`,
		now.Add(time.Hour).Format(time.RFC3339), now.Format(time.RFC3339),
	)
	req := httptest.NewRequest(http.MethodPut, "/events/"+eventID, strings.NewReader(body))
	req.AddCookie(cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestRoute_PutEvent_nonExistentEvent_returns404(t *testing.T) {
	setupRouteTest(t)
	router := buildEventTestRouter()

	cookie := signupAndGetCookie(t, router, "put-event-notfound@ex.com")

	body := `{"title":"x","startAt":"2026-09-01T00:00:00Z","endAt":"2026-09-01T01:00:00Z"}`
	req := httptest.NewRequest(http.MethodPut, "/events/"+uuid.New().String(), strings.NewReader(body))
	req.AddCookie(cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestRoute_PutEvent_otherUsersEvent_returns404(t *testing.T) {
	setupRouteTest(t)
	router := buildEventTestRouter()

	ownerCookie := signupAndGetCookie(t, router, "put-event-owner2@ex.com")
	ownerID := getUserIDFromCookie(t, router, ownerCookie)
	otherCookie := signupAndGetCookie(t, router, "put-event-other@ex.com")

	now := time.Now().UTC().Truncate(time.Second)
	eventID := insertRouteTestEvent(t, ownerID, "Owner's event", now, now.Add(time.Hour))

	body := `{"title":"Hijacked","startAt":"2026-09-01T00:00:00Z","endAt":"2026-09-01T01:00:00Z"}`
	req := httptest.NewRequest(http.MethodPut, "/events/"+eventID, strings.NewReader(body))
	req.AddCookie(otherCookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	// Ownership mismatches are reported as 404, the same as a missing event,
	// to avoid leaking event existence to non-owners.
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestRoute_PutEvent_pastEvent_canBeUpdated(t *testing.T) {
	setupRouteTest(t)
	router := buildEventTestRouter()

	cookie := signupAndGetCookie(t, router, "put-event-past@ex.com")
	userID := getUserIDFromCookie(t, router, cookie)

	past := time.Now().Add(-48 * time.Hour).UTC().Truncate(time.Second)
	eventID := insertRouteTestEvent(t, userID, "Past event", past, past.Add(time.Hour))

	body := fmt.Sprintf(
		`{"title":"Updated past event","startAt":"%s","endAt":"%s"}`,
		past.Format(time.RFC3339), past.Add(2*time.Hour).Format(time.RFC3339),
	)
	req := httptest.NewRequest(http.MethodPut, "/events/"+eventID, strings.NewReader(body))
	req.AddCookie(cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func eventExistsInDB(t *testing.T, id string) bool {
	t.Helper()
	var exists bool
	err := routeTestPool.QueryRow(context.Background(),
		"SELECT EXISTS(SELECT 1 FROM events WHERE id = $1)", id,
	).Scan(&exists)
	if err != nil {
		t.Fatalf("check event existence: %v", err)
	}
	return exists
}

func TestRoute_DeleteEvent_withoutSession_returns401(t *testing.T) {
	setupRouteTest(t)
	router := buildEventTestRouter()

	req := httptest.NewRequest(http.MethodDelete, "/events/"+uuid.New().String(), nil)
	if code := responseCode(t, router, req); code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", code)
	}
}

func TestRoute_DeleteEvent_deletesOwnEvent_returns204(t *testing.T) {
	setupRouteTest(t)
	router := buildEventTestRouter()

	cookie := signupAndGetCookie(t, router, "delete-event-owner@ex.com")
	userID := getUserIDFromCookie(t, router, cookie)

	now := time.Now().UTC().Truncate(time.Second)
	eventID := insertRouteTestEvent(t, userID, "Original title", now, now.Add(time.Hour))

	req := httptest.NewRequest(http.MethodDelete, "/events/"+eventID, nil)
	req.AddCookie(cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", rec.Code, rec.Body.String())
	}
	if eventExistsInDB(t, eventID) {
		t.Errorf("expected event %s to be deleted", eventID)
	}
}

func TestRoute_DeleteEvent_nonExistentEvent_returns404(t *testing.T) {
	setupRouteTest(t)
	router := buildEventTestRouter()

	cookie := signupAndGetCookie(t, router, "delete-event-notfound@ex.com")

	req := httptest.NewRequest(http.MethodDelete, "/events/"+uuid.New().String(), nil)
	req.AddCookie(cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestRoute_DeleteEvent_otherUsersEvent_returns404(t *testing.T) {
	setupRouteTest(t)
	router := buildEventTestRouter()

	ownerCookie := signupAndGetCookie(t, router, "delete-event-owner2@ex.com")
	ownerID := getUserIDFromCookie(t, router, ownerCookie)
	otherCookie := signupAndGetCookie(t, router, "delete-event-other@ex.com")

	now := time.Now().UTC().Truncate(time.Second)
	eventID := insertRouteTestEvent(t, ownerID, "Owner's event", now, now.Add(time.Hour))

	req := httptest.NewRequest(http.MethodDelete, "/events/"+eventID, nil)
	req.AddCookie(otherCookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	// Ownership mismatches are reported as 404, the same as a missing event,
	// to avoid leaking event existence to non-owners (consistent with PUT).
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
	if !eventExistsInDB(t, eventID) {
		t.Errorf("expected event %s to remain after forbidden delete attempt", eventID)
	}
}

func TestRoute_DeleteEvent_pastEvent_canBeDeleted(t *testing.T) {
	setupRouteTest(t)
	router := buildEventTestRouter()

	cookie := signupAndGetCookie(t, router, "delete-event-past@ex.com")
	userID := getUserIDFromCookie(t, router, cookie)

	past := time.Now().Add(-48 * time.Hour).UTC().Truncate(time.Second)
	eventID := insertRouteTestEvent(t, userID, "Past event", past, past.Add(time.Hour))

	req := httptest.NewRequest(http.MethodDelete, "/events/"+eventID, nil)
	req.AddCookie(cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", rec.Code, rec.Body.String())
	}
}

func getUserIDFromCookie(t *testing.T, router *chi.Mux, cookie *http.Cookie) string {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	req.AddCookie(cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GET /auth/me: expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var body struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode me response: %v", err)
	}
	return body.ID
}
