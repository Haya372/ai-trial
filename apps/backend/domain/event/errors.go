package event

import "github.com/Haya372/ai-trial/backend/domain"

const CodeEventNotFound = "EVENT_NOT_FOUND"

var ErrEventNotFound = domain.NewDomainError(CodeEventNotFound, "event not found")
