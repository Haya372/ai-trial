package emailnotification

import (
	"errors"
	"fmt"
	"maps"
	"slices"
	"strings"
	"text/template"
)

// EmailTemplate builds the plain-text subject and body of one kind of email
// with text/template (ADR-041).
type EmailTemplate struct {
	subject *template.Template
	body    *template.Template
	keys    []string
}

// RenderedEmail is the subject and body of one email. It contains the
// template params, so keep it out of logs, traces and storage.
type RenderedEmail struct {
	Subject string
	Body    string
}

var (
	errInvalidTemplateKeys   = errors.New("invalid email template keys")
	errTemplateUninitialized = errors.New("email template is not initialized")
	errSubjectHasLineBreak   = errors.New("rendered email subject contains a line break")
)

// NewEmailTemplate parses the templates and trial-renders them with every
// key set, so a syntax error or a reference to an undeclared key fails at
// startup instead of failing each send with template_render_failed.
func NewEmailTemplate(subject, body string, keys []string) (EmailTemplate, error) {
	sorted := slices.Sorted(slices.Values(keys))
	for i, k := range sorted {
		if k == "" || (i > 0 && k == sorted[i-1]) {
			return EmailTemplate{}, fmt.Errorf("%w: %q", errInvalidTemplateKeys, keys)
		}
	}
	subjectTmpl, err := template.New("subject").Option("missingkey=error").Parse(subject)
	if err != nil {
		return EmailTemplate{}, fmt.Errorf("parse email subject template: %w", err)
	}
	bodyTmpl, err := template.New("body").Option("missingkey=error").Parse(body)
	if err != nil {
		return EmailTemplate{}, fmt.Errorf("parse email body template: %w", err)
	}
	t := EmailTemplate{subject: subjectTmpl, body: bodyTmpl, keys: sorted}

	dummy := make(map[string]string, len(sorted))
	for _, k := range sorted {
		dummy[k] = "x"
	}
	if _, err := t.Render(NewTemplateParams(dummy)); err != nil {
		return EmailTemplate{}, fmt.Errorf("trial-render email template: %w", err)
	}
	return t, nil
}

// Keys returns the parameter keys in ascending order.
func (t EmailTemplate) Keys() []string { return slices.Clone(t.keys) }

// CheckParams returns ErrInvalidTemplateParams unless params has exactly
// the template's keys. The error names keys only, never values.
func (t EmailTemplate) CheckParams(params TemplateParams) error {
	got := params.Keys()
	if slices.Equal(got, t.keys) {
		return nil
	}
	return fmt.Errorf("%w: want keys %q, got %q", ErrInvalidTemplateParams, t.keys, got)
}

func (t EmailTemplate) Render(params TemplateParams) (RenderedEmail, error) {
	if t.subject == nil || t.body == nil {
		return RenderedEmail{}, errTemplateUninitialized
	}
	if err := t.CheckParams(params); err != nil {
		return RenderedEmail{}, err
	}
	values := params.Values()
	var subject, body strings.Builder
	if err := t.subject.Execute(&subject, values); err != nil {
		return RenderedEmail{}, fmt.Errorf("render email subject: %w", err)
	}
	// The subject becomes a mail header, so a line break from a param
	// could inject headers (ADR-041).
	if strings.ContainsAny(subject.String(), "\r\n") {
		return RenderedEmail{}, errSubjectHasLineBreak
	}
	if err := t.body.Execute(&body, values); err != nil {
		return RenderedEmail{}, fmt.Errorf("render email body: %w", err)
	}
	return RenderedEmail{Subject: subject.String(), Body: body.String()}, nil
}

// TemplateRegistry maps each EmailKind to its EmailTemplate. It is built
// once at startup and read concurrently afterwards.
type TemplateRegistry struct {
	templates map[EmailKind]EmailTemplate
}

func NewTemplateRegistry(templates map[EmailKind]EmailTemplate) TemplateRegistry {
	return TemplateRegistry{templates: maps.Clone(templates)}
}

func (r TemplateRegistry) Template(kind EmailKind) (EmailTemplate, error) {
	t, ok := r.templates[kind]
	if !ok {
		return EmailTemplate{}, fmt.Errorf("%w: %q", ErrEmailTemplateNotRegistered, kind)
	}
	return t, nil
}

// CheckRequest validates a send request before a notification is created.
func (r TemplateRegistry) CheckRequest(kind EmailKind, params TemplateParams) error {
	t, err := r.Template(kind)
	if err != nil {
		return err
	}
	return t.CheckParams(params)
}
