package user

import "github.com/Haya372/ai-trial/backend/domain"

const (
	CodeUserNotFound = "USER_NOT_FOUND"
	CodeEmailTaken   = "EMAIL_TAKEN"
)

var (
	ErrUserNotFound = &domain.DomainError{Code: CodeUserNotFound, Message: "user not found"}
	ErrEmailTaken   = &domain.DomainError{Code: CodeEmailTaken, Message: "email already registered"}
)
