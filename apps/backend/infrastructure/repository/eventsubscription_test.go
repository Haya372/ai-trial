//go:build integration

package repository_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/Haya372/ai-trial/backend/domain/eventsubscription"
	"github.com/Haya372/ai-trial/backend/domain/user"
	"github.com/Haya372/ai-trial/backend/infrastructure/repository"
)

func newTestEventSubscription(t *testing.T, eventID, userID uuid.UUID) eventsubscription.EventSubscription {
	t.Helper()
	return eventsubscription.New(uuid.New(), eventID, userID, time.Now().UTC())
}

// setupSharedEvent creates an owner, a subscriber, and an event owned by
// owner that subscriber can subscribe to.
func setupSharedEvent(t *testing.T) (owner, subscriber user.User, eventID uuid.UUID) {
	t.Helper()
	owner = createTestUser(t)
	subscriber = createTestUser(t)
	now := time.Now().UTC().Truncate(time.Second)
	eventID = insertEvent(t, owner.ID(), "Shared event", now, now.Add(time.Hour))
	return owner, subscriber, eventID
}

func TestEventSubscriptionRepository_Create_PersistsAndReturnsEventSubscription(t *testing.T) {
	setupTest(t)
	_, subscriber, eventID := setupSharedEvent(t)

	repo := repository.NewEventSubscriptionRepository(testPool, testTracerProvider)

	saved, err := repo.Create(context.Background(), newTestEventSubscription(t, eventID, subscriber.ID()))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if saved.EventID() != eventID {
		t.Errorf("EventID mismatch: got %v, want %v", saved.EventID(), eventID)
	}
	if saved.UserID() != subscriber.ID() {
		t.Errorf("UserID mismatch: got %v, want %v", saved.UserID(), subscriber.ID())
	}
}

func TestEventSubscriptionRepository_Create_NonExistentEventID_ReturnsError(t *testing.T) {
	setupTest(t)
	subscriber := createTestUser(t)

	repo := repository.NewEventSubscriptionRepository(testPool, testTracerProvider)

	s := newTestEventSubscription(t, uuid.New(), subscriber.ID())
	_, err := repo.Create(context.Background(), s)
	if err == nil {
		t.Fatal("expected an error for a non-existent event_id, got nil")
	}
}

func TestEventSubscriptionRepository_Create_DuplicateEventAndUserID_ReturnsError(t *testing.T) {
	setupTest(t)
	_, subscriber, eventID := setupSharedEvent(t)

	repo := repository.NewEventSubscriptionRepository(testPool, testTracerProvider)

	if _, err := repo.Create(context.Background(), newTestEventSubscription(t, eventID, subscriber.ID())); err != nil {
		t.Fatalf("create first subscription: %v", err)
	}

	_, err := repo.Create(context.Background(), newTestEventSubscription(t, eventID, subscriber.ID()))
	if !errors.Is(err, eventsubscription.ErrAlreadySubscribed) {
		t.Errorf("expected ErrAlreadySubscribed, got %v", err)
	}
}

func TestEventSubscriptionRepository_FindByID_ReturnsEventSubscription(t *testing.T) {
	setupTest(t)
	_, subscriber, eventID := setupSharedEvent(t)

	repo := repository.NewEventSubscriptionRepository(testPool, testTracerProvider)

	created, err := repo.Create(context.Background(), newTestEventSubscription(t, eventID, subscriber.ID()))
	if err != nil {
		t.Fatalf("create subscription: %v", err)
	}

	found, err := repo.FindByID(context.Background(), created.ID())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found.ID() != created.ID() {
		t.Errorf("ID mismatch: got %v, want %v", found.ID(), created.ID())
	}
	if found.EventID() != eventID {
		t.Errorf("EventID mismatch: got %v, want %v", found.EventID(), eventID)
	}
	if found.UserID() != subscriber.ID() {
		t.Errorf("UserID mismatch: got %v, want %v", found.UserID(), subscriber.ID())
	}
}

func TestEventSubscriptionRepository_FindByID_NotFound_ReturnsErrEventSubscriptionNotFound(t *testing.T) {
	setupTest(t)

	repo := repository.NewEventSubscriptionRepository(testPool, testTracerProvider)

	_, err := repo.FindByID(context.Background(), uuid.New())
	if !errors.Is(err, eventsubscription.ErrEventSubscriptionNotFound) {
		t.Errorf("expected ErrEventSubscriptionNotFound, got %v", err)
	}
}

func TestEventSubscriptionRepository_FindByEventAndUserID_ReturnsMatchingSubscription(t *testing.T) {
	setupTest(t)
	_, subscriber, eventID := setupSharedEvent(t)

	repo := repository.NewEventSubscriptionRepository(testPool, testTracerProvider)

	created, err := repo.Create(context.Background(), newTestEventSubscription(t, eventID, subscriber.ID()))
	if err != nil {
		t.Fatalf("create subscription: %v", err)
	}

	found, err := repo.FindByEventAndUserID(context.Background(), eventID, subscriber.ID())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found.ID() != created.ID() {
		t.Errorf("ID mismatch: got %v, want %v", found.ID(), created.ID())
	}
}

func TestEventSubscriptionRepository_FindByEventAndUserID_NotFound_ReturnsErrEventSubscriptionNotFound(t *testing.T) {
	setupTest(t)
	_, subscriber, eventID := setupSharedEvent(t)

	repo := repository.NewEventSubscriptionRepository(testPool, testTracerProvider)

	_, err := repo.FindByEventAndUserID(context.Background(), eventID, subscriber.ID())
	if !errors.Is(err, eventsubscription.ErrEventSubscriptionNotFound) {
		t.Errorf("expected ErrEventSubscriptionNotFound, got %v", err)
	}
}

func TestEventSubscriptionRepository_ListByUserID_ReturnsOnlyThatUsersSubscriptions(t *testing.T) {
	setupTest(t)
	owner := createTestUser(t)
	subscriber := createTestUser(t)
	other := createTestUser(t)
	now := time.Now().UTC().Truncate(time.Second)
	eventA := insertEvent(t, owner.ID(), "Event A", now, now.Add(time.Hour))
	eventB := insertEvent(t, owner.ID(), "Event B", now, now.Add(time.Hour))

	repo := repository.NewEventSubscriptionRepository(testPool, testTracerProvider)

	if _, err := repo.Create(context.Background(), newTestEventSubscription(t, eventA, subscriber.ID())); err != nil {
		t.Fatalf("create subscription A: %v", err)
	}
	if _, err := repo.Create(context.Background(), newTestEventSubscription(t, eventB, subscriber.ID())); err != nil {
		t.Fatalf("create subscription B: %v", err)
	}
	if _, err := repo.Create(context.Background(), newTestEventSubscription(t, eventA, other.ID())); err != nil {
		t.Fatalf("create subscription for other user: %v", err)
	}

	found, err := repo.ListByUserID(context.Background(), subscriber.ID())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(found) != 2 {
		t.Fatalf("expected 2 subscriptions, got %d", len(found))
	}
	for _, s := range found {
		if s.UserID() != subscriber.ID() {
			t.Errorf("expected only subscriber's subscriptions, got one for user %v", s.UserID())
		}
	}
}

func TestEventSubscriptionRepository_Delete_RemovesSubscription(t *testing.T) {
	setupTest(t)
	_, subscriber, eventID := setupSharedEvent(t)

	repo := repository.NewEventSubscriptionRepository(testPool, testTracerProvider)

	created, err := repo.Create(context.Background(), newTestEventSubscription(t, eventID, subscriber.ID()))
	if err != nil {
		t.Fatalf("create subscription: %v", err)
	}

	if err := repo.Delete(context.Background(), created.ID()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = repo.FindByID(context.Background(), created.ID())
	if !errors.Is(err, eventsubscription.ErrEventSubscriptionNotFound) {
		t.Errorf("expected ErrEventSubscriptionNotFound after delete, got %v", err)
	}
}

func TestEventSubscriptionRepository_Delete_NonExistentID_ReturnsNoError(t *testing.T) {
	setupTest(t)

	repo := repository.NewEventSubscriptionRepository(testPool, testTracerProvider)

	if err := repo.Delete(context.Background(), uuid.New()); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestEventSubscriptionsTable_EventIDForeignKey_CascadesOnEventDelete(t *testing.T) {
	setupTest(t)
	_, subscriber, eventID := setupSharedEvent(t)

	repo := repository.NewEventSubscriptionRepository(testPool, testTracerProvider)
	created, err := repo.Create(context.Background(), newTestEventSubscription(t, eventID, subscriber.ID()))
	if err != nil {
		t.Fatalf("create subscription: %v", err)
	}

	if _, err := testPool.Exec(context.Background(), "DELETE FROM events WHERE id = $1", eventID); err != nil {
		t.Fatalf("delete event: %v", err)
	}

	var count int
	if err := testPool.QueryRow(context.Background(),
		"SELECT COUNT(*) FROM event_subscriptions WHERE id = $1", created.ID(),
	).Scan(&count); err != nil {
		t.Fatalf("count event_subscriptions: %v", err)
	}
	if count != 0 {
		t.Errorf("expected event_subscription to be cascade-deleted with its event, but %d row(s) remain", count)
	}
}
