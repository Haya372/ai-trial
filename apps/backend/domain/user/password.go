package user

import (
	"fmt"
	"unicode"

	"golang.org/x/crypto/bcrypt"

	"github.com/Haya372/ai-trial/backend/domain"
)

// Password holds a bcrypt hash for persistence (signup, password change).
type Password struct {
	hash string
}

// LoginPassword holds a validated plaintext password for comparison only. It
// has no Hash() method, so it can never be passed to Repository.Create (which
// requires a Password) — misuse is a compile error rather than a runtime
// failure.
type LoginPassword struct {
	plain string
}

const (
	CodePasswordTooShort               = "TOO_SHORT"
	CodePasswordTooLong                = "TOO_LONG"
	CodePasswordNotASCII               = "INVALID_CHARACTER" //nolint:gosec
	CodePasswordInsufficientComplexity = "INSUFFICIENT_COMPLEXITY"
)

var (
	ErrPasswordTooShort = domain.NewDomainError(
		CodePasswordTooShort, "Password must be at least 8 characters")
	ErrPasswordTooLong = domain.NewDomainError(
		CodePasswordTooLong, "Password must be at most 72 characters")
	ErrPasswordNotASCII = domain.NewDomainError(
		CodePasswordNotASCII, "Password must contain only ASCII printable characters")
	ErrPasswordInsufficientComplexity = domain.NewDomainError(
		CodePasswordInsufficientComplexity, "Password must contain uppercase, lowercase, digit, and symbol")
)

func NewPassword(plain string) (Password, error) {
	if err := validatePlain(plain); err != nil {
		return Password{}, err
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	if err != nil {
		return Password{}, fmt.Errorf("failed to hash password: %w", err)
	}
	return Password{hash: string(hashed)}, nil
}

func NewPasswordFromHash(hash string) Password {
	return Password{hash: hash}
}

// NewLoginPassword validates plain against the same rules as NewPassword but
// skips bcrypt hash generation, so the login path pays bcrypt cost only once
// (the comparison, not also a generation).
func NewLoginPassword(plain string) (LoginPassword, error) {
	if err := validatePlain(plain); err != nil {
		return LoginPassword{}, err
	}
	return LoginPassword{plain: plain}, nil
}

// validatePlain checks the format rules shared by NewPassword and NewLoginPassword.
func validatePlain(plain string) error {
	for _, b := range []byte(plain) {
		if b < 0x20 || b > 0x7E {
			return fmt.Errorf("%w", ErrPasswordNotASCII)
		}
	}
	if len(plain) < 8 {
		return fmt.Errorf("%w", ErrPasswordTooShort)
	}
	// bcrypt silently truncates input beyond 72 bytes; we reject to avoid
	// two different passwords producing the same hash.
	if len(plain) > 72 {
		return fmt.Errorf("%w", ErrPasswordTooLong)
	}
	return checkComplexity(plain)
}

// dummyHash is a bcrypt hash of a fixed, unpublished secret. It is never the
// hash of any real user's password and exists only so CompareDummyPassword
// can spend the same bcrypt cost as a real comparison.
const dummyHash = "$2a$10$XaYWruBb.69NKCrUGOuBUeUpT1vrwFB0cgaNV7itdx3hiBkPeaCBa"

// CompareDummyPassword runs a bcrypt comparison against a fixed dummy hash.
// Call it when no matching user was found, so that the response time for an
// unknown email matches the time for a wrong password and cannot be used to
// enumerate registered emails.
func CompareDummyPassword(password LoginPassword) error {
	return compareHash(dummyHash, password.plain)
}

// compareHash runs the bcrypt comparison shared by ComparePassword and
// CompareDummyPassword, so the real and dummy paths always pay the same cost
// and fail the same way.
func compareHash(hash, plain string) error {
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain)); err != nil {
		return fmt.Errorf("%w", ErrPasswordMismatch)
	}
	return nil
}

func (p Password) Hash() string {
	return p.hash
}

// checkComplexity checks character class coverage. Called only after validatePlain
// has validated that p consists solely of ASCII printable characters (0x20-0x7E).
func checkComplexity(p string) error {
	var hasUpper, hasLower, hasDigit, hasSymbol bool
	for _, r := range p {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsDigit(r):
			hasDigit = true
		default:
			hasSymbol = true
		}
	}
	if !hasUpper || !hasLower || !hasDigit || !hasSymbol {
		return fmt.Errorf("%w", ErrPasswordInsufficientComplexity)
	}
	return nil
}
