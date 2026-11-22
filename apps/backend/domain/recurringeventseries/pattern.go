package recurringeventseries

import "fmt"

type PatternType string

const (
	PatternDaily             PatternType = "daily"
	PatternWeekly            PatternType = "weekly"
	PatternBiweekly          PatternType = "biweekly"
	PatternMonthlyDay        PatternType = "monthly_day"
	PatternMonthlyNthWeekday PatternType = "monthly_nth_weekday"
)

// MonthlyNth is the N of "the Nth weekday of every month". It is stored
// explicitly because a date that is both the 4th and the last weekday of its
// month cannot tell which one the user meant (SPEC-009).
type MonthlyNth string

const (
	MonthlyNth1    MonthlyNth = "1"
	MonthlyNth2    MonthlyNth = "2"
	MonthlyNth3    MonthlyNth = "3"
	MonthlyNth4    MonthlyNth = "4"
	MonthlyNthLast MonthlyNth = "last"
)

var numberedNths = []MonthlyNth{MonthlyNth1, MonthlyNth2, MonthlyNth3, MonthlyNth4}

func NewMonthlyNth(s string) (MonthlyNth, error) {
	switch n := MonthlyNth(s); n {
	case MonthlyNth1, MonthlyNth2, MonthlyNth3, MonthlyNth4, MonthlyNthLast:
		return n, nil
	}
	return "", fmt.Errorf("%w: monthlyNth %q", ErrRecurringPatternInvalid, s)
}

// MonthlyNthCandidates returns the MonthlyNth values that d satisfies, in
// the order the form lists them.
func MonthlyNthCandidates(d LocalDate) []MonthlyNth {
	var out []MonthlyNth
	if week := (d.day - 1) / 7; week < len(numberedNths) {
		out = append(out, numberedNths[week])
	}
	if d.day+7 > daysIn(d.year, d.month) {
		out = append(out, MonthlyNthLast)
	}
	return out
}

type RecurrencePattern struct {
	typ PatternType
	nth MonthlyNth
}

// NewRecurrencePattern requires nth exactly when typ is monthly_nth_weekday.
func NewRecurrencePattern(typ string, nth *string) (RecurrencePattern, error) {
	switch t := PatternType(typ); t {
	case PatternDaily, PatternWeekly, PatternBiweekly, PatternMonthlyDay:
		if nth != nil {
			return RecurrencePattern{}, fmt.Errorf("%w: monthlyNth is only allowed for %s",
				ErrRecurringPatternInvalid, PatternMonthlyNthWeekday)
		}
		return RecurrencePattern{typ: t}, nil
	case PatternMonthlyNthWeekday:
		if nth == nil {
			return RecurrencePattern{}, fmt.Errorf("%w: monthlyNth is required for %s", ErrRecurringPatternInvalid, t)
		}
		n, err := NewMonthlyNth(*nth)
		if err != nil {
			return RecurrencePattern{}, err
		}
		return RecurrencePattern{typ: t, nth: n}, nil
	}
	return RecurrencePattern{}, fmt.Errorf("%w: type %q", ErrRecurringPatternInvalid, typ)
}

func (p RecurrencePattern) Type() PatternType { return p.typ }

func (p RecurrencePattern) MonthlyNth() (MonthlyNth, bool) {
	return p.nth, p.nth != ""
}
