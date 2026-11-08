package tag

import "github.com/Haya372/ai-trial/backend/domain"

const (
	CodeTagNotFound      = "TAG_NOT_FOUND"
	CodeTagNameTaken     = "TAG_NAME_TAKEN"
	CodeTagLimitExceeded = "TAG_LIMIT_EXCEEDED"
)

var (
	ErrTagNotFound      = domain.NewDomainError(CodeTagNotFound, "tag not found")
	ErrTagNameTaken     = domain.NewDomainError(CodeTagNameTaken, "tag name already exists")
	ErrTagLimitExceeded = domain.NewDomainError(CodeTagLimitExceeded, "tag limit exceeded")
)
