package domain

// DomainError is a domain-layer error that carries a stable code for
// presentation-layer mapping without importing domain sub-packages.
// Fields are unexported so package-level sentinel errors (e.g.
// user.ErrPasswordTooShort) can't be mutated by callers that extract
// the pointer via errors.As.
type DomainError struct {
	code    string
	message string
}

// NewDomainError constructs a DomainError. It returns the error interface
// (rather than *DomainError) so callers declare sentinel vars without an
// explicit type annotation.
func NewDomainError(code, message string) error {
	return &DomainError{code: code, message: message}
}

func (e *DomainError) Code() string    { return e.code }
func (e *DomainError) Message() string { return e.message }
func (e *DomainError) Error() string   { return e.message }
