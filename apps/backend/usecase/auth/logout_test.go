package auth_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"

	sessionmock "github.com/Haya372/ai-trial/backend/domain/session/generated"
	authuc "github.com/Haya372/ai-trial/backend/usecase/auth"
)

func TestLogoutCommand_Execute_DeletesCalled(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockSessRepo := sessionmock.NewMockRepository(ctrl)

	id := uuid.New()
	mockSessRepo.EXPECT().Delete(gomock.Any(), id).Return(nil)

	cmd := authuc.NewLogoutCommand(mockSessRepo)
	if err := cmd.Execute(context.Background(), id); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLogoutCommand_Execute_RepoError_Propagates(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockSessRepo := sessionmock.NewMockRepository(ctrl)

	mockSessRepo.EXPECT().Delete(gomock.Any(), gomock.Any()).Return(errDBFailure)

	cmd := authuc.NewLogoutCommand(mockSessRepo)
	if err := cmd.Execute(context.Background(), uuid.New()); err == nil {
		t.Fatal("expected error, got nil")
	}
}
