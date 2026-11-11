package db_test

import (
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/Haya372/ai-trial/backend/infrastructure/db"
)

func TestRegisterUTCTimestamptz_scansInUTC(t *testing.T) {
	m := pgtype.NewMap()
	db.RegisterUTCTimestamptz(m)

	tests := []struct {
		name   string
		format int16
		src    []byte
	}{
		{"text", pgtype.TextFormatCode, []byte("2026-11-01 05:00:00+09")},
		{"binary", pgtype.BinaryFormatCode, encodeTimestamptz(t, time.Date(2026, 10, 31, 20, 0, 0, 0, time.UTC))},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got time.Time
			if err := m.Scan(pgtype.TimestamptzOID, tt.format, tt.src, &got); err != nil {
				t.Fatalf("scan: %v", err)
			}
			if got.Location() != time.UTC {
				t.Errorf("location mismatch: got %v, want UTC", got.Location())
			}
			want := time.Date(2026, 10, 31, 20, 0, 0, 0, time.UTC)
			if !got.Equal(want) {
				t.Errorf("time mismatch: got %v, want %v", got, want)
			}
		})
	}
}

func encodeTimestamptz(t *testing.T, v time.Time) []byte {
	t.Helper()
	buf, err := pgtype.NewMap().Encode(pgtype.TimestamptzOID, pgtype.BinaryFormatCode, v, nil)
	if err != nil {
		t.Fatalf("encode timestamptz: %v", err)
	}
	return buf
}
