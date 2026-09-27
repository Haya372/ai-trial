package middleware_test

import (
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
	"github.com/Haya372/ai-trial/backend/interface/middleware"
)

func TestOptionalAuth_noCookie_callsNextWithoutUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	sessRepo := sessionmock.NewMockRepository(ctrl)
	userRepo := usermock.NewMockRepository(ctrl)

	called := false
	var capturedUser user.User
	mw := middleware.OptionalAuth(sessRepo, userRepo, testLogger)
	h := mw(nextHandlerCapture(&called, &capturedUser))

	req := requestWithCookie(t, "")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if !called {
		t.Error("expected next handler to be called")
	}
	if capturedUser != nil {
		t.Errorf("expected nil user in context, got %v", capturedUser)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestOptionalAuth_invalidUUID_callsNextWithoutUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	sessRepo := sessionmock.NewMockRepository(ctrl)
	userRepo := usermock.NewMockRepository(ctrl)

	called := false
	var capturedUser user.User
	mw := middleware.OptionalAuth(sessRepo, userRepo, testLogger)
	h := mw(nextHandlerCapture(&called, &capturedUser))

	req := requestWithCookie(t, "not-a-uuid")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if !called {
		t.Error("expected next handler to be called")
	}
	if capturedUser != nil {
		t.Errorf("expected nil user in context, got %v", capturedUser)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestOptionalAuth_sessionNotFound_callsNextWithoutUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	sessRepo := sessionmock.NewMockRepository(ctrl)
	userRepo := usermock.NewMockRepository(ctrl)
	sessRepo.EXPECT().FindActiveByID(gomock.Any(), gomock.Any()).Return(nil, nil)

	called := false
	var capturedUser user.User
	mw := middleware.OptionalAuth(sessRepo, userRepo, testLogger)
	h := mw(nextHandlerCapture(&called, &capturedUser))

	req := requestWithCookie(t, uuid.New().String())
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if !called {
		t.Error("expected next handler to be called")
	}
	if capturedUser != nil {
		t.Errorf("expected nil user in context, got %v", capturedUser)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestOptionalAuth_sessionRepoError_callsNextWithoutUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	sessRepo := sessionmock.NewMockRepository(ctrl)
	userRepo := usermock.NewMockRepository(ctrl)
	sessRepo.EXPECT().FindActiveByID(gomock.Any(), gomock.Any()).Return(nil, errDB)

	called := false
	var capturedUser user.User
	mw := middleware.OptionalAuth(sessRepo, userRepo, testLogger)
	h := mw(nextHandlerCapture(&called, &capturedUser))

	req := requestWithCookie(t, uuid.New().String())
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if !called {
		t.Error("expected next handler to be called")
	}
	if capturedUser != nil {
		t.Errorf("expected nil user in context, got %v", capturedUser)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestOptionalAuth_userNotFound_callsNextWithoutUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	sessID := uuid.New()
	userID := uuid.New()

	sess := session.New(sessID, userID, time.Now().Add(time.Hour))
	sessRepo := sessionmock.NewMockRepository(ctrl)
	userRepo := usermock.NewMockRepository(ctrl)
	sessRepo.EXPECT().FindActiveByID(gomock.Any(), sessID).Return(sess, nil)
	userRepo.EXPECT().FindByID(gomock.Any(), userID).Return(nil, nil)

	called := false
	var capturedUser user.User
	mw := middleware.OptionalAuth(sessRepo, userRepo, testLogger)
	h := mw(nextHandlerCapture(&called, &capturedUser))

	req := requestWithCookie(t, sessID.String())
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if !called {
		t.Error("expected next handler to be called")
	}
	if capturedUser != nil {
		t.Errorf("expected nil user in context, got %v", capturedUser)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestOptionalAuth_validSession_setsUserInContext(t *testing.T) {
	assertValidSessionSetsUser(t, middleware.OptionalAuth)
}
