package user

import "github.com/Haya372/ai-trial/backend/domain"

const (
	CodeUserNotFound     = "USER_NOT_FOUND"
	CodeEmailTaken       = "EMAIL_TAKEN"
	CodePasswordMismatch = "PASSWORD_MISMATCH"
)

var (
	ErrUserNotFound     error = &domain.DomainError{Code: CodeUserNotFound, Message: "user not found"}
	ErrEmailTaken       error = &domain.DomainError{Code: CodeEmailTaken, Message: "email already registered"}
	ErrPasswordMismatch error = &domain.DomainError{Code: CodePasswordMismatch, Message: "password does not match"}
)
