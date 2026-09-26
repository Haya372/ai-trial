package event

import "github.com/Haya372/ai-trial/backend/domain"

const CodeEventNotFound = "EVENT_NOT_FOUND"
const CodeEventForbidden = "EVENT_FORBIDDEN"

var ErrEventNotFound = domain.NewDomainError(CodeEventNotFound, "event not found")
var ErrEventForbidden = domain.NewDomainError(CodeEventForbidden, "event belongs to another user")
