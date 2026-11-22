package recurringeventseries

import "fmt"

type EndType string

const (
	EndIndefinite EndType = "indefinite"
	EndUntilDate  EndType = "until_date"
	EndCount      EndType = "count"
)

const minCount = 2

type RecurrenceEnd struct {
	typ       EndType
	untilDate LocalDate
	count     int
}

func NewIndefiniteEnd() RecurrenceEnd {
	return RecurrenceEnd{typ: EndIndefinite}
}

// NewUntilDateEnd does not check the range of d, because the range depends on
// the first occurrence; ExpandOccurrences checks it.
func NewUntilDateEnd(d LocalDate) RecurrenceEnd {
	return RecurrenceEnd{typ: EndUntilDate, untilDate: d}
}

func NewCountEnd(n int) (RecurrenceEnd, error) {
	if n < minCount {
		return RecurrenceEnd{}, fmt.Errorf("%w: count must be at least %d, got %d", ErrRecurringEndInvalid, minCount, n)
	}
	return RecurrenceEnd{typ: EndCount, count: n}, nil
}

func (e RecurrenceEnd) Type() EndType { return e.typ }

func (e RecurrenceEnd) UntilDate() (LocalDate, bool) {
	return e.untilDate, e.typ == EndUntilDate
}

func (e RecurrenceEnd) Count() (int, bool) {
	return e.count, e.typ == EndCount
}
