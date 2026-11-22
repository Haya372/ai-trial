//go:build integration

package repository_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/Haya372/ai-trial/backend/domain/event"
	"github.com/Haya372/ai-trial/backend/domain/user"
	"github.com/Haya372/ai-trial/backend/infrastructure/repository"
	eventuc "github.com/Haya372/ai-trial/backend/usecase/event"
)

func createTestUser(t *testing.T) user.User {
	t.Helper()
	repo := repository.NewUserRepository(testPool, testTracerProvider)
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

	eventRepo := repository.NewEventQueryRepository(testPool, testLogger, testTracerProvider)

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
	if events[0].Location != "" {
		t.Errorf("expected empty location, got %q", events[0].Location)
	}
	if events[0].URL != "" {
		t.Errorf("expected empty url, got %q", events[0].URL)
	}
}

func TestEventQueryRepository_List_ExcludesOutOfRangeEvents(t *testing.T) {
	setupTest(t)
	u := createTestUser(t)

	eventRepo := repository.NewEventQueryRepository(testPool, testLogger, testTracerProvider)

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

	eventRepo := repository.NewEventQueryRepository(testPool, testLogger, testTracerProvider)

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

func TestEventQueryRepository_List_IncludesSubscribedEvents(t *testing.T) {
	setupTest(t)
	owner, subscriber, sharedEventID := setupSharedEvent(t)

	eventRepo := repository.NewEventQueryRepository(testPool, testLogger, testTracerProvider)
	subsRepo := repository.NewEventSubscriptionRepository(testPool, testTracerProvider)

	now := time.Now().UTC().Truncate(time.Second)
	ownEventID := insertEvent(t, subscriber.ID(), "My own event", now, now.Add(time.Hour))

	savedSub, err := subsRepo.Create(
		context.Background(), newTestEventSubscription(t, sharedEventID, subscriber.ID()),
	)
	if err != nil {
		t.Fatalf("create event subscription: %v", err)
	}

	filter := eventuc.ListFilter{
		UserID:    subscriber.ID(),
		StartDate: now.Add(-time.Hour),
		EndDate:   now.Add(2 * time.Hour),
	}
	events, err := eventRepo.List(context.Background(), filter)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("expected 2 events, got %d: %+v", len(events), events)
	}

	byID := map[uuid.UUID]eventuc.EventReadModel{}
	for _, e := range events {
		byID[e.ID] = e
	}
	own, ok := byID[ownEventID]
	if !ok {
		t.Fatalf("expected own event in result: %+v", events)
	}
	if own.IsSubscribed {
		t.Errorf("expected own event IsSubscribed=false, got true")
	}
	if own.SubscriptionID != nil {
		t.Errorf("expected own event SubscriptionID to be nil, got %v", own.SubscriptionID)
	}
	shared, ok := byID[sharedEventID]
	if !ok {
		t.Fatalf("expected subscribed event in result: %+v", events)
	}
	if shared.UserID != owner.ID() {
		t.Errorf("expected subscribed event's UserID to be the owner, got %v", shared.UserID)
	}
	if !shared.IsSubscribed {
		t.Errorf("expected subscribed event IsSubscribed=true, got false")
	}
	if shared.SubscriptionID == nil || *shared.SubscriptionID != savedSub.ID() {
		t.Errorf("expected subscribed event SubscriptionID to be %v, got %v", savedSub.ID(), shared.SubscriptionID)
	}
}

func TestEventQueryRepository_List_ExcludesSubscriptionAfterSourceEventDeleted(t *testing.T) {
	setupTest(t)
	_, subscriber, sharedEventID := setupSharedEvent(t)

	eventRepo := repository.NewEventQueryRepository(testPool, testLogger, testTracerProvider)
	eventWriteRepo := repository.NewEventRepository(testPool, testLogger, testTracerProvider)
	subsRepo := repository.NewEventSubscriptionRepository(testPool, testTracerProvider)

	if _, err := subsRepo.Create(
		context.Background(), newTestEventSubscription(t, sharedEventID, subscriber.ID()),
	); err != nil {
		t.Fatalf("create event subscription: %v", err)
	}

	if err := eventWriteRepo.Delete(context.Background(), sharedEventID); err != nil {
		t.Fatalf("delete source event: %v", err)
	}

	now := time.Now().UTC().Truncate(time.Second)
	filter := eventuc.ListFilter{
		UserID:    subscriber.ID(),
		StartDate: now.Add(-time.Hour),
		EndDate:   now.Add(2 * time.Hour),
	}
	events, err := eventRepo.List(context.Background(), filter)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(events) != 0 {
		t.Errorf("expected deleted shared event to be excluded, got %d: %+v", len(events), events)
	}
}

func insertEvent(t *testing.T, userID uuid.UUID, title string, start, end time.Time) uuid.UUID {
	t.Helper()
	var id uuid.UUID
	err := testPool.QueryRow(context.Background(),
		"INSERT INTO events (user_id, title, start_at, end_at) VALUES ($1, $2, $3, $4) RETURNING id",
		userID, title, start, end,
	).Scan(&id)
	if err != nil {
		t.Fatalf("insert event: %v", err)
	}
	return id
}

func TestEventQueryRepository_List_ReturnsLocationAndUrl(t *testing.T) {
	setupTest(t)
	u := createTestUser(t)

	eventRepo := repository.NewEventQueryRepository(testPool, testLogger, testTracerProvider)

	now := time.Now().UTC().Truncate(time.Second)
	start := now
	end := now.Add(time.Hour)

	_, err := testPool.Exec(context.Background(),
		`INSERT INTO events (user_id, title, start_at, end_at, location, url)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		u.ID(), "Event with location and url", start, end, "Tokyo Office", "https://example.com/meeting",
	)
	if err != nil {
		t.Fatalf("insert event with location/url: %v", err)
	}

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
	if events[0].Location != "Tokyo Office" {
		t.Errorf("location mismatch: got %q, want %q", events[0].Location, "Tokyo Office")
	}
	if events[0].URL != "https://example.com/meeting" {
		t.Errorf("url mismatch: got %q, want %q", events[0].URL, "https://example.com/meeting")
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

func TestEventRepository_FindByID_ReturnsEvent(t *testing.T) {
	setupTest(t)
	u := createTestUser(t)

	eventRepo := repository.NewEventRepository(testPool, testLogger, testTracerProvider)

	now := time.Now().UTC().Truncate(time.Second)
	id := insertEvent(t, u.ID(), "Findable event", now, now.Add(time.Hour))

	e, err := eventRepo.FindByID(context.Background(), id)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if e.Title() != "Findable event" {
		t.Errorf("title mismatch: got %q", e.Title())
	}
	if e.UserID() != u.ID() {
		t.Errorf("userID mismatch: got %v, want %v", e.UserID(), u.ID())
	}
	if len(e.Tags()) != 0 {
		t.Errorf("expected no tags, got %d", len(e.Tags()))
	}
}

func TestEventRepository_FindByID_NotFound_ReturnsErrEventNotFound(t *testing.T) {
	setupTest(t)
	eventRepo := repository.NewEventRepository(testPool, testLogger, testTracerProvider)

	_, err := eventRepo.FindByID(context.Background(), uuid.New())
	if !errors.Is(err, event.ErrEventNotFound) {
		t.Errorf("expected ErrEventNotFound, got %v", err)
	}
}

func TestEventQueryRepository_FindByID_ReturnsReadModel(t *testing.T) {
	setupTest(t)
	u := createTestUser(t)

	eventRepo := repository.NewEventQueryRepository(testPool, testLogger, testTracerProvider)

	now := time.Now().UTC().Truncate(time.Second)
	id := insertEvent(t, u.ID(), "Findable event", now, now.Add(time.Hour))

	e, err := eventRepo.FindByID(context.Background(), id)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if e.Title != "Findable event" {
		t.Errorf("title mismatch: got %q", e.Title)
	}
	if e.UserID != u.ID() {
		t.Errorf("userID mismatch: got %v, want %v", e.UserID, u.ID())
	}
}

func TestEventQueryRepository_FindByID_NotFound_ReturnsErrEventNotFound(t *testing.T) {
	setupTest(t)
	eventRepo := repository.NewEventQueryRepository(testPool, testLogger, testTracerProvider)

	_, err := eventRepo.FindByID(context.Background(), uuid.New())
	if !errors.Is(err, event.ErrEventNotFound) {
		t.Errorf("expected ErrEventNotFound, got %v", err)
	}
}

func TestEventRepository_Update_PersistsChanges(t *testing.T) {
	setupTest(t)
	u := createTestUser(t)

	eventRepo := repository.NewEventRepository(testPool, testLogger, testTracerProvider)

	now := time.Now().UTC().Truncate(time.Second)
	id := insertEvent(t, u.ID(), "Original title", now, now.Add(time.Hour))

	newStart := now.Add(2 * time.Hour)
	newEnd := now.Add(3 * time.Hour)
	updated, err := event.New(
		id, u.ID(), "Updated title", "Updated desc", newStart, newEnd, "Tokyo", "https://example.com",
	)
	if err != nil {
		t.Fatalf("build updated event: %v", err)
	}

	if err := eventRepo.Update(context.Background(), updated); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := eventRepo.FindByID(context.Background(), id)
	if err != nil {
		t.Fatalf("find updated event: %v", err)
	}
	if got.Title() != "Updated title" {
		t.Errorf("title mismatch: got %q", got.Title())
	}
	if got.Description() != "Updated desc" {
		t.Errorf("description mismatch: got %q", got.Description())
	}
	if got.Location() != "Tokyo" {
		t.Errorf("location mismatch: got %q", got.Location())
	}
	if got.URL() != "https://example.com" {
		t.Errorf("url mismatch: got %q", got.URL())
	}
	if !got.StartAt().Equal(newStart) {
		t.Errorf("startAt mismatch: got %v, want %v", got.StartAt(), newStart)
	}
}

func TestEventRepository_Delete_RemovesEvent(t *testing.T) {
	setupTest(t)
	u := createTestUser(t)

	eventRepo := repository.NewEventRepository(testPool, testLogger, testTracerProvider)

	now := time.Now().UTC().Truncate(time.Second)
	id := insertEvent(t, u.ID(), "To be deleted", now, now.Add(time.Hour))

	if err := eventRepo.Delete(context.Background(), id); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err := eventRepo.FindByID(context.Background(), id)
	if !errors.Is(err, event.ErrEventNotFound) {
		t.Errorf("expected ErrEventNotFound after delete, got %v", err)
	}
}

func TestEventRepository_Delete_NonExistentID_ReturnsNoError(t *testing.T) {
	setupTest(t)
	eventRepo := repository.NewEventRepository(testPool, testLogger, testTracerProvider)

	if err := eventRepo.Delete(context.Background(), uuid.New()); err != nil {
		t.Errorf("unexpected error: %v", err)
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

func TestEventRepository_Create_PersistsEventAndReturnsIt(t *testing.T) {
	setupTest(t)
	u := createTestUser(t)

	repo := repository.NewEventRepository(testPool, testLogger, testTracerProvider)

	now := time.Now().UTC().Truncate(time.Second)
	startAt := now
	endAt := now.Add(time.Hour)
	id := uuid.New()

	e, err := event.New(id, u.ID(), "Integration test event", "desc", startAt, endAt, "Tokyo", "https://example.com")
	if err != nil {
		t.Fatalf("create domain event: %v", err)
	}

	saved, err := repo.Create(context.Background(), e)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if saved.ID() != id {
		t.Errorf("ID mismatch: got %v, want %v", saved.ID(), id)
	}
	if saved.UserID() != u.ID() {
		t.Errorf("UserID mismatch: got %v, want %v", saved.UserID(), u.ID())
	}
	if saved.Title() != "Integration test event" {
		t.Errorf("Title mismatch: got %q", saved.Title())
	}
	if saved.Description() != "desc" {
		t.Errorf("Description mismatch: got %q", saved.Description())
	}
	if saved.Location() != "Tokyo" {
		t.Errorf("Location mismatch: got %q", saved.Location())
	}
	if saved.URL() != "https://example.com" {
		t.Errorf("URL mismatch: got %q", saved.URL())
	}
	if !saved.StartAt().Equal(startAt) {
		t.Errorf("StartAt mismatch: got %v, want %v", saved.StartAt(), startAt)
	}
	if !saved.EndAt().Equal(endAt) {
		t.Errorf("EndAt mismatch: got %v, want %v", saved.EndAt(), endAt)
	}
}

func TestEventRepository_Create_NullableFieldsStoredAsNull(t *testing.T) {
	setupTest(t)
	u := createTestUser(t)

	repo := repository.NewEventRepository(testPool, testLogger, testTracerProvider)

	now := time.Now().UTC().Truncate(time.Second)
	id := uuid.New()

	e, err := event.New(id, u.ID(), "No optional fields", "", now, now.Add(time.Hour), "", "")
	if err != nil {
		t.Fatalf("create domain event: %v", err)
	}

	saved, err := repo.Create(context.Background(), e)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if saved.Description() != "" {
		t.Errorf("expected empty description, got %q", saved.Description())
	}
	if saved.Location() != "" {
		t.Errorf("expected empty location, got %q", saved.Location())
	}
	if saved.URL() != "" {
		t.Errorf("expected empty url, got %q", saved.URL())
	}
}

func newEventWithTags(t *testing.T, id, userID uuid.UUID, tags []event.EventTag) event.Event {
	t.Helper()
	now := time.Now().UTC().Truncate(time.Second)
	e, err := event.New(id, userID, "Tagged", "", now, now.Add(time.Hour), "", "")
	if err != nil {
		t.Fatalf("build event: %v", err)
	}
	e, err = e.WithTags(tags)
	if err != nil {
		t.Fatalf("attach tags: %v", err)
	}
	return e
}

func mustEventTag(t *testing.T, tagID uuid.UUID, position int) event.EventTag {
	t.Helper()
	et, err := event.NewEventTag(tagID, position)
	if err != nil {
		t.Fatalf("build event tag: %v", err)
	}
	return et
}

func assertEventTags(t *testing.T, got []event.EventTag, want []event.EventTag) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("expected %d tags, got %d", len(want), len(got))
	}
	for i := range want {
		if got[i].TagID() != want[i].TagID() || got[i].Position() != want[i].Position() {
			t.Errorf("tag[%d] mismatch: got %v@%d, want %v@%d",
				i, got[i].TagID(), got[i].Position(), want[i].TagID(), want[i].Position())
		}
	}
}

func TestEventRepository_Create_PersistsTags(t *testing.T) {
	setupTest(t)
	u := createTestUser(t)
	work := createTestTag(t, u.ID(), "Work")
	private := createTestTag(t, u.ID(), "Private")
	eventRepo := repository.NewEventRepository(testPool, testLogger, testTracerProvider)

	want := []event.EventTag{mustEventTag(t, private.ID(), 1), mustEventTag(t, work.ID(), 2)}
	saved, err := eventRepo.Create(context.Background(), newEventWithTags(t, uuid.New(), u.ID(), want))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertEventTags(t, saved.Tags(), want)

	got, err := eventRepo.FindByID(context.Background(), saved.ID())
	if err != nil {
		t.Fatalf("find created event: %v", err)
	}
	assertEventTags(t, got.Tags(), want)
}

func TestEventRepository_FindByID_RestoresTagsInPositionOrderWithGaps(t *testing.T) {
	setupTest(t)
	u := createTestUser(t)
	work := createTestTag(t, u.ID(), "Work")
	private := createTestTag(t, u.ID(), "Private")
	now := time.Now().UTC()
	eventID := insertEvent(t, u.ID(), "Meeting", now, now.Add(time.Hour))
	insertEventTag(t, eventID, work.ID(), 4)
	insertEventTag(t, eventID, private.ID(), 2)
	eventRepo := repository.NewEventRepository(testPool, testLogger, testTracerProvider)

	got, err := eventRepo.FindByID(context.Background(), eventID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertEventTags(t, got.Tags(), []event.EventTag{mustEventTag(t, private.ID(), 2), mustEventTag(t, work.ID(), 4)})
}

func TestEventRepository_Update_ReplacesTags(t *testing.T) {
	setupTest(t)
	u := createTestUser(t)
	work := createTestTag(t, u.ID(), "Work")
	private := createTestTag(t, u.ID(), "Private")
	travel := createTestTag(t, u.ID(), "Travel")
	eventRepo := repository.NewEventRepository(testPool, testLogger, testTracerProvider)

	id := uuid.New()
	if _, err := eventRepo.Create(context.Background(), newEventWithTags(t, id, u.ID(),
		[]event.EventTag{mustEventTag(t, work.ID(), 1), mustEventTag(t, private.ID(), 2)},
	)); err != nil {
		t.Fatalf("create event: %v", err)
	}

	// Moving work to another position checks that the old rows are gone
	// before the new ones go in, or UNIQUE (event_id, position) would fail.
	want := []event.EventTag{mustEventTag(t, travel.ID(), 1), mustEventTag(t, work.ID(), 2)}
	if err := eventRepo.Update(context.Background(), newEventWithTags(t, id, u.ID(), want)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := eventRepo.FindByID(context.Background(), id)
	if err != nil {
		t.Fatalf("find updated event: %v", err)
	}
	assertEventTags(t, got.Tags(), want)
}

func TestEventRepository_Update_WithNoTags_RemovesAllTags(t *testing.T) {
	setupTest(t)
	u := createTestUser(t)
	work := createTestTag(t, u.ID(), "Work")
	eventRepo := repository.NewEventRepository(testPool, testLogger, testTracerProvider)

	id := uuid.New()
	if _, err := eventRepo.Create(context.Background(), newEventWithTags(t, id, u.ID(),
		[]event.EventTag{mustEventTag(t, work.ID(), 1)},
	)); err != nil {
		t.Fatalf("create event: %v", err)
	}

	if err := eventRepo.Update(context.Background(), newEventWithTags(t, id, u.ID(), nil)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := eventRepo.FindByID(context.Background(), id)
	if err != nil {
		t.Fatalf("find updated event: %v", err)
	}
	if len(got.Tags()) != 0 {
		t.Errorf("expected no tags, got %d", len(got.Tags()))
	}
}

func TestEventRepository_Update_InCallerTx_RollsBackWithCaller(t *testing.T) {
	setupTest(t)
	u := createTestUser(t)
	work := createTestTag(t, u.ID(), "Work")
	eventRepo := repository.NewEventRepository(testPool, testLogger, testTracerProvider)

	id := uuid.New()
	stored := []event.EventTag{mustEventTag(t, work.ID(), 1)}
	if _, err := eventRepo.Create(context.Background(), newEventWithTags(t, id, u.ID(), stored)); err != nil {
		t.Fatalf("create event: %v", err)
	}

	runInTestTx(t, func(ctx context.Context) {
		if err := eventRepo.Update(ctx, newEventWithTags(t, id, u.ID(), nil)); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	got, err := eventRepo.FindByID(context.Background(), id)
	if err != nil {
		t.Fatalf("find event: %v", err)
	}
	assertEventTags(t, got.Tags(), stored)
}

func TestEventRepository_FindByIDForUpdate_ReturnsEventAndLocksRowUntilTxEnds(t *testing.T) {
	setupTest(t)
	u := createTestUser(t)
	work := createTestTag(t, u.ID(), "Work")
	eventRepo := repository.NewEventRepository(testPool, testLogger, testTracerProvider)

	id := uuid.New()
	stored := []event.EventTag{mustEventTag(t, work.ID(), 1)}
	if _, err := eventRepo.Create(context.Background(), newEventWithTags(t, id, u.ID(), stored)); err != nil {
		t.Fatalf("create event: %v", err)
	}

	runInTestTx(t, func(ctx context.Context) {
		got, err := eventRepo.FindByIDForUpdate(ctx, id)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.ID() != id || got.Title() != "Tagged" {
			t.Errorf("event mismatch: got %v %q", got.ID(), got.Title())
		}
		assertEventTags(t, got.Tags(), stored)

		_, err = testPool.Exec(context.Background(), "SELECT id FROM events WHERE id = $1 FOR UPDATE NOWAIT", id)
		if !isLockNotAvailable(err) {
			t.Errorf("expected the event row to be locked, got %v", err)
		}
	})
}

func TestEventRepository_FindByIDForUpdate_NotFound_ReturnsErrEventNotFound(t *testing.T) {
	setupTest(t)
	eventRepo := repository.NewEventRepository(testPool, testLogger, testTracerProvider)

	runInTestTx(t, func(ctx context.Context) {
		if _, err := eventRepo.FindByIDForUpdate(ctx, uuid.New()); !errors.Is(err, event.ErrEventNotFound) {
			t.Errorf("expected ErrEventNotFound, got %v", err)
		}
	})
}

func TestEventRepository_FindByIDForUpdate_OutsideTx_ReturnsError(t *testing.T) {
	setupTest(t)
	u := createTestUser(t)
	now := time.Now().UTC()
	id := insertEvent(t, u.ID(), "Meeting", now, now.Add(time.Hour))
	eventRepo := repository.NewEventRepository(testPool, testLogger, testTracerProvider)

	if _, err := eventRepo.FindByIDForUpdate(context.Background(), id); err == nil {
		t.Error("expected an error outside a transaction")
	}
}

// insertSeries stores a minimal daily series so events can reference it.
func insertSeries(t *testing.T, userID uuid.UUID) uuid.UUID {
	t.Helper()
	var id uuid.UUID
	err := testPool.QueryRow(context.Background(),
		`INSERT INTO recurring_event_series (user_id, pattern_type, end_type, timezone)
		 VALUES ($1, 'daily', 'indefinite', 'Asia/Tokyo') RETURNING id`,
		userID,
	).Scan(&id)
	if err != nil {
		t.Fatalf("insert series: %v", err)
	}
	return id
}

func assertSeriesID(t *testing.T, e event.Event, want uuid.UUID) {
	t.Helper()
	got, ok := e.SeriesID()
	if want == uuid.Nil {
		if ok {
			t.Errorf("expected no series, got %v", got)
		}
		return
	}
	if !ok || got != want {
		t.Errorf("series mismatch: got %v (ok=%v), want %v", got, ok, want)
	}
}

func TestEventRepository_Create_PersistsSeriesID(t *testing.T) {
	setupTest(t)
	u := createTestUser(t)
	seriesID := insertSeries(t, u.ID())
	eventRepo := repository.NewEventRepository(testPool, testLogger, testTracerProvider)

	saved, err := eventRepo.Create(context.Background(),
		newEventWithTags(t, uuid.New(), u.ID(), nil).WithSeriesID(seriesID))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertSeriesID(t, saved, seriesID)

	got, err := eventRepo.FindByID(context.Background(), saved.ID())
	if err != nil {
		t.Fatalf("find event: %v", err)
	}
	assertSeriesID(t, got, seriesID)

	runInTestTx(t, func(ctx context.Context) {
		locked, err := eventRepo.FindByIDForUpdate(ctx, saved.ID())
		if err != nil {
			t.Fatalf("find event for update: %v", err)
		}
		assertSeriesID(t, locked, seriesID)
	})
}

func TestEventRepository_Create_WithoutSeries_StoresNullSeriesID(t *testing.T) {
	setupTest(t)
	u := createTestUser(t)
	eventRepo := repository.NewEventRepository(testPool, testLogger, testTracerProvider)

	saved, err := eventRepo.Create(context.Background(), newEventWithTags(t, uuid.New(), u.ID(), nil))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertSeriesID(t, saved, uuid.Nil)
	if n := countRows(t, "SELECT COUNT(*) FROM events WHERE id = $1 AND series_id IS NULL", saved.ID()); n != 1 {
		t.Errorf("expected series_id to be NULL, got %d matching rows", n)
	}
}

func TestEventRepository_Update_DetachesFromAndAttachesToSeries(t *testing.T) {
	setupTest(t)
	u := createTestUser(t)
	seriesID := insertSeries(t, u.ID())
	eventRepo := repository.NewEventRepository(testPool, testLogger, testTracerProvider)

	e := newEventWithTags(t, uuid.New(), u.ID(), nil).WithSeriesID(seriesID)
	if _, err := eventRepo.Create(context.Background(), e); err != nil {
		t.Fatalf("create event: %v", err)
	}

	if err := eventRepo.Update(context.Background(), e.DetachFromSeries()); err != nil {
		t.Fatalf("detach: %v", err)
	}
	got, err := eventRepo.FindByID(context.Background(), e.ID())
	if err != nil {
		t.Fatalf("find detached event: %v", err)
	}
	assertSeriesID(t, got, uuid.Nil)

	if err := eventRepo.Update(context.Background(), got.WithSeriesID(seriesID)); err != nil {
		t.Fatalf("attach: %v", err)
	}
	got, err = eventRepo.FindByID(context.Background(), e.ID())
	if err != nil {
		t.Fatalf("find attached event: %v", err)
	}
	assertSeriesID(t, got, seriesID)
}

func assertForeignKeyViolation(t *testing.T, err error) {
	t.Helper()
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || pgErr.Code != "23503" {
		t.Errorf("expected foreign key violation, got %v", err)
	}
}

func TestEventRepository_Create_WithAnotherUsersSeries_FailsWithForeignKeyViolation(t *testing.T) {
	setupTest(t)
	owner := createTestUser(t)
	other := createTestUser(t)
	seriesID := insertSeries(t, owner.ID())
	eventRepo := repository.NewEventRepository(testPool, testLogger, testTracerProvider)

	_, err := eventRepo.Create(context.Background(),
		newEventWithTags(t, uuid.New(), other.ID(), nil).WithSeriesID(seriesID))

	assertForeignKeyViolation(t, err)
}

func TestEventRepository_Update_AttachingToAnotherUsersSeries_FailsWithForeignKeyViolation(t *testing.T) {
	setupTest(t)
	owner := createTestUser(t)
	other := createTestUser(t)
	seriesID := insertSeries(t, owner.ID())
	eventRepo := repository.NewEventRepository(testPool, testLogger, testTracerProvider)
	e := newEventWithTags(t, uuid.New(), other.ID(), nil)
	if _, err := eventRepo.Create(context.Background(), e); err != nil {
		t.Fatalf("create event: %v", err)
	}

	err := eventRepo.Update(context.Background(), e.WithSeriesID(seriesID))

	assertForeignKeyViolation(t, err)
	if n := countRows(t, "SELECT COUNT(*) FROM events WHERE id = $1 AND series_id IS NULL", e.ID()); n != 1 {
		t.Errorf("expected the event to stay out of the series, got %d matching rows", n)
	}
}

func TestEventsTable_SeriesIDIndex_CoversStartAt(t *testing.T) {
	setupTest(t)
	var def string
	if err := testPool.QueryRow(context.Background(),
		"SELECT indexdef FROM pg_indexes WHERE tablename = 'events' AND indexname = 'idx_events_series_id'",
	).Scan(&def); err != nil {
		t.Fatalf("read idx_events_series_id: %v", err)
	}
	if !strings.HasSuffix(def, "(series_id, start_at)") {
		t.Errorf("expected idx_events_series_id on (series_id, start_at), got %q", def)
	}
}
