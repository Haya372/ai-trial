package emailnotification_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/Haya372/ai-trial/backend/domain"
	"github.com/Haya372/ai-trial/backend/domain/emailnotification"
)

const kindReset emailnotification.EmailKind = "password_reset"

func mustTemplate(t *testing.T, subject, body string, keys ...string) emailnotification.EmailTemplate {
	t.Helper()
	tmpl, err := emailnotification.NewEmailTemplate(subject, body, keys)
	if err != nil {
		t.Fatalf("NewEmailTemplate error = %v", err)
	}
	return tmpl
}

func TestErrors_areDomainErrorsWithCodes(t *testing.T) {
	tests := map[error]string{
		emailnotification.ErrEmailRecipientNotFound:     emailnotification.CodeEmailRecipientNotFound,
		emailnotification.ErrEmailTemplateNotRegistered: emailnotification.CodeEmailTemplateNotRegistered,
		emailnotification.ErrInvalidTemplateParams:      emailnotification.CodeInvalidTemplateParams,
	}
	for err, code := range tests {
		var de *domain.DomainError
		if !errors.As(err, &de) {
			t.Fatalf("%v is not a DomainError", err)
		}
		if de.Code() != code {
			t.Errorf("Code() = %q, want %q", de.Code(), code)
		}
	}
}

func TestNewEmailTemplate_exposesKeysSorted(t *testing.T) {
	tmpl := mustTemplate(t, "Hi {{.name}}", "{{.url}}", "url", "name")
	got := tmpl.Keys()
	if len(got) != 2 || got[0] != "name" || got[1] != "url" {
		t.Errorf("Keys() = %v, want [name url]", got)
	}
}

func TestNewEmailTemplate_rejectsInvalidDefinitions(t *testing.T) {
	tests := []struct {
		name          string
		subject, body string
		keys          []string
	}{
		{"subject syntax error", "Hi {{.name", "body", []string{"name"}},
		{"body syntax error", "Hi", "{{if}}", nil},
		{"undeclared key in subject", "Hi {{.name}}", "body", nil},
		{"undeclared key in body", "Hi", "{{.url}}", []string{"name"}},
		{"duplicated key", "{{.name}}", "body", []string{"name", "name"}},
		{"empty key", "Hi", "body", []string{""}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := emailnotification.NewEmailTemplate(tt.subject, tt.body, tt.keys); err == nil {
				t.Fatal("NewEmailTemplate error = nil, want error")
			}
		})
	}
}

func TestEmailTemplate_CheckParams(t *testing.T) {
	tmpl := mustTemplate(t, "Hi {{.name}}", "{{.url}}", "name", "url")
	tests := []struct {
		name    string
		params  map[string]string
		wantErr bool
	}{
		{"exact keys", map[string]string{"name": "a", "url": "b"}, false},
		{"missing key", map[string]string{"name": "a"}, true},
		{"extra key", map[string]string{"name": "a", "url": "b", "token": "c"}, true},
		{"empty", nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tmpl.CheckParams(emailnotification.NewTemplateParams(tt.params))
			if tt.wantErr != (err != nil) {
				t.Fatalf("CheckParams error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && !errors.Is(err, emailnotification.ErrInvalidTemplateParams) {
				t.Errorf("error = %v, want ErrInvalidTemplateParams", err)
			}
		})
	}
}

func TestEmailTemplate_CheckParams_errorOmitsValues(t *testing.T) {
	tmpl := mustTemplate(t, "Hi", "{{.url}}", "url")
	err := tmpl.CheckParams(emailnotification.NewTemplateParams(map[string]string{"secret": "https://x/reset?t=abc"}))
	if err == nil || strings.Contains(err.Error(), "abc") {
		t.Errorf("error = %v, want an error without the parameter value", err)
	}
}

func TestEmailTemplate_Render_embedsParamsWithoutEscaping(t *testing.T) {
	tmpl := mustTemplate(t, "Hello {{.name}}", "Reset: {{.url}}\n", "name", "url")
	got, err := tmpl.Render(emailnotification.NewTemplateParams(map[string]string{
		"name": "<sakura>", "url": "https://example.com/r?a=1&b=2",
	}))
	if err != nil {
		t.Fatalf("Render error = %v", err)
	}
	if got.Subject != "Hello <sakura>" {
		t.Errorf("Subject = %q", got.Subject)
	}
	if got.Body != "Reset: https://example.com/r?a=1&b=2\n" {
		t.Errorf("Body = %q", got.Body)
	}
}

func TestEmailTemplate_Render_failsOnMismatchedParams(t *testing.T) {
	tmpl := mustTemplate(t, "Hi", "{{.url}}", "url")
	_, err := tmpl.Render(emailnotification.NewTemplateParams(map[string]string{"other": "x"}))
	if !errors.Is(err, emailnotification.ErrInvalidTemplateParams) {
		t.Errorf("Render error = %v, want ErrInvalidTemplateParams", err)
	}
}

func TestEmailTemplate_Render_rejectsNewlineInSubject(t *testing.T) {
	tmpl := mustTemplate(t, "Hi {{.name}}", "body", "name")
	for _, name := range []string{"a\r\nBcc: x@example.com", "a\nb", "a\rb"} {
		got, err := tmpl.Render(emailnotification.NewTemplateParams(map[string]string{"name": name}))
		if err == nil {
			t.Errorf("Render(%q) = %+v, want error", name, got)
			continue
		}
		if strings.Contains(err.Error(), "Bcc") {
			t.Errorf("error = %v, want an error without the subject", err)
		}
	}
}

func TestEmailTemplate_Render_zeroValueFails(t *testing.T) {
	var tmpl emailnotification.EmailTemplate
	if _, err := tmpl.Render(emailnotification.NewTemplateParams(nil)); err == nil {
		t.Fatal("zero EmailTemplate Render error = nil, want error")
	}
}

func TestTemplateRegistry_Template(t *testing.T) {
	tmpl := mustTemplate(t, "Hi", "{{.url}}", "url")
	registry := emailnotification.NewTemplateRegistry(map[emailnotification.EmailKind]emailnotification.EmailTemplate{
		kindReset: tmpl,
	})

	got, err := registry.Template(kindReset)
	if err != nil {
		t.Fatalf("Template error = %v", err)
	}
	if len(got.Keys()) != 1 || got.Keys()[0] != "url" {
		t.Errorf("Template returned keys %v", got.Keys())
	}

	_, err = registry.Template("account_locked")
	if !errors.Is(err, emailnotification.ErrEmailTemplateNotRegistered) {
		t.Errorf("Template(unregistered) error = %v, want ErrEmailTemplateNotRegistered", err)
	}
}

func TestTemplateRegistry_CheckRequest(t *testing.T) {
	registry := emailnotification.NewTemplateRegistry(map[emailnotification.EmailKind]emailnotification.EmailTemplate{
		kindReset: mustTemplate(t, "Hi", "{{.url}}", "url"),
	})
	tests := []struct {
		name    string
		kind    emailnotification.EmailKind
		params  map[string]string
		wantErr error
	}{
		{"valid", kindReset, map[string]string{"url": "x"}, nil},
		{"unregistered kind", "account_locked", map[string]string{"url": "x"}, emailnotification.ErrEmailTemplateNotRegistered},
		{"mismatched keys", kindReset, map[string]string{"link": "x"}, emailnotification.ErrInvalidTemplateParams},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := registry.CheckRequest(tt.kind, emailnotification.NewTemplateParams(tt.params))
			if tt.wantErr == nil && err != nil {
				t.Fatalf("CheckRequest error = %v, want nil", err)
			}
			if tt.wantErr != nil && !errors.Is(err, tt.wantErr) {
				t.Fatalf("CheckRequest error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewTemplateRegistry_copiesInput(t *testing.T) {
	templates := map[emailnotification.EmailKind]emailnotification.EmailTemplate{
		kindReset: mustTemplate(t, "Hi", "body"),
	}
	registry := emailnotification.NewTemplateRegistry(templates)
	delete(templates, kindReset)
	if _, err := registry.Template(kindReset); err != nil {
		t.Errorf("Template error = %v after mutating the input map", err)
	}
}
