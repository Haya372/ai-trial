package user_test

import (
	"errors"
	"testing"

	"github.com/Haya372/ai-trial/backend/domain/user"
)

func TestUser_fields(t *testing.T) {
	u := user.User{
		Email:       "test@example.com",
		DisplayName: "Test User",
	}
	if u.Email != "test@example.com" {
		t.Errorf("expected email test@example.com, got %s", u.Email)
	}
	if u.DisplayName != "Test User" {
		t.Errorf("expected display name Test User, got %s", u.DisplayName)
	}
}

func TestUserErrors_areDistinct(t *testing.T) {
	if errors.Is(user.ErrNotFound, user.ErrEmailTaken) {
		t.Error("ErrNotFound and ErrEmailTaken must be distinct")
	}
}
