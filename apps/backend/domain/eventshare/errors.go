package eventshare

import "github.com/Haya372/ai-trial/backend/domain"

const CodeEventShareNotFound = "EVENT_SHARE_NOT_FOUND"

var ErrEventShareNotFound = domain.NewDomainError(CodeEventShareNotFound, "event share not found")
