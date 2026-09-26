//go:build integration

package repository_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/Haya372/ai-trial/backend/domain/eventshare"
	"github.com/Haya372/ai-trial/backend/infrastructure/repository"
)

func newTestEventShare(t *testing.T, eventID uuid.UUID, token string, expiresAt time.Time) eventshare.EventShare {
	t.Helper()
	s, err := eventshare.New(uuid.New(), eventID, eventshare.HashToken(token), expiresAt)
	if err != nil {
		t.Fatalf("build domain eventshare: %v", err)
	}
	return s
}

func TestEventShareRepository_Create_PersistsAndReturnsEventShare(t *testing.T) {
	setupTest(t)
	u := createTestUser(t)
	now := time.Now().UTC().Truncate(time.Second)
	eventID := insertEvent(t, u.ID(), "Share target event", now, now.Add(time.Hour))

	repo := repository.NewEventShareRepository(testPool, testTracerProvider)

	id := uuid.New()
	expiresAt := time.Now().UTC().Add(time.Hour).Truncate(time.Second)
	tokenHash := eventshare.HashToken("plain-token-create")
	s, err := eventshare.New(id, eventID, tokenHash, expiresAt)
	if err != nil {
		t.Fatalf("build domain eventshare: %v", err)
	}

	saved, err := repo.Create(context.Background(), s)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if saved.ID() != id {
		t.Errorf("ID mismatch: got %v, want %v", saved.ID(), id)
	}
	if saved.EventID() != eventID {
		t.Errorf("EventID mismatch: got %v, want %v", saved.EventID(), eventID)
	}
	if saved.TokenHash() != tokenHash {
		t.Errorf("TokenHash mismatch: got %q, want %q", saved.TokenHash(), tokenHash)
	}
	if !saved.ExpiresAt().Equal(expiresAt) {
		t.Errorf("ExpiresAt mismatch: got %v, want %v", saved.ExpiresAt(), expiresAt)
	}
}

func TestEventShareRepository_Create_NonExistentEventID_ReturnsError(t *testing.T) {
	setupTest(t)

	repo := repository.NewEventShareRepository(testPool, testTracerProvider)

	s := newTestEventShare(t, uuid.New(), "plain-token-orphan", time.Now().UTC().Add(time.Hour))

	_, err := repo.Create(context.Background(), s)
	if err == nil {
		t.Fatal("expected an error for a non-existent event_id, got nil")
	}
}

func TestEventShareRepository_FindByToken_ReturnsMatchingShare(t *testing.T) {
	setupTest(t)
	u := createTestUser(t)
	now := time.Now().UTC().Truncate(time.Second)
	eventID := insertEvent(t, u.ID(), "Share target event", now, now.Add(time.Hour))

	repo := repository.NewEventShareRepository(testPool, testTracerProvider)

	id := uuid.New()
	expiresAt := time.Now().UTC().Add(time.Hour).Truncate(time.Second)
	const plainToken = "plain-token-findable"
	s, err := eventshare.New(id, eventID, eventshare.HashToken(plainToken), expiresAt)
	if err != nil {
		t.Fatalf("build domain eventshare: %v", err)
	}
	if _, err := repo.Create(context.Background(), s); err != nil {
		t.Fatalf("create eventshare: %v", err)
	}

	found, err := repo.FindByToken(context.Background(), plainToken)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found.ID() != id {
		t.Errorf("ID mismatch: got %v, want %v", found.ID(), id)
	}
	if found.EventID() != eventID {
		t.Errorf("EventID mismatch: got %v, want %v", found.EventID(), eventID)
	}
}

func TestEventShareRepository_FindByToken_UnknownToken_ReturnsErrEventShareNotFound(t *testing.T) {
	setupTest(t)

	repo := repository.NewEventShareRepository(testPool, testTracerProvider)

	_, err := repo.FindByToken(context.Background(), "token-that-was-never-issued")
	if !errors.Is(err, eventshare.ErrEventShareNotFound) {
		t.Errorf("expected ErrEventShareNotFound, got %v", err)
	}
}

func TestEventShareRepository_FindByToken_ExpiredShare_ReturnsShareWithIsExpiredTrue(t *testing.T) {
	setupTest(t)
	u := createTestUser(t)
	now := time.Now().UTC().Truncate(time.Second)
	eventID := insertEvent(t, u.ID(), "Share target event", now, now.Add(time.Hour))

	repo := repository.NewEventShareRepository(testPool, testTracerProvider)

	const plainToken = "plain-token-expired"
	past := time.Now().UTC().Add(-time.Hour).Truncate(time.Second)
	s := newTestEventShare(t, eventID, plainToken, past)
	if _, err := repo.Create(context.Background(), s); err != nil {
		t.Fatalf("create eventshare: %v", err)
	}

	found, err := repo.FindByToken(context.Background(), plainToken)
	if err != nil {
		t.Fatalf("expected an expired share to be returned without error, got: %v", err)
	}
	if !found.IsExpired() {
		t.Error("expected IsExpired() to be true for a share past its expiry")
	}
}

func TestEventShareRepository_FindByToken_MultipleSharesForSameEvent_ReturnsMatchingOne(t *testing.T) {
	setupTest(t)
	u := createTestUser(t)
	now := time.Now().UTC().Truncate(time.Second)
	eventID := insertEvent(t, u.ID(), "Share target event", now, now.Add(time.Hour))

	repo := repository.NewEventShareRepository(testPool, testTracerProvider)

	future := time.Now().UTC().Add(time.Hour).Truncate(time.Second)
	shareA := newTestEventShare(t, eventID, "plain-token-a", future)
	shareB := newTestEventShare(t, eventID, "plain-token-b", future)
	if _, err := repo.Create(context.Background(), shareA); err != nil {
		t.Fatalf("create shareA: %v", err)
	}
	if _, err := repo.Create(context.Background(), shareB); err != nil {
		t.Fatalf("create shareB: %v", err)
	}

	found, err := repo.FindByToken(context.Background(), "plain-token-b")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found.ID() != shareB.ID() {
		t.Errorf("expected to find shareB (%v), got %v", shareB.ID(), found.ID())
	}
}

func TestEventSharesTable_TokenHash_HasUniqueConstraint(t *testing.T) {
	setupTest(t)
	u := createTestUser(t)
	now := time.Now().UTC().Truncate(time.Second)
	eventID := insertEvent(t, u.ID(), "Share target event", now, now.Add(time.Hour))

	expiresAt := time.Now().UTC().Add(time.Hour)
	duplicateHash := eventshare.HashToken("duplicate-source-token")

	_, err := testPool.Exec(context.Background(),
		`INSERT INTO event_shares (event_id, token_hash, expires_at) VALUES ($1, $2, $3)`,
		eventID, duplicateHash, expiresAt,
	)
	if err != nil {
		t.Fatalf("insert first event_share: %v", err)
	}

	_, err = testPool.Exec(context.Background(),
		`INSERT INTO event_shares (event_id, token_hash, expires_at) VALUES ($1, $2, $3)`,
		eventID, duplicateHash, expiresAt,
	)
	if err == nil {
		t.Error("expected a unique constraint violation for a duplicate token_hash, got nil error")
	}
}

func TestEventSharesTable_EventIDForeignKey_CascadesOnEventDelete(t *testing.T) {
	setupTest(t)
	u := createTestUser(t)
	now := time.Now().UTC().Truncate(time.Second)
	eventID := insertEvent(t, u.ID(), "Share target event", now, now.Add(time.Hour))

	var shareID uuid.UUID
	err := testPool.QueryRow(context.Background(),
		`INSERT INTO event_shares (event_id, token_hash, expires_at) VALUES ($1, $2, $3) RETURNING id`,
		eventID, eventshare.HashToken("plain-token-cascade"), time.Now().UTC().Add(time.Hour),
	).Scan(&shareID)
	if err != nil {
		t.Fatalf("insert event_share: %v", err)
	}

	if _, err := testPool.Exec(context.Background(), "DELETE FROM events WHERE id = $1", eventID); err != nil {
		t.Fatalf("delete event: %v", err)
	}

	var count int
	if err := testPool.QueryRow(context.Background(),
		"SELECT COUNT(*) FROM event_shares WHERE id = $1", shareID,
	).Scan(&count); err != nil {
		t.Fatalf("count event_shares: %v", err)
	}
	if count != 0 {
		t.Errorf("expected event_share to be cascade-deleted with its event, but %d row(s) remain", count)
	}
}
