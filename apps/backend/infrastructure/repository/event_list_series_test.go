//go:build integration

package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/Haya372/ai-trial/backend/domain/event"
	rs "github.com/Haya372/ai-trial/backend/domain/recurringeventseries"
	"github.com/Haya372/ai-trial/backend/infrastructure/repository"
	eventuc "github.com/Haya372/ai-trial/backend/usecase/event"
)

// seedSeriesOccurrences saves the series and one occurrence per given day
// offset from base, each carrying the given tags.
func seedSeriesOccurrences(
	t *testing.T, s rs.RecurringEventSeries, base time.Time, days []int, tags ...event.EventTag,
) []event.Event {
	t.Helper()
	createSeries(t, s)
	occs := make([]event.Event, len(days))
	for i, d := range days {
		occs[i] = occurrence(t, s.UserID(), s.ID(), base, d, tags...)
	}
	if err := newSeriesRepo().CreateOccurrences(context.Background(), occs); err != nil {
		t.Fatalf("create occurrences: %v", err)
	}
	return occs
}

func listEvents(t *testing.T, filter eventuc.ListFilter) []eventuc.EventReadModel {
	t.Helper()
	repo := repository.NewEventQueryRepository(testPool, testLogger, testTracerProvider)
	events, err := repo.List(context.Background(), filter)
	if err != nil {
		t.Fatalf("list events: %v", err)
	}
	return events
}

const seriesTZ = "Asia/Tokyo"

func assertSeriesReadModel(t *testing.T, got *eventuc.RecurringSeriesReadModel, want eventuc.RecurringSeriesReadModel) {
	t.Helper()
	if got == nil {
		t.Fatalf("expected RecurringSeries %+v, got nil", want)
	}
	if got.ID != want.ID || got.PatternType != want.PatternType ||
		got.EndType != want.EndType || got.Timezone != want.Timezone {
		t.Errorf("series mismatch: got %+v, want %+v", *got, want)
	}
	assertOptional(t, "MonthlyNth", got.MonthlyNth, want.MonthlyNth)
	assertOptional(t, "UntilDate", got.UntilDate, want.UntilDate)
	assertOptional(t, "Count", got.Count, want.Count)
}

func assertOptional[T comparable](t *testing.T, name string, got, want *T) {
	t.Helper()
	switch {
	case got == nil && want == nil:
	case got == nil || want == nil:
		t.Errorf("%s mismatch: got %v, want %v", name, got, want)
	case *got != *want:
		t.Errorf("%s mismatch: got %v, want %v", name, *got, *want)
	}
}

func TestEventQueryRepository_List_ReturnsRecurringSeriesOfOccurrences(t *testing.T) {
	setupTest(t)
	u := createTestUser(t)
	base := time.Date(2027, 1, 4, 1, 0, 0, 0, time.UTC)

	countSeries := buildSeries(t, uuid.New(), u.ID(), "weekly", nil, countEnd(t, 2))
	seedSeriesOccurrences(t, countSeries, base, []int{0, 7})

	nthSeries := buildSeries(t, uuid.New(), u.ID(), "monthly_nth_weekday", new("1"),
		rs.NewUntilDateEnd(localDate(t, 2027, 3, 31)))
	seedSeriesOccurrences(t, nthSeries, base, []int{1})

	events := listEvents(t, eventuc.ListFilter{
		UserID: u.ID(), StartDate: base.Add(-time.Hour), EndDate: base.AddDate(0, 0, 30),
	})
	if len(events) != 3 {
		t.Fatalf("expected 3 events, got %d", len(events))
	}

	two := 2
	wantCount := eventuc.RecurringSeriesReadModel{
		ID: countSeries.ID(), PatternType: "weekly", EndType: "count", Count: &two, Timezone: seriesTZ,
	}
	wantNth := eventuc.RecurringSeriesReadModel{
		ID: nthSeries.ID(), PatternType: "monthly_nth_weekday", MonthlyNth: new("1"),
		EndType: "until_date", UntilDate: new("2027-03-31"), Timezone: seriesTZ,
	}
	// Ordered by start_at: day 0 (count), day 1 (nth), day 7 (count).
	assertSeriesReadModel(t, events[0].RecurringSeries, wantCount)
	assertSeriesReadModel(t, events[1].RecurringSeries, wantNth)
	assertSeriesReadModel(t, events[2].RecurringSeries, wantCount)
}

func TestEventQueryRepository_List_OccurrenceWithTags_IsAggregatedWithSeries(t *testing.T) {
	setupTest(t)
	u := createTestUser(t)
	tag1 := createTestTag(t, u.ID(), "work")
	tag2 := createTestTag(t, u.ID(), "team")
	base := time.Date(2027, 1, 4, 1, 0, 0, 0, time.UTC)

	s := buildSeries(t, uuid.New(), u.ID(), "daily", nil, rs.NewIndefiniteEnd())
	seedSeriesOccurrences(t, s, base, []int{0},
		mustEventTag(t, tag1.ID(), 1), mustEventTag(t, tag2.ID(), 2))

	events := listEvents(t, eventuc.ListFilter{
		UserID: u.ID(), StartDate: base.Add(-time.Hour), EndDate: base.Add(time.Hour),
	})
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if len(events[0].Tags) != 2 {
		t.Errorf("expected 2 tags, got %d", len(events[0].Tags))
	}
	assertSeriesReadModel(t, events[0].RecurringSeries, eventuc.RecurringSeriesReadModel{
		ID: s.ID(), PatternType: "daily", EndType: "indefinite", Timezone: seriesTZ,
	})
}

func TestEventQueryRepository_List_FilteredByTag_ReturnsRecurringSeries(t *testing.T) {
	setupTest(t)
	u := createTestUser(t)
	tg := createTestTag(t, u.ID(), "work")
	base := time.Date(2027, 1, 4, 1, 0, 0, 0, time.UTC)

	s := buildSeries(t, uuid.New(), u.ID(), "daily", nil, rs.NewIndefiniteEnd())
	seedSeriesOccurrences(t, s, base, []int{0}, mustEventTag(t, tg.ID(), 1))

	tagID := tg.ID()
	events := listEvents(t, eventuc.ListFilter{
		UserID: u.ID(), StartDate: base.Add(-time.Hour), EndDate: base.Add(time.Hour), TagID: &tagID,
	})
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	assertSeriesReadModel(t, events[0].RecurringSeries, eventuc.RecurringSeriesReadModel{
		ID: s.ID(), PatternType: "daily", EndType: "indefinite", Timezone: seriesTZ,
	})
}

func TestEventQueryRepository_List_OneOffEvent_HasNoRecurringSeries(t *testing.T) {
	setupTest(t)
	u := createTestUser(t)
	now := time.Now().UTC().Truncate(time.Second)
	insertEvent(t, u.ID(), "One-off", now, now.Add(time.Hour))

	events := listEvents(t, eventuc.ListFilter{
		UserID: u.ID(), StartDate: now.Add(-time.Hour), EndDate: now.Add(2 * time.Hour),
	})
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].RecurringSeries != nil {
		t.Errorf("expected nil RecurringSeries, got %+v", *events[0].RecurringSeries)
	}
}

func TestEventQueryRepository_List_SubscribedOccurrence_HidesOwnersSeries(t *testing.T) {
	setupTest(t)
	owner := createTestUser(t)
	subscriber := createTestUser(t)
	base := time.Date(2027, 1, 4, 1, 0, 0, 0, time.UTC)

	s := buildSeries(t, uuid.New(), owner.ID(), "daily", nil, rs.NewIndefiniteEnd())
	occs := seedSeriesOccurrences(t, s, base, []int{0})

	subsRepo := repository.NewEventSubscriptionRepository(testPool, testTracerProvider)
	if _, err := subsRepo.Create(context.Background(),
		newTestEventSubscription(t, occs[0].ID(), subscriber.ID())); err != nil {
		t.Fatalf("create event subscription: %v", err)
	}

	events := listEvents(t, eventuc.ListFilter{
		UserID: subscriber.ID(), StartDate: base.Add(-time.Hour), EndDate: base.Add(time.Hour),
	})
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if !events[0].IsSubscribed {
		t.Fatalf("expected subscribed event")
	}
	if events[0].RecurringSeries != nil {
		t.Errorf("expected nil RecurringSeries for subscribed event, got %+v", *events[0].RecurringSeries)
	}
}
