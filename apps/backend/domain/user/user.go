package user

import "github.com/google/uuid"

type User struct {
	ID           uuid.UUID
	Email        string
	DisplayName  string
	PasswordHash string
}
