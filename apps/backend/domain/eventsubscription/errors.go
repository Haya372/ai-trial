package eventsubscription

import "github.com/Haya372/ai-trial/backend/domain"

const (
	CodeEventSubscriptionNotFound      = "EVENT_SUBSCRIPTION_NOT_FOUND"
	CodeEventSubscriptionAlreadyExists = "EVENT_SUBSCRIPTION_ALREADY_EXISTS"
)

var (
	ErrEventSubscriptionNotFound = domain.NewDomainError(CodeEventSubscriptionNotFound, "event subscription not found")
	// ErrAlreadySubscribed is returned when a (event_id, user_id) pair already
	// has a subscription (SPEC-004: duplicate registration is rejected).
	ErrAlreadySubscribed = domain.NewDomainError(
		CodeEventSubscriptionAlreadyExists, "event subscription already exists",
	)
)
