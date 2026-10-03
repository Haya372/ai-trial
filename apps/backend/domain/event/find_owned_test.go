package event_test

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"

	"github.com/Haya372/ai-trial/backend/domain/event"
	eventmock "github.com/Haya372/ai-trial/backend/domain/event/generated"
)

var testLogger = slog.New(slog.DiscardHandler)

func TestFindOwned_Owner_ReturnsEvent(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := eventmock.NewMockRepository(ctrl)

	userID := uuid.New()
	eventID := uuid.New()
	start := time.Now().UTC()
	end := start.Add(time.Hour)
	ev, _ := event.New(eventID, userID, "Meeting", "", start, end, "", "")

	repo.EXPECT().FindByID(gomock.Any(), eventID).Return(ev, nil)

	got, err := event.FindOwned(context.Background(), repo, testLogger, eventID, userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID() != eventID {
		t.Errorf("ID mismatch: got %v, want %v", got.ID(), eventID)
	}
}

func TestFindOwned_NotFound_ReturnsEventNotFoundError(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := eventmock.NewMockRepository(ctrl)

	eventID := uuid.New()
	repo.EXPECT().FindByID(gomock.Any(), eventID).Return(nil, event.ErrEventNotFound)

	_, err := event.FindOwned(context.Background(), repo, testLogger, eventID, uuid.New())
	if !errors.Is(err, event.ErrEventNotFound) {
		t.Errorf("expected ErrEventNotFound, got %v", err)
	}
}

func TestFindOwned_NotOwner_ReturnsEventNotFoundError(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := eventmock.NewMockRepository(ctrl)

	ownerID := uuid.New()
	otherUserID := uuid.New()
	eventID := uuid.New()
	start := time.Now().UTC()
	end := start.Add(time.Hour)
	ev, _ := event.New(eventID, ownerID, "Meeting", "", start, end, "", "")

	repo.EXPECT().FindByID(gomock.Any(), eventID).Return(ev, nil)

	// Ownership mismatches are reported the same way as a missing event
	// to avoid leaking event existence to non-owners.
	_, err := event.FindOwned(context.Background(), repo, testLogger, eventID, otherUserID)
	if !errors.Is(err, event.ErrEventNotFound) {
		t.Errorf("expected ErrEventNotFound, got %v", err)
	}
}
