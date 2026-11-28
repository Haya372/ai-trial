package event_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"

	"github.com/Haya372/ai-trial/backend/domain"
	domainevent "github.com/Haya372/ai-trial/backend/domain/event"
	eventmock "github.com/Haya372/ai-trial/backend/domain/event/generated"
	rs "github.com/Haya372/ai-trial/backend/domain/recurringeventseries"
	rsmock "github.com/Haya372/ai-trial/backend/domain/recurringeventseries/generated"
	"github.com/Haya372/ai-trial/backend/domain/tag"
	tagmock "github.com/Haya372/ai-trial/backend/domain/tag/generated"
	eventuc "github.com/Haya372/ai-trial/backend/usecase/event"
	"github.com/Haya372/ai-trial/backend/usecase/testutil"
)

const (
	scopeSingle         = "single"
	scopeFuture         = "future"
	scopeAll            = "all"
	fieldScope          = "scope"
	codeNotInSeries     = "NOT_IN_SERIES"
	codeScopeNotAllowed = "SCOPE_NOT_ALLOWED"
	fieldTimezone       = "recurrence.timezone"
)

// The second occurrence of a weekly series that started on Monday 2026-11-23.
var secondOccStart = time.Date(2026, 11, 30, 10, 0, 0, 0, time.UTC)

type scopeFixture struct {
	ctrl       *gomock.Controller
	repo       *eventmock.MockRepository
	seriesRepo *rsmock.MockRepository
	tagRepo    *tagmock.MockRepository
	userID     uuid.UUID
	eventID    uuid.UUID
}

func newScopeFixture(t *testing.T) *scopeFixture {
	t.Helper()
	ctrl := gomock.NewController(t)
	return &scopeFixture{
		ctrl:       ctrl,
		repo:       eventmock.NewMockRepository(ctrl),
		seriesRepo: rsmock.NewMockRepository(ctrl),
		tagRepo:    tagmock.NewMockRepository(ctrl),
		userID:     uuid.New(),
		eventID:    uuid.New(),
	}
}

func (f *scopeFixture) command() *eventuc.UpdateEventCommand {
	return eventuc.NewUpdateEventCommand(f.repo, f.seriesRepo, f.tagRepo, &testutil.StubTxManager{}, testLogger)
}

func (f *scopeFixture) oneOff(t *testing.T, start time.Time) domainevent.Event {
	t.Helper()
	ev, err := domainevent.New(f.eventID, f.userID, "Old title", "", start, start.Add(time.Hour), "", "")
	if err != nil {
		t.Fatal(err)
	}
	return ev
}

func weeklyRule(t *testing.T, end rs.RecurrenceEnd) rs.RecurrenceRule {
	t.Helper()
	p, err := rs.NewRecurrencePattern(patternWeekly, nil)
	if err != nil {
		t.Fatal(err)
	}
	tz, err := rs.NewSeriesTimezone(timezoneUTC)
	if err != nil {
		t.Fatal(err)
	}
	return rs.RecurrenceRule{Pattern: p, End: end, Timezone: tz}
}

func weeklyCount(t *testing.T, n int) rs.RecurrenceRule {
	t.Helper()
	end, err := rs.RestoreCountEnd(n)
	if err != nil {
		t.Fatal(err)
	}
	return weeklyRule(t, end)
}

func weeklyCountInput(n int) *eventuc.RecurrenceInput {
	return &eventuc.RecurrenceInput{PatternType: patternWeekly, EndType: endCount, Count: ptr(n), Timezone: timezoneUTC}
}

func newSeries(t *testing.T, userID uuid.UUID, rule rs.RecurrenceRule) rs.RecurringEventSeries {
	t.Helper()
	s, err := rs.New(uuid.New(), userID, rule, nil)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

// expectLockOccurrence stubs lockTarget for an occurrence: the unlocked read,
// then the series row, then the event row, in this order.
func (f *scopeFixture) expectLockOccurrence(occ domainevent.Event, series rs.RecurringEventSeries) {
	gomock.InOrder(
		f.repo.EXPECT().FindByID(gomock.Any(), occ.ID()).Return(occ, nil),
		f.seriesRepo.EXPECT().FindByIDForUpdate(gomock.Any(), series.ID()).Return(series, nil),
		f.repo.EXPECT().FindByIDForUpdate(gomock.Any(), occ.ID()).Return(occ, nil),
	)
}

func (f *scopeFixture) input(start time.Time) eventuc.UpdateEventInput {
	return eventuc.UpdateEventInput{
		ID: f.eventID, Title: newTitle, StartAt: start, EndAt: start.Add(time.Hour),
	}
}

func assertValidationDetail(t *testing.T, err error, field, code string) {
	t.Helper()
	ve, ok := errors.AsType[*domain.ValidationError](err)
	if !ok {
		t.Fatalf("expected *domain.ValidationError, got %T: %v", err, err)
	}
	if len(ve.Details) != 1 || ve.Details[0].Field != field || ve.Details[0].Code != code {
		t.Fatalf("details: got %+v, want [{%s %s}]", ve.Details, field, code)
	}
}

func TestUpdateEventCommand_Execute_InvalidScope_ReturnsInvalidFormat(t *testing.T) {
	f := newScopeFixture(t)
	in := f.input(secondOccStart)
	in.Scope = "weekly"

	// Checked before the transaction: no repository call is expected.
	_, err := f.command().Execute(context.Background(), f.userID, in)
	assertValidationDetail(t, err, fieldScope, codeInvalidFormat)
}

func TestUpdateEventCommand_Execute_OneOff_FutureOrAll_ReturnsNotInSeries(t *testing.T) {
	for _, scope := range []string{scopeFuture, scopeAll} {
		t.Run(scope, func(t *testing.T) {
			f := newScopeFixture(t)
			expectLockTarget(f.repo, f.oneOff(t, secondOccStart))
			in := f.input(secondOccStart)
			in.Scope = scope

			_, err := f.command().Execute(context.Background(), f.userID, in)
			assertValidationDetail(t, err, fieldScope, codeNotInSeries)
		})
	}
}

func TestUpdateEventCommand_Execute_OneOff_NullRecurrence_StaysOneOff(t *testing.T) {
	f := newScopeFixture(t)
	expectLockTarget(f.repo, f.oneOff(t, secondOccStart))
	f.repo.EXPECT().Update(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, e domainevent.Event) error {
			if _, ok := e.SeriesID(); ok {
				t.Error("expected the event to stay one-off")
			}
			return nil
		},
	)
	in := f.input(secondOccStart)
	in.Scope = scopeSingle
	in.Recurrence = eventuc.RecurrenceUpdate{Present: true}

	out, err := f.command().Execute(context.Background(), f.userID, in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Series != nil {
		t.Errorf("Series: got %v, want nil", out.Series)
	}
}

func TestUpdateEventCommand_Execute_OneOff_WithRecurrence_ConvertsToSeries(t *testing.T) {
	f := newScopeFixture(t)
	tagID := uuid.New()
	existing := newTaggedEvent(t, f.eventID, f.userID, tagID)
	expectLockTarget(f.repo, existing)
	f.tagRepo.EXPECT().FindByIDsForShare(gomock.Any(), []uuid.UUID{tagID}).
		Return([]tag.Tag{newStoredTag(t, tagID, f.userID)}, nil)

	var created rs.RecurringEventSeries
	f.seriesRepo.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, s rs.RecurringEventSeries) (rs.RecurringEventSeries, error) {
			if !s.Rule().Equal(weeklyCount(t, 3)) {
				t.Errorf("rule: got %+v", s.Rule())
			}
			created = s
			return s, nil
		},
	)
	f.repo.EXPECT().Update(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, e domainevent.Event) error {
			if e.ID() != f.eventID {
				t.Errorf("first occurrence ID: got %v, want the target %v", e.ID(), f.eventID)
			}
			if id, ok := e.SeriesID(); !ok || id != created.ID() {
				t.Errorf("SeriesID: got (%v, %v), want (%v, true)", id, ok, created.ID())
			}
			if e.Title() != newTitle || !e.StartAt().Equal(secondOccStart) {
				t.Errorf("content: got %q at %v", e.Title(), e.StartAt())
			}
			assertEventTags(t, e.Tags(), existing.Tags())
			return nil
		},
	)
	f.seriesRepo.EXPECT().CreateOccurrences(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, occs []domainevent.Event) error {
			want := []time.Time{secondOccStart.AddDate(0, 0, 7), secondOccStart.AddDate(0, 0, 14)}
			if len(occs) != len(want) {
				t.Fatalf("occurrences: got %d, want %d", len(occs), len(want))
			}
			for i, e := range occs {
				if !e.StartAt().Equal(want[i]) || !e.EndAt().Equal(want[i].Add(time.Hour)) {
					t.Errorf("occurrence %d: got %v-%v", i, e.StartAt(), e.EndAt())
				}
				if id, ok := e.SeriesID(); !ok || id != created.ID() {
					t.Errorf("occurrence %d SeriesID: got (%v, %v)", i, id, ok)
				}
				if e.ID() == f.eventID || e.Title() != newTitle || e.UserID() != f.userID {
					t.Errorf("occurrence %d: id %v title %q user %v", i, e.ID(), e.Title(), e.UserID())
				}
				assertEventTags(t, e.Tags(), existing.Tags())
			}
			return nil
		},
	)

	in := f.input(secondOccStart)
	in.Recurrence = eventuc.RecurrenceUpdate{Present: true, Spec: weeklyCountInput(3)}
	out, err := f.command().Execute(context.Background(), f.userID, in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Event.ID() != f.eventID {
		t.Errorf("returned event: got %v, want %v", out.Event.ID(), f.eventID)
	}
	if out.Series == nil || out.Series.ID() != created.ID() {
		t.Errorf("Series: got %v, want the created series", out.Series)
	}
	if len(out.Tags) != 1 || out.Tags[0].ID != tagID {
		t.Errorf("Tags: got %+v", out.Tags)
	}
}

func TestUpdateEventCommand_Execute_OneOff_InvalidRecurrence_ReturnsValidationError(t *testing.T) {
	tests := []struct {
		name      string
		spec      *eventuc.RecurrenceInput
		wantField string
		wantCode  string
	}{
		// A new rule still needs 2 occurrences, unlike a rule sent back unchanged.
		{"count 1", weeklyCountInput(1), fieldCount, codeOutOfRange},
		{"count 0", weeklyCountInput(0), fieldCount, codeOutOfRange},
		{
			"timezone empty",
			&eventuc.RecurrenceInput{PatternType: patternWeekly, EndType: "indefinite"},
			fieldTimezone, codeRequired,
		},
		{
			"one occurrence before the until date",
			&eventuc.RecurrenceInput{
				PatternType: patternWeekly, EndType: endUntilDate, UntilDate: ptr("2026-12-01"), Timezone: timezoneUTC,
			},
			"recurrence", codeTooFewOccurrences,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newScopeFixture(t)
			// Format errors fail before the lock; range errors after it.
			f.repo.EXPECT().FindByID(gomock.Any(), f.eventID).Return(f.oneOff(t, secondOccStart), nil).MaxTimes(1)
			f.repo.EXPECT().FindByIDForUpdate(gomock.Any(), f.eventID).Return(f.oneOff(t, secondOccStart), nil).MaxTimes(1)
			in := f.input(secondOccStart)
			in.Recurrence = eventuc.RecurrenceUpdate{Present: true, Spec: tt.spec}

			_, err := f.command().Execute(context.Background(), f.userID, in)
			assertValidationDetail(t, err, tt.wantField, tt.wantCode)
		})
	}
}

func (f *scopeFixture) occurrence(t *testing.T, series rs.RecurringEventSeries) domainevent.Event {
	t.Helper()
	return f.oneOff(t, secondOccStart).WithSeriesID(series.ID())
}

func TestUpdateEventCommand_Execute_Occurrence_ScopeOmitted_ReturnsRequired(t *testing.T) {
	f := newScopeFixture(t)
	series := newSeries(t, f.userID, weeklyCount(t, 3))
	f.expectLockOccurrence(f.occurrence(t, series), series)

	_, err := f.command().Execute(context.Background(), f.userID, f.input(secondOccStart))
	assertValidationDetail(t, err, fieldScope, codeRequired)
}

func TestUpdateEventCommand_Execute_Occurrence_Single_Detaches(t *testing.T) {
	f := newScopeFixture(t)
	series := newSeries(t, f.userID, weeklyCount(t, 3))
	tagID := uuid.New()
	occ := newTaggedEvent(t, f.eventID, f.userID, tagID)
	occ, _ = occ.WithContent("Old title", "", secondOccStart, secondOccStart.Add(time.Hour), "", "")
	occ = occ.WithSeriesID(series.ID())
	f.expectLockOccurrence(occ, series)
	f.tagRepo.EXPECT().FindByIDsForShare(gomock.Any(), []uuid.UUID{tagID}).
		Return([]tag.Tag{newStoredTag(t, tagID, f.userID)}, nil)

	// The detached event moves to the next day, which "single" allows.
	moved := secondOccStart.AddDate(0, 0, 1)
	gomock.InOrder(
		f.repo.EXPECT().Update(gomock.Any(), gomock.Any()).DoAndReturn(
			func(_ context.Context, e domainevent.Event) error {
				if _, ok := e.SeriesID(); ok {
					t.Error("expected the event to be detached")
				}
				if e.Title() != newTitle || !e.StartAt().Equal(moved) {
					t.Errorf("content: got %q at %v", e.Title(), e.StartAt())
				}
				assertEventTags(t, e.Tags(), occ.Tags())
				return nil
			},
		),
		f.seriesRepo.EXPECT().CountOccurrences(gomock.Any(), series.ID()).Return(2, nil),
		f.seriesRepo.EXPECT().Update(gomock.Any(), gomock.Any()).DoAndReturn(
			func(_ context.Context, s rs.RecurringEventSeries) error {
				// The excluded date is the occurrence's stored date, not the new one.
				got := s.ExcludedDates()
				if len(got) != 1 || got[0].Date() != series.LocalDateOf(secondOccStart) {
					t.Errorf("excluded dates: got %v, want [2026-11-30]", got)
				}
				return nil
			},
		),
	)

	in := f.input(moved)
	in.Scope = scopeSingle
	out, err := f.command().Execute(context.Background(), f.userID, in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Series != nil {
		t.Errorf("Series: got %v, want nil", out.Series)
	}
	if _, ok := out.Event.SeriesID(); ok {
		t.Error("returned event: expected it to be detached")
	}
}

func TestUpdateEventCommand_Execute_Occurrence_Single_LastOccurrence_DeletesSeries(t *testing.T) {
	f := newScopeFixture(t)
	series := newSeries(t, f.userID, weeklyCount(t, 3))
	f.expectLockOccurrence(f.occurrence(t, series), series)
	f.repo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)
	f.seriesRepo.EXPECT().CountOccurrences(gomock.Any(), series.ID()).Return(0, nil)
	f.seriesRepo.EXPECT().Delete(gomock.Any(), series.ID()).Return(nil)

	in := f.input(secondOccStart)
	in.Scope = scopeSingle
	if _, err := f.command().Execute(context.Background(), f.userID, in); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUpdateEventCommand_Execute_Occurrence_Single_SameRuleWithCount1_Detaches(t *testing.T) {
	f := newScopeFixture(t)
	// A "this and following" change can leave a series with a count of 1.
	series := newSeries(t, f.userID, weeklyCount(t, 1))
	f.expectLockOccurrence(f.occurrence(t, series), series)
	f.repo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)
	f.seriesRepo.EXPECT().CountOccurrences(gomock.Any(), series.ID()).Return(0, nil)
	f.seriesRepo.EXPECT().Delete(gomock.Any(), series.ID()).Return(nil)

	in := f.input(secondOccStart)
	in.Scope = scopeSingle
	in.Recurrence = eventuc.RecurrenceUpdate{Present: true, Spec: weeklyCountInput(1)}
	if _, err := f.command().Execute(context.Background(), f.userID, in); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUpdateEventCommand_Execute_Occurrence_ScopeNotAllowed(t *testing.T) {
	tests := []struct {
		name       string
		scope      string
		recurrence eventuc.RecurrenceUpdate
	}{
		{"single with another rule", scopeSingle, eventuc.RecurrenceUpdate{Present: true, Spec: weeklyCountInput(4)}},
		{"single with null", scopeSingle, eventuc.RecurrenceUpdate{Present: true}},
		{"all with null", scopeAll, eventuc.RecurrenceUpdate{Present: true}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newScopeFixture(t)
			series := newSeries(t, f.userID, weeklyCount(t, 3))
			f.expectLockOccurrence(f.occurrence(t, series), series)
			in := f.input(secondOccStart)
			in.Scope = tt.scope
			in.Recurrence = tt.recurrence

			_, err := f.command().Execute(context.Background(), f.userID, in)
			assertValidationDetail(t, err, fieldScope, codeScopeNotAllowed)
		})
	}
}

// A concurrent "single" can detach the last occurrence and delete the empty
// series while keeping the event, so a missing series does not mean a missing
// event.
func TestUpdateEventCommand_Execute_Occurrence_SeriesDeletedConcurrently_UpdatesAsOneOff(t *testing.T) {
	f := newScopeFixture(t)
	series := newSeries(t, f.userID, weeklyCount(t, 3))
	gomock.InOrder(
		f.repo.EXPECT().FindByID(gomock.Any(), f.eventID).Return(f.occurrence(t, series), nil),
		f.seriesRepo.EXPECT().FindByIDForUpdate(gomock.Any(), series.ID()).Return(nil, rs.ErrRecurringSeriesNotFound),
		f.repo.EXPECT().FindByIDForUpdate(gomock.Any(), f.eventID).Return(f.oneOff(t, secondOccStart), nil),
		f.repo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil),
	)

	in := f.input(secondOccStart)
	in.Scope = scopeSingle
	out, err := f.command().Execute(context.Background(), f.userID, in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Series != nil {
		t.Errorf("Series: got %v, want nil", out.Series)
	}
}

// Deleting the series with "all" deletes its occurrences too.
func TestUpdateEventCommand_Execute_Occurrence_SeriesAndEventDeleted_ReturnsEventNotFound(t *testing.T) {
	f := newScopeFixture(t)
	series := newSeries(t, f.userID, weeklyCount(t, 3))
	f.repo.EXPECT().FindByID(gomock.Any(), f.eventID).Return(f.occurrence(t, series), nil)
	f.seriesRepo.EXPECT().FindByIDForUpdate(gomock.Any(), series.ID()).Return(nil, rs.ErrRecurringSeriesNotFound)
	f.repo.EXPECT().FindByIDForUpdate(gomock.Any(), f.eventID).Return(nil, domainevent.ErrEventNotFound)

	in := f.input(secondOccStart)
	in.Scope = scopeSingle
	_, err := f.command().Execute(context.Background(), f.userID, in)
	if !errors.Is(err, domainevent.ErrEventNotFound) {
		t.Fatalf("expected ErrEventNotFound, got %v", err)
	}
}

func TestUpdateEventCommand_Execute_Occurrence_DetachedConcurrently_UpdatesAsOneOff(t *testing.T) {
	f := newScopeFixture(t)
	series := newSeries(t, f.userID, weeklyCount(t, 3))
	gomock.InOrder(
		f.repo.EXPECT().FindByID(gomock.Any(), f.eventID).Return(f.occurrence(t, series), nil),
		f.seriesRepo.EXPECT().FindByIDForUpdate(gomock.Any(), series.ID()).Return(series, nil),
		// Another request detached the event before this one locked it.
		f.repo.EXPECT().FindByIDForUpdate(gomock.Any(), f.eventID).Return(f.oneOff(t, secondOccStart), nil),
		f.repo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil),
	)

	in := f.input(secondOccStart)
	in.Scope = scopeSingle
	out, err := f.command().Execute(context.Background(), f.userID, in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Series != nil {
		t.Errorf("Series: got %v, want nil", out.Series)
	}
}
