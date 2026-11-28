package emailnotification

import "github.com/Haya372/ai-trial/backend/domain/user"

// RecipientEmail is a snapshot of the user's address at request time, so a
// later change of the user's address does not redirect a pending email.
// It is personal data: keep it out of logs and traces.
type RecipientEmail struct {
	value user.Email
}

func NewRecipientEmail(s string) (RecipientEmail, error) {
	e, err := user.NewEmail(s)
	if err != nil {
		return RecipientEmail{}, err
	}
	return RecipientEmail{value: e}, nil
}

func (r RecipientEmail) String() string { return string(r.value) }
