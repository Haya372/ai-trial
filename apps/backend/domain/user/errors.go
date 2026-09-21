package user

import "github.com/Haya372/ai-trial/backend/domain"

const (
	CodeUserNotFound     = "USER_NOT_FOUND"
	CodeEmailTaken       = "EMAIL_TAKEN"
	CodePasswordMismatch = "PASSWORD_MISMATCH"
)

var (
	ErrUserNotFound     = domain.NewDomainError(CodeUserNotFound, "user not found")
	ErrEmailTaken       = domain.NewDomainError(CodeEmailTaken, "email already registered")
	ErrPasswordMismatch = domain.NewDomainError(CodePasswordMismatch, "password does not match")
)
