package eventsubscription_test

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"

	"github.com/Haya372/ai-trial/backend/domain/eventsubscription"
	subscriptionmock "github.com/Haya372/ai-trial/backend/domain/eventsubscription/generated"
	eventsubscriptionuc "github.com/Haya372/ai-trial/backend/usecase/eventsubscription"
)

var testLogger = slog.New(slog.DiscardHandler)

func TestDeleteSubscriptionCommand_Execute_notFound_propagatesNotFoundError(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := subscriptionmock.NewMockRepository(ctrl)
	subID := uuid.New()

	repo.EXPECT().FindByID(gomock.Any(), subID).Return(nil, eventsubscription.ErrEventSubscriptionNotFound)

	c := eventsubscriptionuc.NewDeleteSubscriptionCommand(repo, testLogger)
	err := c.Execute(context.Background(), uuid.New(), subID)

	if !errors.Is(err, eventsubscription.ErrEventSubscriptionNotFound) {
		t.Errorf("expected ErrEventSubscriptionNotFound, got %v", err)
	}
}

func TestDeleteSubscriptionCommand_Execute_findByIDError_propagates(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := subscriptionmock.NewMockRepository(ctrl)
	subID := uuid.New()

	repo.EXPECT().FindByID(gomock.Any(), subID).Return(nil, errDBFailure)

	c := eventsubscriptionuc.NewDeleteSubscriptionCommand(repo, testLogger)
	err := c.Execute(context.Background(), uuid.New(), subID)

	if !errors.Is(err, errDBFailure) {
		t.Errorf("expected errDBFailure to propagate, got %v", err)
	}
}

func TestDeleteSubscriptionCommand_Execute_notOwner_returnsForbiddenError(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := subscriptionmock.NewMockRepository(ctrl)
	ownerID := uuid.New()
	requesterID := uuid.New()
	subID := uuid.New()
	sub := eventsubscription.New(subID, uuid.New(), ownerID, time.Now())

	repo.EXPECT().FindByID(gomock.Any(), subID).Return(sub, nil)

	c := eventsubscriptionuc.NewDeleteSubscriptionCommand(repo, testLogger)
	err := c.Execute(context.Background(), requesterID, subID)

	if !errors.Is(err, eventsubscription.ErrNotSubscriptionOwner) {
		t.Errorf("expected ErrNotSubscriptionOwner, got %v", err)
	}
}

func TestDeleteSubscriptionCommand_Execute_owner_deletesAndReturnsNoError(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := subscriptionmock.NewMockRepository(ctrl)
	userID := uuid.New()
	subID := uuid.New()
	sub := eventsubscription.New(subID, uuid.New(), userID, time.Now())

	repo.EXPECT().FindByID(gomock.Any(), subID).Return(sub, nil)
	repo.EXPECT().Delete(gomock.Any(), subID).Return(nil)

	c := eventsubscriptionuc.NewDeleteSubscriptionCommand(repo, testLogger)
	err := c.Execute(context.Background(), userID, subID)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteSubscriptionCommand_Execute_deleteError_propagates(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := subscriptionmock.NewMockRepository(ctrl)
	userID := uuid.New()
	subID := uuid.New()
	sub := eventsubscription.New(subID, uuid.New(), userID, time.Now())

	repo.EXPECT().FindByID(gomock.Any(), subID).Return(sub, nil)
	repo.EXPECT().Delete(gomock.Any(), subID).Return(errDBFailure)

	c := eventsubscriptionuc.NewDeleteSubscriptionCommand(repo, testLogger)
	err := c.Execute(context.Background(), userID, subID)

	if !errors.Is(err, errDBFailure) {
		t.Errorf("expected errDBFailure to propagate, got %v", err)
	}
}
