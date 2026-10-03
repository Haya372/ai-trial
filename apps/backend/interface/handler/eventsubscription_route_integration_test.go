//go:build integration

package handler_test

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	domaineventshare "github.com/Haya372/ai-trial/backend/domain/eventshare"
	"github.com/Haya372/ai-trial/backend/infrastructure/db"
	"github.com/Haya372/ai-trial/backend/infrastructure/repository"
	"github.com/Haya372/ai-trial/backend/interface/handler"
	mw "github.com/Haya372/ai-trial/backend/interface/middleware"
	authuc "github.com/Haya372/ai-trial/backend/usecase/auth"
	eventshareuc "github.com/Haya372/ai-trial/backend/usecase/eventshare"
	eventsubscriptionuc "github.com/Haya372/ai-trial/backend/usecase/eventsubscription"
)

type eventSubscriptionTestDeps struct {
	router         *chi.Mux
	eventShareRepo domaineventshare.Repository
}

func buildEventSubscriptionTestRouter() eventSubscriptionTestDeps {
	logger := slog.New(slog.DiscardHandler)
	userRepo := repository.NewUserRepository(routeTestPool, testTracerProvider)
	sessRepo := repository.NewSessionRepository(routeTestPool, testTracerProvider)
	eventQueryRepo := repository.NewEventQueryRepository(routeTestPool, logger, testTracerProvider)
	eventShareRepo := repository.NewEventShareRepository(routeTestPool, testTracerProvider)
	subsRepo := repository.NewEventSubscriptionRepository(routeTestPool, testTracerProvider)
	txMgr := db.NewPgxTxManager(routeTestPool)

	signup := authuc.NewSignupCommand(userRepo, sessRepo, txMgr)
	loader := eventshareuc.NewShareTokenLoader(eventShareRepo, eventQueryRepo)
	subscribeToShare := eventsubscriptionuc.NewSubscribeToShareCommand(loader, subsRepo)

	auth := handler.NewAuthHandler(
		signup, authuc.NewLoginCommand(userRepo, sessRepo), authuc.NewLogoutCommand(sessRepo), logger,
	)
	sub := handler.NewEventSubscriptionHandler(subscribeToShare, logger)

	r := chi.NewRouter()
	r.Post("/auth/signup", auth.Signup)
	r.With(mw.RequireAuth(sessRepo, userRepo, logger)).Get("/auth/me", auth.GetMe)
	r.With(mw.RequireAuth(sessRepo, userRepo, logger)).Post("/shares/{token}/subscriptions", sub.SubscribeToShare)
	return eventSubscriptionTestDeps{router: r, eventShareRepo: eventShareRepo}
}

// insertRouteTestShare creates an event share directly through the
// repository (bypassing CreateShareCommand's future-only validation) so
// tests can set up an already-expired share.
func insertRouteTestShare(
	t *testing.T, repo domaineventshare.Repository, eventID uuid.UUID, expiresAt time.Time,
) string {
	t.Helper()
	token, err := domaineventshare.GenerateToken()
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}
	share, err := domaineventshare.New(uuid.New(), eventID, token.Hash(), expiresAt)
	if err != nil {
		t.Fatalf("build event share: %v", err)
	}
	if _, err := repo.Create(context.Background(), share); err != nil {
		t.Fatalf("insert event share: %v", err)
	}
	return token.String()
}

func subscribeReq(token string, cookie *http.Cookie) *http.Request {
	req := httptest.NewRequest(http.MethodPost, "/shares/"+token+"/subscriptions", nil)
	req.AddCookie(cookie)
	return req
}

func assertEventSubscriptionCountInDB(t *testing.T, eventID, userID string) int {
	t.Helper()
	var count int
	err := routeTestPool.QueryRow(context.Background(),
		"SELECT COUNT(*) FROM event_subscriptions WHERE event_id = $1 AND user_id = $2", eventID, userID,
	).Scan(&count)
	if err != nil {
		t.Fatalf("query event_subscriptions: %v", err)
	}
	return count
}

func TestRoute_SubscribeToShare_withoutSession_returns401(t *testing.T) {
	setupRouteTest(t)
	deps := buildEventSubscriptionTestRouter()

	req := httptest.NewRequest(http.MethodPost, "/shares/sometoken/subscriptions", nil)
	if code := responseCode(t, deps.router, req); code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", code)
	}
}

func TestRoute_SubscribeToShare_tokenNotFound_returns404(t *testing.T) {
	setupRouteTest(t)
	deps := buildEventSubscriptionTestRouter()
	cookie := signupAndGetCookie(t, deps.router, "sub-notfound@ex.com")

	rec := httptest.NewRecorder()
	deps.router.ServeHTTP(rec, subscribeReq("unknown-token", cookie))

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestRoute_SubscribeToShare_expiredToken_returns410(t *testing.T) {
	setupRouteTest(t)
	deps := buildEventSubscriptionTestRouter()
	ownerCookie := signupAndGetCookie(t, deps.router, "sub-owner-expired@ex.com")
	ownerID := getUserIDFromCookie(t, deps.router, ownerCookie)
	viewerCookie := signupAndGetCookie(t, deps.router, "sub-viewer-expired@ex.com")

	now := time.Now().UTC().Truncate(time.Second)
	eventID := insertRouteTestEvent(t, ownerID, "Expired share event", now, now.Add(time.Hour))
	token := insertRouteTestShare(t, deps.eventShareRepo, uuid.MustParse(eventID), now.Add(-time.Minute))

	rec := httptest.NewRecorder()
	deps.router.ServeHTTP(rec, subscribeReq(token, viewerCookie))

	if rec.Code != http.StatusGone {
		t.Fatalf("expected 410, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestRoute_SubscribeToShare_ownEvent_returns403(t *testing.T) {
	setupRouteTest(t)
	deps := buildEventSubscriptionTestRouter()
	ownerCookie := signupAndGetCookie(t, deps.router, "sub-owner-self@ex.com")
	ownerID := getUserIDFromCookie(t, deps.router, ownerCookie)

	now := time.Now().UTC().Truncate(time.Second)
	eventID := insertRouteTestEvent(t, ownerID, "Own event", now, now.Add(time.Hour))
	token := insertRouteTestShare(t, deps.eventShareRepo, uuid.MustParse(eventID), now.Add(time.Hour))

	rec := httptest.NewRecorder()
	deps.router.ServeHTTP(rec, subscribeReq(token, ownerCookie))

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestRoute_SubscribeToShare_notSubscribed_returns201AndPersists(t *testing.T) {
	setupRouteTest(t)
	deps := buildEventSubscriptionTestRouter()
	ownerCookie := signupAndGetCookie(t, deps.router, "sub-owner-new@ex.com")
	ownerID := getUserIDFromCookie(t, deps.router, ownerCookie)
	viewerCookie := signupAndGetCookie(t, deps.router, "sub-viewer-new@ex.com")
	viewerID := getUserIDFromCookie(t, deps.router, viewerCookie)

	now := time.Now().UTC().Truncate(time.Second)
	eventID := insertRouteTestEvent(t, ownerID, "Shared meeting", now, now.Add(time.Hour))
	token := insertRouteTestShare(t, deps.eventShareRepo, uuid.MustParse(eventID), now.Add(time.Hour))

	rec := httptest.NewRecorder()
	deps.router.ServeHTTP(rec, subscribeReq(token, viewerCookie))

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	var body struct {
		ID      string `json:"id"`
		EventID string `json:"eventId"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.EventID != eventID {
		t.Errorf("expected eventId %s, got %s", eventID, body.EventID)
	}
	if count := assertEventSubscriptionCountInDB(t, eventID, viewerID); count != 1 {
		t.Errorf("expected 1 event_subscriptions row, got %d", count)
	}
}

func TestRoute_SubscribeToShare_alreadySubscribed_returns200Idempotent(t *testing.T) {
	setupRouteTest(t)
	deps := buildEventSubscriptionTestRouter()
	ownerCookie := signupAndGetCookie(t, deps.router, "sub-owner-dup@ex.com")
	ownerID := getUserIDFromCookie(t, deps.router, ownerCookie)
	viewerCookie := signupAndGetCookie(t, deps.router, "sub-viewer-dup@ex.com")
	viewerID := getUserIDFromCookie(t, deps.router, viewerCookie)

	now := time.Now().UTC().Truncate(time.Second)
	eventID := insertRouteTestEvent(t, ownerID, "Shared meeting", now, now.Add(time.Hour))
	token := insertRouteTestShare(t, deps.eventShareRepo, uuid.MustParse(eventID), now.Add(time.Hour))

	firstRec := httptest.NewRecorder()
	deps.router.ServeHTTP(firstRec, subscribeReq(token, viewerCookie))
	if firstRec.Code != http.StatusCreated {
		t.Fatalf("expected first request to return 201, got %d: %s", firstRec.Code, firstRec.Body.String())
	}
	var firstBody struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(firstRec.Body).Decode(&firstBody); err != nil {
		t.Fatalf("decode first response: %v", err)
	}

	secondRec := httptest.NewRecorder()
	deps.router.ServeHTTP(secondRec, subscribeReq(token, viewerCookie))
	if secondRec.Code != http.StatusOK {
		t.Fatalf("expected second request to return 200, got %d: %s", secondRec.Code, secondRec.Body.String())
	}
	var secondBody struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(secondRec.Body).Decode(&secondBody); err != nil {
		t.Fatalf("decode second response: %v", err)
	}
	if secondBody.ID != firstBody.ID {
		t.Errorf("expected same subscription id on repeat add, got %s then %s", firstBody.ID, secondBody.ID)
	}
	if count := assertEventSubscriptionCountInDB(t, eventID, viewerID); count != 1 {
		t.Errorf("expected 1 event_subscriptions row after duplicate add, got %d", count)
	}
}
