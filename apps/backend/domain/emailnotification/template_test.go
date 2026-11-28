package emailnotification_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/Haya372/ai-trial/backend/domain"
	"github.com/Haya372/ai-trial/backend/domain/emailnotification"
)

const kindReset emailnotification.EmailKind = "password_reset"

const (
	testURL  = "https://example.com/r"
	keyName  = "name"
	keyURL   = "url"
	bodyText = "body"
)

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
	tmpl := mustTemplate(t, "Hi {{.name}}", "{{.url}}", keyURL, keyName)
	got := tmpl.Keys()
	if len(got) != 2 || got[0] != keyName || got[1] != keyURL {
		t.Errorf("Keys() = %v, want [name url]", got)
	}
}

func TestNewEmailTemplate_rejectsInvalidDefinitions(t *testing.T) {
	tests := []struct {
		name          string
		subject, body string
		keys          []string
	}{
		{"subject syntax error", "Hi {{.name", bodyText, []string{keyName}},
		{"body syntax error", "Hi", "{{if}}", nil},
		{"undeclared key in subject", "Hi {{.name}}", bodyText, nil},
		{"undeclared key in body", "Hi", "{{.url}}", []string{keyName}},
		{"duplicated key", "{{.name}}", bodyText, []string{keyName, keyName}},
		{"empty key", "Hi", bodyText, []string{""}},
		{"undeclared key in an untaken branch", "Hi", `{{if eq .plan "pro"}}{{.pro_url}}{{end}}`, []string{"plan"}},
		{"undeclared key in an else branch", "Hi", `{{if .url}}a{{else}}{{.other}}{{end}}`, []string{keyURL}},
		{"declared but unused key", "Hi", "{{.url}}", []string{keyName, keyURL}},
		{"nested field", "Hi", "{{.url.host}}", []string{keyURL}},
		{"range", "Hi", "{{range .url}}x{{end}}", []string{keyURL}},
		{"with", "Hi", "{{with .url}}{{.}}{{end}}", []string{keyURL}},
		{"variable", "Hi", "{{$u := .url}}{{$u}}", []string{keyURL}},
		{"root variable", "Hi", "{{$.url}}", []string{keyURL}},
		{"bare dot", "Hi", "{{.}}", nil},
		{"define and template", "Hi", `{{define "x"}}{{.url}}{{end}}{{template "x" .}}`, []string{keyURL}},
		{"define named like the body", "Hi", `{{define "body"}}{{.url}}{{end}}`, []string{keyURL}},
		{"define named like the subject", `{{define "subject"}}Hi{{end}}`, bodyText, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := emailnotification.NewEmailTemplate(tt.subject, tt.body, tt.keys); err == nil {
				t.Fatal("NewEmailTemplate error = nil, want error")
			}
		})
	}
}

func TestNewEmailTemplate_acceptsConditionalsAndFunctions(t *testing.T) {
	tmpl := mustTemplate(t,
		`{{if eq .plan "pro"}}Pro{{else}}Free{{end}} notice`,
		`{{if eq .plan "pro"}}{{.pro_url}}{{else if .name}}{{printf "%s" .name}}{{end}}`,
		"plan", "pro_url", keyName,
	)
	got, err := tmpl.Render(emailnotification.NewTemplateParams(map[string]string{
		"plan": "pro", "pro_url": testURL, keyName: "taro",
	}))
	if err != nil {
		t.Fatalf("Render error = %v", err)
	}
	if got.Subject != "Pro notice" || got.Body != testURL {
		t.Errorf("Render = %+v", got)
	}
}

func TestEmailTemplate_CheckParams(t *testing.T) {
	tmpl := mustTemplate(t, "Hi {{.name}}", "{{.url}}", keyName, keyURL)
	tests := []struct {
		name    string
		params  map[string]string
		wantErr bool
	}{
		{"exact keys", map[string]string{keyName: "a", keyURL: "b"}, false},
		{"missing key", map[string]string{keyName: "a"}, true},
		{"extra key", map[string]string{keyName: "a", keyURL: "b", "token": "c"}, true},
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
	tmpl := mustTemplate(t, "Hi", "{{.url}}", keyURL)
	err := tmpl.CheckParams(emailnotification.NewTemplateParams(map[string]string{"unknown": "https://x/reset?t=abc"}))
	if err == nil || strings.Contains(err.Error(), "abc") {
		t.Errorf("error = %v, want an error without the parameter value", err)
	}
}

func TestEmailTemplate_Render_embedsParamsWithoutEscaping(t *testing.T) {
	tmpl := mustTemplate(t, "Hello {{.name}}", "Reset: {{.url}}\n", keyName, keyURL)
	got, err := tmpl.Render(emailnotification.NewTemplateParams(map[string]string{
		keyName: "<sakura>", keyURL: "https://example.com/r?a=1&b=2",
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
	tmpl := mustTemplate(t, "Hi", "{{.url}}", keyURL)
	_, err := tmpl.Render(emailnotification.NewTemplateParams(map[string]string{"other": "x"}))
	if !errors.Is(err, emailnotification.ErrInvalidTemplateParams) {
		t.Errorf("Render error = %v, want ErrInvalidTemplateParams", err)
	}
}

func TestEmailTemplate_Render_rejectsNewlineInSubject(t *testing.T) {
	tmpl := mustTemplate(t, "Hi {{.name}}", bodyText, keyName)
	for _, name := range []string{"a\r\nBcc: x@example.com", "a\nb", "a\rb"} {
		got, err := tmpl.Render(emailnotification.NewTemplateParams(map[string]string{keyName: name}))
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
	tmpl := mustTemplate(t, "Hi", "{{.url}}", keyURL)
	registry := emailnotification.NewTemplateRegistry(map[emailnotification.EmailKind]emailnotification.EmailTemplate{
		kindReset: tmpl,
	})

	got, err := registry.Template(kindReset)
	if err != nil {
		t.Fatalf("Template error = %v", err)
	}
	if len(got.Keys()) != 1 || got.Keys()[0] != keyURL {
		t.Errorf("Template returned keys %v", got.Keys())
	}

	_, err = registry.Template("account_locked")
	if !errors.Is(err, emailnotification.ErrEmailTemplateNotRegistered) {
		t.Errorf("Template(unregistered) error = %v, want ErrEmailTemplateNotRegistered", err)
	}
}

func TestTemplateRegistry_CheckRequest(t *testing.T) {
	registry := emailnotification.NewTemplateRegistry(map[emailnotification.EmailKind]emailnotification.EmailTemplate{
		kindReset: mustTemplate(t, "Hi", "{{.url}}", keyURL),
	})
	tests := []struct {
		name    string
		kind    emailnotification.EmailKind
		params  map[string]string
		wantErr error
	}{
		{"valid", kindReset, map[string]string{keyURL: "x"}, nil},
		{
			"unregistered kind", "account_locked", map[string]string{keyURL: "x"},
			emailnotification.ErrEmailTemplateNotRegistered,
		},
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
		kindReset: mustTemplate(t, "Hi", bodyText),
	}
	registry := emailnotification.NewTemplateRegistry(templates)
	delete(templates, kindReset)
	if _, err := registry.Template(kindReset); err != nil {
		t.Errorf("Template error = %v after mutating the input map", err)
	}
}
