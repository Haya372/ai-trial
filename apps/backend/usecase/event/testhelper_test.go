package event_test

import (
	"context"
	"errors"
	"log/slog"
)

var (
	errDBFailure = errors.New("db error")
	testLogger   = slog.New(slog.DiscardHandler)
)

const (
	testLocation = "Tokyo"
	testURL      = "https://example.com"

	// Error codes shared across Create and Recurrence tests.
	codeRequired          = "REQUIRED"
	codeInvalidFormat     = "INVALID_FORMAT"
	codeOutOfRange        = "OUT_OF_RANGE"
	codeTooFewOccurrences = "TOO_FEW_OCCURRENCES"

	// Recurrence pattern and end types used in multiple recurrence tests.
	patternWeekly     = "weekly"
	endCount          = "count"
	endUntilDate      = "until_date"
	timezoneUTC       = "UTC"
	fieldUntilDateEnd = "recurrence.end.untilDate"
)

// ptr is a generic helper that returns a pointer to v, allocated on the heap.
func ptr[T any](v T) *T {
	p := new(T)
	*p = v
	return p
}

type inTxKey struct{}

// markingTxManager marks the context passed to fn, so a repository mock can
// assert that it was called inside RunInTx.
type markingTxManager struct{}

func (markingTxManager) RunInTx(ctx context.Context, fn func(context.Context) error) error {
	return fn(context.WithValue(ctx, inTxKey{}, true))
}

func inTx(ctx context.Context) bool {
	marked, _ := ctx.Value(inTxKey{}).(bool)
	return marked
}
