//go:build integration

package repository_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/Haya372/ai-trial/backend/domain/event"
	rs "github.com/Haya372/ai-trial/backend/domain/recurringeventseries"
	"github.com/Haya372/ai-trial/backend/infrastructure/repository"
)

func newSeriesRepo() rs.Repository {
	return repository.NewRecurringEventSeriesRepository(testPool, testLogger, testTracerProvider)
}

func localDate(t *testing.T, y int, m time.Month, d int) rs.LocalDate {
	t.Helper()
	ld, err := rs.NewLocalDate(y, m, d)
	if err != nil {
		t.Fatalf("build local date: %v", err)
	}
	return ld
}

func buildSeries(t *testing.T, id, userID uuid.UUID, pattern string, nth *string, end rs.RecurrenceEnd, excluded ...rs.LocalDate) rs.RecurringEventSeries {
	t.Helper()
	p, err := rs.NewRecurrencePattern(pattern, nth)
	if err != nil {
		t.Fatalf("build pattern: %v", err)
	}
	tz, err := rs.NewSeriesTimezone("Asia/Tokyo")
	if err != nil {
		t.Fatalf("build timezone: %v", err)
	}
	ex := make([]rs.ExcludedDate, 0, len(excluded))
	for _, d := range excluded {
		ex = append(ex, rs.NewExcludedDate(d))
	}
	s, err := rs.New(id, userID, rs.RecurrenceRule{Pattern: p, End: end, Timezone: tz}, ex)
	if err != nil {
		t.Fatalf("build series: %v", err)
	}
	return s
}

func countEnd(t *testing.T, n int) rs.RecurrenceEnd {
	t.Helper()
	e, err := rs.RestoreCountEnd(n)
	if err != nil {
		t.Fatalf("build count end: %v", err)
	}
	return e
}

func createSeries(t *testing.T, s rs.RecurringEventSeries) {
	t.Helper()
	if _, err := newSeriesRepo().Create(context.Background(), s); err != nil {
		t.Fatalf("create series: %v", err)
	}
}

func assertSameSeries(t *testing.T, got, want rs.RecurringEventSeries) {
	t.Helper()
	if got.ID() != want.ID() || got.UserID() != want.UserID() {
		t.Errorf("ids mismatch: got %v/%v, want %v/%v", got.ID(), got.UserID(), want.ID(), want.UserID())
	}
	if got.Rule().Pattern != want.Rule().Pattern || got.Rule().End != want.Rule().End {
		t.Errorf("rule mismatch: got %+v/%+v, want %+v/%+v",
			got.Rule().Pattern, got.Rule().End, want.Rule().Pattern, want.Rule().End)
	}
	if got.Rule().Timezone.Name() != want.Rule().Timezone.Name() {
		t.Errorf("timezone mismatch: got %q, want %q", got.Rule().Timezone.Name(), want.Rule().Timezone.Name())
	}
	g, w := got.ExcludedDates(), want.ExcludedDates()
	if len(g) != len(w) {
		t.Fatalf("expected %d excluded dates, got %d", len(w), len(g))
	}
	for i := range w {
		if g[i].Date() != w[i].Date() {
			t.Errorf("excluded[%d] mismatch: got %v, want %v", i, g[i].Date(), w[i].Date())
		}
	}
}

// occurrence builds an event of the series starting `day` days after base.
func occurrence(t *testing.T, userID, seriesID uuid.UUID, base time.Time, day int, tags ...event.EventTag) event.Event {
	t.Helper()
	start := base.AddDate(0, 0, day)
	e, err := event.New(uuid.New(), userID, "Standup", "", start, start.Add(30*time.Minute), "", "")
	if err != nil {
		t.Fatalf("build occurrence: %v", err)
	}
	e, err = e.WithTags(tags)
	if err != nil {
		t.Fatalf("attach tags: %v", err)
	}
	return e.WithSeriesID(seriesID)
}

func TestRecurringSeriesRepository_Create_RoundTripsThroughFindByID(t *testing.T) {
	setupTest(t)
	u := createTestUser(t)
	last := string(rs.MonthlyNthLast)
	until := rs.NewUntilDateEnd(localDate(t, 2027, time.March, 31))

	cases := map[string]rs.RecurringEventSeries{
		"daily indefinite":    buildSeries(t, uuid.New(), u.ID(), "daily", nil, rs.NewIndefiniteEnd()),
		"weekly until date":   buildSeries(t, uuid.New(), u.ID(), "weekly", nil, until),
		"biweekly count":      buildSeries(t, uuid.New(), u.ID(), "biweekly", nil, countEnd(t, 10)),
		"monthly day count 1": buildSeries(t, uuid.New(), u.ID(), "monthly_day", nil, countEnd(t, 1)),
		"monthly nth weekday": buildSeries(t, uuid.New(), u.ID(), "monthly_nth_weekday", &last, until),
		"with excluded dates": buildSeries(t, uuid.New(), u.ID(), "daily", nil, rs.NewIndefiniteEnd(),
			localDate(t, 2026, time.December, 3), localDate(t, 2026, time.December, 1)),
	}
	repo := newSeriesRepo()
	for name, want := range cases {
		t.Run(name, func(t *testing.T) {
			saved, err := repo.Create(context.Background(), want)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			assertSameSeries(t, saved, want)

			got, err := repo.FindByID(context.Background(), want.ID())
			if err != nil {
				t.Fatalf("find series: %v", err)
			}
			assertSameSeries(t, got, want)
		})
	}
}

func TestRecurringSeriesRepository_FindByID_NotFound_ReturnsErrRecurringSeriesNotFound(t *testing.T) {
	setupTest(t)

	if _, err := newSeriesRepo().FindByID(context.Background(), uuid.New()); !errors.Is(err, rs.ErrRecurringSeriesNotFound) {
		t.Errorf("expected ErrRecurringSeriesNotFound, got %v", err)
	}
}

func TestRecurringSeriesRepository_FindByIDForUpdate_ReturnsSeriesAndLocksRowUntilTxEnds(t *testing.T) {
	setupTest(t)
	u := createTestUser(t)
	want := buildSeries(t, uuid.New(), u.ID(), "weekly", nil, countEnd(t, 5), localDate(t, 2026, time.December, 7))
	createSeries(t, want)
	repo := newSeriesRepo()

	runInTestTx(t, func(ctx context.Context) {
		got, err := repo.FindByIDForUpdate(ctx, want.ID())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		assertSameSeries(t, got, want)

		_, err = testPool.Exec(context.Background(),
			"SELECT id FROM recurring_event_series WHERE id = $1 FOR UPDATE NOWAIT", want.ID())
		if !isLockNotAvailable(err) {
			t.Errorf("expected the series row to be locked, got %v", err)
		}
	})
}

func TestRecurringSeriesRepository_FindByIDForUpdate_NotFound_ReturnsErrRecurringSeriesNotFound(t *testing.T) {
	setupTest(t)
	repo := newSeriesRepo()

	runInTestTx(t, func(ctx context.Context) {
		if _, err := repo.FindByIDForUpdate(ctx, uuid.New()); !errors.Is(err, rs.ErrRecurringSeriesNotFound) {
			t.Errorf("expected ErrRecurringSeriesNotFound, got %v", err)
		}
	})
}

func TestRecurringSeriesRepository_FindByIDForUpdate_OutsideTx_ReturnsError(t *testing.T) {
	setupTest(t)
	u := createTestUser(t)
	s := buildSeries(t, uuid.New(), u.ID(), "daily", nil, rs.NewIndefiniteEnd())
	createSeries(t, s)

	if _, err := newSeriesRepo().FindByIDForUpdate(context.Background(), s.ID()); err == nil {
		t.Error("expected an error outside a transaction")
	}
}

func TestRecurringSeriesRepository_Update_ReplacesRuleAndExcludedDates(t *testing.T) {
	setupTest(t)
	u := createTestUser(t)
	id := uuid.New()
	createSeries(t, buildSeries(t, id, u.ID(), "daily", nil, countEnd(t, 10),
		localDate(t, 2026, time.December, 1), localDate(t, 2026, time.December, 2)))
	repo := newSeriesRepo()

	want := buildSeries(t, id, u.ID(), "weekly", nil, rs.NewUntilDateEnd(localDate(t, 2027, time.January, 31)),
		localDate(t, 2026, time.December, 2), localDate(t, 2026, time.December, 9))
	if err := repo.Update(context.Background(), want); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := repo.FindByID(context.Background(), id)
	if err != nil {
		t.Fatalf("find series: %v", err)
	}
	assertSameSeries(t, got, want)
}

func TestRecurringSeriesRepository_Update_WithNoExcludedDates_RemovesAll(t *testing.T) {
	setupTest(t)
	u := createTestUser(t)
	id := uuid.New()
	createSeries(t, buildSeries(t, id, u.ID(), "daily", nil, rs.NewIndefiniteEnd(),
		localDate(t, 2026, time.December, 1), localDate(t, 2026, time.December, 2)))

	if err := newSeriesRepo().Update(context.Background(),
		buildSeries(t, id, u.ID(), "daily", nil, rs.NewIndefiniteEnd())); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if n := countRows(t, "SELECT COUNT(*) FROM recurring_series_exceptions WHERE series_id = $1", id); n != 0 {
		t.Errorf("expected no excluded dates, got %d", n)
	}
}

func TestRecurringSeriesRepository_Update_KeepsCreatedAtOfRetainedExcludedDates(t *testing.T) {
	setupTest(t)
	u := createTestUser(t)
	id := uuid.New()
	kept := localDate(t, 2026, time.December, 1)
	createSeries(t, buildSeries(t, id, u.ID(), "daily", nil, rs.NewIndefiniteEnd(), kept))
	addedAt := time.Date(2026, time.November, 1, 0, 0, 0, 0, time.UTC)
	if _, err := testPool.Exec(context.Background(),
		"UPDATE recurring_series_exceptions SET created_at = $2 WHERE series_id = $1", id, addedAt); err != nil {
		t.Fatalf("backdate created_at: %v", err)
	}

	if err := newSeriesRepo().Update(context.Background(), buildSeries(t, id, u.ID(), "daily", nil,
		rs.NewIndefiniteEnd(), kept, localDate(t, 2026, time.December, 8))); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var got time.Time
	if err := testPool.QueryRow(context.Background(),
		"SELECT created_at FROM recurring_series_exceptions WHERE series_id = $1 AND excluded_date = '2026-12-01'",
		id).Scan(&got); err != nil {
		t.Fatalf("read created_at: %v", err)
	}
	if !got.Equal(addedAt) {
		t.Errorf("expected the retained excluded date to keep created_at %v, got %v", addedAt, got)
	}
}

func TestRecurringSeriesTable_Update_AdvancesUpdatedAt(t *testing.T) {
	setupTest(t)
	u := createTestUser(t)
	s := buildSeries(t, uuid.New(), u.ID(), "daily", nil, rs.NewIndefiniteEnd())
	createSeries(t, s)
	past := time.Now().Add(-time.Hour)
	if _, err := testPool.Exec(context.Background(),
		"UPDATE recurring_event_series SET updated_at = $2 WHERE id = $1", s.ID(), past); err != nil {
		t.Fatalf("backdate updated_at: %v", err)
	}
	// The trigger overwrites the value set above, so read the stored one back.
	var backdated time.Time
	if err := testPool.QueryRow(context.Background(),
		"SELECT updated_at FROM recurring_event_series WHERE id = $1", s.ID()).Scan(&backdated); err != nil {
		t.Fatalf("read updated_at: %v", err)
	}

	if err := newSeriesRepo().Update(context.Background(), s); err != nil {
		t.Fatalf("update series: %v", err)
	}

	var updated time.Time
	if err := testPool.QueryRow(context.Background(),
		"SELECT updated_at FROM recurring_event_series WHERE id = $1", s.ID()).Scan(&updated); err != nil {
		t.Fatalf("read updated_at: %v", err)
	}
	if !updated.After(backdated) {
		t.Errorf("expected updated_at to advance past %v, got %v", backdated, updated)
	}
}

func TestRecurringSeriesRepository_Delete_RemovesOccurrencesAndExcludedDates(t *testing.T) {
	setupTest(t)
	u := createTestUser(t)
	s := buildSeries(t, uuid.New(), u.ID(), "daily", nil, rs.NewIndefiniteEnd(), localDate(t, 2026, time.December, 1))
	createSeries(t, s)
	repo := newSeriesRepo()
	base := time.Date(2026, time.November, 30, 0, 0, 0, 0, time.UTC)
	if err := repo.CreateOccurrences(context.Background(), []event.Event{
		occurrence(t, u.ID(), s.ID(), base, 0), occurrence(t, u.ID(), s.ID(), base, 2),
	}); err != nil {
		t.Fatalf("create occurrences: %v", err)
	}

	if err := repo.Delete(context.Background(), s.ID()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := repo.FindByID(context.Background(), s.ID()); !errors.Is(err, rs.ErrRecurringSeriesNotFound) {
		t.Errorf("expected ErrRecurringSeriesNotFound, got %v", err)
	}
	if n := countRows(t, "SELECT COUNT(*) FROM events WHERE series_id = $1", s.ID()); n != 0 {
		t.Errorf("expected occurrences to be deleted, got %d", n)
	}
	if n := countRows(t, "SELECT COUNT(*) FROM recurring_series_exceptions WHERE series_id = $1", s.ID()); n != 0 {
		t.Errorf("expected excluded dates to be deleted, got %d", n)
	}
}

func TestRecurringSeriesTable_UserDeleted_RemovesSeries(t *testing.T) {
	setupTest(t)
	u := createTestUser(t)
	s := buildSeries(t, uuid.New(), u.ID(), "daily", nil, rs.NewIndefiniteEnd())
	createSeries(t, s)

	if _, err := testPool.Exec(context.Background(), "DELETE FROM users WHERE id = $1", u.ID()); err != nil {
		t.Fatalf("delete user: %v", err)
	}

	if n := countRows(t, "SELECT COUNT(*) FROM recurring_event_series WHERE id = $1", s.ID()); n != 0 {
		t.Errorf("expected series to be deleted with the user, got %d", n)
	}
}

func TestRecurringSeriesRepository_CreateOccurrences_SavesTwoYearsOfDailyOccurrencesWithTags(t *testing.T) {
	setupTest(t)
	u := createTestUser(t)
	work := createTestTag(t, u.ID(), "Work")
	private := createTestTag(t, u.ID(), "Private")
	s := buildSeries(t, uuid.New(), u.ID(), "daily", nil, rs.NewIndefiniteEnd())
	createSeries(t, s)
	repo := newSeriesRepo()

	tags := []event.EventTag{mustEventTag(t, private.ID(), 1), mustEventTag(t, work.ID(), 2)}
	base := time.Date(2026, time.November, 30, 1, 0, 0, 0, time.UTC)
	const n = 731
	occurrences := make([]event.Event, 0, n)
	// Build them in reverse to check that FindOccurrences orders by start.
	for day := n - 1; day >= 0; day-- {
		occurrences = append(occurrences, occurrence(t, u.ID(), s.ID(), base, day, tags...))
	}

	if err := repo.CreateOccurrences(context.Background(), occurrences); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := repo.FindOccurrences(context.Background(), s.ID())
	if err != nil {
		t.Fatalf("find occurrences: %v", err)
	}
	if len(got) != n {
		t.Fatalf("expected %d occurrences, got %d", n, len(got))
	}
	for i, e := range got {
		if want := base.AddDate(0, 0, i); !e.StartAt().Equal(want) {
			t.Fatalf("occurrence[%d] start mismatch: got %v, want %v", i, e.StartAt(), want)
		}
		assertSeriesID(t, e, s.ID())
		assertEventTags(t, e.Tags(), tags)
	}
}

func TestRecurringSeriesRepository_CreateOccurrences_InCallerTx_RollsBackWithCaller(t *testing.T) {
	setupTest(t)
	u := createTestUser(t)
	s := buildSeries(t, uuid.New(), u.ID(), "daily", nil, rs.NewIndefiniteEnd())
	createSeries(t, s)
	repo := newSeriesRepo()
	base := time.Date(2026, time.November, 30, 1, 0, 0, 0, time.UTC)

	runInTestTx(t, func(ctx context.Context) {
		if err := repo.CreateOccurrences(ctx, []event.Event{occurrence(t, u.ID(), s.ID(), base, 0)}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if n := countRows(t, "SELECT COUNT(*) FROM events WHERE series_id = $1", s.ID()); n != 0 {
		t.Errorf("expected the occurrences to roll back, got %d", n)
	}
}

func TestRecurringSeriesRepository_CreateOccurrences_WithoutSeriesID_WritesNothing(t *testing.T) {
	setupTest(t)
	u := createTestUser(t)
	s := buildSeries(t, uuid.New(), u.ID(), "daily", nil, rs.NewIndefiniteEnd())
	createSeries(t, s)
	base := time.Date(2026, time.November, 30, 1, 0, 0, 0, time.UTC)
	standalone := occurrence(t, u.ID(), s.ID(), base, 1).DetachFromSeries()

	err := newSeriesRepo().CreateOccurrences(context.Background(),
		[]event.Event{occurrence(t, u.ID(), s.ID(), base, 0), standalone})
	if err == nil {
		t.Fatal("expected an error for an event without a series")
	}
	if n := countRows(t, "SELECT COUNT(*) FROM events WHERE user_id = $1", u.ID()); n != 0 {
		t.Errorf("expected no events to be written, got %d", n)
	}
}

func TestRecurringSeriesRepository_CreateOccurrences_Empty_DoesNothing(t *testing.T) {
	setupTest(t)

	if err := newSeriesRepo().CreateOccurrences(context.Background(), nil); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// seedOccurrences creates a series with occurrences on days 0..count-1 from
// base, plus a detached event and an occurrence of another series that the
// series operations must leave alone.
func seedOccurrences(t *testing.T, count int) (seriesID, otherSeriesID, detachedID uuid.UUID, base time.Time) {
	t.Helper()
	u := createTestUser(t)
	s := buildSeries(t, uuid.New(), u.ID(), "daily", nil, rs.NewIndefiniteEnd())
	other := buildSeries(t, uuid.New(), u.ID(), "daily", nil, rs.NewIndefiniteEnd())
	createSeries(t, s)
	createSeries(t, other)
	base = time.Date(2026, time.November, 30, 1, 0, 0, 0, time.UTC)

	occurrences := make([]event.Event, 0, count+1)
	for day := range count {
		occurrences = append(occurrences, occurrence(t, u.ID(), s.ID(), base, day))
	}
	occurrences = append(occurrences, occurrence(t, u.ID(), other.ID(), base, 0))
	if err := newSeriesRepo().CreateOccurrences(context.Background(), occurrences); err != nil {
		t.Fatalf("create occurrences: %v", err)
	}

	detached := occurrence(t, u.ID(), s.ID(), base, count).DetachFromSeries()
	if _, err := repository.NewEventRepository(testPool, testLogger, testTracerProvider).
		Create(context.Background(), detached); err != nil {
		t.Fatalf("create detached event: %v", err)
	}
	return s.ID(), other.ID(), detached.ID(), base
}

func TestRecurringSeriesRepository_FindOccurrences_ExcludesOtherSeriesAndDetachedEvents(t *testing.T) {
	setupTest(t)
	seriesID, _, detachedID, _ := seedOccurrences(t, 3)

	got, err := newSeriesRepo().FindOccurrences(context.Background(), seriesID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("expected 3 occurrences, got %d", len(got))
	}
	for _, e := range got {
		assertSeriesID(t, e, seriesID)
		if e.ID() == detachedID {
			t.Error("expected the detached event to be excluded")
		}
	}
}

func TestRecurringSeriesRepository_CountOccurrences(t *testing.T) {
	setupTest(t)
	seriesID, _, _, _ := seedOccurrences(t, 3)
	repo := newSeriesRepo()

	n, err := repo.CountOccurrences(context.Background(), seriesID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 3 {
		t.Errorf("expected 3, got %d", n)
	}

	n, err = repo.CountOccurrences(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 0 {
		t.Errorf("expected 0 for a series without occurrences, got %d", n)
	}
}

func TestRecurringSeriesRepository_DeleteOccurrencesFrom_DeletesFromTheGivenStartOnly(t *testing.T) {
	setupTest(t)
	seriesID, otherSeriesID, detachedID, base := seedOccurrences(t, 4)
	repo := newSeriesRepo()

	if err := repo.DeleteOccurrencesFrom(context.Background(), seriesID, base.AddDate(0, 0, 2)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := repo.FindOccurrences(context.Background(), seriesID)
	if err != nil {
		t.Fatalf("find occurrences: %v", err)
	}
	if len(got) != 2 || !got[0].StartAt().Equal(base) || !got[1].StartAt().Equal(base.AddDate(0, 0, 1)) {
		t.Errorf("expected the first two occurrences to remain, got %d", len(got))
	}
	assertUntouched(t, otherSeriesID, detachedID)
}

func TestRecurringSeriesRepository_DeleteOccurrences_KeepsSeries(t *testing.T) {
	setupTest(t)
	seriesID, otherSeriesID, detachedID, _ := seedOccurrences(t, 3)
	repo := newSeriesRepo()

	if err := repo.DeleteOccurrences(context.Background(), seriesID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if n := countRows(t, "SELECT COUNT(*) FROM events WHERE series_id = $1", seriesID); n != 0 {
		t.Errorf("expected no occurrences, got %d", n)
	}
	if _, err := repo.FindByID(context.Background(), seriesID); err != nil {
		t.Errorf("expected the series to remain, got %v", err)
	}
	assertUntouched(t, otherSeriesID, detachedID)
}

func assertUntouched(t *testing.T, otherSeriesID, detachedID uuid.UUID) {
	t.Helper()
	if n := countRows(t, "SELECT COUNT(*) FROM events WHERE series_id = $1", otherSeriesID); n != 1 {
		t.Errorf("expected the other series' occurrence to remain, got %d", n)
	}
	if n := countRows(t, "SELECT COUNT(*) FROM events WHERE id = $1", detachedID); n != 1 {
		t.Errorf("expected the detached event to remain, got %d", n)
	}
}
