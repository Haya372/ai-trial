package eventshare

import "github.com/Haya372/ai-trial/backend/domain"

const (
	CodeEventShareNotFound = "EVENT_SHARE_NOT_FOUND"
	CodeEventShareExpired  = "EVENT_SHARE_EXPIRED"
)

var (
	ErrEventShareNotFound = domain.NewDomainError(CodeEventShareNotFound, "event share not found")
	ErrEventShareExpired  = domain.NewDomainError(CodeEventShareExpired, "event share expired")
)
