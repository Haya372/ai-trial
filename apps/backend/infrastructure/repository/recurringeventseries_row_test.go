//nolint:testpackage // seriesRow mapping is unexported
package repository

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	rs "github.com/Haya372/ai-trial/backend/domain/recurringeventseries"
)

func mustLocalDate(t *testing.T, y int, m time.Month, d int) rs.LocalDate {
	t.Helper()
	ld, err := rs.NewLocalDate(y, m, d)
	if err != nil {
		t.Fatalf("build local date: %v", err)
	}
	return ld
}

func mustSeries(
	t *testing.T, pattern string, nth *string, end rs.RecurrenceEnd, excluded ...rs.LocalDate,
) rs.RecurringEventSeries {
	t.Helper()
	p, err := rs.NewRecurrencePattern(pattern, nth)
	if err != nil {
		t.Fatalf("build pattern: %v", err)
	}
	tz, err := rs.NewSeriesTimezone("Asia/Tokyo")
	if err != nil {
		t.Fatalf("build timezone: %v", err)
	}
	ex := make([]rs.ExcludedDate, 0, len(excluded))
	for _, d := range excluded {
		ex = append(ex, rs.NewExcludedDate(d))
	}
	s, err := rs.New(uuid.New(), uuid.New(), rs.RecurrenceRule{Pattern: p, End: end, Timezone: tz}, ex)
	if err != nil {
		t.Fatalf("build series: %v", err)
	}
	return s
}

func TestSeriesRow_RoundTrip(t *testing.T) {
	last := string(rs.MonthlyNthLast)
	count, err := rs.NewCountEnd(10)
	if err != nil {
		t.Fatalf("build count end: %v", err)
	}
	until := rs.NewUntilDateEnd(mustLocalDate(t, 2027, time.March, 31))

	cases := map[string]rs.RecurringEventSeries{
		"daily indefinite":    mustSeries(t, "daily", nil, rs.NewIndefiniteEnd()),
		"weekly until date":   mustSeries(t, "weekly", nil, until),
		"biweekly count":      mustSeries(t, "biweekly", nil, count),
		"monthly day":         mustSeries(t, "monthly_day", nil, count),
		"monthly nth weekday": mustSeries(t, "monthly_nth_weekday", &last, until),
		"with excluded dates": mustSeries(t, "daily", nil, count,
			mustLocalDate(t, 2026, time.December, 1), mustLocalDate(t, 2026, time.November, 30)),
	}
	for name, want := range cases {
		t.Run(name, func(t *testing.T) {
			row, excluded := seriesRowOf(want)
			got, err := seriesFromRow(row, excluded)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			assertSeriesEqual(t, got, want)
		})
	}
}

func assertSeriesEqual(t *testing.T, got, want rs.RecurringEventSeries) {
	t.Helper()
	if got.ID() != want.ID() || got.UserID() != want.UserID() {
		t.Errorf("ids mismatch: got %v/%v, want %v/%v", got.ID(), got.UserID(), want.ID(), want.UserID())
	}
	if got.Rule().Pattern != want.Rule().Pattern {
		t.Errorf("pattern mismatch: got %+v, want %+v", got.Rule().Pattern, want.Rule().Pattern)
	}
	if got.Rule().End != want.Rule().End {
		t.Errorf("end mismatch: got %+v, want %+v", got.Rule().End, want.Rule().End)
	}
	if got.Rule().Timezone.Name() != want.Rule().Timezone.Name() {
		t.Errorf("timezone mismatch: got %q, want %q", got.Rule().Timezone.Name(), want.Rule().Timezone.Name())
	}
	g, w := got.ExcludedDates(), want.ExcludedDates()
	if len(g) != len(w) {
		t.Fatalf("expected %d excluded dates, got %d", len(w), len(g))
	}
	for i := range w {
		if g[i].Date() != w[i].Date() {
			t.Errorf("excluded[%d] mismatch: got %v, want %v", i, g[i].Date(), w[i].Date())
		}
	}
}

func TestSeriesRowOf_StoresNullForUnusedColumns(t *testing.T) {
	row, excluded := seriesRowOf(mustSeries(t, "daily", nil, rs.NewIndefiniteEnd()))

	if row.MonthlyNth.Valid || row.EndUntilDate.Valid || row.EndCount.Valid {
		t.Errorf("expected NULL monthly_nth/end_until_date/end_count, got %+v", row)
	}
	if row.PatternType != "daily" || row.EndType != "indefinite" || row.Timezone != "Asia/Tokyo" {
		t.Errorf("unexpected row: %+v", row)
	}
	if len(excluded) != 0 {
		t.Errorf("expected no excluded dates, got %d", len(excluded))
	}
}

func TestSeriesFromRow_RestoresCountOfOne(t *testing.T) {
	row := validSeriesRow()
	row.EndType = string(rs.EndCount)
	row.EndCount = pgtype.Int2{Int16: 1, Valid: true}

	s, err := seriesFromRow(row, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n, ok := s.Rule().End.Count(); !ok || n != 1 {
		t.Errorf("expected count 1, got %d (ok=%v)", n, ok)
	}
}

func validSeriesRow() seriesRow {
	return seriesRow{
		ID:          toPgUUID(uuid.New()),
		UserID:      toPgUUID(uuid.New()),
		PatternType: "daily",
		EndType:     "indefinite",
		Timezone:    "Asia/Tokyo",
	}
}

func TestSeriesFromRow_InvalidStoredValues_ReturnsError(t *testing.T) {
	cases := map[string]func(r *seriesRow){
		"unknown pattern":         func(r *seriesRow) { r.PatternType = "yearly" },
		"nth weekday without nth": func(r *seriesRow) { r.PatternType = "monthly_nth_weekday" },
		"unknown end type":        func(r *seriesRow) { r.EndType = "forever" },
		"until date without date": func(r *seriesRow) { r.EndType = "until_date" },
		"count without count":     func(r *seriesRow) { r.EndType = string(rs.EndCount) },
		"count of zero": func(r *seriesRow) {
			r.EndType = string(rs.EndCount)
			r.EndCount = pgtype.Int2{Valid: true}
		},
		"unknown timezone": func(r *seriesRow) { r.Timezone = "Mars/Olympus" },
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			row := validSeriesRow()
			mutate(&row)
			if _, err := seriesFromRow(row, nil); err == nil {
				t.Error("expected an error")
			}
		})
	}
}

func TestSeriesFromRow_DuplicatedExcludedDates_ReturnsError(t *testing.T) {
	d := pgtype.Date{Time: time.Date(2026, time.December, 1, 0, 0, 0, 0, time.UTC), Valid: true}

	_, err := seriesFromRow(validSeriesRow(), []pgtype.Date{d, d})
	if !errors.Is(err, rs.ErrRecurringExcludedDateDuplicated) {
		t.Errorf("expected ErrRecurringExcludedDateDuplicated, got %v", err)
	}
}
