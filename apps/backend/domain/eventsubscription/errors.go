package eventsubscription

import "github.com/Haya372/ai-trial/backend/domain"

const CodeEventSubscriptionNotFound = "EVENT_SUBSCRIPTION_NOT_FOUND"

var ErrEventSubscriptionNotFound = domain.NewDomainError(CodeEventSubscriptionNotFound, "event subscription not found")
