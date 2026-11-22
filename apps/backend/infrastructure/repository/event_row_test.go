//nolint:testpackage // eventFromRow is unexported row mapping
package repository

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	query "github.com/Haya372/ai-trial/backend/infrastructure/db/generated"
)

func eventRowForTest(seriesID pgtype.UUID) query.InsertEventRow {
	start := time.Date(2026, 11, 22, 1, 0, 0, 0, time.UTC)
	return query.InsertEventRow{
		ID:       toPgUUID(uuid.New()),
		UserID:   toPgUUID(uuid.New()),
		Title:    "Standup",
		StartAt:  pgtype.Timestamptz{Time: start, Valid: true},
		EndAt:    pgtype.Timestamptz{Time: start.Add(time.Hour), Valid: true},
		SeriesID: seriesID,
	}
}

func TestEventFromRow_WithSeriesID_RestoresSeries(t *testing.T) {
	seriesID := uuid.New()

	e, err := eventFromRow(eventRowForTest(toPgUUID(seriesID)), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got, ok := e.SeriesID(); !ok || got != seriesID {
		t.Errorf("series mismatch: got %v (ok=%v), want %v", got, ok, seriesID)
	}
}

func TestEventFromRow_NullSeriesID_IsStandalone(t *testing.T) {
	e, err := eventFromRow(eventRowForTest(pgtype.UUID{}), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got, ok := e.SeriesID(); ok {
		t.Errorf("expected no series, got %v", got)
	}
}
