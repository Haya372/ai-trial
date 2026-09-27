package middleware_test

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"

	"github.com/Haya372/ai-trial/backend/domain/session"
	sessionmock "github.com/Haya372/ai-trial/backend/domain/session/generated"
	"github.com/Haya372/ai-trial/backend/domain/user"
	usermock "github.com/Haya372/ai-trial/backend/domain/user/generated"
	"github.com/Haya372/ai-trial/backend/interface/ctxkey"
	"github.com/Haya372/ai-trial/backend/interface/middleware"
)

// authMiddlewareFactory matches the constructor signature shared by
// middleware.RequireAuth and middleware.OptionalAuth.
type authMiddlewareFactory func(session.Repository, user.Repository, *slog.Logger) func(http.Handler) http.Handler

// setupValidSessionMocks builds a session+user mock pair that resolves
// successfully, shared by the "valid session" tests of both middlewares.
func setupValidSessionMocks(ctrl *gomock.Controller) (
	uuid.UUID, uuid.UUID, *sessionmock.MockRepository, *usermock.MockRepository,
) {
	sessID := uuid.New()
	userID := uuid.New()
	sess := session.New(sessID, userID, time.Now().Add(time.Hour))
	email, _ := user.NewEmail("u@ex.com")
	u := user.New(userID, email, "U", "hash")

	sessRepo := sessionmock.NewMockRepository(ctrl)
	userRepo := usermock.NewMockRepository(ctrl)
	sessRepo.EXPECT().FindActiveByID(gomock.Any(), sessID).Return(sess, nil)
	userRepo.EXPECT().FindByID(gomock.Any(), userID).Return(u, nil)
	return sessID, userID, sessRepo, userRepo
}

// assertValidSessionSetsUser exercises the "valid session" happy path shared
// by RequireAuth and OptionalAuth: next is called with the user in context.
func assertValidSessionSetsUser(t *testing.T, buildMW authMiddlewareFactory) {
	t.Helper()
	ctrl := gomock.NewController(t)
	sessID, userID, sessRepo, userRepo := setupValidSessionMocks(ctrl)

	called := false
	var capturedUser user.User
	mw := buildMW(sessRepo, userRepo, testLogger)
	h := mw(nextHandlerCapture(&called, &capturedUser))

	req := requestWithCookie(t, sessID.String())
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if !called {
		t.Error("expected next handler to be called")
	}
	if capturedUser == nil || capturedUser.ID() != userID {
		t.Errorf("expected user %v in context, got %v", userID, capturedUser)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

const testSessionCookieName = "session_id"

func nextHandlerCapture(called *bool, capturedUser *user.User) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*called = true
		u, _ := r.Context().Value(ctxkey.User).(user.User)
		*capturedUser = u
		w.WriteHeader(http.StatusOK)
	})
}

func requestWithCookie(t *testing.T, cookieValue string) *http.Request {
	t.Helper()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/auth/me", nil)
	if cookieValue != "" {
		req.AddCookie(&http.Cookie{Name: testSessionCookieName, Value: cookieValue}) //nolint:gosec
	}
	return req
}

func TestRequireAuth_validSession_setsUserInContext_and_calls_next(t *testing.T) {
	assertValidSessionSetsUser(t, middleware.RequireAuth)
}

func TestRequireAuth_noSessionCookie_returns401(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockSessRepo := sessionmock.NewMockRepository(ctrl)
	mockUserRepo := usermock.NewMockRepository(ctrl)

	mw := middleware.RequireAuth(mockSessRepo, mockUserRepo, testLogger)
	h := mw(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		t.Error("next handler should not be called")
	}))

	req := requestWithCookie(t, "")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestRequireAuth_invalidUUID_returns401(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockSessRepo := sessionmock.NewMockRepository(ctrl)
	mockUserRepo := usermock.NewMockRepository(ctrl)

	mw := middleware.RequireAuth(mockSessRepo, mockUserRepo, testLogger)
	h := mw(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		t.Error("next handler should not be called")
	}))

	req := requestWithCookie(t, "not-a-uuid")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestRequireAuth_sessionNotFound_returns401(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockSessRepo := sessionmock.NewMockRepository(ctrl)
	mockUserRepo := usermock.NewMockRepository(ctrl)
	mockSessRepo.EXPECT().FindActiveByID(gomock.Any(), gomock.Any()).Return(nil, nil)

	mw := middleware.RequireAuth(mockSessRepo, mockUserRepo, testLogger)
	h := mw(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		t.Error("next handler should not be called")
	}))

	req := requestWithCookie(t, uuid.New().String())
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestRequireAuth_sessionRepoError_returns401(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockSessRepo := sessionmock.NewMockRepository(ctrl)
	mockUserRepo := usermock.NewMockRepository(ctrl)
	mockSessRepo.EXPECT().FindActiveByID(gomock.Any(), gomock.Any()).Return(nil, errDB)

	mw := middleware.RequireAuth(mockSessRepo, mockUserRepo, testLogger)
	h := mw(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		t.Error("next handler should not be called")
	}))

	req := requestWithCookie(t, uuid.New().String())
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestRequireAuth_userNotFound_returns401(t *testing.T) {
	ctrl := gomock.NewController(t)
	sessID := uuid.New()
	userID := uuid.New()

	sess := session.New(sessID, userID, time.Now().Add(time.Hour))
	mockSessRepo := sessionmock.NewMockRepository(ctrl)
	mockUserRepo := usermock.NewMockRepository(ctrl)
	mockSessRepo.EXPECT().FindActiveByID(gomock.Any(), sessID).Return(sess, nil)
	mockUserRepo.EXPECT().FindByID(gomock.Any(), userID).Return(nil, user.ErrUserNotFound)

	mw := middleware.RequireAuth(mockSessRepo, mockUserRepo, testLogger)
	h := mw(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		t.Error("next handler should not be called")
	}))

	req := requestWithCookie(t, sessID.String())
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestRequireAuth_userNilWithNoError_returns401(t *testing.T) {
	ctrl := gomock.NewController(t)
	sessID := uuid.New()
	userID := uuid.New()

	sess := session.New(sessID, userID, time.Now().Add(time.Hour))
	mockSessRepo := sessionmock.NewMockRepository(ctrl)
	mockUserRepo := usermock.NewMockRepository(ctrl)
	mockSessRepo.EXPECT().FindActiveByID(gomock.Any(), sessID).Return(sess, nil)
	mockUserRepo.EXPECT().FindByID(gomock.Any(), userID).Return(nil, nil)

	mw := middleware.RequireAuth(mockSessRepo, mockUserRepo, testLogger)
	h := mw(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		t.Error("next handler should not be called")
	}))

	req := requestWithCookie(t, sessID.String())
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}
