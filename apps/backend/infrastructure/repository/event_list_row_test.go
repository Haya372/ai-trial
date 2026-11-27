//nolint:testpackage // newRecurringSeriesReadModelFromListRow is unexported row mapping
package repository

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	eventuc "github.com/Haya372/ai-trial/backend/usecase/event"
)

const (
	rowTestTokyo      = "Asia/Tokyo"
	rowTestDaily      = "daily"
	rowTestIndefinite = "indefinite"
)

func pgText(s string) pgtype.Text { return pgtype.Text{String: s, Valid: true} }

func TestNewRecurringSeriesReadModelFromListRow_NullSeriesID_ReturnsNil(t *testing.T) {
	if got := newRecurringSeriesReadModelFromListRow(eventListRow{}); got != nil {
		t.Errorf("expected nil, got %+v", *got)
	}
}

func TestNewRecurringSeriesReadModelFromListRow_SetsOnlyNonNullOptionalColumns(t *testing.T) {
	seriesID := uuid.New()
	nth, untilDate, count := "last", "2027-03-31", 1

	tests := []struct {
		name string
		row  eventListRow
		want eventuc.RecurringSeriesReadModel
	}{
		{
			name: "count end",
			row: eventListRow{
				SeriesID: toPgUUID(seriesID), SeriesPatternType: pgText("weekly"),
				SeriesEndType: pgText("count"), SeriesEndCount: pgtype.Int2{Int16: 1, Valid: true},
				SeriesTimezone: pgText(rowTestTokyo),
			},
			want: eventuc.RecurringSeriesReadModel{
				ID: seriesID, PatternType: "weekly", EndType: "count", Count: &count, Timezone: rowTestTokyo,
			},
		},
		{
			name: "monthly nth weekday until date",
			row: eventListRow{
				SeriesID: toPgUUID(seriesID), SeriesPatternType: pgText("monthly_nth_weekday"),
				SeriesMonthlyNth: pgText("last"), SeriesEndType: pgText("until_date"),
				SeriesEndUntilDate: pgtype.Date{Time: time.Date(2027, 3, 31, 0, 0, 0, 0, time.UTC), Valid: true},
				SeriesTimezone:     pgText("UTC"),
			},
			want: eventuc.RecurringSeriesReadModel{
				ID: seriesID, PatternType: "monthly_nth_weekday", MonthlyNth: &nth,
				EndType: "until_date", UntilDate: &untilDate, Timezone: "UTC",
			},
		},
		{
			name: "indefinite end",
			row: eventListRow{
				SeriesID: toPgUUID(seriesID), SeriesPatternType: pgText(rowTestDaily),
				SeriesEndType: pgText(rowTestIndefinite), SeriesTimezone: pgText("UTC"),
			},
			want: eventuc.RecurringSeriesReadModel{
				ID: seriesID, PatternType: rowTestDaily, EndType: rowTestIndefinite, Timezone: "UTC",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := newRecurringSeriesReadModelFromListRow(tt.row)
			if got == nil {
				t.Fatal("expected a read model, got nil")
			}
			if got.ID != tt.want.ID || got.PatternType != tt.want.PatternType ||
				got.EndType != tt.want.EndType || got.Timezone != tt.want.Timezone {
				t.Errorf("mismatch: got %+v, want %+v", *got, tt.want)
			}
			assertSamePtr(t, "MonthlyNth", got.MonthlyNth, tt.want.MonthlyNth)
			assertSamePtr(t, "UntilDate", got.UntilDate, tt.want.UntilDate)
			assertSamePtr(t, "Count", got.Count, tt.want.Count)
		})
	}
}

func assertSamePtr[T comparable](t *testing.T, name string, got, want *T) {
	t.Helper()
	switch {
	case got == nil && want == nil:
	case got == nil || want == nil:
		t.Errorf("%s mismatch: got %v, want %v", name, got, want)
	case *got != *want:
		t.Errorf("%s mismatch: got %v, want %v", name, *got, *want)
	}
}
