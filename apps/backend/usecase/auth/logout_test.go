package auth_test

import (
	"context"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/Haya372/ai-trial/backend/domain/session"
	sessionmock "github.com/Haya372/ai-trial/backend/domain/session/generated"
	authuc "github.com/Haya372/ai-trial/backend/usecase/auth"
)

func TestLogoutCommand_Execute_DeletesCalled(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockSessRepo := sessionmock.NewMockRepository(ctrl)

	token := session.NewToken("plain-session-token")
	mockSessRepo.EXPECT().DeleteByTokenHash(gomock.Any(), token.Hash()).Return(nil)

	cmd := authuc.NewLogoutCommand(mockSessRepo)
	if err := cmd.Execute(context.Background(), token); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLogoutCommand_Execute_RepoError_Propagates(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockSessRepo := sessionmock.NewMockRepository(ctrl)

	mockSessRepo.EXPECT().DeleteByTokenHash(gomock.Any(), gomock.Any()).Return(errDBFailure)

	cmd := authuc.NewLogoutCommand(mockSessRepo)
	if err := cmd.Execute(context.Background(), session.NewToken("t")); err == nil {
		t.Fatal("expected error, got nil")
	}
}
