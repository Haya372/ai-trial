package emailnotification_test

import (
	"errors"
	"slices"
	"testing"

	"github.com/Haya372/ai-trial/backend/domain/emailnotification"
	"github.com/Haya372/ai-trial/backend/domain/user"
)

func TestNewEmailStatus_acceptsKnownValues(t *testing.T) {
	for _, s := range []string{"pending", "sent", "failed"} {
		got, err := emailnotification.NewEmailStatus(s)
		if err != nil {
			t.Fatalf("NewEmailStatus(%q) error = %v", s, err)
		}
		if string(got) != s {
			t.Errorf("NewEmailStatus(%q) = %q", s, got)
		}
	}
}

func TestNewEmailStatus_rejectsUnknownValue(t *testing.T) {
	if _, err := emailnotification.NewEmailStatus("queued"); err == nil {
		t.Fatal("NewEmailStatus(\"queued\") error = nil, want error")
	}
}

func TestEmailStatus_IsTerminal(t *testing.T) {
	tests := map[emailnotification.EmailStatus]bool{
		emailnotification.StatusPending: false,
		emailnotification.StatusSent:    true,
		emailnotification.StatusFailed:  true,
	}
	for status, want := range tests {
		if got := status.IsTerminal(); got != want {
			t.Errorf("%s.IsTerminal() = %v, want %v", status, got, want)
		}
	}
}

func TestFailureCode_classification(t *testing.T) {
	tests := map[string]bool{
		"connection_failed":      true,
		"timeout":                true,
		"rate_limited":           true,
		"temporary_rejection":    true,
		"invalid_recipient":      false,
		"recipient_rejected":     false,
		"auth_failed":            false,
		"permanent_rejection":    false,
		"template_render_failed": false,
		"worker_interrupted":     false,
	}
	for s, wantTransient := range tests {
		code, err := emailnotification.NewFailureCode(s)
		if err != nil {
			t.Fatalf("NewFailureCode(%q) error = %v", s, err)
		}
		if string(code) != s {
			t.Errorf("NewFailureCode(%q) = %q", s, code)
		}
		if got := code.IsTransient(); got != wantTransient {
			t.Errorf("%s.IsTransient() = %v, want %v", s, got, wantTransient)
		}
	}
}

func TestNewFailureCode_rejectsUnknownValue(t *testing.T) {
	if _, err := emailnotification.NewFailureCode("smtp 550 user@example.com"); err == nil {
		t.Fatal("NewFailureCode(unknown) error = nil, want error")
	}
}

func TestNewEmailKind_rejectsEmpty(t *testing.T) {
	if _, err := emailnotification.NewEmailKind(""); err == nil {
		t.Fatal("NewEmailKind(\"\") error = nil, want error")
	}
	kind, err := emailnotification.NewEmailKind("password_reset")
	if err != nil {
		t.Fatalf("NewEmailKind error = %v", err)
	}
	if string(kind) != "password_reset" {
		t.Errorf("NewEmailKind = %q", kind)
	}
}

func TestNewRecipientEmail_validatesFormat(t *testing.T) {
	got, err := emailnotification.NewRecipientEmail("sakura@example.com")
	if err != nil {
		t.Fatalf("NewRecipientEmail error = %v", err)
	}
	if got.String() != "sakura@example.com" {
		t.Errorf("String() = %q", got.String())
	}

	_, err = emailnotification.NewRecipientEmail("not-an-email")
	if !errors.Is(err, user.ErrInvalidEmail) {
		t.Errorf("NewRecipientEmail(invalid) error = %v, want %v", err, user.ErrInvalidEmail)
	}
}

func TestTemplateParams_copiesInputAndOutput(t *testing.T) {
	src := map[string]string{"reset_url": "https://example.com/r", "name": "sakura"}
	params := emailnotification.NewTemplateParams(src)

	src["reset_url"] = "mutated"
	if params.Values()["reset_url"] != "https://example.com/r" {
		t.Error("mutating the input map changed TemplateParams")
	}

	out := params.Values()
	out["name"] = "mutated"
	if params.Values()["name"] != "sakura" {
		t.Error("mutating Values() changed TemplateParams")
	}

	if got, want := params.Keys(), []string{"name", "reset_url"}; !slices.Equal(got, want) {
		t.Errorf("Keys() = %v, want %v", got, want)
	}
	if params.IsEmpty() {
		t.Error("IsEmpty() = true, want false")
	}
}

func TestTemplateParams_emptyWhenNilOrEmpty(t *testing.T) {
	if !emailnotification.NewTemplateParams(nil).IsEmpty() {
		t.Error("NewTemplateParams(nil).IsEmpty() = false")
	}
	if !emailnotification.NewTemplateParams(map[string]string{}).IsEmpty() {
		t.Error("NewTemplateParams({}).IsEmpty() = false")
	}
	if got := emailnotification.NewTemplateParams(nil).Values(); got == nil || len(got) != 0 {
		t.Errorf("Values() = %v, want empty non-nil map", got)
	}
}
