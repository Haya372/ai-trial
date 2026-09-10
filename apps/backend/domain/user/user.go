package user

import "github.com/google/uuid"

type User interface {
	ID() uuid.UUID
	Email() Email
	DisplayName() string
	PasswordHash() string
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

func (u *userEntity) ID() uuid.UUID        { return u.id }
func (u *userEntity) Email() Email         { return u.email }
func (u *userEntity) DisplayName() string  { return u.displayName }
func (u *userEntity) PasswordHash() string { return u.passwordHash }
