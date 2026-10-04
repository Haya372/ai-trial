package usecase

import "time"

// Clock abstracts the current time so time-dependent logic (e.g. session
// expiry) can be controlled from tests instead of depending on the real
// wall clock.
type Clock interface {
	Now() time.Time
}

// RealClock is the production Clock backed by the system wall clock.
type RealClock struct{}

func (RealClock) Now() time.Time { return time.Now() }
