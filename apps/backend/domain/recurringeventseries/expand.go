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

// initialized is false when a part of the rule skipped its constructor.
func (r RecurrenceRule) initialized() bool {
	return r.Pattern.typ != "" && r.End.typ != "" && r.Timezone.loc != nil
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
	// A zero-value pattern would never advance past the first date, so the
	// loop in occurrenceDates would not terminate.
	if !rule.initialized() {
		return nil, ErrRecurrenceRuleUninitialized
	}
	loc := rule.Timezone.Location()
	localStart := firstStart.In(loc)
	first := LocalDateOf(localStart)

	stop, err := stopDate(first, rule)
	if err != nil {
		return nil, err
	}

	startAt := func(d LocalDate) (time.Time, bool) {
		// The first occurrence keeps the instant the user sent, which already
		// tells which of two repeated wall-clock times was meant.
		if d == first {
			return firstStart.UTC(), true
		}
		return atWallClock(d, localStart, loc)
	}
	count, hasCount := rule.End.Count()
	starts, candidates := occurrenceStarts(rule.Pattern, first, stop, count, excluded, startAt)
	if hasCount {
		if candidates < count {
			return nil, fmt.Errorf("%w: %d occurrences do not fit by %s", ErrRecurringSeriesSpanExceeded, count, stop)
		}
	} else if candidates < minCount {
		return nil, fmt.Errorf("%w: got %d", ErrRecurringSeriesTooFewOccurrences, candidates)
	}

	duration := firstEnd.Sub(firstStart)
	occs := make([]Occurrence, len(starts))
	for i, start := range starts {
		occs[i] = Occurrence{StartAt: start, EndAt: start.Add(duration)}
	}
	return occs, nil
}

// occurrenceStarts returns the start of each non-excluded candidate up to
// stop, and the number of candidates before exclusion. A positive limit stops
// after that many candidates, excluded ones included, as a count end counts
// them. A date whose start does not exist (a day removed from the calendar)
// is not a candidate, like a month without the day.
func occurrenceStarts(
	p RecurrencePattern, first, stop LocalDate, limit int, excluded []ExcludedDate,
	startAt func(LocalDate) (time.Time, bool),
) ([]time.Time, int) {
	isExcluded := make(map[LocalDate]bool, len(excluded))
	for _, e := range excluded {
		isExcluded[e.Date()] = true
	}
	var starts []time.Time
	candidates := 0
	for k := 0; limit <= 0 || candidates < limit; k++ {
		c := candidateAt(p, first, k)
		if c.periodStart.After(stop) || (c.exists && c.date.After(stop)) {
			return starts, candidates
		}
		if !c.exists {
			continue
		}
		start, ok := startAt(c.date)
		if !ok {
			continue
		}
		candidates++
		if !isExcluded[c.date] {
			starts = append(starts, start)
		}
	}
	return starts, candidates
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

// atWallClock returns the instant of clock's wall-clock time on d, resolved
// like RFC 5545 and common calendars: a time repeated on a fall-back day is
// the first instance, and a time in a gap uses the offset before the gap,
// which moves it forward. time.Date resolves both differently by zone. ok is
// false when d itself was removed from the calendar.
func atWallClock(d LocalDate, clock time.Time, loc *time.Location) (time.Time, bool) {
	want := time.Date(d.year, d.month, d.day,
		clock.Hour(), clock.Minute(), clock.Second(), clock.Nanosecond(), time.UTC)
	t := time.Date(d.year, d.month, d.day,
		clock.Hour(), clock.Minute(), clock.Second(), clock.Nanosecond(), loc)

	// Every instant that reads as want lies in t's zone period or a
	// neighboring one; the earliest of them is the first instance.
	start, end := t.ZoneBounds()
	var first time.Time
	for _, o := range []time.Time{start.Add(-time.Nanosecond), t, end} {
		_, offset := o.Zone()
		inst := want.Add(-time.Duration(offset) * time.Second)
		if naiveWallClock(inst.In(loc)).Equal(want) && (first.IsZero() || inst.Before(first)) {
			first = inst
		}
	}
	if !first.IsZero() {
		return first.UTC(), true
	}

	// want falls in a gap: use the offset before it.
	transition := end
	if naiveWallClock(t).After(want) {
		transition = start
	}
	_, offsetBefore := transition.Add(-time.Nanosecond).Zone()
	inst := want.Add(-time.Duration(offsetBefore) * time.Second)
	return inst, LocalDateOf(inst.In(loc)) == d
}

// naiveWallClock returns t's wall-clock reading as if it were UTC, so wall
// clocks in different offsets can be compared.
func naiveWallClock(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), time.UTC)
}
