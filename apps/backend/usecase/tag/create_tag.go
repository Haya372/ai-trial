package tag

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/Haya372/ai-trial/backend/domain"
	domaintag "github.com/Haya372/ai-trial/backend/domain/tag"
	"github.com/Haya372/ai-trial/backend/usecase"
)

type CreateTagInput struct {
	Name  string
	Color string
}

type CreateTagCommand struct {
	tagRepo   domaintag.Repository
	txManager usecase.TransactionManager
}

func NewCreateTagCommand(r domaintag.Repository, tx usecase.TransactionManager) *CreateTagCommand {
	return &CreateTagCommand{tagRepo: r, txManager: tx}
}

func (c *CreateTagCommand) Execute(ctx context.Context, userID uuid.UUID, in CreateTagInput) (domaintag.Tag, error) {
	var details []domain.ValidationDetail
	name, nameErr := domaintag.NewTagName(in.Name)
	if nameErr != nil {
		details = append(details, toValidationDetail("name", nameErr))
	}
	color, colorErr := domaintag.NewTagColor(in.Color)
	if colorErr != nil {
		details = append(details, toValidationDetail("color", colorErr))
	}
	if len(details) > 0 {
		return nil, &domain.ValidationError{Details: details}
	}

	// Locking the owner row before counting serializes concurrent creates by
	// the same user, so they cannot both pass the limit check (ADR-024).
	var saved domaintag.Tag
	if err := c.txManager.RunInTx(ctx, func(ctx context.Context) error {
		if err := c.tagRepo.LockOwner(ctx, userID); err != nil {
			return err
		}
		count, err := c.tagRepo.CountByUserID(ctx, userID)
		if err != nil {
			return err
		}
		if count >= domaintag.MaxTagsPerUser {
			return domaintag.ErrTagLimitExceeded
		}
		saved, err = c.tagRepo.Create(ctx, domaintag.New(uuid.New(), userID, name, color))
		return err
	}); err != nil {
		return nil, fmt.Errorf("create tag: %w", err)
	}
	return saved, nil
}

func toValidationDetail(field string, err error) domain.ValidationDetail {
	if domErr, ok := errors.AsType[*domain.DomainError](err); ok {
		return domain.ValidationDetail{Field: field, Code: domErr.Code(), Message: domErr.Message()}
	}
	return domain.ValidationDetail{Field: field, Code: "INVALID", Message: err.Error()}
}
