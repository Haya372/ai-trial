package eventshare

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
)

// tokenByteLength is 32 bytes (256 bits) of randomness per ADR-028.
const tokenByteLength = 32

// GenerateToken returns a new plaintext share token. Callers must not persist
// the plaintext; only HashToken's output is stored (ADR-028).
func GenerateToken() (string, error) {
	b := make([]byte, tokenByteLength)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// HashToken returns the SHA-256 hash of a plaintext token, hex-encoded, for
// storage and lookup (ADR-028).
func HashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
