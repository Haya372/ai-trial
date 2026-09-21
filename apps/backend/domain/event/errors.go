package event

import "github.com/Haya372/ai-trial/backend/domain"

const (
	CodeEventNotFound = "EVENT_NOT_FOUND"
	CodeForbidden     = "FORBIDDEN"
)

var (
	ErrEventNotFound = domain.NewDomainError(CodeEventNotFound, "event not found")
	ErrForbidden     = domain.NewDomainError(CodeForbidden, "you do not have permission to access this event")
)
