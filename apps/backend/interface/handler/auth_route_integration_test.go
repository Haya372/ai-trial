//go:build integration

package handler_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/mock/gomock"

	"github.com/Haya372/ai-trial/backend/domain/session"
	sessionmock "github.com/Haya372/ai-trial/backend/domain/session/generated"
	"github.com/Haya372/ai-trial/backend/domain/user"
	usermock "github.com/Haya372/ai-trial/backend/domain/user/generated"
	"github.com/Haya372/ai-trial/backend/interface/handler"
	mw "github.com/Haya372/ai-trial/backend/interface/middleware"
)

func buildRouter(auth *handler.AuthHandler, sessRepo session.Repository, userRepo user.Repository) *chi.Mux {
	r := chi.NewRouter()
	r.Post("/auth/signup", auth.Signup)
	r.Post("/auth/login", auth.Login)
	r.With(mw.RequireAuth(sessRepo, userRepo)).Post("/auth/logout", auth.Logout)
	r.With(mw.RequireAuth(sessRepo, userRepo)).Get("/auth/me", auth.GetMe)
	return r
}

func TestRoute_Logout_withoutSession_returns401(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockSessRepo := sessionmock.NewMockRepository(ctrl)
	mockUserRepo := usermock.NewMockRepository(ctrl)

	router := buildRouter(
		handler.NewAuthHandler(&stubSignupExec{}, &stubLoginExec{}, &stubLogoutExec{}),
		mockSessRepo,
		mockUserRepo,
	)

	req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestRoute_Logout_withValidSession_returns204(t *testing.T) {
	ctrl := gomock.NewController(t)
	sessID := uuid.New()
	userID := uuid.New()

	sess := session.New(sessID, userID, time.Now().Add(time.Hour))
	email, _ := user.NewEmail("u@ex.com")
	u := user.New(userID, email, "U", "hash")

	mockSessRepo := sessionmock.NewMockRepository(ctrl)
	mockUserRepo := usermock.NewMockRepository(ctrl)
	mockSessRepo.EXPECT().FindActiveByID(gomock.Any(), sessID).Return(sess, nil)
	mockUserRepo.EXPECT().FindByID(gomock.Any(), userID).Return(u, nil)

	logoutStub := &stubLogoutExec{fn: func(_ context.Context, _ uuid.UUID) error { return nil }}

	router := buildRouter(
		handler.NewAuthHandler(&stubSignupExec{}, &stubLoginExec{}, logoutStub),
		mockSessRepo,
		mockUserRepo,
	)

	req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
	req.AddCookie(&http.Cookie{Name: testSessionCookieName, Value: sessID.String()}) //nolint:gosec
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestRoute_GetMe_withoutSession_returns401(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockSessRepo := sessionmock.NewMockRepository(ctrl)
	mockUserRepo := usermock.NewMockRepository(ctrl)

	router := buildRouter(
		handler.NewAuthHandler(&stubSignupExec{}, &stubLoginExec{}, &stubLogoutExec{}),
		mockSessRepo,
		mockUserRepo,
	)

	req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}
