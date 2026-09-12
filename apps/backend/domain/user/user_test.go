package user_test

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/Haya372/ai-trial/backend/domain/user"
)

func TestNewUser_returnsAccessibleFields(t *testing.T) {
	email, _ := user.NewEmail("test@example.com")
	id := uuid.New()
	u := user.New(id, email, "Test User", "$2a$10$hash")

	if u.ID() != id {
		t.Errorf("ID() = %v, want %v", u.ID(), id)
	}
	if u.Email() != email {
		t.Errorf("Email() = %v, want %v", u.Email(), email)
	}
	if u.DisplayName() != "Test User" {
		t.Errorf("DisplayName() = %q, want %q", u.DisplayName(), "Test User")
	}
}

func TestUser_ComparePassword_correct(t *testing.T) {
	plain := "SecurePass1!"
	pw, err := user.NewPassword(plain)
	if err != nil {
		t.Fatalf("NewPassword() unexpected error: %v", err)
	}
	email, _ := user.NewEmail("test@example.com")
	u := user.New(uuid.New(), email, "Test User", pw.Hash())

	if err := u.ComparePassword(pw); err != nil {
		t.Errorf("ComparePassword() with matching password returned error: %v", err)
	}
}

func TestUser_ComparePassword_wrong(t *testing.T) {
	pw, _ := user.NewPassword("SecurePass1!")
	wrongPw, _ := user.NewPassword("WrongPass1!")
	email, _ := user.NewEmail("test@example.com")
	u := user.New(uuid.New(), email, "Test User", pw.Hash())

	if err := u.ComparePassword(wrongPw); err == nil {
		t.Error("ComparePassword() with wrong password expected error, got nil")
	}
}

func TestUserErrors_areDistinct(t *testing.T) {
	if errors.Is(user.ErrUserNotFound, user.ErrEmailTaken) {
		t.Error("ErrUserNotFound and ErrEmailTaken must be distinct")
	}
}
