package emailnotification

import "errors"

// EmailKind identifies which EmailTemplate a notification uses. The values
// are defined by the specs that add each kind of email (ADR-019), so this
// package only rejects the empty kind; TemplateRegistry rejects kinds that
// have no template.
type EmailKind string

var errEmptyEmailKind = errors.New("email kind must not be empty")

func NewEmailKind(s string) (EmailKind, error) {
	if s == "" {
		return "", errEmptyEmailKind
	}
	return EmailKind(s), nil
}
