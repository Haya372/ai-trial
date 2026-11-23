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

// Recurrence-specific constants for this test file.
// patternWeekly, endCount, endUntilDate, timezoneUTC, codeRequired,
// codeInvalidFormat, codeOutOfRange, codeTooFewOccurrences, fieldUntilDateEnd
// are shared from testhelper_test.go.
const (
	patternMonthlyDay = "monthly_day"
	patternMonthlyNth = "monthly_nth_weekday"
	fieldMonthlyNth   = "recurrence.pattern.monthlyNth"
	fieldCount        = "recurrence.end.count"
)

// expectReadBack makes eventRepo.FindByID return the saved occurrence with the
// given ID, standing in for reading the first occurrence back after the bulk
// insert. occs points at the slice the CreateOccurrences stub captured.
func expectReadBack(t *testing.T, eventRepo *eventmock.MockRepository, occs *[]domainevent.Event) {
	t.Helper()
	eventRepo.EXPECT().FindByID(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, id uuid.UUID) (domainevent.Event, error) {
			for _, e := range *occs {
				if e.ID() == id {
					return e, nil
				}
			}
			t.Errorf("FindByID(%v): not a saved occurrence", id)
			return nil, domainevent.ErrEventNotFound
		},
	)
}

// TestCreateEventCommand_Execute_WithRecurrence_Weekly_Count3 creates a weekly
// series with count=3 and verifies that the series and all three occurrences are
// written inside a single transaction, each sharing the same SeriesID.
func TestCreateEventCommand_Execute_WithRecurrence_Weekly_Count3(t *testing.T) {
	ctrl := gomock.NewController(t)
	eventRepo := eventmock.NewMockRepository(ctrl)
	seriesRepo := rsmock.NewMockRepository(ctrl)
	tagRepo := tagmock.NewMockRepository(ctrl)

	userID := uuid.New()
	// 2026-11-23 Monday 10:00 UTC
	startAt := time.Date(2026, 11, 23, 10, 0, 0, 0, time.UTC)
	endAt := startAt.Add(time.Hour)

	var capturedSeries rs.RecurringEventSeries
	var capturedOccs []domainevent.Event

	seriesRepo.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, s rs.RecurringEventSeries) (rs.RecurringEventSeries, error) {
			capturedSeries = s
			return s, nil
		},
	)
	seriesRepo.EXPECT().CreateOccurrences(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, occs []domainevent.Event) error {
			capturedOccs = occs
			return nil
		},
	)
	expectReadBack(t, eventRepo, &capturedOccs)
	// eventRepo.Create must NOT be called for a recurring event
	eventRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Times(0)

	cmd := eventuc.NewCreateEventCommand(eventRepo, seriesRepo, tagRepo, &testutil.StubTxManager{})
	out, err := cmd.Execute(context.Background(), userID, eventuc.CreateEventInput{
		Title:   testCreateTitle,
		StartAt: startAt,
		EndAt:   endAt,
		Recurrence: &eventuc.RecurrenceInput{
			PatternType: patternWeekly,
			EndType:     endCount,
			Count:       ptr(3),
			Timezone:    timezoneUTC,
		},
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Series == nil {
		t.Fatal("expected Series in output, got nil")
	}
	if capturedSeries == nil {
		t.Fatal("expected series passed to repository, got nil")
	}

	// 3 occurrences created
	if len(capturedOccs) != 3 {
		t.Fatalf("expected 3 occurrences, got %d", len(capturedOccs))
	}

	// All share the same SeriesID
	seriesID := out.Series.ID()
	for i, occ := range capturedOccs {
		sid, ok := occ.SeriesID()
		if !ok {
			t.Errorf("occurrence[%d] has no SeriesID", i)
		} else if sid != seriesID {
			t.Errorf("occurrence[%d] SeriesID mismatch: got %v, want %v", i, sid, seriesID)
		}
		if occ.UserID() != userID {
			t.Errorf("occurrence[%d] UserID mismatch", i)
		}
		if occ.Title() != testCreateTitle {
			t.Errorf("occurrence[%d] Title mismatch: got %q", i, occ.Title())
		}
	}

	// First occurrence's startAt equals the request startAt
	if !capturedOccs[0].StartAt().Equal(startAt) {
		t.Errorf("first occurrence startAt mismatch: got %v, want %v", capturedOccs[0].StartAt(), startAt)
	}

	// Second occurrence is 1 week later
	wantSecond := startAt.Add(7 * 24 * time.Hour)
	if !capturedOccs[1].StartAt().Equal(wantSecond) {
		t.Errorf("second occurrence startAt mismatch: got %v, want %v", capturedOccs[1].StartAt(), wantSecond)
	}

	// Output event is the first occurrence
	if !out.Event.StartAt().Equal(startAt) {
		t.Errorf("output event StartAt mismatch: got %v, want %v", out.Event.StartAt(), startAt)
	}
}

// TestCreateEventCommand_Execute_WithRecurrence_InTransaction verifies that
// seriesRepo.Create and seriesRepo.CreateOccurrences are both called inside a tx.
func TestCreateEventCommand_Execute_WithRecurrence_InTransaction(t *testing.T) {
	ctrl := gomock.NewController(t)
	eventRepo := eventmock.NewMockRepository(ctrl)
	seriesRepo := rsmock.NewMockRepository(ctrl)
	tagRepo := tagmock.NewMockRepository(ctrl)

	startAt := time.Date(2026, 11, 23, 10, 0, 0, 0, time.UTC)
	endAt := startAt.Add(time.Hour)
	var capturedOccs []domainevent.Event

	seriesRepo.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, s rs.RecurringEventSeries) (rs.RecurringEventSeries, error) {
			if !inTx(ctx) {
				t.Error("series Create must be called inside a transaction")
			}
			return s, nil
		},
	)
	seriesRepo.EXPECT().CreateOccurrences(gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, occs []domainevent.Event) error {
			if !inTx(ctx) {
				t.Error("CreateOccurrences must be called inside a transaction")
			}
			capturedOccs = occs
			return nil
		},
	)
	expectReadBack(t, eventRepo, &capturedOccs)

	cmd := eventuc.NewCreateEventCommand(eventRepo, seriesRepo, tagRepo, markingTxManager{})
	_, err := cmd.Execute(context.Background(), uuid.New(), eventuc.CreateEventInput{
		Title:   testCreateTitle,
		StartAt: startAt,
		EndAt:   endAt,
		Recurrence: &eventuc.RecurrenceInput{
			PatternType: patternWeekly,
			EndType:     endCount,
			Count:       ptr(3),
			Timezone:    timezoneUTC,
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestCreateEventCommand_Execute_WithRecurrence_WithTags all occurrences receive
// the same tags; a foreign tag returns tagIds/NOT_FOUND without saving the series.
func TestCreateEventCommand_Execute_WithRecurrence_WithTags_AllOccurrencesTagged(t *testing.T) {
	ctrl := gomock.NewController(t)
	eventRepo := eventmock.NewMockRepository(ctrl)
	seriesRepo := rsmock.NewMockRepository(ctrl)
	tagRepo := tagmock.NewMockRepository(ctrl)

	userID := uuid.New()
	tagID := uuid.New()
	startAt := time.Date(2026, 11, 23, 10, 0, 0, 0, time.UTC)
	endAt := startAt.Add(time.Hour)

	tagRepo.EXPECT().FindByIDsForShare(gomock.Any(), []uuid.UUID{tagID}).Return(
		[]tag.Tag{newStoredTag(t, tagID, userID)}, nil,
	)
	seriesRepo.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, s rs.RecurringEventSeries) (rs.RecurringEventSeries, error) { return s, nil },
	)

	var capturedOccs []domainevent.Event
	seriesRepo.EXPECT().CreateOccurrences(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, occs []domainevent.Event) error {
			capturedOccs = occs
			return nil
		},
	)
	expectReadBack(t, eventRepo, &capturedOccs)

	cmd := eventuc.NewCreateEventCommand(eventRepo, seriesRepo, tagRepo, &testutil.StubTxManager{})
	out, err := cmd.Execute(context.Background(), userID, eventuc.CreateEventInput{
		Title:   testCreateTitle,
		StartAt: startAt,
		EndAt:   endAt,
		TagIDs:  []uuid.UUID{tagID},
		Recurrence: &eventuc.RecurrenceInput{
			PatternType: patternWeekly,
			EndType:     endCount,
			Count:       ptr(2),
			Timezone:    timezoneUTC,
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out.Tags) != 1 || out.Tags[0].ID != tagID {
		t.Errorf("expected tag in output, got %v", out.Tags)
	}
	for i, occ := range capturedOccs {
		if len(occ.Tags()) != 1 || occ.Tags()[0].TagID() != tagID {
			t.Errorf("occurrence[%d] missing tag", i)
		}
	}
}

// TestCreateEventCommand_Execute_WithRecurrence_ForeignTag returns NOT_FOUND
// and must not save the series.
func TestCreateEventCommand_Execute_WithRecurrence_ForeignTag_ReturnsNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	eventRepo := eventmock.NewMockRepository(ctrl)
	seriesRepo := rsmock.NewMockRepository(ctrl)
	tagRepo := tagmock.NewMockRepository(ctrl)

	userID := uuid.New()
	foreignTagID := uuid.New()
	startAt := time.Date(2026, 11, 23, 10, 0, 0, 0, time.UTC)
	endAt := startAt.Add(time.Hour)

	tagRepo.EXPECT().FindByIDsForShare(gomock.Any(), []uuid.UUID{foreignTagID}).Return(
		[]tag.Tag{newStoredTag(t, foreignTagID, uuid.New())}, nil, // different owner
	)
	// series Create and CreateOccurrences must NOT be called
	seriesRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Times(0)
	seriesRepo.EXPECT().CreateOccurrences(gomock.Any(), gomock.Any()).Times(0)

	cmd := eventuc.NewCreateEventCommand(eventRepo, seriesRepo, tagRepo, &testutil.StubTxManager{})
	_, err := cmd.Execute(context.Background(), userID, eventuc.CreateEventInput{
		Title:   testCreateTitle,
		StartAt: startAt,
		EndAt:   endAt,
		TagIDs:  []uuid.UUID{foreignTagID},
		Recurrence: &eventuc.RecurrenceInput{
			PatternType: patternWeekly,
			EndType:     endCount,
			Count:       ptr(2),
			Timezone:    timezoneUTC,
		},
	})
	var ve *domain.ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected *domain.ValidationError, got %T: %v", err, err)
	}
	if len(ve.Details) == 0 || ve.Details[0].Field != fieldTagIDs || ve.Details[0].Code != codeNotFound {
		t.Errorf("unexpected validation detail: %v", ve.Details)
	}
}

// TestCreateEventCommand_Execute_NilRecurrence_UsesExistingFlow confirms the
// one-off creation path is unchanged when Recurrence is nil.
func TestCreateEventCommand_Execute_NilRecurrence_UsesExistingFlow(t *testing.T) {
	ctrl := gomock.NewController(t)
	eventRepo := eventmock.NewMockRepository(ctrl)
	seriesRepo := rsmock.NewMockRepository(ctrl)
	tagRepo := tagmock.NewMockRepository(ctrl)

	now := time.Now().UTC()
	eventRepo.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, e domainevent.Event) (domainevent.Event, error) { return e, nil },
	)
	// Series-related calls must NOT happen
	seriesRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Times(0)
	seriesRepo.EXPECT().CreateOccurrences(gomock.Any(), gomock.Any()).Times(0)

	cmd := eventuc.NewCreateEventCommand(eventRepo, seriesRepo, tagRepo, &testutil.StubTxManager{})
	out, err := cmd.Execute(context.Background(), uuid.New(), eventuc.CreateEventInput{
		Title:      testCreateTitle,
		StartAt:    now,
		EndAt:      now.Add(time.Hour),
		Recurrence: nil,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Series != nil {
		t.Errorf("expected nil Series for one-off event, got %v", out.Series)
	}
}

// TestCreateEventCommand_Execute_RecurrenceValidation is a table-driven test
// that verifies the validation order and error codes per the design document.
func TestCreateEventCommand_Execute_RecurrenceValidation(t *testing.T) {
	startAt := time.Date(2026, 11, 23, 10, 0, 0, 0, time.UTC)
	endAt := startAt.Add(time.Hour)

	tests := []struct {
		name      string
		in        eventuc.RecurrenceInput
		wantField string
		wantCode  string
	}{
		{
			name:      "pattern.type empty",
			in:        eventuc.RecurrenceInput{PatternType: "", EndType: endCount, Count: ptr(2), Timezone: timezoneUTC},
			wantField: "recurrence.pattern.type",
			wantCode:  codeRequired,
		},
		{
			name: "pattern.type invalid",
			in: eventuc.RecurrenceInput{
				PatternType: "badpattern", EndType: endCount, Count: ptr(2), Timezone: timezoneUTC,
			},
			wantField: "recurrence.pattern.type",
			wantCode:  codeInvalidFormat,
		},
		{
			name: "monthly_nth_weekday without monthlyNth",
			in: eventuc.RecurrenceInput{
				PatternType: patternMonthlyNth, EndType: endCount, Count: ptr(2), Timezone: timezoneUTC,
			},
			wantField: "recurrence.pattern.monthlyNth",
			wantCode:  codeRequired,
		},
		{
			name:      "end.type empty",
			in:        eventuc.RecurrenceInput{PatternType: patternWeekly, EndType: "", Timezone: timezoneUTC},
			wantField: "recurrence.end.type",
			wantCode:  codeRequired,
		},
		{
			name:      "end.type invalid",
			in:        eventuc.RecurrenceInput{PatternType: patternWeekly, EndType: "badend", Timezone: timezoneUTC},
			wantField: "recurrence.end.type",
			wantCode:  codeInvalidFormat,
		},
		{
			name:      "until_date without untilDate",
			in:        eventuc.RecurrenceInput{PatternType: patternWeekly, EndType: endUntilDate, Timezone: timezoneUTC},
			wantField: fieldUntilDateEnd,
			wantCode:  codeRequired,
		},
		{
			name:      "count end without count",
			in:        eventuc.RecurrenceInput{PatternType: patternWeekly, EndType: endCount, Timezone: timezoneUTC},
			wantField: fieldCount,
			wantCode:  codeRequired,
		},
		{
			name: "untilDate invalid format",
			in: eventuc.RecurrenceInput{
				PatternType: patternWeekly, EndType: endUntilDate,
				UntilDate: ptr("not-a-date"), Timezone: timezoneUTC,
			},
			wantField: fieldUntilDateEnd,
			wantCode:  codeInvalidFormat,
		},
		{
			name: "count less than 2",
			in: eventuc.RecurrenceInput{
				PatternType: patternWeekly, EndType: endCount, Count: ptr(1), Timezone: timezoneUTC,
			},
			wantField: fieldCount,
			wantCode:  codeOutOfRange,
		},
		{
			name:      "timezone empty",
			in:        eventuc.RecurrenceInput{PatternType: patternWeekly, EndType: endCount, Count: ptr(2), Timezone: ""},
			wantField: "recurrence.timezone",
			wantCode:  codeRequired,
		},
		{
			name: "timezone invalid",
			in: eventuc.RecurrenceInput{
				PatternType: patternWeekly, EndType: endCount, Count: ptr(2), Timezone: "Invalid/Zone",
			},
			wantField: "recurrence.timezone",
			wantCode:  codeInvalidFormat,
		},
		{
			name: "untilDate before startAt (OUT_OF_RANGE from ExpandOccurrences)",
			in: eventuc.RecurrenceInput{
				PatternType: patternWeekly,
				EndType:     endUntilDate,
				UntilDate:   ptr("2026-11-22"), // day before startAt
				Timezone:    timezoneUTC,
			},
			wantField: fieldUntilDateEnd,
			wantCode:  codeOutOfRange,
		},
		{
			// Monthly from 2026-11-23: the next candidate, 2026-12-23, is after
			// the until date, so only the first occurrence fits.
			name: "too few occurrences",
			in: eventuc.RecurrenceInput{
				PatternType: patternMonthlyDay,
				EndType:     endUntilDate,
				UntilDate:   ptr("2026-11-30"),
				Timezone:    timezoneUTC,
			},
			wantField: "recurrence",
			wantCode:  codeTooFewOccurrences,
		},
		{
			name: "monthlyNth not in the enum",
			in: eventuc.RecurrenceInput{
				PatternType: patternMonthlyNth, MonthlyNth: ptr("5"),
				EndType: endCount, Count: ptr(2), Timezone: timezoneUTC,
			},
			wantField: fieldMonthlyNth,
			wantCode:  codeInvalidFormat,
		},
		{
			name: "monthlyNth with a pattern other than monthly_nth_weekday",
			in: eventuc.RecurrenceInput{
				PatternType: patternWeekly, MonthlyNth: ptr("1"),
				EndType: endCount, Count: ptr(2), Timezone: timezoneUTC,
			},
			wantField: fieldMonthlyNth,
			wantCode:  codeInvalidFormat,
		},
		{
			// 2026-11-23 is the 4th Monday of November, not the 1st.
			name: "monthlyNth does not fall on the first occurrence",
			in: eventuc.RecurrenceInput{
				PatternType: patternMonthlyNth, MonthlyNth: ptr("1"),
				EndType: endCount, Count: ptr(2), Timezone: timezoneUTC,
			},
			wantField: fieldMonthlyNth,
			wantCode:  codeInvalidFormat,
		},
		{
			name: "untilDate not on the calendar",
			in: eventuc.RecurrenceInput{
				PatternType: patternWeekly, EndType: endUntilDate,
				UntilDate: ptr("2027-02-30"), Timezone: timezoneUTC,
			},
			wantField: fieldUntilDateEnd,
			wantCode:  codeInvalidFormat,
		},
		{
			// The span ends on 2028-11-23.
			name: "untilDate more than 2 years after startAt",
			in: eventuc.RecurrenceInput{
				PatternType: patternWeekly, EndType: endUntilDate,
				UntilDate: ptr("2028-11-24"), Timezone: timezoneUTC,
			},
			wantField: fieldUntilDateEnd,
			wantCode:  codeOutOfRange,
		},
		{
			// About 105 weekly occurrences fit in 2 years.
			name: "count does not fit in 2 years",
			in: eventuc.RecurrenceInput{
				PatternType: patternWeekly, EndType: endCount, Count: ptr(200), Timezone: timezoneUTC,
			},
			wantField: fieldCount,
			wantCode:  codeOutOfRange,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			eventRepo := eventmock.NewMockRepository(ctrl)
			seriesRepo := rsmock.NewMockRepository(ctrl)
			tagRepo := tagmock.NewMockRepository(ctrl)

			// No repository calls should happen during validation
			eventRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Times(0)
			seriesRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Times(0)
			seriesRepo.EXPECT().CreateOccurrences(gomock.Any(), gomock.Any()).Times(0)

			cmd := eventuc.NewCreateEventCommand(eventRepo, seriesRepo, tagRepo, &testutil.StubTxManager{})
			_, err := cmd.Execute(context.Background(), uuid.New(), eventuc.CreateEventInput{
				Title:      testCreateTitle,
				StartAt:    startAt,
				EndAt:      endAt,
				Recurrence: &tt.in,
			})
			if err == nil {
				t.Fatal("expected validation error, got nil")
			}
			var ve *domain.ValidationError
			if !errors.As(err, &ve) {
				t.Fatalf("expected *domain.ValidationError, got %T: %v", err, err)
			}
			if len(ve.Details) == 0 {
				t.Fatal("expected at least one validation detail")
			}
			d := ve.Details[0]
			if d.Field != tt.wantField {
				t.Errorf("field mismatch: got %q, want %q", d.Field, tt.wantField)
			}
			if d.Code != tt.wantCode {
				t.Errorf("code mismatch: got %q, want %q", d.Code, tt.wantCode)
			}
		})
	}
}

// TestCreateEventCommand_Execute_SeriesRepoError_PropagatesError verifies that
// errors from seriesRepo.Create are propagated to the caller.
func TestCreateEventCommand_Execute_SeriesRepoError_PropagatesError(t *testing.T) {
	ctrl := gomock.NewController(t)
	eventRepo := eventmock.NewMockRepository(ctrl)
	seriesRepo := rsmock.NewMockRepository(ctrl)
	tagRepo := tagmock.NewMockRepository(ctrl)

	startAt := time.Date(2026, 11, 23, 10, 0, 0, 0, time.UTC)
	endAt := startAt.Add(time.Hour)

	seriesRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil, errDBFailure)

	cmd := eventuc.NewCreateEventCommand(eventRepo, seriesRepo, tagRepo, &testutil.StubTxManager{})
	_, err := cmd.Execute(context.Background(), uuid.New(), eventuc.CreateEventInput{
		Title:   testCreateTitle,
		StartAt: startAt,
		EndAt:   endAt,
		Recurrence: &eventuc.RecurrenceInput{
			PatternType: patternWeekly,
			EndType:     endCount,
			Count:       ptr(2),
			Timezone:    timezoneUTC,
		},
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, errDBFailure) {
		t.Errorf("expected wrapped errDBFailure, got %v", err)
	}
}

// TestCreateEventCommand_Execute_CreateOccurrencesError_PropagatesError verifies that
// errors from seriesRepo.CreateOccurrences are propagated to the caller.
func TestCreateEventCommand_Execute_CreateOccurrencesError_PropagatesError(t *testing.T) {
	ctrl := gomock.NewController(t)
	eventRepo := eventmock.NewMockRepository(ctrl)
	seriesRepo := rsmock.NewMockRepository(ctrl)
	tagRepo := tagmock.NewMockRepository(ctrl)

	startAt := time.Date(2026, 11, 23, 10, 0, 0, 0, time.UTC)
	endAt := startAt.Add(time.Hour)

	seriesRepo.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, s rs.RecurringEventSeries) (rs.RecurringEventSeries, error) { return s, nil },
	)
	seriesRepo.EXPECT().CreateOccurrences(gomock.Any(), gomock.Any()).Return(errDBFailure)

	cmd := eventuc.NewCreateEventCommand(eventRepo, seriesRepo, tagRepo, &testutil.StubTxManager{})
	_, err := cmd.Execute(context.Background(), uuid.New(), eventuc.CreateEventInput{
		Title:   testCreateTitle,
		StartAt: startAt,
		EndAt:   endAt,
		Recurrence: &eventuc.RecurrenceInput{
			PatternType: patternWeekly,
			EndType:     endCount,
			Count:       ptr(2),
			Timezone:    timezoneUTC,
		},
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, errDBFailure) {
		t.Errorf("expected wrapped errDBFailure, got %v", err)
	}
}

// TestCreateEventCommand_Execute_WithRecurrence_ReturnsStoredFirstOccurrence
// verifies that the output is the first occurrence as read back after saving,
// so a sub-microsecond startAt matches what later reads return.
func TestCreateEventCommand_Execute_WithRecurrence_ReturnsStoredFirstOccurrence(t *testing.T) {
	ctrl := gomock.NewController(t)
	eventRepo := eventmock.NewMockRepository(ctrl)
	seriesRepo := rsmock.NewMockRepository(ctrl)
	tagRepo := tagmock.NewMockRepository(ctrl)

	userID := uuid.New()
	startAt := time.Date(2026, 11, 23, 10, 0, 0, 123456789, time.UTC)
	endAt := startAt.Add(time.Hour)
	storedStart := startAt.Truncate(time.Microsecond)

	var firstID uuid.UUID
	seriesRepo.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, s rs.RecurringEventSeries) (rs.RecurringEventSeries, error) { return s, nil },
	)
	seriesRepo.EXPECT().CreateOccurrences(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, occs []domainevent.Event) error {
			firstID = occs[0].ID()
			return nil
		},
	)
	eventRepo.EXPECT().FindByID(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, id uuid.UUID) (domainevent.Event, error) {
			if id != firstID {
				t.Errorf("FindByID id = %v, want first occurrence %v", id, firstID)
			}
			return domainevent.New(id, userID, testCreateTitle, "", storedStart, storedStart.Add(time.Hour), "", "")
		},
	)

	cmd := eventuc.NewCreateEventCommand(eventRepo, seriesRepo, tagRepo, &testutil.StubTxManager{})
	out, err := cmd.Execute(context.Background(), userID, eventuc.CreateEventInput{
		Title:   testCreateTitle,
		StartAt: startAt,
		EndAt:   endAt,
		Recurrence: &eventuc.RecurrenceInput{
			PatternType: patternWeekly, EndType: endCount, Count: ptr(3), Timezone: timezoneUTC,
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !out.Event.StartAt().Equal(storedStart) {
		t.Errorf("output StartAt = %v, want stored %v", out.Event.StartAt(), storedStart)
	}
}
