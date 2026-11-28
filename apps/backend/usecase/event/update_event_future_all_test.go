package event_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"

	domainevent "github.com/Haya372/ai-trial/backend/domain/event"
	rs "github.com/Haya372/ai-trial/backend/domain/recurringeventseries"
	"github.com/Haya372/ai-trial/backend/domain/tag"
	eventuc "github.com/Haya372/ai-trial/backend/usecase/event"
)

// firstOccStart is the first occurrence of the weekly series the tests use;
// secondOccStart is a week later.
var firstOccStart = time.Date(2026, 11, 23, 10, 0, 0, 0, time.UTC)

func weekStart(i int) time.Time { return firstOccStart.AddDate(0, 0, 7*i) }

// seriesWith returns a weekly UTC series with the given end, excluding the
// dates of the weeks in excludedWeeks.
func seriesWith(t *testing.T, userID uuid.UUID, rule rs.RecurrenceRule, excludedWeeks ...int) rs.RecurringEventSeries {
	t.Helper()
	excluded := make([]rs.ExcludedDate, len(excludedWeeks))
	for i, w := range excludedWeeks {
		excluded[i] = rs.NewExcludedDate(rs.LocalDateOf(weekStart(w)))
	}
	s, err := rs.New(uuid.New(), userID, rule, excluded)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

// occurrences returns the events of the series on the given weeks, tagged
// with tagIDs. The one on targetWeek has the fixture's event ID.
func (f *scopeFixture) occurrences(
	t *testing.T, series rs.RecurringEventSeries, targetWeek int, weeks []int, tagIDs ...uuid.UUID,
) []domainevent.Event {
	t.Helper()
	tags, err := domainevent.NewEventTags(tagIDs)
	if err != nil {
		t.Fatal(err)
	}
	out := make([]domainevent.Event, len(weeks))
	for i, w := range weeks {
		id := uuid.New()
		if w == targetWeek {
			id = f.eventID
		}
		start := weekStart(w)
		ev, err := domainevent.New(id, f.userID, "Old title", "", start, start.Add(time.Hour), "", "")
		if err != nil {
			t.Fatal(err)
		}
		if ev, err = ev.WithTags(tags); err != nil {
			t.Fatal(err)
		}
		out[i] = ev.WithSeriesID(series.ID())
	}
	return out
}

// expectOccurrences stubs the lock of the target and the read of the series' occurrences.
func (f *scopeFixture) expectOccurrences(series rs.RecurringEventSeries, occs []domainevent.Event) {
	f.expectLockOccurrence(target(occs, f.eventID), series)
	f.seriesRepo.EXPECT().FindOccurrences(gomock.Any(), series.ID()).Return(occs, nil)
}

func target(events []domainevent.Event, id uuid.UUID) domainevent.Event {
	for _, e := range events {
		if e.ID() == id {
			return e
		}
	}
	return nil
}

func (f *scopeFixture) keepTags(t *testing.T, tagIDs ...uuid.UUID) {
	t.Helper()
	stored := make([]tag.Tag, len(tagIDs))
	for i, id := range tagIDs {
		stored[i] = newStoredTag(t, id, f.userID)
	}
	f.tagRepo.EXPECT().FindByIDsForShare(gomock.Any(), tagIDs).Return(stored, nil)
}

func assertUntil(t *testing.T, s rs.RecurringEventSeries, want time.Time) {
	t.Helper()
	d, ok := s.Rule().End.UntilDate()
	if !ok || d != rs.LocalDateOf(want) {
		t.Errorf("end: got %+v, want until %s", s.Rule().End, rs.LocalDateOf(want))
	}
}

func assertCount(t *testing.T, s rs.RecurringEventSeries, want int) {
	t.Helper()
	if n, ok := s.Rule().End.Count(); !ok || n != want {
		t.Errorf("end: got %+v, want count %d", s.Rule().End, want)
	}
}

func excludedDateStrings(s rs.RecurringEventSeries) []string {
	dates := s.ExcludedDates()
	out := make([]string, 0, len(dates))
	for _, e := range dates {
		out = append(out, e.Date().String())
	}
	return out
}

func TestUpdateEventCommand_Execute_Future_NullRecurrence_EndsSeriesBeforeTheTarget(t *testing.T) {
	f := newScopeFixture(t)
	series := seriesWith(t, f.userID, weeklyCount(t, 5), 3)
	tagID := uuid.New()
	occs := f.occurrences(t, series, 1, []int{0, 1, 2, 4}, tagID)
	f.expectOccurrences(series, occs)
	f.keepTags(t, tagID)

	// Stopping the repetition changes the rule, so the date may change.
	moved := secondOccStart.AddDate(0, 0, 1)
	gomock.InOrder(
		f.repo.EXPECT().Update(gomock.Any(), gomock.Any()).DoAndReturn(
			func(_ context.Context, e domainevent.Event) error {
				if _, ok := e.SeriesID(); ok {
					t.Error("expected the target to become one-off")
				}
				if e.ID() != f.eventID || e.Title() != newTitle || !e.StartAt().Equal(moved) {
					t.Errorf("target: got %v %q at %v", e.ID(), e.Title(), e.StartAt())
				}
				assertEventTags(t, e.Tags(), occs[1].Tags())
				return nil
			},
		),
		f.seriesRepo.EXPECT().Update(gomock.Any(), gomock.Any()).DoAndReturn(
			func(_ context.Context, s rs.RecurringEventSeries) error {
				assertUntil(t, s, secondOccStart.AddDate(0, 0, -1))
				if got := excludedDateStrings(s); len(got) != 0 {
					t.Errorf("excluded dates: got %v, want none", got)
				}
				return nil
			},
		),
		f.seriesRepo.EXPECT().DeleteOccurrencesFrom(gomock.Any(), series.ID(), secondOccStart).Return(nil),
	)

	in := f.input(moved)
	in.Scope = scopeFuture
	in.Recurrence = eventuc.RecurrenceUpdate{Present: true}
	out, err := f.command().Execute(context.Background(), f.userID, in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Series != nil || out.Event.ID() != f.eventID {
		t.Errorf("got event %v series %v, want the one-off target", out.Event.ID(), out.Series)
	}
}

// "all" with null is SCOPE_NOT_ALLOWED, but the client offers only "this and
// following" when the repetition stops, so the first occurrence must work.
func TestUpdateEventCommand_Execute_Future_NullRecurrence_FirstOccurrence_DeletesSeries(t *testing.T) {
	f := newScopeFixture(t)
	series := seriesWith(t, f.userID, weeklyCount(t, 3))
	occs := f.occurrences(t, series, 0, []int{0, 1, 2})
	f.expectOccurrences(series, occs)
	gomock.InOrder(
		f.repo.EXPECT().Update(gomock.Any(), gomock.Any()).DoAndReturn(
			func(_ context.Context, e domainevent.Event) error {
				if _, ok := e.SeriesID(); ok {
					t.Error("expected the target to become one-off")
				}
				return nil
			},
		),
		f.seriesRepo.EXPECT().Delete(gomock.Any(), series.ID()).Return(nil),
	)

	in := f.input(firstOccStart)
	in.Scope = scopeFuture
	in.Recurrence = eventuc.RecurrenceUpdate{Present: true}
	if _, err := f.command().Execute(context.Background(), f.userID, in); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUpdateEventCommand_Execute_Future_KeepingRule_MovesFollowingOccurrencesToANewSeries(t *testing.T) {
	f := newScopeFixture(t)
	// Week 3 was deleted or detached, so it stays excluded in the new series.
	series := seriesWith(t, f.userID, weeklyCount(t, 5), 3)
	tagID := uuid.New()
	occs := f.occurrences(t, series, 1, []int{0, 1, 2, 4}, tagID)
	f.expectOccurrences(series, occs)
	f.keepTags(t, tagID)

	newStart := secondOccStart.Add(3 * time.Hour)
	var created rs.RecurringEventSeries
	f.seriesRepo.EXPECT().Update(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, s rs.RecurringEventSeries) error {
			if s.ID() != series.ID() {
				t.Errorf("updated series: got %v, want the current one", s.ID())
			}
			assertUntil(t, s, secondOccStart.AddDate(0, 0, -1))
			return nil
		},
	)
	gomock.InOrder(
		f.seriesRepo.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(
			func(_ context.Context, s rs.RecurringEventSeries) (rs.RecurringEventSeries, error) {
				// Weeks 1, 2 and 4 remain, plus the excluded week 3.
				assertCount(t, s, 4)
				if got := excludedDateStrings(s); len(got) != 1 || got[0] != "2026-12-14" {
					t.Errorf("excluded dates: got %v, want [2026-12-14]", got)
				}
				if s.Rule().Pattern != series.Rule().Pattern || s.ID() == series.ID() {
					t.Errorf("new series: got %+v id %v", s.Rule(), s.ID())
				}
				created = s
				return s, nil
			},
		),
		f.seriesRepo.EXPECT().UpdateOccurrences(gomock.Any(), gomock.Any()).DoAndReturn(
			func(_ context.Context, got []domainevent.Event) error {
				wantIDs := []uuid.UUID{occs[1].ID(), occs[2].ID(), occs[3].ID()}
				wantStarts := []time.Time{newStart, weekStart(2).Add(3 * time.Hour), weekStart(4).Add(3 * time.Hour)}
				if len(got) != len(wantIDs) {
					t.Fatalf("occurrences: got %d, want %d", len(got), len(wantIDs))
				}
				for i, e := range got {
					if e.ID() != wantIDs[i] || e.Title() != newTitle {
						t.Errorf("occurrence %d: got %v %q", i, e.ID(), e.Title())
					}
					if !e.StartAt().Equal(wantStarts[i]) || !e.EndAt().Equal(wantStarts[i].Add(time.Hour)) {
						t.Errorf("occurrence %d: got %v-%v, want start %v", i, e.StartAt(), e.EndAt(), wantStarts[i])
					}
					if id, ok := e.SeriesID(); !ok || id != created.ID() {
						t.Errorf("occurrence %d SeriesID: got (%v, %v), want the new series", i, id, ok)
					}
					assertEventTags(t, e.Tags(), occs[1].Tags())
				}
				return nil
			},
		),
	)

	in := f.input(newStart)
	in.Scope = scopeFuture
	out, err := f.command().Execute(context.Background(), f.userID, in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Event.ID() != f.eventID || !out.Event.StartAt().Equal(newStart) {
		t.Errorf("returned event: got %v at %v", out.Event.ID(), out.Event.StartAt())
	}
	if out.Series == nil || out.Series.ID() != created.ID() {
		t.Errorf("Series: got %v, want the new series", out.Series)
	}
	if len(out.Tags) != 1 || out.Tags[0].ID != tagID {
		t.Errorf("Tags: got %+v", out.Tags)
	}
}

// The client sends the series' rule back unchanged; for the last occurrence
// the new series gets a count of 1, which a new rule could not have.
func TestUpdateEventCommand_Execute_Future_SameRuleObject_LastOccurrence_KeepsRuleWithCount1(t *testing.T) {
	f := newScopeFixture(t)
	series := seriesWith(t, f.userID, weeklyCount(t, 2))
	occs := f.occurrences(t, series, 1, []int{0, 1})
	f.expectOccurrences(series, occs)
	f.seriesRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)
	f.seriesRepo.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, s rs.RecurringEventSeries) (rs.RecurringEventSeries, error) {
			assertCount(t, s, 1)
			return s, nil
		},
	)
	f.seriesRepo.EXPECT().UpdateOccurrences(gomock.Any(), gomock.Len(1)).Return(nil)

	in := f.input(secondOccStart)
	in.Scope = scopeFuture
	in.Recurrence = eventuc.RecurrenceUpdate{Present: true, Spec: weeklyCountInput(2)}
	if _, err := f.command().Execute(context.Background(), f.userID, in); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUpdateEventCommand_Execute_FutureOrAll_KeepingRule_DateChanged_ReturnsDateChangeNotAllowed(t *testing.T) {
	for _, scope := range []string{scopeFuture, scopeAll} {
		t.Run(scope, func(t *testing.T) {
			f := newScopeFixture(t)
			series := seriesWith(t, f.userID, weeklyCount(t, 3))
			occs := f.occurrences(t, series, 1, []int{0, 1, 2})
			f.expectOccurrences(series, occs)

			// No write is expected.
			in := f.input(secondOccStart.AddDate(0, 0, 1))
			in.Scope = scope
			_, err := f.command().Execute(context.Background(), f.userID, in)
			if !errors.Is(err, rs.ErrRecurringOccurrenceDateChangeNotAllowed) {
				t.Fatalf("expected ErrRecurringOccurrenceDateChangeNotAllowed, got %v", err)
			}
		})
	}
}

func TestUpdateEventCommand_Execute_Future_NewRule_StartsANewSeriesFromTheTarget(t *testing.T) {
	f := newScopeFixture(t)
	series := seriesWith(t, f.userID, weeklyCount(t, 5), 3)
	occs := f.occurrences(t, series, 1, []int{0, 1, 2, 4})
	f.expectOccurrences(series, occs)

	// A new rule may move the date.
	moved := secondOccStart.AddDate(0, 0, 1)
	var created rs.RecurringEventSeries
	f.seriesRepo.EXPECT().Update(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, s rs.RecurringEventSeries) error {
			assertUntil(t, s, secondOccStart.AddDate(0, 0, -1))
			return nil
		},
	)
	createSeries := f.seriesRepo.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, s rs.RecurringEventSeries) (rs.RecurringEventSeries, error) {
			assertCount(t, s, 3)
			if len(s.ExcludedDates()) != 0 {
				t.Errorf("excluded dates: got %v, want none", excludedDateStrings(s))
			}
			created = s
			return s, nil
		},
	)
	moveTarget := f.repo.EXPECT().Update(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, e domainevent.Event) error {
			if id, ok := e.SeriesID(); !ok || id != created.ID() || e.ID() != f.eventID {
				t.Errorf("target: got %v in series (%v, %v)", e.ID(), id, ok)
			}
			return nil
		},
	).After(createSeries)
	// The target already moved to the new series, so this keeps it.
	f.seriesRepo.EXPECT().DeleteOccurrencesFrom(gomock.Any(), series.ID(), secondOccStart).Return(nil).After(moveTarget)
	f.seriesRepo.EXPECT().CreateOccurrences(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, got []domainevent.Event) error {
			want := []time.Time{moved.AddDate(0, 0, 7), moved.AddDate(0, 0, 14)}
			if len(got) != len(want) {
				t.Fatalf("occurrences: got %d, want %d", len(got), len(want))
			}
			for i, e := range got {
				if !e.StartAt().Equal(want[i]) {
					t.Errorf("occurrence %d: got %v, want %v", i, e.StartAt(), want[i])
				}
				if id, ok := e.SeriesID(); !ok || id != created.ID() {
					t.Errorf("occurrence %d SeriesID: got (%v, %v)", i, id, ok)
				}
			}
			return nil
		},
	).After(createSeries)

	in := f.input(moved)
	in.Scope = scopeFuture
	in.Recurrence = eventuc.RecurrenceUpdate{Present: true, Spec: weeklyCountInput(3)}
	out, err := f.command().Execute(context.Background(), f.userID, in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Event.ID() != f.eventID || out.Series == nil || out.Series.ID() != created.ID() {
		t.Errorf("got event %v series %v, want the target in the new series", out.Event.ID(), out.Series)
	}
}

func TestUpdateEventCommand_Execute_FutureOrAll_NewRuleOutOfRange_ReturnsValidationError(t *testing.T) {
	for _, scope := range []string{scopeFuture, scopeAll} {
		t.Run(scope, func(t *testing.T) {
			f := newScopeFixture(t)
			series := seriesWith(t, f.userID, weeklyCount(t, 5))
			occs := f.occurrences(t, series, 1, []int{0, 1, 2})
			f.expectOccurrences(series, occs)

			in := f.input(secondOccStart)
			in.Scope = scope
			in.Recurrence = eventuc.RecurrenceUpdate{Present: true, Spec: weeklyCountInput(1)}
			_, err := f.command().Execute(context.Background(), f.userID, in)
			assertValidationDetail(t, err, fieldCount, codeOutOfRange)
		})
	}
}

func TestUpdateEventCommand_Execute_Future_FirstOccurrence_KeepingRule_UpdatesEveryOccurrence(t *testing.T) {
	f := newScopeFixture(t)
	series := seriesWith(t, f.userID, weeklyCount(t, 3))
	occs := f.occurrences(t, series, 0, []int{0, 1, 2})
	f.expectOccurrences(series, occs)
	// No Create or series Update: "all" keeps the series as it is.
	f.seriesRepo.EXPECT().UpdateOccurrences(gomock.Any(), gomock.Len(3)).Return(nil)

	in := f.input(firstOccStart)
	in.Scope = scopeFuture
	out, err := f.command().Execute(context.Background(), f.userID, in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Series == nil || out.Series.ID() != series.ID() {
		t.Errorf("Series: got %v, want the current series", out.Series)
	}
}

func TestUpdateEventCommand_Execute_All_KeepingRule_UpdatesEveryOccurrence(t *testing.T) {
	f := newScopeFixture(t)
	series := seriesWith(t, f.userID, weeklyCount(t, 5), 3)
	oldTag, newTag := uuid.New(), uuid.New()
	occs := f.occurrences(t, series, 1, []int{0, 1, 2, 4}, oldTag)
	f.expectOccurrences(series, occs)
	f.keepTags(t, newTag)

	newStart := secondOccStart.Add(-2 * time.Hour)
	f.seriesRepo.EXPECT().UpdateOccurrences(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, got []domainevent.Event) error {
			if len(got) != len(occs) {
				t.Fatalf("occurrences: got %d, want %d", len(got), len(occs))
			}
			wantTags, _ := domainevent.NewEventTags([]uuid.UUID{newTag})
			for i, e := range got {
				want := occs[i].StartAt().Add(-2 * time.Hour)
				if e.ID() != occs[i].ID() || e.Title() != newTitle || !e.StartAt().Equal(want) {
					t.Errorf("occurrence %d: got %v %q at %v, want %v at %v", i, e.ID(), e.Title(), e.StartAt(), occs[i].ID(), want)
				}
				if id, ok := e.SeriesID(); !ok || id != series.ID() {
					t.Errorf("occurrence %d SeriesID: got (%v, %v), want the current series", i, id, ok)
				}
				assertEventTags(t, e.Tags(), wantTags)
			}
			return nil
		},
	)

	in := f.input(newStart)
	in.Scope = scopeAll
	in.TagIDs = &[]uuid.UUID{newTag}
	out, err := f.command().Execute(context.Background(), f.userID, in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Event.ID() != f.eventID || !out.Event.StartAt().Equal(newStart) {
		t.Errorf("returned event: got %v at %v", out.Event.ID(), out.Event.StartAt())
	}
	if out.Series == nil || out.Series.ID() != series.ID() {
		t.Errorf("Series: got %v, want the current series", out.Series)
	}
}

func TestUpdateEventCommand_Execute_All_TagOfAnotherUser_ReturnsNotFoundWithoutWriting(t *testing.T) {
	f := newScopeFixture(t)
	series := seriesWith(t, f.userID, weeklyCount(t, 3))
	occs := f.occurrences(t, series, 1, []int{0, 1, 2})
	f.expectOccurrences(series, occs)
	foreign := uuid.New()
	f.tagRepo.EXPECT().FindByIDsForShare(gomock.Any(), []uuid.UUID{foreign}).
		Return([]tag.Tag{newStoredTag(t, foreign, uuid.New())}, nil)

	in := f.input(secondOccStart)
	in.Scope = scopeAll
	in.TagIDs = &[]uuid.UUID{foreign}
	_, err := f.command().Execute(context.Background(), f.userID, in)
	assertValidationDetail(t, err, "tagIds", "NOT_FOUND")
}

// An "all" change covers the earlier occurrences too, so the regenerated
// series starts from the first one, moved as far as the target moved.
func TestUpdateEventCommand_Execute_All_NewRule_RegeneratesFromTheFirstOccurrence(t *testing.T) {
	f := newScopeFixture(t)
	series := seriesWith(t, f.userID, weeklyCount(t, 5), 3)
	occs := f.occurrences(t, series, 1, []int{0, 1, 2, 4})
	f.expectOccurrences(series, occs)

	// The target moves a day later and three hours later.
	moved := secondOccStart.AddDate(0, 0, 1).Add(3 * time.Hour)
	var createdEvents []domainevent.Event
	gomock.InOrder(
		f.seriesRepo.EXPECT().Update(gomock.Any(), gomock.Any()).DoAndReturn(
			func(_ context.Context, s rs.RecurringEventSeries) error {
				if s.ID() != series.ID() {
					t.Errorf("series ID: got %v, want %v", s.ID(), series.ID())
				}
				assertCount(t, s, 3)
				if len(s.ExcludedDates()) != 0 {
					t.Errorf("excluded dates: got %v, want none", excludedDateStrings(s))
				}
				return nil
			},
		),
		f.seriesRepo.EXPECT().DeleteOccurrences(gomock.Any(), series.ID()).Return(nil),
		f.seriesRepo.EXPECT().CreateOccurrences(gomock.Any(), gomock.Any()).DoAndReturn(
			func(_ context.Context, got []domainevent.Event) error {
				first := firstOccStart.AddDate(0, 0, 1).Add(3 * time.Hour)
				want := []time.Time{first, first.AddDate(0, 0, 7), first.AddDate(0, 0, 14)}
				if len(got) != len(want) {
					t.Fatalf("occurrences: got %d, want %d", len(got), len(want))
				}
				for i, e := range got {
					if !e.StartAt().Equal(want[i]) || e.Title() != newTitle {
						t.Errorf("occurrence %d: got %q at %v, want %v", i, e.Title(), e.StartAt(), want[i])
					}
					if id, ok := e.SeriesID(); !ok || id != series.ID() {
						t.Errorf("occurrence %d SeriesID: got (%v, %v)", i, id, ok)
					}
				}
				createdEvents = got
				return nil
			},
		),
	)

	in := f.input(moved)
	in.Scope = scopeAll
	in.Recurrence = eventuc.RecurrenceUpdate{Present: true, Spec: weeklyCountInput(3)}
	out, err := f.command().Execute(context.Background(), f.userID, in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(createdEvents) == 0 || out.Event.ID() != createdEvents[0].ID() {
		t.Errorf("returned event: got %v, want the first regenerated occurrence", out.Event.ID())
	}
	if out.Series == nil || out.Series.ID() != series.ID() {
		t.Errorf("Series: got %v, want the regenerated series", out.Series)
	}
}
