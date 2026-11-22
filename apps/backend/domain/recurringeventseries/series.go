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
	// A new series that keeps the rule takes them over, so deleted and
	// detached occurrences do not come back. Treating the first occurrence as
	// "all" is the caller's decision, since the series does not know its
	// first occurrence.
	EndBefore(occurrenceStart time.Time) (ended RecurringEventSeries, dropped []ExcludedDate)
	// FollowingEnd returns the end the new series takes over in a "this and
	// following" change that keeps the rule, so the calendar shows the same
	// occurrences before and after. A count becomes the number of remaining
	// candidates from target on, excluded dates included; an indefinite end
	// becomes an until date on the last occurrence; an until date carries
	// over. A single remaining occurrence makes a non-count end a count of 1.
	FollowingEnd(occurrenceStarts []time.Time, target time.Time) (RecurrenceEnd, error)
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
	from := s.LocalDateOf(occurrenceStart)
	// s.excluded is sorted, so the dates from the target on form its tail.
	i, _ := slices.BinarySearchFunc(s.excluded, from, func(e ExcludedDate, d LocalDate) int {
		return e.date.compare(d)
	})
	copied := *s
	copied.rule.End = NewUntilDateEnd(from.addDays(-1))
	copied.excluded = slices.Clone(s.excluded[:i])
	return &copied, slices.Clone(s.excluded[i:])
}

// errTargetNotAnOccurrence is not a DomainError: the caller passes target
// from the series' own occurrences, so missing it is a bug, not user input.
var errTargetNotAnOccurrence = errors.New("the target is not the start of an occurrence of the series")

func (s *seriesEntity) FollowingEnd(occurrenceStarts []time.Time, target time.Time) (RecurrenceEnd, error) {
	// Checking membership catches a caller that passes the edited start
	// instead of the stored one, which would count from the wrong instant.
	if !slices.ContainsFunc(occurrenceStarts, target.Equal) {
		return RecurrenceEnd{}, fmt.Errorf("%w: %s", errTargetNotAnOccurrence, target)
	}
	remaining, last := countFrom(occurrenceStarts, target)
	// One remaining occurrence is a single candidate up to an until date,
	// which ExpandOccurrences rejects; a count of 1 keeps that occurrence.
	if remaining == 1 && s.rule.End.Type() != EndCount {
		return RestoreCountEnd(1)
	}
	switch s.rule.End.Type() {
	case EndCount:
		// A count counts excluded dates, and the new series takes over
		// those from the target on.
		from := s.LocalDateOf(target)
		i, _ := slices.BinarySearchFunc(s.excluded, from, func(e ExcludedDate, d LocalDate) int {
			return e.date.compare(d)
		})
		return RestoreCountEnd(remaining + len(s.excluded) - i)
	case EndIndefinite:
		// Carrying "indefinite" over would restart its 2-year window at
		// the target and add occurrences the calendar did not show.
		return NewUntilDateEnd(s.LocalDateOf(last)), nil
	case EndUntilDate:
	}
	return s.rule.End, nil
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
