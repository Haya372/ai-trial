package event_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/Haya372/ai-trial/backend/domain"
	domainevent "github.com/Haya372/ai-trial/backend/domain/event"
	eventuc "github.com/Haya372/ai-trial/backend/usecase/event"
)

const testCreateTitle = "Meeting"

// stubCreateRepo is a stub for domain/event.Repository used in create_event tests.
type stubCreateRepo struct {
	createFn func(ctx context.Context, e domainevent.Event) (domainevent.Event, error)
}

func (s *stubCreateRepo) Create(ctx context.Context, e domainevent.Event) (domainevent.Event, error) {
	if s.createFn == nil {
		// By default, echo back the same event (simulates successful persistence).
		return e, nil
	}
	return s.createFn(ctx, e)
}

func (s *stubCreateRepo) FindByID(_ context.Context, _ uuid.UUID) (domainevent.Event, error) {
	return nil, nil
}

func (s *stubCreateRepo) Update(_ context.Context, _ domainevent.Event) error {
	return nil
}

func TestCreateEventCommand_Execute_Success(t *testing.T) {
	userID := uuid.New()
	now := time.Now().UTC().Truncate(time.Second)
	startAt := now
	endAt := now.Add(time.Hour)

	var capturedEvent domainevent.Event
	repo := &stubCreateRepo{
		createFn: func(_ context.Context, e domainevent.Event) (domainevent.Event, error) {
			capturedEvent = e
			return e, nil
		},
	}

	cmd := eventuc.NewCreateEventCommand(repo)
	result, err := cmd.Execute(context.Background(), userID, eventuc.CreateEventInput{
		Title:       testCreateTitle,
		Description: "Team sync",
		StartAt:     startAt,
		EndAt:       endAt,
		Location:    testLocation,
		URL:         testURL,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedEvent == nil {
		t.Fatal("expected event passed to repository, got nil")
	}
	if capturedEvent.UserID() != userID {
		t.Errorf("userID mismatch: got %v, want %v", capturedEvent.UserID(), userID)
	}
	if capturedEvent.Title() != testCreateTitle {
		t.Errorf("title mismatch: got %q", capturedEvent.Title())
	}
	if result.Title() != testCreateTitle {
		t.Errorf("result title mismatch: got %q", result.Title())
	}
	if result.Description() != "Team sync" {
		t.Errorf("result description mismatch: got %q", result.Description())
	}
	if result.Location() != testLocation {
		t.Errorf("result location mismatch: got %q", result.Location())
	}
	if result.URL() != testURL {
		t.Errorf("result url mismatch: got %q", result.URL())
	}
	if result.UserID() != userID {
		t.Errorf("result userID mismatch: got %v, want %v", result.UserID(), userID)
	}
}

func TestCreateEventCommand_Execute_EmptyTitle_ReturnsValidationError(t *testing.T) {
	userID := uuid.New()
	now := time.Now().UTC()
	repo := &stubCreateRepo{}

	cmd := eventuc.NewCreateEventCommand(repo)
	_, err := cmd.Execute(context.Background(), userID, eventuc.CreateEventInput{
		Title:   "",
		StartAt: now,
		EndAt:   now.Add(time.Hour),
	})

	if err == nil {
		t.Fatal("expected validation error for empty title, got nil")
	}
	var ve *domain.ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected *domain.ValidationError, got %T: %v", err, err)
	}
	if len(ve.Details) == 0 || ve.Details[0].Field != "title" {
		t.Errorf("expected validation error on field title, got: %v", ve.Details)
	}
	if ve.Details[0].Code != "REQUIRED" {
		t.Errorf("expected code REQUIRED, got %q", ve.Details[0].Code)
	}
}

func TestCreateEventCommand_Execute_EndAtEqualStartAt_ReturnsValidationError(t *testing.T) {
	userID := uuid.New()
	now := time.Now().UTC()
	repo := &stubCreateRepo{}

	cmd := eventuc.NewCreateEventCommand(repo)
	_, err := cmd.Execute(context.Background(), userID, eventuc.CreateEventInput{
		Title:   testCreateTitle,
		StartAt: now,
		EndAt:   now, // equal: invalid per SPEC-003
	})

	if err == nil {
		t.Fatal("expected validation error for end == start, got nil")
	}
	var ve *domain.ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected *domain.ValidationError, got %T", err)
	}
	if len(ve.Details) == 0 || ve.Details[0].Field != "endAt" {
		t.Errorf("expected validation error on field endAt, got: %v", ve.Details)
	}
	if ve.Details[0].Code != "INVALID_DATE_RANGE" {
		t.Errorf("expected code INVALID_DATE_RANGE, got %q", ve.Details[0].Code)
	}
}

func TestCreateEventCommand_Execute_RepositoryError_WrapsError(t *testing.T) {
	userID := uuid.New()
	now := time.Now().UTC()
	repo := &stubCreateRepo{
		createFn: func(_ context.Context, _ domainevent.Event) (domainevent.Event, error) {
			return nil, errDBFailure
		},
	}

	cmd := eventuc.NewCreateEventCommand(repo)
	_, err := cmd.Execute(context.Background(), userID, eventuc.CreateEventInput{
		Title:   testCreateTitle,
		StartAt: now,
		EndAt:   now.Add(time.Hour),
	})

	if err == nil {
		t.Fatal("expected error from repository, got nil")
	}
	if !errors.Is(err, errDBFailure) {
		t.Errorf("expected wrapped errDBFailure, got %v", err)
	}
}
