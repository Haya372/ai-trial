package recurringeventseries

import (
	"fmt"
	"slices"
	"time"
)

const spanYears = 2

type RecurrenceRule struct {
	Pattern  RecurrencePattern
	End      RecurrenceEnd
	Timezone SeriesTimezone
}

// Occurrence is the start and end of one occurrence, in UTC (ADR-036).
type Occurrence struct {
	StartAt time.Time
	EndAt   time.Time
}

// ExpandOccurrences returns every occurrence of the series whose first
// occurrence is [firstStart, firstEnd], skipping excluded dates. Each
// occurrence keeps the first one's wall-clock start in the series' timezone,
// so its UTC time shifts across DST transitions (ADR-037). The end is the
// start plus the first occurrence's duration: rebuilding the end from its
// wall-clock time could put it before the start when it falls in a DST gap.
// firstEnd must not be before firstStart; the Event aggregate guarantees it.
func ExpandOccurrences(
	firstStart, firstEnd time.Time, rule RecurrenceRule, excluded []ExcludedDate,
) ([]Occurrence, error) {
	loc := rule.Timezone.Location()
	localStart := firstStart.In(loc)
	first := LocalDateOf(localStart)

	stop, err := stopDate(first, rule)
	if err != nil {
		return nil, err
	}

	dates, candidates := occurrenceDates(rule.Pattern, first, stop, excluded)
	if count, ok := rule.End.Count(); ok {
		if len(dates) < count {
			return nil, fmt.Errorf("%w: %d occurrences do not fit by %s", ErrRecurringSeriesSpanExceeded, count, stop)
		}
		dates = dates[:count]
	} else if candidates < minCount {
		return nil, fmt.Errorf("%w: got %d", ErrRecurringSeriesTooFewOccurrences, candidates)
	}

	duration := firstEnd.Sub(firstStart)
	occs := make([]Occurrence, len(dates))
	for i, d := range dates {
		// The first occurrence keeps the instant the user sent: on a fall-back
		// day its wall-clock time is ambiguous and could resolve an hour off.
		start := firstStart.UTC()
		if d != first {
			start = atWallClock(d, localStart, loc)
		}
		occs[i] = Occurrence{StartAt: start, EndAt: start.Add(duration)}
	}
	return occs, nil
}

// occurrenceDates returns the non-excluded dates up to stop, and the number
// of candidate dates before exclusion.
func occurrenceDates(p RecurrencePattern, first, stop LocalDate, excluded []ExcludedDate) ([]LocalDate, int) {
	isExcluded := make(map[LocalDate]bool, len(excluded))
	for _, e := range excluded {
		isExcluded[e.Date()] = true
	}
	var dates []LocalDate
	candidates := 0
	for k := 0; ; k++ {
		c := candidateAt(p, first, k)
		if c.periodStart.After(stop) || (c.exists && c.date.After(stop)) {
			return dates, candidates
		}
		if !c.exists {
			continue
		}
		candidates++
		if !isExcluded[c.date] {
			dates = append(dates, c.date)
		}
	}
}

// stopDate validates the parts of rule that depend on the first occurrence
// and returns the last date an occurrence may start on.
func stopDate(first LocalDate, rule RecurrenceRule) (LocalDate, error) {
	if nth, ok := rule.Pattern.MonthlyNth(); ok && !slices.Contains(MonthlyNthCandidates(first), nth) {
		return LocalDate{}, fmt.Errorf("%w: %s does not fall on monthlyNth %q", ErrRecurringPatternInvalid, first, nth)
	}
	limit := spanLimit(first)
	until, ok := rule.End.UntilDate()
	if !ok {
		return limit, nil
	}
	if until.Before(first) || until.After(limit) {
		return LocalDate{}, fmt.Errorf("%w: untilDate %s must be between %s and %s",
			ErrRecurringEndInvalid, until, first, limit)
	}
	return until, nil
}

// candidate is the k-th date the pattern produces. exists is false when the
// period has no such date (a month without the day); periodStart is the
// first day of the period so the caller can still tell when to stop.
type candidate struct {
	date        LocalDate
	exists      bool
	periodStart LocalDate
}

func candidateAt(p RecurrencePattern, first LocalDate, k int) candidate {
	var stepDays int
	switch p.typ {
	case PatternDaily:
		stepDays = 1
	case PatternWeekly:
		stepDays = 7
	case PatternBiweekly:
		stepDays = 14
	case PatternMonthlyDay, PatternMonthlyNthWeekday:
		return monthlyCandidateAt(p, first, k)
	}
	d := first.addDays(stepDays * k)
	return candidate{date: d, exists: true, periodStart: d}
}

// monthlyCandidateAt computes months by hand because time.AddDate
// normalizes 1/31 + 1 month to 3/3 instead of skipping February.
func monthlyCandidateAt(p RecurrencePattern, first LocalDate, k int) candidate {
	monthStart := LocalDateOf(time.Date(first.year, first.month+time.Month(k), 1, 0, 0, 0, 0, time.UTC))
	days := daysIn(monthStart.year, monthStart.month)
	day := first.day
	if p.typ == PatternMonthlyNthWeekday {
		firstWeekdayDay := 1 + (int(first.weekday())-int(monthStart.weekday())+7)%7
		if p.nth == MonthlyNthLast {
			day = firstWeekdayDay + 7*((days-firstWeekdayDay)/7)
		} else {
			day = firstWeekdayDay + 7*int(p.nth[0]-'1')
		}
	}
	if day > days {
		return candidate{periodStart: monthStart}
	}
	d := LocalDate{year: monthStart.year, month: monthStart.month, day: day}
	return candidate{date: d, exists: true, periodStart: monthStart}
}

// spanLimit is the same date 2 years later, or Feb 28 when first is Feb 29
// and the target year is not a leap year, so the span never exceeds 2 years.
func spanLimit(first LocalDate) LocalDate {
	y := first.year + spanYears
	return LocalDate{year: y, month: first.month, day: min(first.day, daysIn(y, first.month))}
}

func atWallClock(d LocalDate, clock time.Time, loc *time.Location) time.Time {
	return time.Date(d.year, d.month, d.day,
		clock.Hour(), clock.Minute(), clock.Second(), clock.Nanosecond(), loc).UTC()
}
