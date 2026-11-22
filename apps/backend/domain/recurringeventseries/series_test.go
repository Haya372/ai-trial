package recurringeventseries_test

import (
	"errors"
	"slices"
	"testing"
	"time"

	"github.com/google/uuid"

	res "github.com/Haya372/ai-trial/backend/domain/recurringeventseries"
)

const dec7 = "2026-12-07"

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
	dates := s.ExcludedDates()
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

	cases := map[string]struct {
		starts   []time.Time
		excluded []res.ExcludedDate
		target   time.Time
		want     int
	}{
		"4th of 10":        {starts, nil, starts[3], 7},
		"first":            {starts, nil, starts[0], 10},
		"last":             {starts, nil, starts[9], 1},
		"earlier one gone": {slices.Delete(slices.Clone(starts), 1, 2), excludedOn(t, dateOf(1)), starts[3], 7},
		"later one gone":   {slices.Delete(slices.Clone(starts), 5, 6), excludedOn(t, dateOf(5)), starts[3], 7},
		"last one gone":    {slices.Delete(slices.Clone(starts), 9, 10), excludedOn(t, dateOf(9)), starts[8], 2},
		"unordered starts": {[]time.Time{starts[5], starts[0], starts[3], starts[4]}, nil, starts[3], 3},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			s := mustSeries(t, rule, c.excluded...)
			got, err := s.FollowingEnd(c.starts, c.target)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if n, ok := got.Count(); !ok || n != c.want {
				t.Errorf("got (%d, %v), want (%d, true)", n, ok, c.want)
			}
		})
	}
}

func TestFollowingEnd_keepsUntilDateEnd(t *testing.T) {
	start := localAt(t, tokyo, 2026, 12, 7, 10, 0)
	next := localAt(t, tokyo, 2026, 12, 14, 10, 0)
	end := until(t, 2027, 6, 30)
	s := mustSeries(t, mustRule(t, weekly, nil, end, tokyo))

	got, err := s.FollowingEnd([]time.Time{start, next}, start)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != end {
		t.Errorf("got %+v, want %+v", got, end)
	}
}

// Carrying an indefinite end over would restart its 2-year window at the
// target and add occurrences the calendar did not show before the change.
func TestFollowingEnd_indefiniteEndsOnTheLastOccurrence(t *testing.T) {
	s := mustSeries(t, mustRule(t, daily, nil, res.NewIndefiniteEnd(), tokyo))
	target := localAt(t, tokyo, 2027, 6, 1, 10, 0)
	// 00:30 JST on the last day is the previous day in UTC.
	last := localAt(t, tokyo, 2028, 1, 1, 0, 30)
	starts := []time.Time{last, localAt(t, tokyo, 2026, 1, 1, 10, 0), target}

	got, err := s.FollowingEnd(starts, target)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d, ok := got.UntilDate(); !ok || d.String() != "2028-01-01" {
		t.Errorf("got %+v, want until 2028-01-01", got)
	}
}

// An until date that leaves one candidate would fail ExpandOccurrences'
// minimum of 2, which only user-entered rules must meet.
func TestFollowingEnd_indefiniteWithOneOccurrenceLeftBecomesCountOne(t *testing.T) {
	s := mustSeries(t, mustRule(t, weekly, nil, res.NewIndefiniteEnd(), tokyo))
	prev := localAt(t, tokyo, 2026, 12, 28, 10, 0)
	last := localAt(t, tokyo, 2027, 1, 4, 10, 0)

	got, err := s.FollowingEnd([]time.Time{prev, last}, last)
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

	got, err := s.FollowingEnd([]time.Time{prev, last}, last)
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
			if _, err := s.FollowingEnd([]time.Time{first, second}, target); err == nil {
				t.Errorf("%s, %s end: expected an error, got nil", name, s.Rule().End.Type())
			}
		}
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
