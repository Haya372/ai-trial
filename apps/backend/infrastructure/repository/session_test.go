//go:build integration

package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/Haya372/ai-trial/backend/domain/session"
	"github.com/Haya372/ai-trial/backend/domain/user"
	"github.com/Haya372/ai-trial/backend/infrastructure/repository"
)

func createSessionTestUser(t *testing.T, email string) user.User {
	t.Helper()
	userRepo := repository.NewUserRepository(testPool, testTracerProvider)
	e, _ := user.NewEmail(email)
	password, _ := user.NewPassword("SecurePass1!")
	u, err := userRepo.Create(context.Background(), e, "Sess User", password)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	return u
}

func TestSessionRepository_Create_and_FindActiveByTokenHash(t *testing.T) {
	setupTest(t)
	sessRepo := repository.NewSessionRepository(testPool, testTracerProvider)
	u := createSessionTestUser(t, "sess@example.com")

	token := session.NewToken("create-find-token")
	expiresAt := time.Now().Add(30 * 24 * time.Hour).UTC().Truncate(time.Microsecond)
	sess, err := sessRepo.Create(context.Background(), u.ID(), token.Hash(), expiresAt)
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	if sess.UserID() != u.ID() {
		t.Errorf("expected user ID %v, got %v", u.ID(), sess.UserID())
	}

	found, err := sessRepo.FindActiveByTokenHash(context.Background(), token.Hash())
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

func TestSessionRepository_Create_storesOnlyTheHash(t *testing.T) {
	setupTest(t)
	sessRepo := repository.NewSessionRepository(testPool, testTracerProvider)
	u := createSessionTestUser(t, "hash-only@example.com")

	token := session.NewToken("plaintext-must-not-be-stored")
	expiresAt := time.Now().Add(time.Hour)
	if _, err := sessRepo.Create(context.Background(), u.ID(), token.Hash(), expiresAt); err != nil {
		t.Fatalf("create session: %v", err)
	}

	var plaintextRows int
	err := testPool.QueryRow(context.Background(),
		`SELECT COUNT(*) FROM sessions WHERE token_hash = $1 OR id::text = $1`, token.String(),
	).Scan(&plaintextRows)
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	if plaintextRows != 0 {
		t.Errorf("plaintext token must not be stored, found %d matching rows", plaintextRows)
	}

	var hashRows int
	err = testPool.QueryRow(context.Background(),
		`SELECT COUNT(*) FROM sessions WHERE token_hash = $1`, token.Hash().String(),
	).Scan(&hashRows)
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	if hashRows != 1 {
		t.Errorf("expected 1 row stored under the token hash, got %d", hashRows)
	}
}

func TestSessionRepository_FindActiveByTokenHash_unknownHash_returnsNil(t *testing.T) {
	setupTest(t)
	sessRepo := repository.NewSessionRepository(testPool, testTracerProvider)

	found, err := sessRepo.FindActiveByTokenHash(context.Background(), session.NewToken("unknown").Hash())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found != nil {
		t.Error("expected nil for an unknown token hash")
	}
}

func TestSessionRepository_FindActiveByTokenHash_plaintextToken_returnsNil(t *testing.T) {
	setupTest(t)
	sessRepo := repository.NewSessionRepository(testPool, testTracerProvider)
	u := createSessionTestUser(t, "plaintext-lookup@example.com")

	token := session.NewToken("lookup-by-plaintext")
	if _, err := sessRepo.Create(context.Background(), u.ID(), token.Hash(), time.Now().Add(time.Hour)); err != nil {
		t.Fatalf("create session: %v", err)
	}

	// Looking up by the plaintext value (as if it were a hash) must not match.
	found, err := sessRepo.FindActiveByTokenHash(context.Background(), session.NewTokenHash(token.String()))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found != nil {
		t.Error("expected nil when looking up by the plaintext token")
	}
}

func TestSessionRepository_FindActiveByTokenHash_expiredSession_returnsNil(t *testing.T) {
	setupTest(t)
	sessRepo := repository.NewSessionRepository(testPool, testTracerProvider)
	u := createSessionTestUser(t, "expired-active@example.com")

	token := session.NewToken("expired-token")
	expiresAt := time.Now().Add(-1 * time.Hour).UTC().Truncate(time.Microsecond)
	if _, err := sessRepo.Create(context.Background(), u.ID(), token.Hash(), expiresAt); err != nil {
		t.Fatalf("create session: %v", err)
	}

	found, err := sessRepo.FindActiveByTokenHash(context.Background(), token.Hash())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found != nil {
		t.Error("expected nil for expired session, but found one")
	}
}

func TestSessionRepository_DeleteByTokenHash_removesSession(t *testing.T) {
	setupTest(t)
	sessRepo := repository.NewSessionRepository(testPool, testTracerProvider)
	u := createSessionTestUser(t, "del@example.com")

	token := session.NewToken("delete-token")
	if _, err := sessRepo.Create(context.Background(), u.ID(), token.Hash(), time.Now().Add(time.Hour)); err != nil {
		t.Fatalf("create session: %v", err)
	}

	if err := sessRepo.DeleteByTokenHash(context.Background(), token.Hash()); err != nil {
		t.Fatalf("delete: %v", err)
	}

	found, err := sessRepo.FindActiveByTokenHash(context.Background(), token.Hash())
	if err != nil {
		t.Fatalf("unexpected error after delete: %v", err)
	}
	if found != nil {
		t.Error("expected nil session after delete, but found one")
	}
}

func TestSessionRepository_DeleteByTokenHash_onlyRemovesMatchingSession(t *testing.T) {
	setupTest(t)
	sessRepo := repository.NewSessionRepository(testPool, testTracerProvider)
	u := createSessionTestUser(t, "del-other@example.com")

	keep := session.NewToken("keep-token")
	remove := session.NewToken("remove-token")
	for _, tk := range []session.Token{keep, remove} {
		if _, err := sessRepo.Create(context.Background(), u.ID(), tk.Hash(), time.Now().Add(time.Hour)); err != nil {
			t.Fatalf("create session: %v", err)
		}
	}

	if err := sessRepo.DeleteByTokenHash(context.Background(), remove.Hash()); err != nil {
		t.Fatalf("delete: %v", err)
	}

	found, err := sessRepo.FindActiveByTokenHash(context.Background(), keep.Hash())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found == nil {
		t.Error("expected the other session to remain")
	}
}

func TestSessionRepository_Create_duplicateTokenHash_returnsError(t *testing.T) {
	setupTest(t)
	sessRepo := repository.NewSessionRepository(testPool, testTracerProvider)
	u := createSessionTestUser(t, "dup@example.com")

	token := session.NewToken("duplicate-token")
	if _, err := sessRepo.Create(context.Background(), u.ID(), token.Hash(), time.Now().Add(time.Hour)); err != nil {
		t.Fatalf("first create: %v", err)
	}
	if _, err := sessRepo.Create(context.Background(), u.ID(), token.Hash(), time.Now().Add(time.Hour)); err == nil {
		t.Fatal("expected a unique-constraint error for a duplicate token hash")
	}
}
