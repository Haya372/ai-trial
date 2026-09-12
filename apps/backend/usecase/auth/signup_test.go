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

func TestSignupCommand_Execute_ValidInput_ReturnsAuthOutput(t *testing.T) {
	ctrl := gomock.NewController(t)

	fixedUserID := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	fixedSessID := uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")
	email, _ := user.NewEmail(testEmail)

	mockUserRepo := usermock.NewMockRepository(ctrl)
	mockSessRepo := sessionmock.NewMockRepository(ctrl)

	u := user.New(fixedUserID, email, "test", "hash")
	sess := session.New(fixedSessID, fixedUserID, time.Now().Add(30*24*time.Hour))

	mockUserRepo.EXPECT().
		Create(gomock.Any(), email, "test", gomock.Any()).
		Return(u, nil)
	mockSessRepo.EXPECT().
		Create(gomock.Any(), fixedUserID, gomock.Any()).
		Return(sess, nil)

	cmd := authuc.NewSignupCommand(mockUserRepo, mockSessRepo)
	out, err := cmd.Execute(context.Background(), authuc.SignupInput{
		Email:    testEmail,
		Password: testPassword,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.User.Email() != email {
		t.Errorf("expected email %v, got %v", email, out.User.Email())
	}
	if out.SessionID != fixedSessID {
		t.Errorf("expected session ID %v, got %v", fixedSessID, out.SessionID)
	}
}

func TestSignupCommand_Execute_DisplayNameDefaultsToEmailLocalPart(t *testing.T) {
	ctrl := gomock.NewController(t)
	email, _ := user.NewEmail("hello@example.com")

	mockUserRepo := usermock.NewMockRepository(ctrl)
	mockSessRepo := sessionmock.NewMockRepository(ctrl)

	u := user.New(uuid.New(), email, "hello", "hash")
	mockUserRepo.EXPECT().
		Create(gomock.Any(), email, "hello", gomock.Any()).
		Return(u, nil)
	mockSessRepo.EXPECT().
		Create(gomock.Any(), u.ID(), gomock.Any()).
		Return(session.New(uuid.New(), u.ID(), time.Now().Add(30*24*time.Hour)), nil)

	cmd := authuc.NewSignupCommand(mockUserRepo, mockSessRepo)
	_, err := cmd.Execute(context.Background(), authuc.SignupInput{
		Email:    "hello@example.com",
		Password: testPassword,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSignupCommand_Execute_InvalidEmail_ReturnsValidationError(t *testing.T) {
	ctrl := gomock.NewController(t)
	cmd := authuc.NewSignupCommand(
		usermock.NewMockRepository(ctrl),
		sessionmock.NewMockRepository(ctrl),
	)
	_, err := cmd.Execute(context.Background(), authuc.SignupInput{
		Email:    "not-an-email",
		Password: testPassword,
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var ve *authuc.ValidationError
	if !errors.As(err, &ve) {
		t.Errorf("expected *ValidationError, got %T: %v", err, err)
	}
}

func TestSignupCommand_Execute_ShortPassword_ReturnsValidationError(t *testing.T) {
	ctrl := gomock.NewController(t)
	cmd := authuc.NewSignupCommand(
		usermock.NewMockRepository(ctrl),
		sessionmock.NewMockRepository(ctrl),
	)
	_, err := cmd.Execute(context.Background(), authuc.SignupInput{
		Email:    testEmail,
		Password: "Ab1!",
	})
	var ve *authuc.ValidationError
	if !errors.As(err, &ve) {
		t.Errorf("expected *ValidationError, got %T: %v", err, err)
	}
}

func TestSignupCommand_Execute_PasswordMissingComplexity_ReturnsValidationError(t *testing.T) {
	ctrl := gomock.NewController(t)
	cmd := authuc.NewSignupCommand(
		usermock.NewMockRepository(ctrl),
		sessionmock.NewMockRepository(ctrl),
	)
	_, err := cmd.Execute(context.Background(), authuc.SignupInput{
		Email:    testEmail,
		Password: "alllowercase",
	})
	var ve *authuc.ValidationError
	if !errors.As(err, &ve) {
		t.Errorf("expected *ValidationError, got %T: %v", err, err)
	}
	if len(ve.Details) == 0 || ve.Details[0].Code != "INSUFFICIENT_COMPLEXITY" {
		t.Errorf("expected INSUFFICIENT_COMPLEXITY, got %v", ve.Details)
	}
}

func TestSignupCommand_Execute_EmailTaken_ReturnsEmailTakenError(t *testing.T) {
	ctrl := gomock.NewController(t)
	email, _ := user.NewEmail("dup@example.com")

	mockUserRepo := usermock.NewMockRepository(ctrl)
	mockSessRepo := sessionmock.NewMockRepository(ctrl)

	mockUserRepo.EXPECT().
		Create(gomock.Any(), email, gomock.Any(), gomock.Any()).
		Return(nil, user.ErrEmailTaken)

	cmd := authuc.NewSignupCommand(mockUserRepo, mockSessRepo)
	_, err := cmd.Execute(context.Background(), authuc.SignupInput{
		Email:    "dup@example.com",
		Password: testPassword,
	})
	if !errors.Is(err, user.ErrEmailTaken) {
		t.Errorf("expected ErrEmailTaken, got %v", err)
	}
}
