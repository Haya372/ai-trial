package eventshare_test

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/google/uuid"

	domaineventshare "github.com/Haya372/ai-trial/backend/domain/eventshare"
	"github.com/Haya372/ai-trial/backend/domain/user"
	eventuc "github.com/Haya372/ai-trial/backend/usecase/event"
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
	s, err := domaineventshare.New(uuid.New(), eventID, domaineventshare.NewTokenHash("hash"), expiresAt)
	if err != nil {
		panic(err)
	}
	return s
}

func newTestEvent(ownerID uuid.UUID) eventuc.EventReadModel {
	now := time.Now()
	return eventuc.EventReadModel{
		ID:      uuid.New(),
		UserID:  ownerID,
		Title:   "Test Event",
		StartAt: now,
		EndAt:   now.Add(time.Hour),
	}
}

// stubEventQueryService implements eventuc.QueryService with only the method
// ShareTokenLoader needs (FindByID); List is unused by this package's tests.
type stubEventQueryService struct {
	findByIDFn func(context.Context, uuid.UUID) (eventuc.EventReadModel, error)
}

func (s *stubEventQueryService) List(
	context.Context, eventuc.ListFilter,
) ([]eventuc.EventReadModel, error) {
	return nil, nil
}

func (s *stubEventQueryService) ListByIDs(
	context.Context, []uuid.UUID, time.Time, time.Time,
) ([]eventuc.EventReadModel, error) {
	return nil, nil
}

func (s *stubEventQueryService) FindByID(ctx context.Context, id uuid.UUID) (eventuc.EventReadModel, error) {
	if s.findByIDFn == nil {
		return eventuc.EventReadModel{}, nil
	}
	return s.findByIDFn(ctx, id)
}

func newTestUser(id uuid.UUID) user.User {
	email, _ := user.NewEmail("u@ex.com")
	return user.New(id, email, "U", "hash")
}
