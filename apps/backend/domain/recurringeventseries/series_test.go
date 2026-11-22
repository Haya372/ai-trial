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

func TestEndBefore_switchesToUntilTheDayBeforeAndDropsLaterExcludedDates(t *testing.T) {
	for name, end := range map[string]res.RecurrenceEnd{
		"count":      mustCount(t, 10),
		"indefinite": res.NewIndefiniteEnd(),
		"until_date": until(t, 2027, 6, 30),
	} {
		t.Run(name, func(t *testing.T) {
			s := mustSeries(t, mustRule(t, weekly, nil, end, tokyo),
				excludedOn(t, "2026-11-23", "2026-11-30", dec7)...)

			// Midnight of Nov 30 in Tokyo is still Nov 29 in UTC.
			got := s.EndBefore(localAt(t, tokyo, 2026, 11, 30, 0, 0))

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

	got := s.EndBefore(localAt(t, tokyo, 2027, 3, 1, 9, 0))

	if d, _ := got.Rule().End.UntilDate(); d.String() != "2027-02-28" {
		t.Errorf("got until %s, want 2027-02-28", d)
	}
}

func TestFollowingEnd_countIsRemainingOccurrencesFromTarget(t *testing.T) {
	s := mustSeries(t, mustRule(t, weekly, nil, mustCount(t, 10), tokyo))
	starts := make([]time.Time, 10)
	for i := range starts {
		starts[i] = localAt(t, tokyo, 2026, 12, 7+7*i, 10, 0)
	}

	cases := map[string]struct {
		starts []time.Time
		target time.Time
		want   int
	}{
		"4th of 10":        {starts, starts[3], 7},
		"first":            {starts, starts[0], 10},
		"last":             {starts, starts[9], 1},
		"earlier one gone": {slices.Delete(slices.Clone(starts), 1, 2), starts[3], 7},
		"later one gone":   {slices.Delete(slices.Clone(starts), 5, 6), starts[3], 6},
		"unordered starts": {[]time.Time{starts[5], starts[0], starts[3], starts[4]}, starts[3], 3},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
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

func TestFollowingEnd_keepsNonCountEnd(t *testing.T) {
	start := localAt(t, tokyo, 2026, 12, 7, 10, 0)
	for name, end := range map[string]res.RecurrenceEnd{
		"indefinite": res.NewIndefiniteEnd(),
		"until_date": until(t, 2027, 6, 30),
	} {
		s := mustSeries(t, mustRule(t, weekly, nil, end, tokyo))

		got, err := s.FollowingEnd([]time.Time{start}, start)
		if err != nil {
			t.Fatalf("%s: unexpected error: %v", name, err)
		}
		if got != end {
			t.Errorf("%s: got %+v, want %+v", name, got, end)
		}
	}
}

func TestFollowingEnd_rejectsTargetAfterEveryOccurrence(t *testing.T) {
	s := mustSeries(t, mustRule(t, weekly, nil, mustCount(t, 10), tokyo))
	start := localAt(t, tokyo, 2026, 12, 7, 10, 0)

	_, err := s.FollowingEnd([]time.Time{start}, start.Add(time.Minute))
	if err == nil {
		t.Fatal("expected an error, got nil")
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
