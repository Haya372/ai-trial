package eventshare_test

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/google/uuid"

	domainevent "github.com/Haya372/ai-trial/backend/domain/event"
	domaineventshare "github.com/Haya372/ai-trial/backend/domain/eventshare"
	"github.com/Haya372/ai-trial/backend/domain/user"
)

var (
	errDBFailure = errors.New("db error")
	testLogger   = slog.New(slog.DiscardHandler)
)

// failingTxManager returns errDBFailure without ever invoking fn, simulating
// a transaction that fails to start.
type failingTxManager struct{}

func (f *failingTxManager) RunInTx(context.Context, func(context.Context) error) error {
	return errDBFailure
}

func newTime() time.Time { return time.Now() }

func newTestShare(eventID uuid.UUID, expired bool) domaineventshare.EventShare {
	expiresAt := time.Now().Add(time.Hour)
	if expired {
		expiresAt = time.Now().Add(-time.Hour)
	}
	s, err := domaineventshare.New(uuid.New(), eventID, "hash", expiresAt)
	if err != nil {
		panic(err)
	}
	return s
}

func newTestEvent(ownerID uuid.UUID) domainevent.Event {
	ev, err := domainevent.New(
		uuid.New(), ownerID, "Test Event", "", time.Now(), time.Now().Add(time.Hour), "", "",
	)
	if err != nil {
		panic(err)
	}
	return ev
}

func newTestUser(id uuid.UUID) user.User {
	email, _ := user.NewEmail("u@ex.com")
	return user.New(id, email, "U", "hash")
}
