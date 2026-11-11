package handler

import (
	"errors"
	"strings"
	"time"
)

var errNotUTCDateTime = errors.New("datetime must be RFC 3339 in UTC ending with Z")

// parseUTCDateTime accepts only RFC 3339 strings ending with "Z" so that the API
// carries a single datetime notation (ADR-036), even for offsets equivalent to UTC.
func parseUTCDateTime(s string) (time.Time, error) {
	if !strings.HasSuffix(s, "Z") {
		return time.Time{}, errNotUTCDateTime
	}
	t, err := time.Parse(time.RFC3339Nano, s)
	if err != nil {
		return time.Time{}, errNotUTCDateTime
	}
	return t, nil
}
