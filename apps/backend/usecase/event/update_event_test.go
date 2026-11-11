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

const newTitle = "New title"

func TestUpdateEventCommand_Execute_ValidInput_UpdatesAndReturnsEvent(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := eventmock.NewMockRepository(ctrl)

	userID := uuid.New()
	eventID := uuid.New()
	start := time.Now().UTC()
	end := start.Add(time.Hour)

	existing, _ := domainevent.New(eventID, userID, "Old title", "old desc", start, end, "", "")

	repo.EXPECT().FindByIDForUpdate(gomock.Any(), eventID).Return(existing, nil)
	repo.EXPECT().Update(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, e domainevent.Event) error {
			if e.Title() != newTitle {
				t.Errorf("title mismatch: got %q", e.Title())
			}
			return nil
		},
	)

	cmd := eventuc.NewUpdateEventCommand(repo, tagmock.NewMockRepository(ctrl), &testutil.StubTxManager{}, testLogger)
	out, err := cmd.Execute(context.Background(), userID, eventuc.UpdateEventInput{
		ID:          eventID,
		Title:       newTitle,
		Description: "new desc",
		StartAt:     start,
		EndAt:       end,
		Location:    testLocation,
		URL:         testURL,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Event.Title() != newTitle {
		t.Errorf("title mismatch: got %q", out.Event.Title())
	}
	if out.Event.Location() != testLocation {
		t.Errorf("location mismatch: got %q", out.Event.Location())
	}
	if out.Event.URL() != testURL {
		t.Errorf("url mismatch: got %q", out.Event.URL())
	}
}

func TestUpdateEventCommand_Execute_EndAtBeforeStartAt_ReturnsValidationError(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := eventmock.NewMockRepository(ctrl)

	userID := uuid.New()
	eventID := uuid.New()
	start := time.Now().UTC()
	end := start.Add(time.Hour)

	// Body validation fails before entering the transaction, so FindByIDForUpdate
	// must NOT be called.
	cmd := eventuc.NewUpdateEventCommand(repo, tagmock.NewMockRepository(ctrl), &testutil.StubTxManager{}, testLogger)
	_, err := cmd.Execute(context.Background(), userID, eventuc.UpdateEventInput{
		ID:      eventID,
		Title:   newTitle,
		StartAt: end,
		EndAt:   start,
	})
	if _, ok := errors.AsType[*domain.ValidationError](err); !ok {
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

	// Body validation fails before entering the transaction, so FindByIDForUpdate
	// must NOT be called.
	cmd := eventuc.NewUpdateEventCommand(repo, tagmock.NewMockRepository(ctrl), &testutil.StubTxManager{}, testLogger)
	_, err := cmd.Execute(context.Background(), userID, eventuc.UpdateEventInput{
		ID:      eventID,
		Title:   "",
		StartAt: start,
		EndAt:   end,
	})
	if _, ok := errors.AsType[*domain.ValidationError](err); !ok {
		t.Fatalf("expected *ValidationError, got %T: %v", err, err)
	}
}

func TestUpdateEventCommand_Execute_EventNotFound_ReturnsEventNotFoundError(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := eventmock.NewMockRepository(ctrl)

	userID := uuid.New()
	eventID := uuid.New()

	repo.EXPECT().FindByIDForUpdate(gomock.Any(), eventID).Return(nil, domainevent.ErrEventNotFound)

	cmd := eventuc.NewUpdateEventCommand(repo, tagmock.NewMockRepository(ctrl), &testutil.StubTxManager{}, testLogger)
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
	repo.EXPECT().FindByIDForUpdate(gomock.Any(), eventID).Return(existing, nil)

	cmd := eventuc.NewUpdateEventCommand(repo, tagmock.NewMockRepository(ctrl), &testutil.StubTxManager{}, testLogger)
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

	repo.EXPECT().FindByIDForUpdate(gomock.Any(), eventID).Return(existing, nil)
	repo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)

	cmd := eventuc.NewUpdateEventCommand(repo, tagmock.NewMockRepository(ctrl), &testutil.StubTxManager{}, testLogger)
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

	repo.EXPECT().FindByIDForUpdate(gomock.Any(), eventID).Return(existing, nil)
	repo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(errDBFailure)

	cmd := eventuc.NewUpdateEventCommand(repo, tagmock.NewMockRepository(ctrl), &testutil.StubTxManager{}, testLogger)
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

func newTaggedEvent(t *testing.T, eventID, userID uuid.UUID, tagIDs ...uuid.UUID) domainevent.Event {
	t.Helper()
	start := time.Now().UTC()
	tags, err := domainevent.NewEventTags(tagIDs)
	if err != nil {
		t.Fatalf("build tags: %v", err)
	}
	e, _ := domainevent.New(eventID, userID, "Old title", "", start, start.Add(time.Hour), "", "")
	e, err = e.WithTags(tags)
	if err != nil {
		t.Fatalf("attach tags: %v", err)
	}
	return e
}

func newStoredTag(t *testing.T, id, userID uuid.UUID) tag.Tag {
	t.Helper()
	name, _ := tag.NewTagName("Tag " + id.String()[:8])
	return tag.New(id, userID, name, tag.ColorRed)
}

func assertEventTags(t *testing.T, got, want []domainevent.EventTag) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("expected %d tags, got %d", len(want), len(got))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("tag[%d] mismatch: got %v, want %v", i, got[i], want[i])
		}
	}
}

func TestUpdateEventCommand_Execute_KeepsExistingTagsLockedForShare(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := eventmock.NewMockRepository(ctrl)
	tagRepo := tagmock.NewMockRepository(ctrl)

	userID := uuid.New()
	eventID := uuid.New()
	t1, t2 := uuid.New(), uuid.New()
	existing := newTaggedEvent(t, eventID, userID, t1, t2)

	repo.EXPECT().FindByIDForUpdate(gomock.Any(), eventID).Return(existing, nil)
	tagRepo.EXPECT().FindByIDsForShare(gomock.Any(), []uuid.UUID{t1, t2}).DoAndReturn(
		func(ctx context.Context, _ []uuid.UUID) ([]tag.Tag, error) {
			if !inTx(ctx) {
				t.Error("expected FindByIDsForShare to be called inside a transaction")
			}
			return []tag.Tag{newStoredTag(t, t1, userID), newStoredTag(t, t2, userID)}, nil
		},
	)
	repo.EXPECT().Update(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, e domainevent.Event) error {
			assertEventTags(t, e.Tags(), existing.Tags())
			return nil
		},
	)

	cmd := eventuc.NewUpdateEventCommand(repo, tagRepo, markingTxManager{}, testLogger)
	out, err := cmd.Execute(context.Background(), userID, eventuc.UpdateEventInput{
		ID: eventID, Title: newTitle, StartAt: existing.StartAt(), EndAt: existing.EndAt(),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertEventTags(t, out.Event.Tags(), existing.Tags())
}

func TestUpdateEventCommand_Execute_DropsTagDeletedConcurrently(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := eventmock.NewMockRepository(ctrl)
	tagRepo := tagmock.NewMockRepository(ctrl)

	userID := uuid.New()
	eventID := uuid.New()
	t1, t2 := uuid.New(), uuid.New()
	existing := newTaggedEvent(t, eventID, userID, t1, t2)

	// t1 was deleted after the event was read, so it is missing from the
	// locked rows; writing it back would fail the event_tags foreign key.
	repo.EXPECT().FindByIDForUpdate(gomock.Any(), eventID).Return(existing, nil)
	tagRepo.EXPECT().FindByIDsForShare(gomock.Any(), []uuid.UUID{t1, t2}).
		Return([]tag.Tag{newStoredTag(t, t2, userID)}, nil)
	want := existing.Tags()[1:]
	repo.EXPECT().Update(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, e domainevent.Event) error {
			assertEventTags(t, e.Tags(), want)
			return nil
		},
	)

	cmd := eventuc.NewUpdateEventCommand(repo, tagRepo, &testutil.StubTxManager{}, testLogger)
	out, err := cmd.Execute(context.Background(), userID, eventuc.UpdateEventInput{
		ID: eventID, Title: newTitle, StartAt: existing.StartAt(), EndAt: existing.EndAt(),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertEventTags(t, out.Event.Tags(), want)
}

func TestUpdateEventCommand_Execute_TagLockError_ReturnsErrorWithoutUpdate(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := eventmock.NewMockRepository(ctrl)
	tagRepo := tagmock.NewMockRepository(ctrl)

	userID := uuid.New()
	eventID := uuid.New()
	existing := newTaggedEvent(t, eventID, userID, uuid.New())

	repo.EXPECT().FindByIDForUpdate(gomock.Any(), eventID).Return(existing, nil)
	tagRepo.EXPECT().FindByIDsForShare(gomock.Any(), gomock.Any()).Return(nil, errDBFailure)

	cmd := eventuc.NewUpdateEventCommand(repo, tagRepo, &testutil.StubTxManager{}, testLogger)
	_, err := cmd.Execute(context.Background(), userID, eventuc.UpdateEventInput{
		ID: eventID, Title: newTitle, StartAt: existing.StartAt(), EndAt: existing.EndAt(),
	})
	if !errors.Is(err, errDBFailure) {
		t.Errorf("expected errDBFailure, got %v", err)
	}
}

func TestUpdateEventCommand_Execute_FindsAndUpdatesInsideTransaction(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := eventmock.NewMockRepository(ctrl)

	userID := uuid.New()
	eventID := uuid.New()
	start := time.Now().UTC()
	end := start.Add(time.Hour)
	existing, _ := domainevent.New(eventID, userID, "Old title", "", start, end, "", "")

	repo.EXPECT().FindByIDForUpdate(gomock.Any(), eventID).DoAndReturn(
		func(ctx context.Context, _ uuid.UUID) (domainevent.Event, error) {
			if !inTx(ctx) {
				t.Error("expected FindByIDForUpdate to be called inside a transaction")
			}
			return existing, nil
		},
	)
	repo.EXPECT().Update(gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, _ domainevent.Event) error {
			if !inTx(ctx) {
				t.Error("expected Update to be called inside a transaction")
			}
			return nil
		},
	)

	cmd := eventuc.NewUpdateEventCommand(repo, tagmock.NewMockRepository(ctrl), markingTxManager{}, testLogger)
	if _, err := cmd.Execute(context.Background(), userID, eventuc.UpdateEventInput{
		ID:      eventID,
		Title:   newTitle,
		StartAt: start,
		EndAt:   end,
	}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUpdateEventCommand_Execute_NilTagIDs_KeepsExistingTags(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := eventmock.NewMockRepository(ctrl)
	tagRepo := tagmock.NewMockRepository(ctrl)

	userID := uuid.New()
	eventID := uuid.New()
	t1 := uuid.New()
	existing := newTaggedEvent(t, eventID, userID, t1)

	repo.EXPECT().FindByIDForUpdate(gomock.Any(), eventID).Return(existing, nil)
	tagRepo.EXPECT().FindByIDsForShare(gomock.Any(), []uuid.UUID{t1}).Return(
		[]tag.Tag{newStoredTag(t, t1, userID)},
		nil,
	)
	repo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)

	cmd := eventuc.NewUpdateEventCommand(repo, tagRepo, &testutil.StubTxManager{}, testLogger)
	out, err := cmd.Execute(context.Background(), userID, eventuc.UpdateEventInput{
		ID: eventID, Title: newTitle, StartAt: existing.StartAt(), EndAt: existing.EndAt(),
		TagIDs: nil, // nil = keep existing
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertEventTags(t, out.Event.Tags(), existing.Tags())
}

func TestUpdateEventCommand_Execute_NilTagIDs_ReturnsKeptTagViewsInOrder(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := eventmock.NewMockRepository(ctrl)
	tagRepo := tagmock.NewMockRepository(ctrl)

	userID := uuid.New()
	eventID := uuid.New()
	t1, t2, t3 := uuid.New(), uuid.New(), uuid.New()
	existing := newTaggedEvent(t, eventID, userID, t1, t2, t3)
	stored1, stored3 := newStoredTag(t, t1, userID), newStoredTag(t, t3, userID)

	// Locked rows come back in a different order and t2 was deleted concurrently.
	repo.EXPECT().FindByIDForUpdate(gomock.Any(), eventID).Return(existing, nil)
	tagRepo.EXPECT().FindByIDsForShare(gomock.Any(), []uuid.UUID{t1, t2, t3}).
		Return([]tag.Tag{stored3, stored1}, nil)
	repo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)

	cmd := eventuc.NewUpdateEventCommand(repo, tagRepo, &testutil.StubTxManager{}, testLogger)
	out, err := cmd.Execute(context.Background(), userID, eventuc.UpdateEventInput{
		ID: eventID, Title: newTitle, StartAt: existing.StartAt(), EndAt: existing.EndAt(),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []eventuc.EventTagView{
		{ID: t1, Name: string(stored1.Name()), Color: string(stored1.Color())},
		{ID: t3, Name: string(stored3.Name()), Color: string(stored3.Color())},
	}
	if len(out.Tags) != len(want) {
		t.Fatalf("expected %d tag views, got %d", len(want), len(out.Tags))
	}
	for i := range want {
		if out.Tags[i] != want[i] {
			t.Errorf("tags[%d] mismatch: got %+v, want %+v", i, out.Tags[i], want[i])
		}
	}
}

func TestUpdateEventCommand_Execute_EmptyTagIDs_RemovesAllTags(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := eventmock.NewMockRepository(ctrl)
	tagRepo := tagmock.NewMockRepository(ctrl)

	userID := uuid.New()
	eventID := uuid.New()
	existing := newTaggedEvent(t, eventID, userID, uuid.New())

	repo.EXPECT().FindByIDForUpdate(gomock.Any(), eventID).Return(existing, nil)
	// FindByIDsForShare should NOT be called for an explicit empty list replacement.
	repo.EXPECT().Update(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, e domainevent.Event) error {
			if len(e.Tags()) != 0 {
				t.Errorf("expected event to have no tags after empty tagIds, got %d", len(e.Tags()))
			}
			return nil
		},
	)

	emptyTagIDs := []uuid.UUID{}
	cmd := eventuc.NewUpdateEventCommand(repo, tagRepo, &testutil.StubTxManager{}, testLogger)
	out, err := cmd.Execute(context.Background(), userID, eventuc.UpdateEventInput{
		ID: eventID, Title: newTitle, StartAt: existing.StartAt(), EndAt: existing.EndAt(),
		TagIDs: &emptyTagIDs,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out.Event.Tags()) != 0 {
		t.Errorf("expected empty tags after empty tagIds, got %d", len(out.Event.Tags()))
	}
}

func TestUpdateEventCommand_Execute_ReplaceTagIDs_UpdatesToNewTags(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := eventmock.NewMockRepository(ctrl)
	tagRepo := tagmock.NewMockRepository(ctrl)

	userID := uuid.New()
	eventID := uuid.New()
	oldTag := uuid.New()
	newTag1, newTag2 := uuid.New(), uuid.New()
	existing := newTaggedEvent(t, eventID, userID, oldTag)

	repo.EXPECT().FindByIDForUpdate(gomock.Any(), eventID).Return(existing, nil)
	tagRepo.EXPECT().FindByIDsForShare(gomock.Any(), []uuid.UUID{newTag1, newTag2}).Return(
		[]tag.Tag{newStoredTag(t, newTag1, userID), newStoredTag(t, newTag2, userID)},
		nil,
	)
	repo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)

	newTagIDs := []uuid.UUID{newTag1, newTag2}
	cmd := eventuc.NewUpdateEventCommand(repo, tagRepo, &testutil.StubTxManager{}, testLogger)
	out, err := cmd.Execute(context.Background(), userID, eventuc.UpdateEventInput{
		ID: eventID, Title: newTitle, StartAt: existing.StartAt(), EndAt: existing.EndAt(),
		TagIDs: &newTagIDs,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out.Tags) != 2 {
		t.Fatalf("expected 2 tags, got %d", len(out.Tags))
	}
	if out.Tags[0].ID != newTag1 {
		t.Errorf("tags[0] mismatch: got %v, want %v", out.Tags[0].ID, newTag1)
	}
	if out.Tags[1].ID != newTag2 {
		t.Errorf("tags[1] mismatch: got %v, want %v", out.Tags[1].ID, newTag2)
	}
}

func TestUpdateEventCommand_Execute_TooManyReplacementTags_ReturnsValidationError(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := eventmock.NewMockRepository(ctrl)
	tagRepo := tagmock.NewMockRepository(ctrl)

	userID := uuid.New()
	eventID := uuid.New()
	start := time.Now().UTC()
	end := start.Add(time.Hour)

	// Tag count validation fails before entering the transaction, so FindByIDForUpdate
	// must NOT be called.
	tooMany := []uuid.UUID{uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New()}
	cmd := eventuc.NewUpdateEventCommand(repo, tagRepo, &testutil.StubTxManager{}, testLogger)
	_, err := cmd.Execute(context.Background(), userID, eventuc.UpdateEventInput{
		ID: eventID, Title: newTitle, StartAt: start, EndAt: end,
		TagIDs: &tooMany,
	})
	if err == nil {
		t.Fatal("expected validation error for too many tags, got nil")
	}
	var ve *domain.ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected *domain.ValidationError, got %T: %v", err, err)
	}
	if len(ve.Details) == 0 || ve.Details[0].Field != fieldTagIDs || ve.Details[0].Code != codeTooMany {
		t.Errorf("expected tagIds TOO_MANY, got: %v", ve.Details)
	}
}

func TestUpdateEventCommand_Execute_ForeignReplacementTag_ReturnsValidationError(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := eventmock.NewMockRepository(ctrl)
	tagRepo := tagmock.NewMockRepository(ctrl)

	userID := uuid.New()
	otherUserID := uuid.New()
	eventID := uuid.New()
	foreignTag := uuid.New()
	existing, _ := domainevent.New(eventID, userID, "Title", "", time.Now(), time.Now().Add(time.Hour), "", "")

	repo.EXPECT().FindByIDForUpdate(gomock.Any(), eventID).Return(existing, nil)
	// FindByIDsForShare returns tag owned by another user.
	tagRepo.EXPECT().FindByIDsForShare(gomock.Any(), []uuid.UUID{foreignTag}).Return(
		[]tag.Tag{newStoredTag(t, foreignTag, otherUserID)},
		nil,
	)

	tagIDs := []uuid.UUID{foreignTag}
	cmd := eventuc.NewUpdateEventCommand(repo, tagRepo, &testutil.StubTxManager{}, testLogger)
	_, err := cmd.Execute(context.Background(), userID, eventuc.UpdateEventInput{
		ID: eventID, Title: newTitle, StartAt: existing.StartAt(), EndAt: existing.EndAt(),
		TagIDs: &tagIDs,
	})
	if err == nil {
		t.Fatal("expected validation error for foreign tag, got nil")
	}
	var ve *domain.ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected *domain.ValidationError, got %T: %v", err, err)
	}
	if len(ve.Details) == 0 || ve.Details[0].Field != fieldTagIDs || ve.Details[0].Code != codeNotFound {
		t.Errorf("expected tagIds NOT_FOUND, got: %v", ve.Details)
	}
}
