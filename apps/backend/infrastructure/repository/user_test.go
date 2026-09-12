//go:build integration

package repository_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Haya372/ai-trial/backend/domain/user"
	"github.com/Haya372/ai-trial/backend/infrastructure/repository"
)

func TestUserRepository_Create_success(t *testing.T) {
	setupTest(t)
	repo := repository.NewUserRepository(testPool)

	email, _ := user.NewEmail("test@example.com")
	password, _ := user.NewPassword("SecurePass1!")

	u, err := repo.Create(context.Background(), email, "Test User", password)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(u.Email()) != "test@example.com" {
		t.Errorf("expected email test@example.com, got %s", u.Email())
	}
	if u.DisplayName() != "Test User" {
		t.Errorf("expected display name Test User, got %s", u.DisplayName())
	}
}

func TestUserRepository_FindByEmail_success(t *testing.T) {
	setupTest(t)
	repo := repository.NewUserRepository(testPool)

	email, _ := user.NewEmail("find@example.com")
	password, _ := user.NewPassword("SecurePass1!")
	_, err := repo.Create(context.Background(), email, "Find User", password)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}

	found, err := repo.FindByEmail(context.Background(), email)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(found.Email()) != "find@example.com" {
		t.Errorf("expected email find@example.com, got %s", found.Email())
	}
}

func TestUserRepository_FindByEmail_notFound(t *testing.T) {
	setupTest(t)
	repo := repository.NewUserRepository(testPool)

	email, _ := user.NewEmail("notfound@example.com")
	_, err := repo.FindByEmail(context.Background(), email)
	if !errors.Is(err, user.ErrUserNotFound) {
		t.Errorf("expected ErrUserNotFound, got %v", err)
	}
}

func TestUserRepository_Create_duplicateEmail_returnsEmailTaken(t *testing.T) {
	setupTest(t)
	repo := repository.NewUserRepository(testPool)

	email, _ := user.NewEmail("dup@example.com")
	password, _ := user.NewPassword("SecurePass1!")

	_, err := repo.Create(context.Background(), email, "User1", password)
	if err != nil {
		t.Fatalf("first create: %v", err)
	}

	_, err = repo.Create(context.Background(), email, "User2", password)
	if !errors.Is(err, user.ErrEmailTaken) {
		t.Errorf("expected ErrEmailTaken, got %v", err)
	}
}
