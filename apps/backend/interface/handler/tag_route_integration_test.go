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
	"sync"
	"testing"

	"github.com/go-chi/chi/v5"

	domaintag "github.com/Haya372/ai-trial/backend/domain/tag"
	"github.com/Haya372/ai-trial/backend/infrastructure/db"
	"github.com/Haya372/ai-trial/backend/infrastructure/repository"
	"github.com/Haya372/ai-trial/backend/interface/handler"
	mw "github.com/Haya372/ai-trial/backend/interface/middleware"
	"github.com/Haya372/ai-trial/backend/usecase"
	authuc "github.com/Haya372/ai-trial/backend/usecase/auth"
	taguc "github.com/Haya372/ai-trial/backend/usecase/tag"
)

func buildTagTestRouter() *chi.Mux {
	logger := slog.New(slog.DiscardHandler)
	userRepo := repository.NewUserRepository(routeTestPool, testTracerProvider)
	sessRepo := repository.NewSessionRepository(routeTestPool, testTracerProvider)
	tagRepo := repository.NewTagRepository(routeTestPool, logger, testTracerProvider)
	txMgr := db.NewPgxTxManager(routeTestPool)

	signup := authuc.NewSignupCommand(userRepo, sessRepo, txMgr, usecase.RealClock{})
	login := authuc.NewLoginCommand(userRepo, sessRepo, usecase.RealClock{})
	logout := authuc.NewLogoutCommand(sessRepo)

	auth := handler.NewAuthHandler(signup, login, logout, logger)
	tg := handler.NewTagHandler(taguc.NewCreateTagCommand(tagRepo, txMgr), logger)

	r := chi.NewRouter()
	r.Post("/auth/signup", auth.Signup)
	r.With(mw.RequireAuth(sessRepo, userRepo, logger)).Get("/auth/me", auth.GetMe)
	r.With(mw.RequireAuth(sessRepo, userRepo, logger)).Post("/tags", tg.CreateTag)
	return r
}

func postTag(t *testing.T, router *chi.Mux, cookie *http.Cookie, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/tags", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if cookie != nil {
		req.AddCookie(cookie)
	}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func insertRouteTestTags(t *testing.T, userID string, n int) {
	t.Helper()
	for i := range n {
		if _, err := routeTestPool.Exec(context.Background(),
			"INSERT INTO tags (user_id, name, color) VALUES ($1, $2, 'red')",
			userID, fmt.Sprintf("tag-%02d", i),
		); err != nil {
			t.Fatalf("insert tag: %v", err)
		}
	}
}

func tagRowCount(t *testing.T, userID string) int {
	t.Helper()
	var count int
	if err := routeTestPool.QueryRow(context.Background(),
		"SELECT COUNT(*) FROM tags WHERE user_id = $1", userID,
	).Scan(&count); err != nil {
		t.Fatalf("count tags: %v", err)
	}
	return count
}

func TestRoute_CreateTag_unauthenticated_returns401(t *testing.T) {
	setupRouteTest(t)
	router := buildTagTestRouter()

	rec := postTag(t, router, nil, `{"name":"work","color":"red"}`)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestRoute_CreateTag_success_returns201AndStoresTrimmedName(t *testing.T) {
	setupRouteTest(t)
	router := buildTagTestRouter()
	cookie := signupAndGetCookie(t, router, "tag-create@ex.com")
	userID := getUserIDFromCookie(t, router, cookie)

	rec := postTag(t, router, cookie, `{"name":"  Work  ","color":"indigo"}`)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	var body struct {
		ID         string `json:"id"`
		Name       string `json:"name"`
		Color      string `json:"color"`
		EventCount *int   `json:"eventCount"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Name != testTagName || body.Color != "indigo" {
		t.Errorf("got name=%q color=%q, want Work/indigo", body.Name, body.Color)
	}
	if body.EventCount == nil || *body.EventCount != 0 {
		t.Errorf("eventCount: got %v, want 0", body.EventCount)
	}

	var storedUserID, storedName, storedColor string
	if err := routeTestPool.QueryRow(context.Background(),
		"SELECT user_id, name, color FROM tags WHERE id = $1", body.ID,
	).Scan(&storedUserID, &storedName, &storedColor); err != nil {
		t.Fatalf("select stored tag: %v", err)
	}
	if storedUserID != userID || storedName != testTagName || storedColor != "indigo" {
		t.Errorf("stored row: got (%s, %q, %q), want (%s, Work, indigo)", storedUserID, storedName, storedColor, userID)
	}
}

func TestRoute_CreateTag_invalidInput_returns400WithDetails(t *testing.T) {
	setupRouteTest(t)
	router := buildTagTestRouter()
	cookie := signupAndGetCookie(t, router, "tag-invalid@ex.com")

	rec := postTag(t, router, cookie, `{"name":"   ","color":"black"}`)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
	var body struct {
		Code    string `json:"code"`
		Details []struct {
			Field string `json:"field"`
			Code  string `json:"code"`
		} `json:"details"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Code != codeValidationError || len(body.Details) != 2 {
		t.Fatalf("got %+v, want VALIDATION_ERROR with 2 details", body)
	}
	if body.Details[0].Field != fieldTagName || body.Details[0].Code != "REQUIRED" ||
		body.Details[1].Field != fieldTagColor || body.Details[1].Code != codeInvalidFormat {
		t.Errorf("details: got %+v", body.Details)
	}
}

func TestRoute_CreateTag_duplicateNameIgnoringCase_returns409(t *testing.T) {
	setupRouteTest(t)
	router := buildTagTestRouter()
	cookie := signupAndGetCookie(t, router, "tag-dup@ex.com")

	if rec := postTag(t, router, cookie, `{"name":"Work","color":"red"}`); rec.Code != http.StatusCreated {
		t.Fatalf("first create: expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	rec := postTag(t, router, cookie, `{"name":"WORK","color":"blue"}`)

	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestRoute_CreateTag_sameNameAsAnotherUsersTag_returns201(t *testing.T) {
	setupRouteTest(t)
	router := buildTagTestRouter()
	ownerCookie := signupAndGetCookie(t, router, "tag-owner@ex.com")
	otherCookie := signupAndGetCookie(t, router, "tag-other@ex.com")
	otherID := getUserIDFromCookie(t, router, otherCookie)
	insertRouteTestTags(t, otherID, domaintag.MaxTagsPerUser)

	if rec := postTag(t, router, otherCookie, `{"name":"Work","color":"red"}`); rec.Code != http.StatusConflict {
		t.Fatalf("other user at the limit: expected 409, got %d", rec.Code)
	}
	rec := postTag(t, router, ownerCookie, `{"name":"tag-00","color":"red"}`)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 since names and the limit are per user, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestRoute_CreateTag_atLimit_returns409(t *testing.T) {
	setupRouteTest(t)
	router := buildTagTestRouter()
	cookie := signupAndGetCookie(t, router, "tag-limit@ex.com")
	userID := getUserIDFromCookie(t, router, cookie)
	insertRouteTestTags(t, userID, domaintag.MaxTagsPerUser)

	rec := postTag(t, router, cookie, `{"name":"one more","color":"red"}`)

	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d: %s", rec.Code, rec.Body.String())
	}
	if got := tagRowCount(t, userID); got != domaintag.MaxTagsPerUser {
		t.Errorf("tag count: got %d, want %d", got, domaintag.MaxTagsPerUser)
	}
}

func TestRoute_CreateTag_concurrentCreates_doNotExceedLimit(t *testing.T) {
	setupRouteTest(t)
	router := buildTagTestRouter()
	cookie := signupAndGetCookie(t, router, "tag-race@ex.com")
	userID := getUserIDFromCookie(t, router, cookie)
	insertRouteTestTags(t, userID, domaintag.MaxTagsPerUser-1)

	const workers = 5
	codes := make([]int, workers)
	var wg sync.WaitGroup
	for i := range workers {
		wg.Go(func() {
			codes[i] = postTag(t, router, cookie, fmt.Sprintf(`{"name":"race-%d","color":"red"}`, i)).Code
		})
	}
	wg.Wait()

	created := 0
	for _, c := range codes {
		switch c {
		case http.StatusCreated:
			created++
		case http.StatusConflict:
		default:
			t.Errorf("unexpected status %d", c)
		}
	}
	if created != 1 {
		t.Errorf("expected exactly 1 create to succeed, got %d (codes: %v)", created, codes)
	}
	if got := tagRowCount(t, userID); got != domaintag.MaxTagsPerUser {
		t.Errorf("tag count: got %d, want %d", got, domaintag.MaxTagsPerUser)
	}
}
