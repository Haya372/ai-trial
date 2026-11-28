package emailnotification

import "github.com/Haya372/ai-trial/backend/domain"

const (
	CodeEmailRecipientNotFound     = "EMAIL_RECIPIENT_NOT_FOUND"
	CodeEmailTemplateNotRegistered = "EMAIL_TEMPLATE_NOT_REGISTERED"
	CodeInvalidTemplateParams      = "EMAIL_INVALID_TEMPLATE_PARAMS"
)

var (
	// ErrEmailRecipientNotFound is returned when the user a send request
	// names does not exist; no notification is created.
	ErrEmailRecipientNotFound = domain.NewDomainError(
		CodeEmailRecipientNotFound, "email recipient not found",
	)
	ErrEmailTemplateNotRegistered = domain.NewDomainError(
		CodeEmailTemplateNotRegistered, "email template is not registered for the kind",
	)
	// ErrInvalidTemplateParams is returned when the parameter keys do not
	// match the template's keys exactly (SPEC-010: missing or extra keys).
	ErrInvalidTemplateParams = domain.NewDomainError(
		CodeInvalidTemplateParams, "template params do not match the template keys",
	)
)
