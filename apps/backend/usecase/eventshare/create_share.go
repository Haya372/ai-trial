package eventshare

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"github.com/Haya372/ai-trial/backend/domain"
	domainevent "github.com/Haya372/ai-trial/backend/domain/event"
	domaineventshare "github.com/Haya372/ai-trial/backend/domain/eventshare"
)

type CreateShareInput struct {
	EventID   uuid.UUID
	ExpiresAt *time.Time
}

// CreateShareResult carries both the persisted share and the plaintext token,
// since the plaintext is only ever available at creation time (ADR-028).
type CreateShareResult struct {
	Share domaineventshare.EventShare
	Token string
}

type CreateShareCommand struct {
	eventRepo domainevent.Repository
	shareRepo domaineventshare.Repository
	logger    *slog.Logger
}

func NewCreateShareCommand(
	eventRepo domainevent.Repository, shareRepo domaineventshare.Repository, logger *slog.Logger,
) *CreateShareCommand {
	return &CreateShareCommand{eventRepo: eventRepo, shareRepo: shareRepo, logger: logger}
}

func (c *CreateShareCommand) Execute(
	ctx context.Context, userID uuid.UUID, in CreateShareInput,
) (*CreateShareResult, error) {
	ev, err := c.eventRepo.FindByID(ctx, in.EventID)
	if err != nil {
		return nil, fmt.Errorf("find event: %w", err)
	}
	if ev.UserID() != userID {
		c.logger.Warn("attempted to share event owned by another user",
			"event_id", in.EventID, "user_id", userID, "owner_id", ev.UserID())
		return nil, domainevent.ErrEventForbidden
	}

	expiresAt := ev.EndAt()
	if in.ExpiresAt != nil {
		expiresAt = *in.ExpiresAt
	}
	if expiresAt.Before(ev.StartAt()) {
		return nil, &domain.ValidationError{Details: []domain.ValidationDetail{
			{Field: "expiresAt", Code: "BEFORE_EVENT_START", Message: "expiresAt must not be before the event's start time"},
		}}
	}

	token, err := domaineventshare.GenerateToken()
	if err != nil {
		return nil, fmt.Errorf("generate token: %w", err)
	}
	share, err := domaineventshare.New(uuid.New(), in.EventID, domaineventshare.HashToken(token), expiresAt)
	if err != nil {
		return nil, err
	}

	created, err := c.shareRepo.Create(ctx, share)
	if err != nil {
		return nil, fmt.Errorf("create event share: %w", err)
	}
	return &CreateShareResult{Share: created, Token: token}, nil
}
