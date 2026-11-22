package recurringeventseries

import (
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/google/uuid"
)

// RecurringEventSeries is the aggregate root of a recurring series. Its
// occurrences are separate Event aggregates that refer to it by ID.
// Operations take an occurrence's start instant rather than a date so that
// the conversion to a local date always uses the series' timezone.
type RecurringEventSeries interface {
	ID() uuid.UUID
	UserID() uuid.UUID
	Rule() RecurrenceRule
	// ExcludedDates returns the excluded dates in ascending order.
	ExcludedDates() []ExcludedDate
	// LocalDateOf returns the date of t in the series' timezone.
	LocalDateOf(t time.Time) LocalDate
	// ExcludeOccurrence returns a copy with the occurrence's local date
	// excluded, for "this occurrence only" deletes and detaches.
	ExcludeOccurrence(occurrenceStart time.Time) (RecurringEventSeries, error)
	// EndBefore returns a copy that ends the day before the occurrence's local
	// date, for the original series of a "this and following" change or
	// delete, and the excluded dates from that day on, which the copy drops.
	// FollowingEnd returns those a new series that keeps the rule takes
	// over, if any, so deleted and detached occurrences do not come back.
	// Treating the first occurrence as "all" is the caller's decision, since
	// the series does not know its first occurrence.
	EndBefore(occurrenceStart time.Time) (ended RecurringEventSeries, dropped []ExcludedDate)
	// FollowingEnd returns the end and the excluded dates the new series
	// takes over in a "this and following" change that keeps the rule, so
	// the calendar shows the same occurrences before and after. newStart is
	// the new series' first occurrence. On the target's local date, a count
	// becomes the number of remaining candidates, excluded dates included; an
	// indefinite end becomes an until date on the last occurrence; an until
	// date carries over; a single remaining occurrence makes a non-count end
	// a count of 1. On another date, the end is an until date on the last
	// occurrence shifted by the same days (see movedEnd).
	FollowingEnd(occurrenceStarts []time.Time, target, newStart time.Time) (RecurrenceEnd, []ExcludedDate, error)
}

type seriesEntity struct {
	id       uuid.UUID
	userID   uuid.UUID
	rule     RecurrenceRule
	excluded []ExcludedDate
}

// New both creates a series and restores a stored one. It does not check the
// rule against the first occurrence (the until date's range, monthlyNth, at
// least 2 occurrences): that needs the first occurrence's start, so
// ExpandOccurrences checks it.
func New(id, userID uuid.UUID, rule RecurrenceRule, excluded []ExcludedDate) (RecurringEventSeries, error) {
	if !rule.initialized() {
		return nil, ErrRecurrenceRuleUninitialized
	}
	sorted := slices.Clone(excluded)
	slices.SortFunc(sorted, func(a, b ExcludedDate) int { return a.date.compare(b.date) })
	for i := 1; i < len(sorted); i++ {
		if sorted[i] == sorted[i-1] {
			return nil, fmt.Errorf("%w: %s", ErrRecurringExcludedDateDuplicated, sorted[i].date)
		}
	}
	return &seriesEntity{id: id, userID: userID, rule: rule, excluded: sorted}, nil
}

func (s *seriesEntity) ID() uuid.UUID                 { return s.id }
func (s *seriesEntity) UserID() uuid.UUID             { return s.userID }
func (s *seriesEntity) Rule() RecurrenceRule          { return s.rule }
func (s *seriesEntity) ExcludedDates() []ExcludedDate { return slices.Clone(s.excluded) }

func (s *seriesEntity) LocalDateOf(t time.Time) LocalDate {
	return LocalDateOf(t.In(s.rule.Timezone.Location()))
}

func (s *seriesEntity) ExcludeOccurrence(occurrenceStart time.Time) (RecurringEventSeries, error) {
	return New(s.id, s.userID, s.rule,
		append(slices.Clone(s.excluded), NewExcludedDate(s.LocalDateOf(occurrenceStart))))
}

func (s *seriesEntity) EndBefore(occurrenceStart time.Time) (RecurringEventSeries, []ExcludedDate) {
	i := s.excludedFrom(occurrenceStart)
	copied := *s
	copied.rule.End = NewUntilDateEnd(s.LocalDateOf(occurrenceStart).addDays(-1))
	copied.excluded = slices.Clone(s.excluded[:i])
	return &copied, slices.Clone(s.excluded[i:])
}

// errTargetNotAnOccurrence is not a DomainError: the caller passes target
// from the series' own occurrences, so missing it is a bug, not user input.
var errTargetNotAnOccurrence = errors.New("the target is not the start of an occurrence of the series")

func (s *seriesEntity) FollowingEnd(
	occurrenceStarts []time.Time, target, newStart time.Time,
) (RecurrenceEnd, []ExcludedDate, error) {
	// Checking membership catches a caller that passes the edited start
	// instead of the stored one, which would count from the wrong instant.
	if !slices.ContainsFunc(occurrenceStarts, target.Equal) {
		return RecurrenceEnd{}, nil, fmt.Errorf("%w: %s", errTargetNotAnOccurrence, target)
	}
	remaining, last := countFrom(occurrenceStarts, target)
	excluded := slices.Clone(s.excluded[s.excludedFrom(target):])
	if s.LocalDateOf(newStart) != s.LocalDateOf(target) {
		return s.movedEnd(newStart, target, last, excluded)
	}
	if remaining == 1 && s.rule.End.Type() != EndCount {
		// A single remaining occurrence is a single candidate up to an
		// until date, which ExpandOccurrences rejects.
		end, err := RestoreCountEnd(1)
		return end, nil, err
	}
	switch s.rule.End.Type() {
	case EndCount:
		// A count counts the excluded dates the new series takes over.
		end, err := RestoreCountEnd(remaining + len(excluded))
		return end, excluded, err
	case EndIndefinite:
		// Carrying "indefinite" over would restart its 2-year window at
		// the target and add occurrences the calendar did not show.
		return NewUntilDateEnd(s.LocalDateOf(last)), excluded, nil
	case EndUntilDate:
	}
	return s.rule.End, excluded, nil
}

// movedEnd handles a target moved to another date, which moves the weekday
// or day of the month of the new series. Its candidates differ from the
// original ones, so a count or the original until date may not fit in its
// 2-year window; the last occurrence shifted by the same days always does.
// Excluded dates that are no longer candidates are ignored by an until date,
// and those that still are keep deleted occurrences gone, except one on the
// new first occurrence, which would drop it.
func (s *seriesEntity) movedEnd(
	newStart, target, last time.Time, excluded []ExcludedDate,
) (RecurrenceEnd, []ExcludedDate, error) {
	first := s.LocalDateOf(newStart)
	shift := first.daysSince(s.LocalDateOf(target))
	excluded = slices.DeleteFunc(excluded, func(e ExcludedDate) bool { return e.date == first })
	end := NewUntilDateEnd(s.LocalDateOf(last).addDays(shift))
	rule := RecurrenceRule{Pattern: s.rule.Pattern, End: end, Timezone: s.rule.Timezone}
	if _, err := ExpandOccurrences(newStart, newStart, rule, nil); errors.Is(err, ErrRecurringSeriesTooFewOccurrences) {
		one, err := RestoreCountEnd(1)
		return one, nil, err
	}
	return end, excluded, nil
}

// excludedFrom returns the index of the first excluded date on or after the
// occurrence's local date; s.excluded is sorted, so those dates form its tail.
func (s *seriesEntity) excludedFrom(occurrenceStart time.Time) int {
	i, _ := slices.BinarySearchFunc(s.excluded, s.LocalDateOf(occurrenceStart), func(e ExcludedDate, d LocalDate) int {
		return e.date.compare(d)
	})
	return i
}

// countFrom returns how many starts are not before target, and the latest one.
func countFrom(starts []time.Time, target time.Time) (int, time.Time) {
	n, last := 0, target
	for _, start := range starts {
		if !start.Before(target) {
			n++
		}
		if start.After(last) {
			last = start
		}
	}
	return n, last
}
