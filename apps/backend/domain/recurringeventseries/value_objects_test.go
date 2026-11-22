package recurringeventseries_test

import (
	"errors"
	"slices"
	"testing"
	"time"

	res "github.com/Haya372/ai-trial/backend/domain/recurringeventseries"
)

func mustDate(t *testing.T, y int, m time.Month, d int) res.LocalDate {
	t.Helper()
	ld, err := res.NewLocalDate(y, m, d)
	if err != nil {
		t.Fatalf("NewLocalDate(%d, %d, %d): %v", y, m, d, err)
	}
	return ld
}

func TestNewLocalDate(t *testing.T) {
	d, err := res.NewLocalDate(2028, time.February, 29)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d.String() != "2028-02-29" {
		t.Errorf("got %s", d)
	}
	for _, c := range []struct {
		y int
		m time.Month
		d int
	}{{2026, time.February, 29}, {2026, time.April, 31}, {2026, time.January, 0}, {2026, 13, 1}} {
		if _, err := res.NewLocalDate(c.y, c.m, c.d); !errors.Is(err, res.ErrLocalDateInvalid) {
			t.Errorf("NewLocalDate(%d, %d, %d): got %v, want ErrLocalDateInvalid", c.y, c.m, c.d, err)
		}
	}
}

func TestLocalDateOf_usesLocationOfTime(t *testing.T) {
	tokyo, _ := time.LoadLocation("Asia/Tokyo")
	tm := time.Date(2026, 3, 1, 16, 0, 0, 0, time.UTC).In(tokyo)
	if got := res.LocalDateOf(tm).String(); got != "2026-03-02" {
		t.Errorf("got %s, want 2026-03-02", got)
	}
}

func TestLocalDate_compare(t *testing.T) {
	a := mustDate(t, 2026, time.January, 31)
	b := mustDate(t, 2026, time.February, 1)
	if !a.Before(b) || a.After(b) || !b.After(a) || b.Before(a) || a.Before(a) || a.After(a) {
		t.Error("unexpected comparison result")
	}
}

func TestNewMonthlyNth(t *testing.T) {
	for _, s := range []string{"1", "2", "3", "4", "last"} {
		n, err := res.NewMonthlyNth(s)
		if err != nil || string(n) != s {
			t.Errorf("NewMonthlyNth(%q) = %q, %v", s, n, err)
		}
	}
	for _, s := range []string{"", "0", "5", "Last", "first"} {
		if _, err := res.NewMonthlyNth(s); !errors.Is(err, res.ErrRecurringPatternInvalid) {
			t.Errorf("NewMonthlyNth(%q): got %v, want ErrRecurringPatternInvalid", s, err)
		}
	}
}

func TestMonthlyNthCandidates(t *testing.T) {
	cases := []struct {
		name string
		date res.LocalDate
		want []res.MonthlyNth
	}{
		{"first week", mustDate(t, 2026, time.November, 2), []res.MonthlyNth{"1"}},
		{"4th and last", mustDate(t, 2026, time.November, 26), []res.MonthlyNth{res.MonthlyNth4, res.MonthlyNthLast}},
		{"4th but not last", mustDate(t, 2026, time.November, 23), []res.MonthlyNth{"4"}},
		{"5th week", mustDate(t, 2026, time.November, 30), []res.MonthlyNth{res.MonthlyNthLast}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := res.MonthlyNthCandidates(c.date); !slices.Equal(got, c.want) {
				t.Errorf("got %v, want %v", got, c.want)
			}
		})
	}
}

func TestNewRecurrencePattern(t *testing.T) {
	for _, typ := range []string{"daily", "weekly", "biweekly", "monthly_day"} {
		p, err := res.NewRecurrencePattern(typ, nil)
		if err != nil {
			t.Fatalf("%s: unexpected error: %v", typ, err)
		}
		if string(p.Type()) != typ {
			t.Errorf("Type() = %q, want %q", p.Type(), typ)
		}
		if _, ok := p.MonthlyNth(); ok {
			t.Errorf("%s: MonthlyNth should be absent", typ)
		}
	}

	p, err := res.NewRecurrencePattern("monthly_nth_weekday", new(string(res.MonthlyNthLast)))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n, ok := p.MonthlyNth(); !ok || n != res.MonthlyNthLast {
		t.Errorf("MonthlyNth() = %q, %v", n, ok)
	}

	invalid := []struct {
		typ string
		nth *string
	}{
		{"", nil},
		{"yearly", nil},
		{"monthly_nth_weekday", new("5")},
		{"weekly", new("1")},
	}
	for _, c := range invalid {
		if _, err := res.NewRecurrencePattern(c.typ, c.nth); !errors.Is(err, res.ErrRecurringPatternInvalid) {
			t.Errorf("NewRecurrencePattern(%q, %v): got %v, want ErrRecurringPatternInvalid", c.typ, c.nth, err)
		}
	}
}

func TestNewRecurrencePattern_monthlyNthRequired(t *testing.T) {
	// SPEC-009 reports a missing monthlyNth as REQUIRED, not INVALID_FORMAT.
	_, err := res.NewRecurrencePattern(string(res.PatternMonthlyNthWeekday), nil)
	if !errors.Is(err, res.ErrRecurringPatternMonthlyNthRequired) {
		t.Errorf("got %v, want ErrRecurringPatternMonthlyNthRequired", err)
	}
}

func TestRecurrenceEnd(t *testing.T) {
	e := res.NewIndefiniteEnd()
	if e.Type() != res.EndIndefinite {
		t.Errorf("Type() = %q", e.Type())
	}
	if _, ok := e.UntilDate(); ok {
		t.Error("UntilDate should be absent")
	}
	if _, ok := e.Count(); ok {
		t.Error("Count should be absent")
	}

	until := mustDate(t, 2026, time.December, 31)
	e = res.NewUntilDateEnd(until)
	if d, ok := e.UntilDate(); e.Type() != res.EndUntilDate || !ok || d != until {
		t.Errorf("until end = %q, %v, %v", e.Type(), d, ok)
	}

	e, err := res.NewCountEnd(2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n, ok := e.Count(); e.Type() != res.EndCount || !ok || n != 2 {
		t.Errorf("count end = %q, %d, %v", e.Type(), n, ok)
	}
	for _, n := range []int{1, 0, -1} {
		if _, err := res.NewCountEnd(n); !errors.Is(err, res.ErrRecurringEndInvalid) {
			t.Errorf("NewCountEnd(%d): got %v, want ErrRecurringEndInvalid", n, err)
		}
	}
}

func TestNewSeriesTimezone(t *testing.T) {
	for _, name := range []string{"Asia/Tokyo", "America/New_York", "UTC"} {
		z, err := res.NewSeriesTimezone(name)
		if err != nil {
			t.Fatalf("%s: unexpected error: %v", name, err)
		}
		if z.Name() != name || z.Location().String() != name {
			t.Errorf("Name() = %q, Location() = %q", z.Name(), z.Location())
		}
	}
	// "" and "Local" are accepted by time.LoadLocation but are not IANA names
	// and would make the result depend on the server's environment.
	for _, name := range []string{"", "Local", "Invalid/Zone", "Asia/Tokyo "} {
		if _, err := res.NewSeriesTimezone(name); !errors.Is(err, res.ErrRecurringTimezoneInvalid) {
			t.Errorf("NewSeriesTimezone(%q): got %v, want ErrRecurringTimezoneInvalid", name, err)
		}
	}
}

func TestExcludedDate(t *testing.T) {
	d := mustDate(t, 2026, time.March, 9)
	if got := res.NewExcludedDate(d).Date(); got != d {
		t.Errorf("Date() = %v, want %v", got, d)
	}
}

func TestErrorCodes(t *testing.T) {
	cases := []struct {
		err  error
		code string
	}{
		{res.ErrRecurringPatternInvalid, "INVALID_FORMAT"},
		{res.ErrRecurringEndInvalid, "OUT_OF_RANGE"},
		{res.ErrRecurringTimezoneInvalid, "INVALID_FORMAT"},
		{res.ErrRecurringSeriesTooFewOccurrences, "TOO_FEW_OCCURRENCES"},
		{res.ErrRecurringSeriesSpanExceeded, "OUT_OF_RANGE"},
		{res.ErrRecurringPatternMonthlyNthRequired, "REQUIRED"},
		{res.ErrRecurringOccurrenceDateChangeNotAllowed, "DATE_CHANGE_NOT_ALLOWED"},
	}
	for _, c := range cases {
		var de interface{ Code() string }
		if !errors.As(c.err, &de) || de.Code() != c.code {
			t.Errorf("%v: code mismatch, want %s", c.err, c.code)
		}
	}
}
