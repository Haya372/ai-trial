package user

import (
	"fmt"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type User interface {
	ID() uuid.UUID
	Email() Email
	DisplayName() string
	ComparePassword(password Password) error
}

type userEntity struct {
	id           uuid.UUID
	email        Email
	displayName  string
	passwordHash string
}

func New(id uuid.UUID, email Email, displayName, passwordHash string) User {
	return &userEntity{id: id, email: email, displayName: displayName, passwordHash: passwordHash}
}

func (u *userEntity) ID() uuid.UUID       { return u.id }
func (u *userEntity) Email() Email        { return u.email }
func (u *userEntity) DisplayName() string { return u.displayName }

func (u *userEntity) ComparePassword(password Password) error {
	if err := bcrypt.CompareHashAndPassword([]byte(u.passwordHash), []byte(password.plain)); err != nil {
		return fmt.Errorf("password mismatch: %w", err)
	}
	return nil
}
