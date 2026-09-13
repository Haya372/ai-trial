//go:build integration

package handler_test

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"

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
	txMgr := db.NewPgxTxManager(routeTestPool)

	signup := authuc.NewSignupCommand(userRepo, sessRepo, txMgr)
	login := authuc.NewLoginCommand(userRepo, sessRepo)
	logout := authuc.NewLogoutCommand(sessRepo)
	listEvents := eventuc.NewListEventsQuery(eventQueryRepo)

	auth := handler.NewAuthHandler(signup, login, logout)
	ev := handler.NewEventHandler(listEvents, logger)

	r := chi.NewRouter()
	r.Post("/auth/signup", auth.Signup)
	r.Post("/auth/login", auth.Login)
	r.With(mw.RequireAuth(sessRepo, userRepo)).Get("/auth/me", auth.GetMe)
	r.With(mw.RequireAuth(sessRepo, userRepo)).Get("/events", ev.ServeHTTP)
	return r
}

func insertRouteTestEvent(t *testing.T, userID, title string, start, end time.Time) {
	t.Helper()
	_, err := routeTestPool.Exec(context.Background(),
		"INSERT INTO events (user_id, title, start_at, end_at) VALUES ($1, $2, $3, $4)",
		userID, title, start, end,
	)
	if err != nil {
		t.Fatalf("insert event: %v", err)
	}
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
