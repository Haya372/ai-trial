package auth_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"

	"github.com/Haya372/ai-trial/backend/domain/session"
	sessionmock "github.com/Haya372/ai-trial/backend/domain/session/generated"
	"github.com/Haya372/ai-trial/backend/domain/user"
	usermock "github.com/Haya372/ai-trial/backend/domain/user/generated"
	authuc "github.com/Haya372/ai-trial/backend/usecase/auth"
)

// stubUser implements user.User with controlled ComparePassword behavior.
type stubUser struct {
	id          uuid.UUID
	email       user.Email
	displayName string
	compareErr  error
}

func (u *stubUser) ID() uuid.UUID                         { return u.id }
func (u *stubUser) Email() user.Email                     { return u.email }
func (u *stubUser) DisplayName() string                   { return u.displayName }
func (u *stubUser) ComparePassword(_ user.Password) error { return u.compareErr }

func TestLoginCommand_Execute_ValidCredentials_ReturnsAuthOutput(t *testing.T) {
	ctrl := gomock.NewController(t)

	fixedUserID := uuid.New()
	fixedSessID := uuid.New()
	email, _ := user.NewEmail("u@ex.com")

	stub := &stubUser{id: fixedUserID, email: email, displayName: "U", compareErr: nil}

	mockUserRepo := usermock.NewMockRepository(ctrl)
	mockSessRepo := sessionmock.NewMockRepository(ctrl)

	mockUserRepo.EXPECT().
		FindByEmail(gomock.Any(), email).
		Return(stub, nil)
	mockSessRepo.EXPECT().
		Create(gomock.Any(), fixedUserID, gomock.Any()).
		Return(session.New(fixedSessID, fixedUserID, time.Now().Add(30*24*time.Hour)), nil)

	cmd := authuc.NewLoginCommand(mockUserRepo, mockSessRepo)
	out, err := cmd.Execute(context.Background(), authuc.LoginInput{
		Email:    "u@ex.com",
		Password: testPassword,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.SessionID != fixedSessID {
		t.Errorf("expected session %v, got %v", fixedSessID, out.SessionID)
	}
}

func TestLoginCommand_Execute_WrongPassword_ReturnsInvalidCredentials(t *testing.T) {
	ctrl := gomock.NewController(t)

	email, _ := user.NewEmail("u@ex.com")
	stub := &stubUser{id: uuid.New(), email: email, compareErr: errPasswordMismatch}

	mockUserRepo := usermock.NewMockRepository(ctrl)
	mockSessRepo := sessionmock.NewMockRepository(ctrl)

	mockUserRepo.EXPECT().
		FindByEmail(gomock.Any(), email).
		Return(stub, nil)

	cmd := authuc.NewLoginCommand(mockUserRepo, mockSessRepo)
	_, err := cmd.Execute(context.Background(), authuc.LoginInput{
		Email:    "u@ex.com",
		Password: "WrongPass1!",
	})
	if !errors.Is(err, authuc.ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestLoginCommand_Execute_UnknownEmail_ReturnsInvalidCredentials(t *testing.T) {
	ctrl := gomock.NewController(t)

	email, _ := user.NewEmail("no@ex.com")
	mockUserRepo := usermock.NewMockRepository(ctrl)

	mockUserRepo.EXPECT().
		FindByEmail(gomock.Any(), email).
		Return(nil, user.ErrUserNotFound)

	cmd := authuc.NewLoginCommand(mockUserRepo, sessionmock.NewMockRepository(ctrl))
	_, err := cmd.Execute(context.Background(), authuc.LoginInput{
		Email:    "no@ex.com",
		Password: testPassword,
	})
	if !errors.Is(err, authuc.ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestLoginCommand_Execute_InvalidEmailFormat_ReturnsInvalidCredentials(t *testing.T) {
	ctrl := gomock.NewController(t)
	cmd := authuc.NewLoginCommand(
		usermock.NewMockRepository(ctrl),
		sessionmock.NewMockRepository(ctrl),
	)
	_, err := cmd.Execute(context.Background(), authuc.LoginInput{
		Email:    "not-an-email",
		Password: testPassword,
	})
	if !errors.Is(err, authuc.ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}
}
