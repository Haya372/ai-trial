package emailnotification_test

import (
	"testing"

	"github.com/Haya372/ai-trial/backend/domain/emailnotification"
)

const (
	senderName    = "AI Trial"
	senderAddress = "no-reply@example.com"
)

func TestSendResult_accepted(t *testing.T) {
	r := emailnotification.SendAccepted()
	if r.Outcome() != emailnotification.OutcomeAccepted {
		t.Errorf("Outcome() = %v, want accepted", r.Outcome())
	}
	if _, ok := r.FailureCode(); ok {
		t.Error("FailureCode() is set for an accepted result")
	}
}

func TestSendResult_failedOutcomeFollowsFailureCode(t *testing.T) {
	tests := map[emailnotification.FailureCode]emailnotification.SendOutcome{
		emailnotification.FailureTimeout:           emailnotification.OutcomeTransientFailure,
		emailnotification.FailureRecipientRejected: emailnotification.OutcomePermanentFailure,
	}
	for code, want := range tests {
		r := emailnotification.SendFailed(code)
		if r.Outcome() != want {
			t.Errorf("SendFailed(%s).Outcome() = %v, want %v", code, r.Outcome(), want)
		}
		if got, ok := r.FailureCode(); !ok || got != code {
			t.Errorf("SendFailed(%s).FailureCode() = %v, %v", code, got, ok)
		}
	}
}

func TestNewSender(t *testing.T) {
	s, err := emailnotification.NewSender(senderName, senderAddress)
	if err != nil {
		t.Fatalf("NewSender error = %v", err)
	}
	if s.Name() != senderName || s.Address() != senderAddress {
		t.Errorf("NewSender = %q <%s>", s.Name(), s.Address())
	}
}

func TestNewSender_rejectsInvalidValues(t *testing.T) {
	tests := map[string][2]string{
		"empty name":         {"", senderAddress},
		"line break in name": {"AI Trial\r\nBcc: x@example.com", senderAddress},
		"invalid address":    {senderName, "no-reply"},
		"empty address":      {senderName, ""},
	}
	for name, v := range tests {
		t.Run(name, func(t *testing.T) {
			if _, err := emailnotification.NewSender(v[0], v[1]); err == nil {
				t.Fatal("NewSender error = nil, want error")
			}
		})
	}
}
