package recurringeventseries

import (
	"fmt"
	"time"
)

// LocalDate is a calendar date in a SeriesTimezone, without time or zone.
type LocalDate struct {
	year  int
	month time.Month
	day   int
}

func NewLocalDate(year int, month time.Month, day int) (LocalDate, error) {
	if month < time.January || month > time.December || day < 1 || day > daysIn(year, month) {
		return LocalDate{}, fmt.Errorf("%w: %04d-%02d-%02d", ErrLocalDateInvalid, year, month, day)
	}
	return LocalDate{year: year, month: month, day: day}, nil
}

// LocalDateOf returns the date of t in t's own location.
func LocalDateOf(t time.Time) LocalDate {
	y, m, d := t.Date()
	return LocalDate{year: y, month: m, day: d}
}

func (d LocalDate) Year() int               { return d.year }
func (d LocalDate) Month() time.Month       { return d.month }
func (d LocalDate) Day() int                { return d.day }
func (d LocalDate) Before(o LocalDate) bool { return d.compare(o) < 0 }
func (d LocalDate) After(o LocalDate) bool  { return d.compare(o) > 0 }

func (d LocalDate) String() string {
	return fmt.Sprintf("%04d-%02d-%02d", d.year, d.month, d.day)
}

func (d LocalDate) compare(o LocalDate) int {
	switch {
	case d.year != o.year:
		return d.year - o.year
	case d.month != o.month:
		return int(d.month - o.month)
	default:
		return d.day - o.day
	}
}

// addDays is safe to delegate to time.Date normalization, unlike month
// arithmetic, because days never overflow into a different meaning.
func (d LocalDate) addDays(n int) LocalDate {
	return LocalDateOf(time.Date(d.year, d.month, d.day+n, 0, 0, 0, 0, time.UTC))
}

// daysUntil returns how many days o is after d, negative when it is before.
func (d LocalDate) daysUntil(o LocalDate) int {
	from := time.Date(d.year, d.month, d.day, 0, 0, 0, 0, time.UTC)
	to := time.Date(o.year, o.month, o.day, 0, 0, 0, 0, time.UTC)
	return int(to.Sub(from).Hours() / 24)
}

func (d LocalDate) weekday() time.Weekday {
	return time.Date(d.year, d.month, d.day, 0, 0, 0, 0, time.UTC).Weekday()
}

func daysIn(year int, month time.Month) int {
	return time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
}
