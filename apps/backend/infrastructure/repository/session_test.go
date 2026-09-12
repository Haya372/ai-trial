//go:build integration

package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/Haya372/ai-trial/backend/domain/user"
	"github.com/Haya372/ai-trial/backend/infrastructure/repository"
)

func TestSessionRepository_Create_and_FindByID(t *testing.T) {
	setupTest(t)
	userRepo := repository.NewUserRepository(testPool)
	sessRepo := repository.NewSessionRepository(testPool)

	email, _ := user.NewEmail("sess@example.com")
	password, _ := user.NewPassword("SecurePass1!")
	u, err := userRepo.Create(context.Background(), email, "Sess User", password)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}

	expiresAt := time.Now().Add(30 * 24 * time.Hour).UTC().Truncate(time.Microsecond)
	sess, err := sessRepo.Create(context.Background(), u.ID(), expiresAt)
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	if sess.UserID() != u.ID() {
		t.Errorf("expected user ID %v, got %v", u.ID(), sess.UserID())
	}

	found, err := sessRepo.FindByID(context.Background(), sess.ID())
	if err != nil {
		t.Fatalf("find session: %v", err)
	}
	if found == nil {
		t.Fatal("expected session, got nil")
	}
	if found.ID() != sess.ID() {
		t.Errorf("expected session ID %v, got %v", sess.ID(), found.ID())
	}
}

func TestSessionRepository_Delete_removes_session(t *testing.T) {
	setupTest(t)
	userRepo := repository.NewUserRepository(testPool)
	sessRepo := repository.NewSessionRepository(testPool)

	email, _ := user.NewEmail("del@example.com")
	password, _ := user.NewPassword("SecurePass1!")
	u, _ := userRepo.Create(context.Background(), email, "Del User", password)

	expiresAt := time.Now().Add(30 * 24 * time.Hour)
	sess, _ := sessRepo.Create(context.Background(), u.ID(), expiresAt)

	if err := sessRepo.Delete(context.Background(), sess.ID()); err != nil {
		t.Fatalf("delete: %v", err)
	}

	found, err := sessRepo.FindByID(context.Background(), sess.ID())
	if err != nil {
		t.Fatalf("unexpected error after delete: %v", err)
	}
	if found != nil {
		t.Error("expected nil session after delete, but found one")
	}
}

func TestSessionRepository_FindByID_expiredSession_returnsNil(t *testing.T) {
	setupTest(t)
	userRepo := repository.NewUserRepository(testPool)
	sessRepo := repository.NewSessionRepository(testPool)

	email, _ := user.NewEmail("expired@example.com")
	password, _ := user.NewPassword("SecurePass1!")
	u, _ := userRepo.Create(context.Background(), email, "Expired User", password)

	expiresAt := time.Now().Add(-1 * time.Hour)
	sess, _ := sessRepo.Create(context.Background(), u.ID(), expiresAt)

	found, err := sessRepo.FindByID(context.Background(), sess.ID())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found != nil {
		t.Error("expected nil for expired session, but found one")
	}
}
