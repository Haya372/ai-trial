//go:build integration

package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/Haya372/ai-trial/backend/domain/user"
	"github.com/Haya372/ai-trial/backend/infrastructure/repository"
	eventuc "github.com/Haya372/ai-trial/backend/usecase/event"
)

func createTestUser(t *testing.T) user.User {
	t.Helper()
	repo := repository.NewUserRepository(testPool)
	email, _ := user.NewEmail("event-user-" + uuid.New().String() + "@example.com")
	password, _ := user.NewPassword("SecurePass1!")
	u, err := repo.Create(context.Background(), email, "Event User", password)
	if err != nil {
		t.Fatalf("create test user: %v", err)
	}
	return u
}

func TestEventQueryRepository_List_ReturnsEventsInRange(t *testing.T) {
	setupTest(t)
	u := createTestUser(t)

	eventRepo := repository.NewEventQueryRepository(testPool, testLogger)

	now := time.Now().UTC().Truncate(time.Second)
	start := now
	end := now.Add(time.Hour)

	insertEvent(t, u.ID(), "In-range event", start, end)

	filter := eventuc.ListFilter{
		UserID:    u.ID(),
		StartDate: now.Add(-time.Hour),
		EndDate:   now.Add(2 * time.Hour),
	}
	events, err := eventRepo.List(context.Background(), filter)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].Title != "In-range event" {
		t.Errorf("title mismatch: got %q", events[0].Title)
	}
}

func TestEventQueryRepository_List_ExcludesOutOfRangeEvents(t *testing.T) {
	setupTest(t)
	u := createTestUser(t)

	eventRepo := repository.NewEventQueryRepository(testPool, testLogger)

	now := time.Now().UTC().Truncate(time.Second)

	insertEvent(t, u.ID(), "Future event", now.Add(10*time.Hour), now.Add(11*time.Hour))

	filter := eventuc.ListFilter{
		UserID:    u.ID(),
		StartDate: now,
		EndDate:   now.Add(time.Hour),
	}
	events, err := eventRepo.List(context.Background(), filter)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(events) != 0 {
		t.Errorf("expected 0 events, got %d", len(events))
	}
}

func TestEventQueryRepository_List_OnlyReturnsUserEvents(t *testing.T) {
	setupTest(t)
	u1 := createTestUser(t)
	u2 := createTestUser(t)

	eventRepo := repository.NewEventQueryRepository(testPool, testLogger)

	now := time.Now().UTC().Truncate(time.Second)
	insertEvent(t, u1.ID(), "User1 event", now, now.Add(time.Hour))
	insertEvent(t, u2.ID(), "User2 event", now, now.Add(time.Hour))

	filter := eventuc.ListFilter{
		UserID:    u1.ID(),
		StartDate: now.Add(-time.Hour),
		EndDate:   now.Add(2 * time.Hour),
	}
	events, err := eventRepo.List(context.Background(), filter)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].Title != "User1 event" {
		t.Errorf("title mismatch: got %q", events[0].Title)
	}
}

func insertEvent(t *testing.T, userID uuid.UUID, title string, start, end time.Time) {
	t.Helper()
	_, err := testPool.Exec(context.Background(),
		"INSERT INTO events (user_id, title, start_at, end_at) VALUES ($1, $2, $3, $4)",
		userID, title, start, end,
	)
	if err != nil {
		t.Fatalf("insert event: %v", err)
	}
}

func TestEventsTable_LocationAndUrlColumns_AcceptValues(t *testing.T) {
	setupTest(t)
	u := createTestUser(t)

	now := time.Now().UTC().Truncate(time.Second)
	var gotLocation, gotURL *string
	err := testPool.QueryRow(context.Background(),
		`INSERT INTO events (user_id, title, start_at, end_at, location, url)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING location, url`,
		u.ID(), "Event with location and url", now, now.Add(time.Hour),
		"Tokyo Office", "https://example.com/meeting",
	).Scan(&gotLocation, &gotURL)
	if err != nil {
		t.Fatalf("insert event with location/url: %v", err)
	}
	if gotLocation == nil || *gotLocation != "Tokyo Office" {
		t.Errorf("location mismatch: got %v", gotLocation)
	}
	if gotURL == nil || *gotURL != "https://example.com/meeting" {
		t.Errorf("url mismatch: got %v", gotURL)
	}
}

func TestEventsTable_LocationAndUrlColumns_AreNullable(t *testing.T) {
	setupTest(t)
	u := createTestUser(t)

	now := time.Now().UTC().Truncate(time.Second)
	insertEvent(t, u.ID(), "Event without location and url", now, now.Add(time.Hour))

	var gotLocation, gotURL *string
	err := testPool.QueryRow(context.Background(),
		"SELECT location, url FROM events WHERE user_id = $1", u.ID(),
	).Scan(&gotLocation, &gotURL)
	if err != nil {
		t.Fatalf("select event: %v", err)
	}
	if gotLocation != nil {
		t.Errorf("expected location to be NULL, got %v", *gotLocation)
	}
	if gotURL != nil {
		t.Errorf("expected url to be NULL, got %v", *gotURL)
	}
}
