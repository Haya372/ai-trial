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

// requireUTCDateTimes applies parseUTCDateTime to datetime fields of a JSON request body.
// encoding/json decodes any RFC 3339 offset into time.Time, so the body has to be
// re-read with these fields as strings to enforce ADR-036. nil means the field is absent.
func requireUTCDateTimes(values ...*string) error {
	for _, v := range values {
		if v == nil {
			continue
		}
		if _, err := parseUTCDateTime(*v); err != nil {
			return err
		}
	}
	return nil
}
