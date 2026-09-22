package event_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"

	domainevent "github.com/Haya372/ai-trial/backend/domain/event"
	eventmock "github.com/Haya372/ai-trial/backend/domain/event/generated"
	eventuc "github.com/Haya372/ai-trial/backend/usecase/event"
)

func TestDeleteEventCommand_Execute_ValidInput_DeletesEvent(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := eventmock.NewMockRepository(ctrl)

	userID := uuid.New()
	eventID := uuid.New()
	start := time.Now().UTC()
	end := start.Add(time.Hour)

	existing, _ := domainevent.New(eventID, userID, "Title", "", start, end, "", "")

	repo.EXPECT().FindByID(gomock.Any(), eventID).Return(existing, nil)
	repo.EXPECT().Delete(gomock.Any(), eventID).Return(nil)

	cmd := eventuc.NewDeleteEventCommand(repo, testLogger)
	err := cmd.Execute(context.Background(), userID, eventID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteEventCommand_Execute_EventNotFound_ReturnsEventNotFoundError(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := eventmock.NewMockRepository(ctrl)

	userID := uuid.New()
	eventID := uuid.New()

	repo.EXPECT().FindByID(gomock.Any(), eventID).Return(nil, domainevent.ErrEventNotFound)

	cmd := eventuc.NewDeleteEventCommand(repo, testLogger)
	err := cmd.Execute(context.Background(), userID, eventID)
	if !errors.Is(err, domainevent.ErrEventNotFound) {
		t.Errorf("expected ErrEventNotFound, got %v", err)
	}
}

func TestDeleteEventCommand_Execute_NotOwner_ReturnsEventNotFoundError(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := eventmock.NewMockRepository(ctrl)

	ownerID := uuid.New()
	otherUserID := uuid.New()
	eventID := uuid.New()
	start := time.Now().UTC()
	end := start.Add(time.Hour)

	existing, _ := domainevent.New(eventID, ownerID, "Title", "", start, end, "", "")
	repo.EXPECT().FindByID(gomock.Any(), eventID).Return(existing, nil)

	cmd := eventuc.NewDeleteEventCommand(repo, testLogger)
	err := cmd.Execute(context.Background(), otherUserID, eventID)
	// Ownership mismatches are reported the same way as a missing event
	// (ErrEventNotFound) to avoid leaking event existence to non-owners,
	// consistent with UpdateEventCommand.
	if !errors.Is(err, domainevent.ErrEventNotFound) {
		t.Errorf("expected ErrEventNotFound, got %v", err)
	}
}

func TestDeleteEventCommand_Execute_PastEvent_AllowsDelete(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := eventmock.NewMockRepository(ctrl)

	userID := uuid.New()
	eventID := uuid.New()
	past := time.Now().Add(-48 * time.Hour)
	existing, _ := domainevent.New(eventID, userID, "Past event", "", past, past.Add(time.Hour), "", "")

	repo.EXPECT().FindByID(gomock.Any(), eventID).Return(existing, nil)
	repo.EXPECT().Delete(gomock.Any(), eventID).Return(nil)

	cmd := eventuc.NewDeleteEventCommand(repo, testLogger)
	err := cmd.Execute(context.Background(), userID, eventID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteEventCommand_Execute_RepoDeleteError_ReturnsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := eventmock.NewMockRepository(ctrl)

	userID := uuid.New()
	eventID := uuid.New()
	start := time.Now().UTC()
	end := start.Add(time.Hour)
	existing, _ := domainevent.New(eventID, userID, "Title", "", start, end, "", "")

	repo.EXPECT().FindByID(gomock.Any(), eventID).Return(existing, nil)
	repo.EXPECT().Delete(gomock.Any(), eventID).Return(errDBFailure)

	cmd := eventuc.NewDeleteEventCommand(repo, testLogger)
	err := cmd.Execute(context.Background(), userID, eventID)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
