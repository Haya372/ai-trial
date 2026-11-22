package recurringeventseries_test

import (
	"errors"
	"slices"
	"testing"
	"time"

	res "github.com/Haya372/ai-trial/backend/domain/recurringeventseries"
)

const (
	tokyo   = "Asia/Tokyo"
	newYork = "America/New_York"
	utc     = "UTC"
)

var (
	daily       = string(res.PatternDaily)
	weekly      = string(res.PatternWeekly)
	biweekly    = string(res.PatternBiweekly)
	monthlyDay  = string(res.PatternMonthlyDay)
	monthlyNthW = string(res.PatternMonthlyNthWeekday)
)

func mustRule(t *testing.T, typ string, nth *string, end res.RecurrenceEnd, tz string) res.RecurrenceRule {
	t.Helper()
	p, err := res.NewRecurrencePattern(typ, nth)
	if err != nil {
		t.Fatalf("NewRecurrencePattern: %v", err)
	}
	z, err := res.NewSeriesTimezone(tz)
	if err != nil {
		t.Fatalf("NewSeriesTimezone: %v", err)
	}
	return res.RecurrenceRule{Pattern: p, End: end, Timezone: z}
}

func mustCount(t *testing.T, n int) res.RecurrenceEnd {
	t.Helper()
	e, err := res.NewCountEnd(n)
	if err != nil {
		t.Fatalf("NewCountEnd: %v", err)
	}
	return e
}

func until(t *testing.T, y int, m time.Month, d int) res.RecurrenceEnd {
	t.Helper()
	return res.NewUntilDateEnd(mustDate(t, y, m, d))
}

func localAt(t *testing.T, tz string, y int, m time.Month, d, hh, mm int) time.Time {
	t.Helper()
	loc, err := time.LoadLocation(tz)
	if err != nil {
		t.Fatal(err)
	}
	return time.Date(y, m, d, hh, mm, 0, 0, loc).UTC()
}

// expandHour expands a one-hour first occurrence starting at start.
func expandHour(start time.Time, rule res.RecurrenceRule, excluded ...res.ExcludedDate) ([]res.Occurrence, error) {
	return res.ExpandOccurrences(start, start.Add(time.Hour), rule, excluded)
}

func startDates(occs []res.Occurrence, rule res.RecurrenceRule) []string {
	out := make([]string, len(occs))
	for i, o := range occs {
		out[i] = o.StartAt.In(rule.Timezone.Location()).Format(time.DateOnly)
	}
	return out
}

func assertDates(t *testing.T, start time.Time, rule res.RecurrenceRule, want []string, excluded ...res.ExcludedDate) {
	t.Helper()
	occs, err := expandHour(start, rule, excluded...)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := startDates(occs, rule); !slices.Equal(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func assertErr(t *testing.T, start time.Time, rule res.RecurrenceRule, want error) {
	t.Helper()
	if _, err := expandHour(start, rule); !errors.Is(err, want) {
		t.Errorf("got %v, want %v", err, want)
	}
}

func TestExpandOccurrences_weeklyKeepsWallClockAcrossDST(t *testing.T) {
	start := localAt(t, newYork, 2026, time.March, 2, 10, 0)
	end := localAt(t, newYork, 2026, time.March, 2, 11, 30)
	rule := mustRule(t, weekly, nil, mustCount(t, 36), newYork)

	occs, err := res.ExpandOccurrences(start, end, rule, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(occs) != 36 {
		t.Fatalf("got %d occurrences, want 36", len(occs))
	}
	loc := rule.Timezone.Location()
	for _, o := range occs {
		if o.StartAt.Location() != time.UTC || o.EndAt.Location() != time.UTC {
			t.Fatalf("occurrence not in UTC: %v", o)
		}
		s, e := o.StartAt.In(loc), o.EndAt.In(loc)
		if s.Weekday() != time.Monday || s.Hour() != 10 || s.Minute() != 0 || e.Hour() != 11 || e.Minute() != 30 {
			t.Errorf("wall clock drifted: %v - %v", s, e)
		}
	}
	// 10:00 EST = 15:00Z until DST starts on 3/8, 10:00 EDT = 14:00Z until it
	// ends on 11/1, then 15:00Z again.
	if got := occs[0].StartAt.Hour(); got != 15 {
		t.Errorf("first occurrence UTC hour = %d, want 15", got)
	}
	if got := occs[1].StartAt.Hour(); got != 14 {
		t.Errorf("second occurrence UTC hour = %d, want 14", got)
	}
	if got := occs[35].StartAt; !got.Equal(time.Date(2026, 11, 2, 15, 0, 0, 0, time.UTC)) {
		t.Errorf("36th occurrence = %v, want 2026-11-02 15:00Z", got)
	}
}

func TestExpandOccurrences_endDateSpansMidnight(t *testing.T) {
	start := localAt(t, tokyo, 2026, time.January, 5, 23, 0)
	end := localAt(t, tokyo, 2026, time.January, 6, 1, 0)
	occs, err := res.ExpandOccurrences(start, end, mustRule(t, daily, nil, mustCount(t, 3), tokyo), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if want := localAt(t, tokyo, 2026, time.January, 8, 1, 0); !occs[2].EndAt.Equal(want) {
		t.Errorf("third EndAt = %v, want %v", occs[2].EndAt, want)
	}
}

func TestExpandOccurrences_monthlyDaySkipsMissingDays(t *testing.T) {
	start := localAt(t, tokyo, 2026, time.January, 31, 9, 0)
	assertDates(t, start, mustRule(t, monthlyDay, nil, mustCount(t, 4), tokyo),
		[]string{"2026-01-31", "2026-03-31", "2026-05-31", "2026-07-31"})
}

func TestExpandOccurrences_monthlyDayFeb29SkipsCommonYears(t *testing.T) {
	start := localAt(t, utc, 2028, time.February, 29, 9, 0)
	rule := mustRule(t, monthlyDay, nil, res.NewIndefiniteEnd(), utc)
	occs, err := expandHour(start, rule)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	dates := startDates(occs, rule)
	for _, d := range dates {
		if d[8:] != "29" {
			t.Errorf("occurrence %s is not on the 29th", d)
		}
	}
	// 2029-02 has no 29th, and the 2-year limit from 2028-02-29 is 2030-02-28.
	if slices.Contains(dates, "2029-03-01") || dates[len(dates)-1] != "2030-01-29" || len(dates) != 23 {
		t.Errorf("got %d occurrences ending %s, want 23 ending 2030-01-29", len(dates), dates[len(dates)-1])
	}
}

func TestExpandOccurrences_monthlyNthWeekday(t *testing.T) {
	t.Run("last on a 5th-week Monday", func(t *testing.T) {
		start := localAt(t, tokyo, 2026, time.November, 30, 9, 0)
		assertDates(t, start, mustRule(t, monthlyNthW, new(string(res.MonthlyNthLast)), mustCount(t, 4), tokyo),
			[]string{"2026-11-30", "2026-12-28", "2027-01-25", "2027-02-22"})
	})
	t.Run("4th Monday", func(t *testing.T) {
		start := localAt(t, tokyo, 2026, time.November, 23, 9, 0)
		assertDates(t, start, mustRule(t, monthlyNthW, new("4"), mustCount(t, 3), tokyo),
			[]string{"2026-11-23", "2026-12-28", "2027-01-25"})
	})
	t.Run("nth not matching the first date", func(t *testing.T) {
		start := localAt(t, tokyo, 2026, time.November, 30, 9, 0)
		assertErr(t, start, mustRule(t, monthlyNthW, new("4"), mustCount(t, 3), tokyo), res.ErrRecurringPatternInvalid)
	})
}

func TestExpandOccurrences_indefiniteCoversTwoYearsInclusive(t *testing.T) {
	// From 2026-01-05 the limit is 2028-01-05 (730 days later).
	cases := []struct {
		typ  string
		want int
		last string
	}{
		{daily, 731, "2028-01-05"},
		{weekly, 105, "2028-01-03"},
		{biweekly, 53, "2028-01-03"},
		{monthlyDay, 25, "2028-01-05"},
	}
	start := localAt(t, utc, 2026, time.January, 5, 9, 0)
	for _, c := range cases {
		t.Run(c.typ, func(t *testing.T) {
			rule := mustRule(t, c.typ, nil, res.NewIndefiniteEnd(), utc)
			occs, err := expandHour(start, rule)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			dates := startDates(occs, rule)
			if len(dates) != c.want || dates[len(dates)-1] != c.last {
				t.Errorf("got %d ending %s, want %d ending %s", len(dates), dates[len(dates)-1], c.want, c.last)
			}
		})
	}
}

func TestExpandOccurrences_count(t *testing.T) {
	t.Run("excluded dates and skipped months are not counted", func(t *testing.T) {
		start := localAt(t, tokyo, 2026, time.January, 31, 9, 0)
		excluded := res.NewExcludedDate(mustDate(t, 2026, time.March, 31))
		assertDates(t, start, mustRule(t, monthlyDay, nil, mustCount(t, 3), tokyo),
			[]string{"2026-01-31", "2026-05-31", "2026-07-31"}, excluded)
	})

	start := localAt(t, utc, 2026, time.January, 5, 9, 0)
	t.Run("last occurrence on the 2-year limit", func(t *testing.T) {
		if _, err := expandHour(start, mustRule(t, monthlyDay, nil, mustCount(t, 25), utc)); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	t.Run("beyond the 2-year limit", func(t *testing.T) {
		assertErr(t, start, mustRule(t, monthlyDay, nil, mustCount(t, 26), utc), res.ErrRecurringSeriesSpanExceeded)
	})
}

func TestExpandOccurrences_untilDate(t *testing.T) {
	// 00:30 JST is the previous day in UTC; the local date must be used.
	start := localAt(t, tokyo, 2026, time.January, 5, 0, 30)

	t.Run("includes the occurrence on the until date", func(t *testing.T) {
		assertDates(t, start, mustRule(t, daily, nil, until(t, 2026, time.January, 7), tokyo),
			[]string{"2026-01-05", "2026-01-06", "2026-01-07"})
	})
	t.Run("excluded dates are skipped", func(t *testing.T) {
		excluded := res.NewExcludedDate(mustDate(t, 2026, time.January, 6))
		assertDates(t, start, mustRule(t, daily, nil, until(t, 2026, time.January, 7), tokyo),
			[]string{"2026-01-05", "2026-01-07"}, excluded)
	})
	t.Run("accepts the 2-year limit", func(t *testing.T) {
		if _, err := expandHour(start, mustRule(t, weekly, nil, until(t, 2028, time.January, 5), tokyo)); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	t.Run("before the first date", func(t *testing.T) {
		assertErr(t, start, mustRule(t, daily, nil, until(t, 2026, time.January, 4), tokyo), res.ErrRecurringEndInvalid)
	})
	t.Run("beyond the 2-year limit", func(t *testing.T) {
		assertErr(t, start, mustRule(t, daily, nil, until(t, 2028, time.January, 6), tokyo), res.ErrRecurringEndInvalid)
	})
}

func TestExpandOccurrences_tooFewOccurrences(t *testing.T) {
	start := localAt(t, tokyo, 2026, time.January, 30, 9, 0)
	cases := map[string]res.RecurrenceRule{
		"monthly_day until the next month without the day": mustRule(t, monthlyDay, nil,
			until(t, 2026, time.February, 28), tokyo),
		"daily until the first date": mustRule(t, daily, nil,
			until(t, 2026, time.January, 30), tokyo),
		"weekly until before the 2nd occurrence": mustRule(t, weekly, nil,
			until(t, 2026, time.February, 5), tokyo),
	}
	for name, rule := range cases {
		t.Run(name, func(t *testing.T) {
			assertErr(t, start, rule, res.ErrRecurringSeriesTooFewOccurrences)
		})
	}
}
