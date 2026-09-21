package event_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"

	"github.com/Haya372/ai-trial/backend/domain"
	domainevent "github.com/Haya372/ai-trial/backend/domain/event"
	eventmock "github.com/Haya372/ai-trial/backend/domain/event/generated"
	eventuc "github.com/Haya372/ai-trial/backend/usecase/event"
)

const newTitle = "New title"

func TestUpdateEventCommand_Execute_ValidInput_UpdatesAndReturnsEvent(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := eventmock.NewMockRepository(ctrl)

	userID := uuid.New()
	eventID := uuid.New()
	start := time.Now().UTC()
	end := start.Add(time.Hour)

	existing, _ := domainevent.New(eventID, userID, "Old title", "old desc", start, end, "", "")

	repo.EXPECT().FindByID(gomock.Any(), eventID).Return(existing, nil)
	repo.EXPECT().Update(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, e domainevent.Event) error {
			if e.Title() != newTitle {
				t.Errorf("title mismatch: got %q", e.Title())
			}
			return nil
		},
	)

	cmd := eventuc.NewUpdateEventCommand(repo, testLogger)
	out, err := cmd.Execute(context.Background(), userID, eventuc.UpdateEventInput{
		ID:          eventID,
		Title:       newTitle,
		Description: "new desc",
		StartAt:     start,
		EndAt:       end,
		Location:    "Tokyo",
		URL:         "https://example.com",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Title() != newTitle {
		t.Errorf("title mismatch: got %q", out.Title())
	}
	if out.Location() != "Tokyo" {
		t.Errorf("location mismatch: got %q", out.Location())
	}
	if out.URL() != "https://example.com" {
		t.Errorf("url mismatch: got %q", out.URL())
	}
}

func TestUpdateEventCommand_Execute_EndAtBeforeStartAt_ReturnsValidationError(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := eventmock.NewMockRepository(ctrl)

	userID := uuid.New()
	eventID := uuid.New()
	start := time.Now().UTC()
	end := start.Add(time.Hour)

	existing, _ := domainevent.New(eventID, userID, "Old title", "", start, end, "", "")
	repo.EXPECT().FindByID(gomock.Any(), eventID).Return(existing, nil)

	cmd := eventuc.NewUpdateEventCommand(repo, testLogger)
	_, err := cmd.Execute(context.Background(), userID, eventuc.UpdateEventInput{
		ID:      eventID,
		Title:   newTitle,
		StartAt: end,
		EndAt:   start,
	})
	var ve *domain.ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected *ValidationError, got %T: %v", err, err)
	}
}

func TestUpdateEventCommand_Execute_EmptyTitle_ReturnsValidationError(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := eventmock.NewMockRepository(ctrl)

	userID := uuid.New()
	eventID := uuid.New()
	start := time.Now().UTC()
	end := start.Add(time.Hour)

	existing, _ := domainevent.New(eventID, userID, "Old title", "", start, end, "", "")
	repo.EXPECT().FindByID(gomock.Any(), eventID).Return(existing, nil)

	cmd := eventuc.NewUpdateEventCommand(repo, testLogger)
	_, err := cmd.Execute(context.Background(), userID, eventuc.UpdateEventInput{
		ID:      eventID,
		Title:   "",
		StartAt: start,
		EndAt:   end,
	})
	var ve *domain.ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected *ValidationError, got %T: %v", err, err)
	}
}

func TestUpdateEventCommand_Execute_EventNotFound_ReturnsEventNotFoundError(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := eventmock.NewMockRepository(ctrl)

	userID := uuid.New()
	eventID := uuid.New()

	repo.EXPECT().FindByID(gomock.Any(), eventID).Return(nil, domainevent.ErrEventNotFound)

	cmd := eventuc.NewUpdateEventCommand(repo, testLogger)
	_, err := cmd.Execute(context.Background(), userID, eventuc.UpdateEventInput{
		ID:      eventID,
		Title:   newTitle,
		StartAt: time.Now(),
		EndAt:   time.Now().Add(time.Hour),
	})
	if !errors.Is(err, domainevent.ErrEventNotFound) {
		t.Errorf("expected ErrEventNotFound, got %v", err)
	}
}

func TestUpdateEventCommand_Execute_NotOwner_ReturnsEventNotFoundError(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := eventmock.NewMockRepository(ctrl)

	ownerID := uuid.New()
	otherUserID := uuid.New()
	eventID := uuid.New()
	start := time.Now().UTC()
	end := start.Add(time.Hour)

	existing, _ := domainevent.New(eventID, ownerID, "Title", "", start, end, "", "")
	repo.EXPECT().FindByID(gomock.Any(), eventID).Return(existing, nil)

	cmd := eventuc.NewUpdateEventCommand(repo, testLogger)
	_, err := cmd.Execute(context.Background(), otherUserID, eventuc.UpdateEventInput{
		ID:      eventID,
		Title:   newTitle,
		StartAt: start,
		EndAt:   end,
	})
	// Ownership mismatches are reported the same way as a missing event
	// (ErrEventNotFound) to avoid leaking event existence to non-owners.
	if !errors.Is(err, domainevent.ErrEventNotFound) {
		t.Errorf("expected ErrEventNotFound, got %v", err)
	}
}

func TestUpdateEventCommand_Execute_PastEvent_AllowsUpdate(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := eventmock.NewMockRepository(ctrl)

	userID := uuid.New()
	eventID := uuid.New()
	past := time.Now().Add(-48 * time.Hour)
	existing, _ := domainevent.New(eventID, userID, "Old title", "", past, past.Add(time.Hour), "", "")

	repo.EXPECT().FindByID(gomock.Any(), eventID).Return(existing, nil)
	repo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)

	cmd := eventuc.NewUpdateEventCommand(repo, testLogger)
	_, err := cmd.Execute(context.Background(), userID, eventuc.UpdateEventInput{
		ID:      eventID,
		Title:   "Updated past event",
		StartAt: past,
		EndAt:   past.Add(2 * time.Hour),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUpdateEventCommand_Execute_RepoUpdateError_ReturnsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := eventmock.NewMockRepository(ctrl)

	userID := uuid.New()
	eventID := uuid.New()
	start := time.Now().UTC()
	end := start.Add(time.Hour)
	existing, _ := domainevent.New(eventID, userID, "Old title", "", start, end, "", "")

	repo.EXPECT().FindByID(gomock.Any(), eventID).Return(existing, nil)
	repo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(errDBFailure)

	cmd := eventuc.NewUpdateEventCommand(repo, testLogger)
	_, err := cmd.Execute(context.Background(), userID, eventuc.UpdateEventInput{
		ID:      eventID,
		Title:   newTitle,
		StartAt: start,
		EndAt:   end,
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
