package eventshare_test

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
	domaineventshare "github.com/Haya372/ai-trial/backend/domain/eventshare"
	eventsharemock "github.com/Haya372/ai-trial/backend/domain/eventshare/generated"
	eventshareuc "github.com/Haya372/ai-trial/backend/usecase/eventshare"
)

func TestCreateShareCommand_Execute_ValidInput_CreatesAndReturnsShareWithToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	eventRepo := eventmock.NewMockRepository(ctrl)
	shareRepo := eventsharemock.NewMockRepository(ctrl)

	userID := uuid.New()
	eventID := uuid.New()
	start := time.Now().UTC()
	end := start.Add(time.Hour)
	ev, _ := domainevent.New(eventID, userID, "Meeting", "", start, end, "", "")

	eventRepo.EXPECT().FindByID(gomock.Any(), eventID).Return(ev, nil)
	shareRepo.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, s domaineventshare.EventShare) (domaineventshare.EventShare, error) {
			if s.EventID() != eventID {
				t.Errorf("EventID mismatch: got %v, want %v", s.EventID(), eventID)
			}
			if !s.ExpiresAt().Equal(end) {
				t.Errorf("ExpiresAt mismatch: got %v, want event end %v", s.ExpiresAt(), end)
			}
			return s, nil
		},
	)

	cmd := eventshareuc.NewCreateShareCommand(eventRepo, shareRepo, testLogger)
	out, err := cmd.Execute(context.Background(), userID, eventshareuc.CreateShareInput{EventID: eventID})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Token == "" {
		t.Error("expected a non-empty plaintext token")
	}
	if out.Share.TokenHash() != domaineventshare.HashToken(out.Token) {
		t.Error("stored token hash does not match the returned plaintext token")
	}
}

func TestCreateShareCommand_Execute_WithExpiresAtOverride_UsesGivenExpiresAt(t *testing.T) {
	ctrl := gomock.NewController(t)
	eventRepo := eventmock.NewMockRepository(ctrl)
	shareRepo := eventsharemock.NewMockRepository(ctrl)

	userID := uuid.New()
	eventID := uuid.New()
	start := time.Now().UTC()
	end := start.Add(time.Hour)
	override := end.Add(48 * time.Hour)
	ev, _ := domainevent.New(eventID, userID, "Meeting", "", start, end, "", "")

	eventRepo.EXPECT().FindByID(gomock.Any(), eventID).Return(ev, nil)
	shareRepo.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, s domaineventshare.EventShare) (domaineventshare.EventShare, error) {
			if !s.ExpiresAt().Equal(override) {
				t.Errorf("ExpiresAt mismatch: got %v, want override %v", s.ExpiresAt(), override)
			}
			return s, nil
		},
	)

	cmd := eventshareuc.NewCreateShareCommand(eventRepo, shareRepo, testLogger)
	_, err := cmd.Execute(context.Background(), userID, eventshareuc.CreateShareInput{
		EventID: eventID, ExpiresAt: &override,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCreateShareCommand_Execute_EventNotFound_ReturnsEventNotFoundError(t *testing.T) {
	ctrl := gomock.NewController(t)
	eventRepo := eventmock.NewMockRepository(ctrl)
	shareRepo := eventsharemock.NewMockRepository(ctrl)

	eventID := uuid.New()
	eventRepo.EXPECT().FindByID(gomock.Any(), eventID).Return(nil, domainevent.ErrEventNotFound)

	cmd := eventshareuc.NewCreateShareCommand(eventRepo, shareRepo, testLogger)
	_, err := cmd.Execute(context.Background(), uuid.New(), eventshareuc.CreateShareInput{EventID: eventID})
	if !errors.Is(err, domainevent.ErrEventNotFound) {
		t.Errorf("expected ErrEventNotFound, got %v", err)
	}
}

func TestCreateShareCommand_Execute_NotOwner_ReturnsForbiddenError(t *testing.T) {
	ctrl := gomock.NewController(t)
	eventRepo := eventmock.NewMockRepository(ctrl)
	shareRepo := eventsharemock.NewMockRepository(ctrl)

	ownerID := uuid.New()
	otherUserID := uuid.New()
	eventID := uuid.New()
	start := time.Now().UTC()
	end := start.Add(time.Hour)
	ev, _ := domainevent.New(eventID, ownerID, "Meeting", "", start, end, "", "")

	eventRepo.EXPECT().FindByID(gomock.Any(), eventID).Return(ev, nil)

	cmd := eventshareuc.NewCreateShareCommand(eventRepo, shareRepo, testLogger)
	_, err := cmd.Execute(context.Background(), otherUserID, eventshareuc.CreateShareInput{EventID: eventID})
	if !errors.Is(err, domainevent.ErrEventForbidden) {
		t.Errorf("expected ErrEventForbidden, got %v", err)
	}
}

func TestCreateShareCommand_Execute_ExpiresAtBeforeEventStart_ReturnsValidationError(t *testing.T) {
	ctrl := gomock.NewController(t)
	eventRepo := eventmock.NewMockRepository(ctrl)
	shareRepo := eventsharemock.NewMockRepository(ctrl)

	userID := uuid.New()
	eventID := uuid.New()
	start := time.Now().UTC()
	end := start.Add(time.Hour)
	tooEarly := start.Add(-time.Minute)
	ev, _ := domainevent.New(eventID, userID, "Meeting", "", start, end, "", "")

	eventRepo.EXPECT().FindByID(gomock.Any(), eventID).Return(ev, nil)

	cmd := eventshareuc.NewCreateShareCommand(eventRepo, shareRepo, testLogger)
	_, err := cmd.Execute(context.Background(), userID, eventshareuc.CreateShareInput{
		EventID: eventID, ExpiresAt: &tooEarly,
	})
	var ve *domain.ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected *domain.ValidationError, got %T: %v", err, err)
	}
}

func TestCreateShareCommand_Execute_RepoCreateError_ReturnsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	eventRepo := eventmock.NewMockRepository(ctrl)
	shareRepo := eventsharemock.NewMockRepository(ctrl)

	userID := uuid.New()
	eventID := uuid.New()
	start := time.Now().UTC()
	end := start.Add(time.Hour)
	ev, _ := domainevent.New(eventID, userID, "Meeting", "", start, end, "", "")

	eventRepo.EXPECT().FindByID(gomock.Any(), eventID).Return(ev, nil)
	shareRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil, errDBFailure)

	cmd := eventshareuc.NewCreateShareCommand(eventRepo, shareRepo, testLogger)
	_, err := cmd.Execute(context.Background(), userID, eventshareuc.CreateShareInput{EventID: eventID})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
