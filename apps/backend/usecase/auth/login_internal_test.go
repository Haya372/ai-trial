package auth

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/Haya372/ai-trial/backend/domain/user"
)

// fakeUser implements user.User with a controllable ComparePassword result,
// for white-box testing of authenticate without needing real bcrypt hashes.
type fakeUser struct {
	id         uuid.UUID
	compareErr error
}

func (f *fakeUser) ID() uuid.UUID                              { return f.id }
func (f *fakeUser) Email() user.Email                          { return "" }
func (f *fakeUser) DisplayName() string                        { return "" }
func (f *fakeUser) ComparePassword(_ user.LoginPassword) error { return f.compareErr }

func validLoginPassword(t *testing.T) user.LoginPassword {
	t.Helper()
	p, err := user.NewLoginPassword("SecurePass1!")
	if err != nil {
		t.Fatalf("NewLoginPassword() unexpected error: %v", err)
	}
	return p
}

func TestAuthenticate_found_correctPassword_returnsUser(t *testing.T) {
	u := &fakeUser{id: uuid.New(), compareErr: nil}

	got, err := authenticate(false, u, validLoginPassword(t), nil)
	if err != nil {
		t.Fatalf("authenticate() unexpected error: %v", err)
	}
	if got != u {
		t.Errorf("authenticate() = %v, want %v", got, u)
	}
}

func TestAuthenticate_found_wrongPassword_returnsPasswordMismatch(t *testing.T) {
	u := &fakeUser{id: uuid.New(), compareErr: user.ErrPasswordMismatch}

	got, err := authenticate(false, u, validLoginPassword(t), nil)
	if !errors.Is(err, user.ErrPasswordMismatch) {
		t.Errorf("authenticate() error = %v, want ErrPasswordMismatch", err)
	}
	if got != nil {
		t.Errorf("authenticate() user = %v, want nil on error", got)
	}
}

func TestAuthenticate_notFound_runsCompareAgainstDummyHash_returnsUserNotFound(t *testing.T) {
	// u is nil, mirroring the not-found path in Execute. If authenticate ever
	// touched u before confirming the user was found, this would panic.
	got, err := authenticate(true, nil, validLoginPassword(t), user.ErrUserNotFound)
	if !errors.Is(err, user.ErrUserNotFound) {
		t.Errorf("authenticate() error = %v, want ErrUserNotFound", err)
	}
	if got != nil {
		t.Errorf("authenticate() user = %v, want nil", got)
	}
}
