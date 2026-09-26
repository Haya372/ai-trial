//go:build integration

package handler_test

import (
	"encoding/json"
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
	eventshareuc "github.com/Haya372/ai-trial/backend/usecase/eventshare"
)

func buildEventShareTestRouter() *chi.Mux {
	logger := slog.New(slog.DiscardHandler)
	userRepo := repository.NewUserRepository(routeTestPool, testTracerProvider)
	sessRepo := repository.NewSessionRepository(routeTestPool, testTracerProvider)
	eventRepo := repository.NewEventRepository(routeTestPool, logger, testTracerProvider)
	eventShareRepo := repository.NewEventShareRepository(routeTestPool, testTracerProvider)
	txMgr := db.NewPgxTxManager(routeTestPool)

	signup := authuc.NewSignupCommand(userRepo, sessRepo, txMgr)
	createShare := eventshareuc.NewCreateShareCommand(eventRepo, eventShareRepo, logger)

	auth := handler.NewAuthHandler(signup, authuc.NewLoginCommand(userRepo, sessRepo), authuc.NewLogoutCommand(sessRepo), logger)
	es := handler.NewEventShareHandler(createShare, logger)

	r := chi.NewRouter()
	r.Post("/auth/signup", auth.Signup)
	r.With(mw.RequireAuth(sessRepo, userRepo, logger)).Get("/auth/me", auth.GetMe)
	r.With(mw.RequireAuth(sessRepo, userRepo, logger)).Post("/events/{id}/shares", es.CreateShare)
	return r
}

func TestRoute_CreateEventShare_withoutSession_returns401(t *testing.T) {
	setupRouteTest(t)
	router := buildEventShareTestRouter()

	req := httptest.NewRequest(http.MethodPost, "/events/"+uuid.New().String()+"/shares", nil)
	if code := responseCode(t, router, req); code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", code)
	}
}

func TestRoute_CreateEventShare_ownEvent_returns201WithShareURL(t *testing.T) {
	setupRouteTest(t)
	router := buildEventShareTestRouter()

	cookie := signupAndGetCookie(t, router, "share-owner@ex.com")
	userID := getUserIDFromCookie(t, router, cookie)

	now := time.Now().UTC().Truncate(time.Second)
	eventID := insertRouteTestEvent(t, userID, "Shared meeting", now, now.Add(time.Hour))

	req := httptest.NewRequest(http.MethodPost, "/events/"+eventID+"/shares", nil)
	req.AddCookie(cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	var body struct {
		URL       string    `json:"url"`
		ExpiresAt time.Time `json:"expiresAt"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !strings.HasPrefix(body.URL, "/share/") || len(body.URL) <= len("/share/") {
		t.Errorf("expected a non-empty share url, got %q", body.URL)
	}
	if !body.ExpiresAt.Equal(now.Add(time.Hour)) {
		t.Errorf("expected expiresAt to default to event end (%v), got %v", now.Add(time.Hour), body.ExpiresAt)
	}
}

func TestRoute_CreateEventShare_otherUsersEvent_returns403(t *testing.T) {
	setupRouteTest(t)
	router := buildEventShareTestRouter()

	ownerCookie := signupAndGetCookie(t, router, "share-owner2@ex.com")
	ownerID := getUserIDFromCookie(t, router, ownerCookie)
	otherCookie := signupAndGetCookie(t, router, "share-other@ex.com")

	now := time.Now().UTC().Truncate(time.Second)
	eventID := insertRouteTestEvent(t, ownerID, "Owner's meeting", now, now.Add(time.Hour))

	req := httptest.NewRequest(http.MethodPost, "/events/"+eventID+"/shares", nil)
	req.AddCookie(otherCookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestRoute_CreateEventShare_nonExistentEvent_returns404(t *testing.T) {
	setupRouteTest(t)
	router := buildEventShareTestRouter()

	cookie := signupAndGetCookie(t, router, "share-notfound@ex.com")

	req := httptest.NewRequest(http.MethodPost, "/events/"+uuid.New().String()+"/shares", nil)
	req.AddCookie(cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}
