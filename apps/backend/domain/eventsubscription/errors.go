package eventsubscription

import "github.com/Haya372/ai-trial/backend/domain"

const (
	CodeEventSubscriptionNotFound      = "EVENT_SUBSCRIPTION_NOT_FOUND"
	CodeEventSubscriptionAlreadyExists = "EVENT_SUBSCRIPTION_ALREADY_EXISTS"
	CodeCannotSubscribeToOwnEvent      = "EVENT_SUBSCRIPTION_OWN_EVENT"
)

var (
	ErrEventSubscriptionNotFound = domain.NewDomainError(CodeEventSubscriptionNotFound, "event subscription not found")
	// ErrAlreadySubscribed is returned when a (event_id, user_id) pair already
	// has a subscription (SPEC-004: duplicate registration is rejected).
	ErrAlreadySubscribed = domain.NewDomainError(
		CodeEventSubscriptionAlreadyExists, "event subscription already exists",
	)
	// ErrCannotSubscribeToOwnEvent is returned when a user tries to add their
	// own event to their calendar via a share link (SPEC-004: the "add to my
	// calendar" button isn't shown to the event's owner).
	ErrCannotSubscribeToOwnEvent = domain.NewDomainError(
		CodeCannotSubscribeToOwnEvent, "cannot subscribe to your own event",
	)
)
