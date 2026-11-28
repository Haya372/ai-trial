package recurringeventseries_test

import (
	"errors"
	"slices"
	"testing"
	"time"

	"github.com/google/uuid"

	res "github.com/Haya372/ai-trial/backend/domain/recurringeventseries"
)

const (
	dec7  = "2026-12-07"
	dec28 = "2026-12-28"
)

func excludedOn(t *testing.T, dates ...string) []res.ExcludedDate {
	t.Helper()
	out := make([]res.ExcludedDate, len(dates))
	for i, s := range dates {
		d, err := time.Parse(time.DateOnly, s)
		if err != nil {
			t.Fatal(err)
		}
		out[i] = res.NewExcludedDate(mustDate(t, d.Year(), d.Month(), d.Day()))
	}
	return out
}

func excludedStrings(s res.RecurringEventSeries) []string {
	return dateStrings(s.ExcludedDates())
}

// dateStrings returns nil for no dates, so it compares equal to a nil want.
func dateStrings(dates []res.ExcludedDate) []string {
	if len(dates) == 0 {
		return nil
	}
	out := make([]string, len(dates))
	for i, e := range dates {
		out[i] = e.Date().String()
	}
	return out
}

func mustSeries(t *testing.T, rule res.RecurrenceRule, excluded ...res.ExcludedDate) res.RecurringEventSeries {
	t.Helper()
	s, err := res.New(uuid.New(), uuid.New(), rule, excluded)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return s
}

func TestNew_keepsFieldsAndSortsExcludedDates(t *testing.T) {
	id, userID := uuid.New(), uuid.New()
	rule := mustRule(t, weekly, nil, mustCount(t, 10), tokyo)

	s, err := res.New(id, userID, rule, excludedOn(t, "2026-12-14", dec7))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if s.ID() != id || s.UserID() != userID {
		t.Errorf("ID/UserID: got (%v, %v), want (%v, %v)", s.ID(), s.UserID(), id, userID)
	}
	if s.Rule() != rule {
		t.Errorf("Rule: got %+v, want %+v", s.Rule(), rule)
	}
	if got, want := excludedStrings(s), []string{dec7, "2026-12-14"}; !slices.Equal(got, want) {
		t.Errorf("ExcludedDates: got %v, want %v", got, want)
	}
}

func TestNew_rejectsDuplicatedExcludedDates(t *testing.T) {
	rule := mustRule(t, weekly, nil, res.NewIndefiniteEnd(), tokyo)

	_, err := res.New(uuid.New(), uuid.New(), rule, excludedOn(t, dec7, dec7))
	if !errors.Is(err, res.ErrRecurringExcludedDateDuplicated) {
		t.Errorf("got %v, want ErrRecurringExcludedDateDuplicated", err)
	}
}

func TestNew_rejectsUninitializedRule(t *testing.T) {
	valid := mustRule(t, weekly, nil, res.NewIndefiniteEnd(), tokyo)
	for name, rule := range map[string]res.RecurrenceRule{
		"pattern":  {End: valid.End, Timezone: valid.Timezone},
		"end":      {Pattern: valid.Pattern, Timezone: valid.Timezone},
		"timezone": {Pattern: valid.Pattern, End: valid.End},
	} {
		_, err := res.New(uuid.New(), uuid.New(), rule, nil)
		if !errors.Is(err, res.ErrRecurrenceRuleUninitialized) {
			t.Errorf("%s: got %v, want ErrRecurrenceRuleUninitialized", name, err)
		}
	}
}

func TestExcludedDates_returnsCopy(t *testing.T) {
	s := mustSeries(t, mustRule(t, weekly, nil, res.NewIndefiniteEnd(), tokyo), excludedOn(t, dec7)...)

	got := s.ExcludedDates()
	got[0] = excludedOn(t, "2027-01-04")[0]

	if excludedStrings(s)[0] != dec7 {
		t.Errorf("mutating the returned slice changed the series")
	}
}

func TestLocalDateOf_usesSeriesTimezone(t *testing.T) {
	s := mustSeries(t, mustRule(t, daily, nil, res.NewIndefiniteEnd(), tokyo))

	// 15:00 UTC is midnight of the next day in Tokyo.
	got := s.LocalDateOf(time.Date(2026, 12, 6, 15, 0, 0, 0, time.UTC))

	if got.String() != dec7 {
		t.Errorf("got %s, want 2026-12-07", got)
	}
}

func TestExcludeOccurrence_addsLocalDateAndKeepsOriginal(t *testing.T) {
	s := mustSeries(t, mustRule(t, daily, nil, res.NewIndefiniteEnd(), tokyo), excludedOn(t, "2026-12-10")...)

	got, err := s.ExcludeOccurrence(time.Date(2026, 12, 6, 15, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if want := []string{dec7, "2026-12-10"}; !slices.Equal(excludedStrings(got), want) {
		t.Errorf("ExcludedDates: got %v, want %v", excludedStrings(got), want)
	}
	if got.ID() != s.ID() || got.Rule() != s.Rule() {
		t.Errorf("ExcludeOccurrence changed other fields")
	}
	if want := []string{"2026-12-10"}; !slices.Equal(excludedStrings(s), want) {
		t.Errorf("original series was mutated: got %v", excludedStrings(s))
	}
}

func TestExcludeOccurrence_rejectsDuplicate(t *testing.T) {
	s := mustSeries(t, mustRule(t, daily, nil, res.NewIndefiniteEnd(), tokyo), excludedOn(t, dec7)...)

	_, err := s.ExcludeOccurrence(localAt(t, tokyo, 2026, 12, 7, 10, 0))
	if !errors.Is(err, res.ErrRecurringExcludedDateDuplicated) {
		t.Errorf("got %v, want ErrRecurringExcludedDateDuplicated", err)
	}
}

func TestEndBefore_switchesToUntilTheDayBeforeAndReturnsLaterExcludedDates(t *testing.T) {
	for name, end := range map[string]res.RecurrenceEnd{
		"count":      mustCount(t, 10),
		"indefinite": res.NewIndefiniteEnd(),
		"until_date": until(t, 2027, 6, 30),
	} {
		t.Run(name, func(t *testing.T) {
			s := mustSeries(t, mustRule(t, weekly, nil, end, tokyo),
				excludedOn(t, "2026-11-23", "2026-11-30", dec7)...)

			// Midnight of Nov 30 in Tokyo is still Nov 29 in UTC.
			got, dropped := s.EndBefore(localAt(t, tokyo, 2026, 11, 30, 0, 0))

			d, ok := got.Rule().End.UntilDate()
			if !ok || d.String() != "2026-11-29" {
				t.Errorf("End: got (%v, %v), want until 2026-11-29", d, ok)
			}
			if _, ok := got.Rule().End.Count(); ok {
				t.Errorf("End still has a count")
			}
			if want := []string{"2026-11-23"}; !slices.Equal(excludedStrings(got), want) {
				t.Errorf("ExcludedDates: got %v, want %v", excludedStrings(got), want)
			}
			droppedDates := make([]string, len(dropped))
			for i, e := range dropped {
				droppedDates[i] = e.Date().String()
			}
			if want := []string{"2026-11-30", dec7}; !slices.Equal(droppedDates, want) {
				t.Errorf("dropped: got %v, want %v", droppedDates, want)
			}
			if got.Rule().Pattern != s.Rule().Pattern || got.Rule().Timezone != s.Rule().Timezone {
				t.Errorf("EndBefore changed the pattern or timezone")
			}
			if len(s.ExcludedDates()) != 3 || s.Rule().End != end {
				t.Errorf("original series was mutated")
			}
		})
	}
}

func TestEndBefore_crossesMonthBoundary(t *testing.T) {
	s := mustSeries(t, mustRule(t, monthlyDay, nil, res.NewIndefiniteEnd(), tokyo))

	got, _ := s.EndBefore(localAt(t, tokyo, 2027, 3, 1, 9, 0))

	if d, _ := got.Rule().End.UntilDate(); d.String() != "2027-02-28" {
		t.Errorf("got until %s, want 2027-02-28", d)
	}
}

// A count end counts excluded dates, so the remaining count adds the excluded
// dates the new series takes over to the occurrences from the target on.
func TestFollowingEnd_countIsRemainingCandidatesFromTarget(t *testing.T) {
	rule := mustRule(t, weekly, nil, mustCount(t, 10), tokyo)
	starts := make([]time.Time, 10)
	for i := range starts {
		starts[i] = localAt(t, tokyo, 2026, 12, 7+7*i, 10, 0)
	}
	dateOf := func(i int) string { return starts[i].In(rule.Timezone.Location()).Format(time.DateOnly) }
	gone := func(i int) []time.Time { return slices.Delete(slices.Clone(starts), i, i+1) }

	cases := map[string]struct {
		starts   []time.Time
		excluded []res.ExcludedDate
		target   time.Time
		want     int
		// wantOver is the excluded dates the new series takes over.
		wantOver []string
	}{
		"4th of 10":        {starts, nil, starts[3], 7, nil},
		"first":            {starts, nil, starts[0], 10, nil},
		"last":             {starts, nil, starts[9], 1, nil},
		"earlier one gone": {gone(1), excludedOn(t, dateOf(1)), starts[3], 7, nil},
		"later one gone":   {gone(5), excludedOn(t, dateOf(5)), starts[3], 7, []string{dateOf(5)}},
		"last one gone":    {gone(9), excludedOn(t, dateOf(9)), starts[8], 2, []string{dateOf(9)}},
		"unordered starts": {[]time.Time{starts[5], starts[0], starts[3], starts[4]}, nil, starts[3], 3, nil},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			s := mustSeries(t, rule, c.excluded...)
			got, excluded, err := s.FollowingEnd(c.starts, c.target)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if n, ok := got.Count(); !ok || n != c.want {
				t.Errorf("got (%d, %v), want (%d, true)", n, ok, c.want)
			}
			if over := dateStrings(excluded); !slices.Equal(over, c.wantOver) {
				t.Errorf("got excluded %v, want %v", over, c.wantOver)
			}
		})
	}
}

func TestFollowingEnd_keepsUntilDateEnd(t *testing.T) {
	start := localAt(t, tokyo, 2026, 12, 7, 10, 0)
	next := localAt(t, tokyo, 2026, 12, 14, 10, 0)
	end := until(t, 2027, 6, 30)
	s := mustSeries(t, mustRule(t, weekly, nil, end, tokyo), excludedOn(t, dec28)...)

	got, excluded, err := s.FollowingEnd([]time.Time{start, next}, start)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != end {
		t.Errorf("got %+v, want %+v", got, end)
	}
	if over, want := dateStrings(excluded), []string{dec28}; !slices.Equal(over, want) {
		t.Errorf("got excluded %v, want %v", over, want)
	}
}

// Carrying an indefinite end over would restart its 2-year window at the
// target and add occurrences the calendar did not show before the change.
func TestFollowingEnd_indefiniteEndsOnTheLastOccurrence(t *testing.T) {
	s := mustSeries(t, mustRule(t, daily, nil, res.NewIndefiniteEnd(), tokyo),
		excludedOn(t, "2027-05-31", "2027-12-31")...)
	target := localAt(t, tokyo, 2027, 6, 1, 10, 0)
	// 00:30 JST on the last day is the previous day in UTC.
	last := localAt(t, tokyo, 2028, 1, 1, 0, 30)
	starts := []time.Time{last, localAt(t, tokyo, 2026, 1, 1, 10, 0), target}

	got, excluded, err := s.FollowingEnd(starts, target)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d, ok := got.UntilDate(); !ok || d.String() != "2028-01-01" {
		t.Errorf("got %+v, want until 2028-01-01", got)
	}
	if over, want := dateStrings(excluded), []string{"2027-12-31"}; !slices.Equal(over, want) {
		t.Errorf("got excluded %v, want %v", over, want)
	}
}

// An until date that leaves one candidate would fail ExpandOccurrences'
// minimum of 2, which only user-entered rules must meet.
func TestFollowingEnd_indefiniteWithOneOccurrenceLeftBecomesCountOne(t *testing.T) {
	s := mustSeries(t, mustRule(t, weekly, nil, res.NewIndefiniteEnd(), tokyo))
	prev := localAt(t, tokyo, 2026, 12, 28, 10, 0)
	last := localAt(t, tokyo, 2027, 1, 4, 10, 0)

	got, _, err := s.FollowingEnd([]time.Time{prev, last}, last)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n, ok := got.Count(); !ok || n != 1 {
		t.Errorf("got %+v, want count 1", got)
	}
}

// An until date that leaves one candidate would fail ExpandOccurrences'
// minimum of 2, which only user-entered rules must meet.
func TestFollowingEnd_untilDateWithOneOccurrenceLeftBecomesCountOne(t *testing.T) {
	end := until(t, 2027, 1, 4)
	s := mustSeries(t, mustRule(t, weekly, nil, end, tokyo))
	prev := localAt(t, tokyo, 2026, 12, 28, 10, 0)
	last := localAt(t, tokyo, 2027, 1, 4, 10, 0)

	got, _, err := s.FollowingEnd([]time.Time{prev, last}, last)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n, ok := got.Count(); !ok || n != 1 {
		t.Fatalf("got %+v, want count 1", got)
	}
	occs, err := res.ExpandOccurrences(last, last.Add(time.Hour), res.RecurrenceRule{
		Pattern: s.Rule().Pattern, End: got, Timezone: s.Rule().Timezone,
	}, nil)
	if err != nil || len(occs) != 1 {
		t.Errorf("ExpandOccurrences: got %d occurrences, err %v; want 1, nil", len(occs), err)
	}
}

func TestFollowingEnd_takesOverExcludedDatesFromTheTargetOn(t *testing.T) {
	s := mustSeries(t, mustRule(t, weekly, nil, mustCount(t, 10), tokyo),
		excludedOn(t, "2026-12-14", dec28, "2027-01-11")...)
	target := localAt(t, tokyo, 2026, 12, 21, 10, 0)
	starts := []time.Time{localAt(t, tokyo, 2026, 12, 7, 10, 0), target, localAt(t, tokyo, 2027, 1, 4, 10, 0)}

	_, excluded, err := s.FollowingEnd(starts, target)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := dateStrings(excluded)
	if want := []string{dec28, "2027-01-11"}; !slices.Equal(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestFollowingEnd_rejectsTargetThatIsNotAnOccurrence(t *testing.T) {
	count := mustSeries(t, mustRule(t, weekly, nil, mustCount(t, 10), tokyo))
	indefinite := mustSeries(t, mustRule(t, weekly, nil, res.NewIndefiniteEnd(), tokyo))
	first := localAt(t, tokyo, 2026, 12, 7, 10, 0)
	second := localAt(t, tokyo, 2026, 12, 14, 10, 0)

	for name, target := range map[string]time.Time{
		"between occurrences":    first.Add(30 * time.Minute),
		"after every occurrence": second.Add(time.Minute),
	} {
		for _, s := range []res.RecurringEventSeries{count, indefinite} {
			if _, _, err := s.FollowingEnd([]time.Time{first, second}, target); err == nil {
				t.Errorf("%s, %s end: expected an error, got nil", name, s.Rule().End.Type())
			}
		}
	}
}

// The local date is judged in the series' timezone, where the weekday and
// day of the month of the series come from.
func TestCheckDateKept(t *testing.T) {
	s := mustSeries(t, mustRule(t, weekly, nil, res.NewIndefiniteEnd(), tokyo))
	target := localAt(t, tokyo, 2026, 12, 21, 10, 0)

	for name, c := range map[string]struct {
		newStart time.Time
		wantErr  bool
	}{
		"same instant":                   {target, false},
		"another time on the same date":  {localAt(t, tokyo, 2026, 12, 21, 23, 30), false},
		"the next date":                  {localAt(t, tokyo, 2026, 12, 22, 10, 0), true},
		"the previous date":              {localAt(t, tokyo, 2026, 12, 20, 10, 0), true},
		"same UTC date, next local date": {localAt(t, tokyo, 2026, 12, 22, 0, 30), true},
	} {
		t.Run(name, func(t *testing.T) {
			err := s.CheckDateKept(target, c.newStart)
			if c.wantErr != errors.Is(err, res.ErrRecurringOccurrenceDateChangeNotAllowed) || (!c.wantErr && err != nil) {
				t.Errorf("got %v, want error %v", err, c.wantErr)
			}
		})
	}
}

func TestRestoreCountEnd(t *testing.T) {
	e, err := res.RestoreCountEnd(1)
	if err != nil {
		t.Fatalf("RestoreCountEnd(1): unexpected error: %v", err)
	}
	if n, ok := e.Count(); !ok || n != 1 {
		t.Errorf("got (%d, %v), want (1, true)", n, ok)
	}

	if _, err := res.RestoreCountEnd(0); !errors.Is(err, res.ErrRecurringEndInvalid) {
		t.Errorf("RestoreCountEnd(0): got %v, want ErrRecurringEndInvalid", err)
	}
}

func TestMoveOccurrences_keepsEachDateAtTheNewTimeAndDuration(t *testing.T) {
	s := mustSeries(t, mustRule(t, weekly, nil, mustCount(t, 3), tokyo))
	starts := []time.Time{
		localAt(t, tokyo, 2026, 12, 7, 10, 0),
		localAt(t, tokyo, 2026, 12, 14, 10, 0),
		localAt(t, tokyo, 2026, 12, 21, 10, 0),
	}
	newStart := localAt(t, tokyo, 2026, 12, 14, 13, 30)
	newEnd := localAt(t, tokyo, 2026, 12, 14, 15, 0)

	got, err := s.MoveOccurrences(starts, starts[1], newStart, newEnd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []time.Time{
		localAt(t, tokyo, 2026, 12, 7, 13, 30),
		newStart,
		localAt(t, tokyo, 2026, 12, 21, 13, 30),
	}
	if len(got) != len(want) {
		t.Fatalf("got %d occurrences, want %d", len(got), len(want))
	}
	for i, w := range want {
		if !got[i].StartAt.Equal(w) || !got[i].EndAt.Equal(w.Add(90*time.Minute)) {
			t.Errorf("occurrence %d: got [%v, %v], want [%v, %v]", i, got[i].StartAt, got[i].EndAt, w, w.Add(90*time.Minute))
		}
	}
}

// Adding the UTC difference would put the occurrences after the DST change
// an hour off; each one keeps the wall-clock time instead.
func TestMoveOccurrences_keepsWallClockAcrossDST(t *testing.T) {
	s := mustSeries(t, mustRule(t, weekly, nil, mustCount(t, 2), newYork))
	before := localAt(t, newYork, 2026, time.March, 2, 10, 0)
	after := localAt(t, newYork, 2026, time.March, 9, 10, 0)
	newStart := localAt(t, newYork, 2026, time.March, 2, 11, 0)

	got, err := s.MoveOccurrences([]time.Time{before, after}, before, newStart, newStart.Add(time.Hour))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if want := localAt(t, newYork, 2026, time.March, 9, 11, 0); !got[1].StartAt.Equal(want) {
		t.Errorf("got %v, want %v", got[1].StartAt, want)
	}
}

func TestMoveOccurrences_rejectsTargetThatIsNotAnOccurrence(t *testing.T) {
	s := mustSeries(t, mustRule(t, weekly, nil, mustCount(t, 2), tokyo))
	first := localAt(t, tokyo, 2026, 12, 7, 10, 0)

	if _, err := s.MoveOccurrences([]time.Time{first}, first.Add(time.Minute), first, first.Add(time.Hour)); err == nil {
		t.Error("expected an error, got nil")
	}
}

func TestRegeneratedFirstStart(t *testing.T) {
	tz, err := res.NewSeriesTimezone(tokyo)
	if err != nil {
		t.Fatal(err)
	}
	first := localAt(t, tokyo, 2026, 12, 7, 10, 0)
	fifth := localAt(t, tokyo, 2027, 1, 4, 10, 0)

	for name, c := range map[string]struct {
		target, newStart, want time.Time
	}{
		"the target is the first": {first, localAt(t, tokyo, 2026, 12, 8, 9, 0), localAt(t, tokyo, 2026, 12, 8, 9, 0)},
		"only the time changes":   {fifth, localAt(t, tokyo, 2027, 1, 4, 14, 0), localAt(t, tokyo, 2026, 12, 7, 14, 0)},
		"two days later":          {fifth, localAt(t, tokyo, 2027, 1, 6, 9, 0), localAt(t, tokyo, 2026, 12, 9, 9, 0)},
		"a day earlier":           {fifth, localAt(t, tokyo, 2027, 1, 3, 9, 0), localAt(t, tokyo, 2026, 12, 6, 9, 0)},
		// 00:30 JST on 2027-01-05 is 2027-01-04 in UTC.
		"the next local date, same UTC date": {fifth, localAt(t, tokyo, 2027, 1, 5, 0, 30), localAt(t, tokyo, 2026, 12, 8, 0, 30)},
	} {
		t.Run(name, func(t *testing.T) {
			got, err := res.RegeneratedFirstStart(first, c.target, c.newStart, tz)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !got.Equal(c.want) {
				t.Errorf("got %v, want %v", got, c.want)
			}
		})
	}
}

func TestRegeneratedFirstStart_crossesMonthAndYear(t *testing.T) {
	tz, err := res.NewSeriesTimezone(tokyo)
	if err != nil {
		t.Fatal(err)
	}
	first := localAt(t, tokyo, 2026, 12, 30, 10, 0)
	target := localAt(t, tokyo, 2027, 2, 24, 10, 0)

	got, err := res.RegeneratedFirstStart(first, target, localAt(t, tokyo, 2027, 3, 1, 10, 0), tz)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if want := localAt(t, tokyo, 2027, 1, 4, 10, 0); !got.Equal(want) {
		t.Errorf("got %v, want %v", got, want)
	}
}
