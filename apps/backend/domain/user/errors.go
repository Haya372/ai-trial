package user

import "github.com/Haya372/ai-trial/backend/domain"

var (
	ErrUserNotFound = &domain.DomainError{Code: "USER_NOT_FOUND", Message: "user not found"}
	ErrEmailTaken   = &domain.DomainError{Code: "EMAIL_TAKEN", Message: "email already registered"}
)
