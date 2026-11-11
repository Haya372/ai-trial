package tag_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"

	domaintag "github.com/Haya372/ai-trial/backend/domain/tag"
	tagmock "github.com/Haya372/ai-trial/backend/domain/tag/generated"
	taguc "github.com/Haya372/ai-trial/backend/usecase/tag"
)

func TestDeleteTagCommand_Execute_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := tagmock.NewMockRepository(ctrl)
	userID := uuid.New()
	tagID := uuid.New()

	gomock.InOrder(
		repo.EXPECT().FindByID(gomock.Any(), tagID).DoAndReturn(
			func(ctx context.Context, _ uuid.UUID) (domaintag.Tag, error) {
				assertInTx(t, ctx, "FindByID")
				return newStoredTag(t, tagID, userID), nil
			},
		),
		repo.EXPECT().Delete(gomock.Any(), tagID).DoAndReturn(
			func(ctx context.Context, _ uuid.UUID) error {
				assertInTx(t, ctx, "Delete")
				return nil
			},
		),
	)

	cmd := taguc.NewDeleteTagCommand(repo, markingTxManager{})
	if err := cmd.Execute(context.Background(), userID, tagID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteTagCommand_Execute_TagNotFound_ReturnsErrTagNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := tagmock.NewMockRepository(ctrl)
	tagID := uuid.New()

	repo.EXPECT().FindByID(gomock.Any(), tagID).Return(nil, domaintag.ErrTagNotFound)

	cmd := taguc.NewDeleteTagCommand(repo, markingTxManager{})
	err := cmd.Execute(context.Background(), uuid.New(), tagID)

	if !errors.Is(err, domaintag.ErrTagNotFound) {
		t.Errorf("got %v, want ErrTagNotFound", err)
	}
}

func TestDeleteTagCommand_Execute_AnotherUsersTag_ReturnsErrTagNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := tagmock.NewMockRepository(ctrl)
	tagID := uuid.New()

	repo.EXPECT().FindByID(gomock.Any(), tagID).Return(newStoredTag(t, tagID, uuid.New()), nil)
	// Delete must not be called for a tag owned by someone else.

	cmd := taguc.NewDeleteTagCommand(repo, markingTxManager{})
	err := cmd.Execute(context.Background(), uuid.New(), tagID)

	if !errors.Is(err, domaintag.ErrTagNotFound) {
		t.Errorf("got %v, want ErrTagNotFound", err)
	}
}

func TestDeleteTagCommand_Execute_RepositoryErrors(t *testing.T) {
	t.Run("FindByID fails", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		repo := tagmock.NewMockRepository(ctrl)
		tagID := uuid.New()

		repo.EXPECT().FindByID(gomock.Any(), tagID).Return(nil, errDBFailure)

		cmd := taguc.NewDeleteTagCommand(repo, markingTxManager{})
		err := cmd.Execute(context.Background(), uuid.New(), tagID)

		if !errors.Is(err, errDBFailure) {
			t.Errorf("got %v, want %v", err, errDBFailure)
		}
	})

	t.Run("Delete fails", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		repo := tagmock.NewMockRepository(ctrl)
		userID := uuid.New()
		tagID := uuid.New()

		repo.EXPECT().FindByID(gomock.Any(), tagID).Return(newStoredTag(t, tagID, userID), nil)
		repo.EXPECT().Delete(gomock.Any(), tagID).Return(errDBFailure)

		cmd := taguc.NewDeleteTagCommand(repo, markingTxManager{})
		err := cmd.Execute(context.Background(), userID, tagID)

		if !errors.Is(err, errDBFailure) {
			t.Errorf("got %v, want %v", err, errDBFailure)
		}
	})
}
