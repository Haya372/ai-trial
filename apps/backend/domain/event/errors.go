package event

import "github.com/Haya372/ai-trial/backend/domain"

const CodeEventNotFound = "EVENT_NOT_FOUND"

// The EventTag codes double as details[].code of the tagIds validation error
// (SPEC-008). The position code is the exception: only stored data can
// violate it, so it surfaces as an internal error.
const (
	CodeEventTagLimitExceeded   = "TOO_MANY"
	CodeEventTagDuplicated      = "DUPLICATED"
	CodeEventTagForeign         = "NOT_FOUND"
	CodeEventTagPositionInvalid = "EVENT_TAG_POSITION_INVALID"
)

var (
	ErrEventNotFound         = domain.NewDomainError(CodeEventNotFound, "event not found")
	ErrEventTagLimitExceeded = domain.NewDomainError(
		CodeEventTagLimitExceeded, "An event can have at most 5 tags")
	ErrEventTagDuplicated = domain.NewDomainError(
		CodeEventTagDuplicated, "The same tag cannot be added to an event more than once")
	// ErrEventTagForeign covers both a missing tag and another user's tag, so
	// the response never reveals whether someone else's tag exists.
	ErrEventTagForeign = domain.NewDomainError(
		CodeEventTagForeign, "Tag not found")
	ErrEventTagPositionInvalid = domain.NewDomainError(
		CodeEventTagPositionInvalid, "Event tag positions must be unique within 1-5")
)
