package eventsubscription

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	domainevent "github.com/Haya372/ai-trial/backend/domain/event"
	domaineventshare "github.com/Haya372/ai-trial/backend/domain/eventshare"
	domaineventsubscription "github.com/Haya372/ai-trial/backend/domain/eventsubscription"
	eventshareuc "github.com/Haya372/ai-trial/backend/usecase/eventshare"
)

type SubscribeToShareInput struct {
	Token  domaineventshare.Token
	UserID uuid.UUID
}

type SubscribeToShareOutput struct {
	SubscriptionID uuid.UUID
	EventID        uuid.UUID
	CreatedAt      time.Time
	Created        bool
}

type SubscribeToShareCommand struct {
	loader   *eventshareuc.ShareTokenLoader
	subsRepo domaineventsubscription.Repository
}

func NewSubscribeToShareCommand(
	l *eventshareuc.ShareTokenLoader, s domaineventsubscription.Repository,
) *SubscribeToShareCommand {
	return &SubscribeToShareCommand{loader: l, subsRepo: s}
}

// Execute implements SPEC-004's idempotent "add to my calendar": a repeat
// request for an already-subscribed (event, user) pair returns the existing
// subscription instead of erroring.
func (c *SubscribeToShareCommand) Execute(
	ctx context.Context, in SubscribeToShareInput,
) (SubscribeToShareOutput, error) {
	_, ev, err := c.loader.Load(ctx, in.Token)
	if err != nil {
		return SubscribeToShareOutput{}, err
	}

	if ev.UserID == in.UserID {
		return SubscribeToShareOutput{}, domainevent.ErrEventForbidden
	}

	existing, err := c.subsRepo.FindByEventAndUserID(ctx, ev.ID, in.UserID)
	if err == nil {
		return toSubscribeOutput(existing, false), nil
	}
	if !errors.Is(err, domaineventsubscription.ErrEventSubscriptionNotFound) {
		return SubscribeToShareOutput{}, err
	}

	sub := domaineventsubscription.New(uuid.New(), ev.ID, in.UserID, time.Now())
	created, err := c.subsRepo.Create(ctx, sub)
	if err != nil {
		// A concurrent request may have created the subscription between the
		// FindByEventAndUserID check above and this Create; treat that race
		// the same as the idempotent path instead of surfacing a conflict.
		if errors.Is(err, domaineventsubscription.ErrAlreadySubscribed) {
			existing, ferr := c.subsRepo.FindByEventAndUserID(ctx, ev.ID, in.UserID)
			if ferr != nil {
				return SubscribeToShareOutput{}, ferr
			}
			return toSubscribeOutput(existing, false), nil
		}
		return SubscribeToShareOutput{}, err
	}
	return toSubscribeOutput(created, true), nil
}

func toSubscribeOutput(s domaineventsubscription.EventSubscription, created bool) SubscribeToShareOutput {
	return SubscribeToShareOutput{
		SubscriptionID: s.ID(),
		EventID:        s.EventID(),
		CreatedAt:      s.CreatedAt(),
		Created:        created,
	}
}
