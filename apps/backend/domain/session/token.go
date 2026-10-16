package session

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
)

// tokenByteLength is 32 bytes (256 bits) of randomness per ADR-033.
const tokenByteLength = 32

// Token is a plaintext session token (ADR-033). It is sent to the client as a
// cookie value; callers must not persist the plaintext, only its Hash().
type Token struct {
	value string
}

// GenerateToken returns a new random plaintext session token.
func GenerateToken() (Token, error) {
	b := make([]byte, tokenByteLength)
	if _, err := rand.Read(b); err != nil {
		return Token{}, err
	}
	return Token{value: base64.RawURLEncoding.EncodeToString(b)}, nil
}

// NewToken wraps a plaintext token value supplied by a caller (e.g. one read
// from an incoming cookie) so it can be hashed and looked up.
func NewToken(value string) Token {
	return Token{value: value}
}

func (t Token) String() string { return t.value }

// Hash returns the hex-encoded SHA-256 hash of the token for storage and
// lookup (ADR-033).
func (t Token) Hash() TokenHash {
	sum := sha256.Sum256([]byte(t.value))
	return TokenHash{value: hex.EncodeToString(sum[:])}
}

// TokenHash is the stored hash of a plaintext Token (ADR-033).
type TokenHash struct {
	value string
}

// NewTokenHash wraps a hash value read from storage (e.g. a database column).
func NewTokenHash(value string) TokenHash {
	return TokenHash{value: value}
}

func (h TokenHash) String() string { return h.value }
