package recurringeventseries

import (
	"fmt"
	"time"
)

// SeriesTimezone is the IANA time zone of the device that created the
// series. Occurrence wall-clock times and excluded dates are interpreted in
// it (ADR-037).
type SeriesTimezone struct {
	name string
	loc  *time.Location
}

func NewSeriesTimezone(name string) (SeriesTimezone, error) {
	// time.LoadLocation maps "" to UTC and "Local" to the server's zone, which
	// would make occurrences depend on where the server runs.
	if name == "" || name == "Local" {
		return SeriesTimezone{}, fmt.Errorf("%w: %q", ErrRecurringTimezoneInvalid, name)
	}
	loc, err := time.LoadLocation(name)
	if err != nil {
		return SeriesTimezone{}, fmt.Errorf("%w: %q", ErrRecurringTimezoneInvalid, name)
	}
	return SeriesTimezone{name: name, loc: loc}, nil
}

func (z SeriesTimezone) Name() string             { return z.name }
func (z SeriesTimezone) Location() *time.Location { return z.loc }
