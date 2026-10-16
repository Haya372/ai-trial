package session_test

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"testing"

	"github.com/Haya372/ai-trial/backend/domain/session"
)

func TestGenerateToken_returnsURLSafeToken(t *testing.T) {
	token, err := session.GenerateToken()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token.String() == "" {
		t.Fatal("GenerateToken() returned an empty token")
	}
	for _, c := range token.String() {
		isURLSafe := (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-' || c == '_'
		if !isURLSafe {
			t.Fatalf("token contains non URL-safe character: %q in %q", c, token.String())
		}
	}
}

func TestGenerateToken_carries256BitsOfEntropy(t *testing.T) {
	token, err := session.GenerateToken()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	raw, err := base64.RawURLEncoding.DecodeString(token.String())
	if err != nil {
		t.Fatalf("token is not unpadded base64url: %v", err)
	}
	if len(raw) != 32 {
		t.Errorf("decoded token length = %d bytes, want 32", len(raw))
	}
}

func TestGenerateToken_returnsDifferentTokensEachCall(t *testing.T) {
	first, err := session.GenerateToken()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	second, err := session.GenerateToken()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if first.String() == second.String() {
		t.Fatalf("expected two calls to differ, both were %q", first.String())
	}
}

func TestNewToken_roundTripsTheGivenValue(t *testing.T) {
	token := session.NewToken("plain-token-value")
	if token.String() != "plain-token-value" {
		t.Errorf("String() = %q, want %q", token.String(), "plain-token-value")
	}
}

func TestToken_Hash_isHexEncodedSHA256(t *testing.T) {
	sum := sha256.Sum256([]byte("some-token"))
	want := hex.EncodeToString(sum[:])

	got := session.NewToken("some-token").Hash()

	if got.String() != want {
		t.Errorf("Hash() = %q, want %q", got.String(), want)
	}
}

func TestToken_Hash_differsForDifferentTokens(t *testing.T) {
	hashA := session.NewToken("token-a").Hash()
	hashB := session.NewToken("token-b").Hash()
	if hashA.String() == hashB.String() {
		t.Fatal("Hash() returned the same hash for different tokens")
	}
}

func TestToken_Hash_doesNotExposePlaintext(t *testing.T) {
	token := session.NewToken("plaintext-secret")
	if token.Hash().String() == token.String() {
		t.Fatal("Hash() must not equal the plaintext token")
	}
}

func TestNewTokenHash_roundTripsTheGivenValue(t *testing.T) {
	hash := session.NewTokenHash("stored-hash-value")
	if hash.String() != "stored-hash-value" {
		t.Errorf("String() = %q, want %q", hash.String(), "stored-hash-value")
	}
}
