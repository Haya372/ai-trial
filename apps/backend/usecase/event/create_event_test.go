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
	"github.com/Haya372/ai-trial/backend/domain/tag"
	tagmock "github.com/Haya372/ai-trial/backend/domain/tag/generated"
	eventuc "github.com/Haya372/ai-trial/backend/usecase/event"
	"github.com/Haya372/ai-trial/backend/usecase/testutil"
)

const (
	testCreateTitle = "Meeting"
	fieldTagIDs     = "tagIds"
	codeNotFound    = "NOT_FOUND"
	codeTooMany     = "TOO_MANY"
	codeDuplicated  = "DUPLICATED"
)

func TestCreateEventCommand_Execute_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := eventmock.NewMockRepository(ctrl)

	userID := uuid.New()
	now := time.Now().UTC().Truncate(time.Second)
	startAt := now
	endAt := now.Add(time.Hour)

	var capturedEvent domainevent.Event
	repo.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, e domainevent.Event) (domainevent.Event, error) {
			capturedEvent = e
			return e, nil
		},
	)

	cmd := eventuc.NewCreateEventCommand(repo, tagmock.NewMockRepository(ctrl), &testutil.StubTxManager{})
	out, err := cmd.Execute(context.Background(), userID, eventuc.CreateEventInput{
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
	if out.Event.Title() != testCreateTitle {
		t.Errorf("result title mismatch: got %q", out.Event.Title())
	}
	if out.Event.Description() != "Team sync" {
		t.Errorf("result description mismatch: got %q", out.Event.Description())
	}
	if out.Event.Location() != testLocation {
		t.Errorf("result location mismatch: got %q", out.Event.Location())
	}
	if out.Event.URL() != testURL {
		t.Errorf("result url mismatch: got %q", out.Event.URL())
	}
	if out.Event.UserID() != userID {
		t.Errorf("result userID mismatch: got %v, want %v", out.Event.UserID(), userID)
	}
	if len(out.Tags) != 0 {
		t.Errorf("expected empty tags for untagged event, got %d", len(out.Tags))
	}
}

func TestCreateEventCommand_Execute_EmptyTitle_ReturnsValidationError(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := eventmock.NewMockRepository(ctrl)
	// Create must not be called: validation fails before reaching the repository.

	userID := uuid.New()
	now := time.Now().UTC()

	cmd := eventuc.NewCreateEventCommand(repo, tagmock.NewMockRepository(ctrl), &testutil.StubTxManager{})
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
	ctrl := gomock.NewController(t)
	repo := eventmock.NewMockRepository(ctrl)
	// Create must not be called: validation fails before reaching the repository.

	userID := uuid.New()
	now := time.Now().UTC()

	cmd := eventuc.NewCreateEventCommand(repo, tagmock.NewMockRepository(ctrl), &testutil.StubTxManager{})
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
	ctrl := gomock.NewController(t)
	repo := eventmock.NewMockRepository(ctrl)

	userID := uuid.New()
	now := time.Now().UTC()
	repo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil, errDBFailure)

	cmd := eventuc.NewCreateEventCommand(repo, tagmock.NewMockRepository(ctrl), &testutil.StubTxManager{})
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

func TestCreateEventCommand_Execute_CreatesInsideTransaction(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := eventmock.NewMockRepository(ctrl)

	repo.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, e domainevent.Event) (domainevent.Event, error) {
			if !inTx(ctx) {
				t.Error("expected Create to be called inside a transaction")
			}
			return e, nil
		},
	)

	now := time.Now().UTC()
	cmd := eventuc.NewCreateEventCommand(repo, tagmock.NewMockRepository(ctrl), markingTxManager{})
	if _, err := cmd.Execute(context.Background(), uuid.New(), eventuc.CreateEventInput{
		Title:   testCreateTitle,
		StartAt: now,
		EndAt:   now.Add(time.Hour),
	}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCreateEventCommand_Execute_WithTags_ReturnsTagsInOrder(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := eventmock.NewMockRepository(ctrl)
	tagRepo := tagmock.NewMockRepository(ctrl)

	userID := uuid.New()
	t1, t2 := uuid.New(), uuid.New()
	now := time.Now().UTC()

	repo.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, e domainevent.Event) (domainevent.Event, error) {
			return e, nil
		},
	)
	tagRepo.EXPECT().FindByIDsForShare(gomock.Any(), []uuid.UUID{t1, t2}).Return(
		[]tag.Tag{newStoredTag(t, t1, userID), newStoredTag(t, t2, userID)},
		nil,
	)

	cmd := eventuc.NewCreateEventCommand(repo, tagRepo, &testutil.StubTxManager{})
	out, err := cmd.Execute(context.Background(), userID, eventuc.CreateEventInput{
		Title:   testCreateTitle,
		StartAt: now,
		EndAt:   now.Add(time.Hour),
		TagIDs:  []uuid.UUID{t1, t2},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out.Tags) != 2 {
		t.Fatalf("expected 2 tags, got %d", len(out.Tags))
	}
	if out.Tags[0].ID != t1 {
		t.Errorf("tags[0] ID mismatch: got %v, want %v", out.Tags[0].ID, t1)
	}
	if out.Tags[1].ID != t2 {
		t.Errorf("tags[1] ID mismatch: got %v, want %v", out.Tags[1].ID, t2)
	}
}

func TestCreateEventCommand_Execute_TooManyTags_ReturnsValidationError(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := eventmock.NewMockRepository(ctrl)
	// Create must not be called when tagIds validation fails.
	tagRepo := tagmock.NewMockRepository(ctrl)

	userID := uuid.New()
	now := time.Now().UTC()
	tagIDs := []uuid.UUID{uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New()} // 6 tags

	cmd := eventuc.NewCreateEventCommand(repo, tagRepo, &testutil.StubTxManager{})
	_, err := cmd.Execute(context.Background(), userID, eventuc.CreateEventInput{
		Title:   testCreateTitle,
		StartAt: now,
		EndAt:   now.Add(time.Hour),
		TagIDs:  tagIDs,
	})
	if err == nil {
		t.Fatal("expected validation error for too many tags, got nil")
	}
	var ve *domain.ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected *domain.ValidationError, got %T: %v", err, err)
	}
	if len(ve.Details) == 0 || ve.Details[0].Field != fieldTagIDs {
		t.Errorf("expected error on field tagIds, got: %v", ve.Details)
	}
	if ve.Details[0].Code != codeTooMany {
		t.Errorf("expected code TOO_MANY, got %q", ve.Details[0].Code)
	}
}

func TestCreateEventCommand_Execute_DuplicateTags_ReturnsValidationError(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := eventmock.NewMockRepository(ctrl)
	tagRepo := tagmock.NewMockRepository(ctrl)

	userID := uuid.New()
	now := time.Now().UTC()
	dup := uuid.New()

	cmd := eventuc.NewCreateEventCommand(repo, tagRepo, &testutil.StubTxManager{})
	_, err := cmd.Execute(context.Background(), userID, eventuc.CreateEventInput{
		Title:   testCreateTitle,
		StartAt: now,
		EndAt:   now.Add(time.Hour),
		TagIDs:  []uuid.UUID{dup, dup},
	})
	if err == nil {
		t.Fatal("expected validation error for duplicate tags, got nil")
	}
	var ve *domain.ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected *domain.ValidationError, got %T: %v", err, err)
	}
	if len(ve.Details) == 0 || ve.Details[0].Field != fieldTagIDs {
		t.Errorf("expected error on field tagIds, got: %v", ve.Details)
	}
	if ve.Details[0].Code != codeDuplicated {
		t.Errorf("expected code DUPLICATED, got %q", ve.Details[0].Code)
	}
}

func TestCreateEventCommand_Execute_TagNotFound_ReturnsValidationError(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := eventmock.NewMockRepository(ctrl)
	// Create must not be called when tag not found.
	tagRepo := tagmock.NewMockRepository(ctrl)

	userID := uuid.New()
	missingTagID := uuid.New()
	now := time.Now().UTC()

	// FindByIDsForShare returns empty slice (tag not found / foreign).
	tagRepo.EXPECT().FindByIDsForShare(gomock.Any(), []uuid.UUID{missingTagID}).Return([]tag.Tag{}, nil)

	cmd := eventuc.NewCreateEventCommand(repo, tagRepo, &testutil.StubTxManager{})
	_, err := cmd.Execute(context.Background(), userID, eventuc.CreateEventInput{
		Title:   testCreateTitle,
		StartAt: now,
		EndAt:   now.Add(time.Hour),
		TagIDs:  []uuid.UUID{missingTagID},
	})
	if err == nil {
		t.Fatal("expected validation error for missing tag, got nil")
	}
	var ve *domain.ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected *domain.ValidationError, got %T: %v", err, err)
	}
	if len(ve.Details) == 0 || ve.Details[0].Field != fieldTagIDs {
		t.Errorf("expected error on field tagIds, got: %v", ve.Details)
	}
	if ve.Details[0].Code != codeNotFound {
		t.Errorf("expected code NOT_FOUND, got %q", ve.Details[0].Code)
	}
}

func TestCreateEventCommand_Execute_ForeignTag_ReturnsValidationError(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := eventmock.NewMockRepository(ctrl)
	tagRepo := tagmock.NewMockRepository(ctrl)

	userID := uuid.New()
	otherUserID := uuid.New()
	foreignTagID := uuid.New()
	now := time.Now().UTC()

	// FindByIDsForShare returns tag owned by another user.
	tagRepo.EXPECT().FindByIDsForShare(gomock.Any(), []uuid.UUID{foreignTagID}).Return(
		[]tag.Tag{newStoredTag(t, foreignTagID, otherUserID)},
		nil,
	)

	cmd := eventuc.NewCreateEventCommand(repo, tagRepo, &testutil.StubTxManager{})
	_, err := cmd.Execute(context.Background(), userID, eventuc.CreateEventInput{
		Title:   testCreateTitle,
		StartAt: now,
		EndAt:   now.Add(time.Hour),
		TagIDs:  []uuid.UUID{foreignTagID},
	})
	if err == nil {
		t.Fatal("expected validation error for foreign tag, got nil")
	}
	var ve *domain.ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected *domain.ValidationError, got %T: %v", err, err)
	}
	if len(ve.Details) == 0 || ve.Details[0].Field != fieldTagIDs {
		t.Errorf("expected error on field tagIds, got: %v", ve.Details)
	}
	if ve.Details[0].Code != codeNotFound {
		t.Errorf("expected code NOT_FOUND, got %q", ve.Details[0].Code)
	}
}

func TestCreateEventCommand_Execute_FindByIDsForShareCalledInsideTransaction(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := eventmock.NewMockRepository(ctrl)
	tagRepo := tagmock.NewMockRepository(ctrl)

	userID := uuid.New()
	tagID := uuid.New()
	now := time.Now().UTC()

	tagRepo.EXPECT().FindByIDsForShare(gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, _ []uuid.UUID) ([]tag.Tag, error) {
			if !inTx(ctx) {
				t.Error("expected FindByIDsForShare to be called inside a transaction")
			}
			return []tag.Tag{newStoredTag(t, tagID, userID)}, nil
		},
	)
	repo.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, e domainevent.Event) (domainevent.Event, error) {
			return e, nil
		},
	)

	cmd := eventuc.NewCreateEventCommand(repo, tagRepo, markingTxManager{})
	if _, err := cmd.Execute(context.Background(), userID, eventuc.CreateEventInput{
		Title:   testCreateTitle,
		StartAt: now,
		EndAt:   now.Add(time.Hour),
		TagIDs:  []uuid.UUID{tagID},
	}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCreateEventCommand_Execute_NilTagIDs_ReturnsEmptyTags(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := eventmock.NewMockRepository(ctrl)

	userID := uuid.New()
	now := time.Now().UTC()
	repo.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, e domainevent.Event) (domainevent.Event, error) {
			return e, nil
		},
	)

	// No FindByIDsForShare call expected when TagIDs is nil.
	cmd := eventuc.NewCreateEventCommand(repo, tagmock.NewMockRepository(ctrl), &testutil.StubTxManager{})
	out, err := cmd.Execute(context.Background(), userID, eventuc.CreateEventInput{
		Title:   testCreateTitle,
		StartAt: now,
		EndAt:   now.Add(time.Hour),
		TagIDs:  nil,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out.Tags) != 0 {
		t.Errorf("expected empty tags for nil TagIDs, got %d", len(out.Tags))
	}
}
