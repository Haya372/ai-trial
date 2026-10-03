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
	"github.com/Haya372/ai-trial/backend/usecase"
)

type CreateShareInput struct {
	EventID   uuid.UUID
	ExpiresAt *time.Time
}

// CreateShareResult carries both the persisted share and the plaintext token,
// since the plaintext is only ever available at creation time (ADR-028).
type CreateShareResult struct {
	Share domaineventshare.EventShare
	Token domaineventshare.Token
}

type CreateShareCommand struct {
	eventRepo domainevent.Repository
	shareRepo domaineventshare.Repository
	txManager usecase.TransactionManager
	logger    *slog.Logger
}

func NewCreateShareCommand(
	eventRepo domainevent.Repository,
	shareRepo domaineventshare.Repository,
	txManager usecase.TransactionManager,
	logger *slog.Logger,
) *CreateShareCommand {
	return &CreateShareCommand{eventRepo: eventRepo, shareRepo: shareRepo, txManager: txManager, logger: logger}
}

// Execute runs inside a transaction because it reads the event via eventRepo
// and writes the share via shareRepo: without a shared transaction, the
// event could be deleted between the two calls, and shareRepo.Create's
// resulting FK-violation error is translated back into ErrEventNotFound by
// the repository (see eventShareRepository.Create).
func (c *CreateShareCommand) Execute(
	ctx context.Context, userID uuid.UUID, in CreateShareInput,
) (*CreateShareResult, error) {
	var result *CreateShareResult
	err := c.txManager.RunInTx(ctx, func(ctx context.Context) error {
		ev, err := domainevent.FindOwned(ctx, c.eventRepo, c.logger, in.EventID, userID)
		if err != nil {
			return err
		}

		expiresAt := ev.EndAt()
		if in.ExpiresAt != nil {
			expiresAt = *in.ExpiresAt
			// SPEC-004 requires an explicit expiresAt to be both on/after the
			// event's start and a future date; the default (event end) is
			// exempted so sharing a past event keeps working as before.
			if expiresAt.Before(ev.StartAt()) {
				return &domain.ValidationError{Details: []domain.ValidationDetail{
					{Field: "expiresAt", Code: "BEFORE_EVENT_START", Message: "expiresAt must not be before the event's start time"},
				}}
			}
			if !time.Now().Before(expiresAt) {
				return &domain.ValidationError{Details: []domain.ValidationDetail{
					{Field: "expiresAt", Code: "NOT_IN_FUTURE", Message: "expiresAt must be in the future"},
				}}
			}
		}

		token, err := domaineventshare.GenerateToken()
		if err != nil {
			return fmt.Errorf("generate token: %w", err)
		}
		share, err := domaineventshare.New(uuid.New(), in.EventID, token.Hash(), expiresAt)
		if err != nil {
			return err
		}

		created, err := c.shareRepo.Create(ctx, share)
		if err != nil {
			return fmt.Errorf("create event share: %w", err)
		}
		result = &CreateShareResult{Share: created, Token: token}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}
