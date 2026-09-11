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
	if u.PasswordHash() != "$2a$10$hash" {
		t.Errorf("PasswordHash() = %q, want %q", u.PasswordHash(), "$2a$10$hash")
	}
}

func TestUserErrors_areDistinct(t *testing.T) {
	if errors.Is(user.ErrUserNotFound, user.ErrEmailTaken) {
		t.Error("ErrUserNotFound and ErrEmailTaken must be distinct")
	}
}
