package domain

// DomainError is a domain-layer error that carries a stable code for
// presentation-layer mapping without importing domain sub-packages.
type DomainError struct {
	Code    string
	Message string
}

func (e *DomainError) Error() string { return e.Message }
