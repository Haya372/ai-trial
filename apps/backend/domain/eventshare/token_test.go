package eventshare_test

import (
	"testing"

	"github.com/Haya372/ai-trial/backend/domain/eventshare"
)

func TestGenerateToken_returnsNonEmptyToken(t *testing.T) {
	token, err := eventshare.GenerateToken()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token.String() == "" {
		t.Fatal("GenerateToken() returned an empty token")
	}
}

func TestGenerateToken_returnsURLSafeCharactersOnly(t *testing.T) {
	token, err := eventshare.GenerateToken()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, c := range token.String() {
		isURLSafe := (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-' || c == '_'
		if !isURLSafe {
			t.Fatalf("token contains non URL-safe character: %q in %q", c, token.String())
		}
	}
}

func TestGenerateToken_returnsDifferentTokensEachCall(t *testing.T) {
	token1, err := eventshare.GenerateToken()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	token2, err := eventshare.GenerateToken()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token1.String() == token2.String() {
		t.Fatalf("expected two calls to GenerateToken() to differ, both were %q", token1.String())
	}
}

func TestNewToken_roundTripsTheGivenValue(t *testing.T) {
	token := eventshare.NewToken("plain-token-value")
	if token.String() != "plain-token-value" {
		t.Errorf("String() = %q, want %q", token.String(), "plain-token-value")
	}
}

func TestToken_Hash_isDeterministic(t *testing.T) {
	token := eventshare.NewToken("same-token")
	first := token.Hash()
	second := token.Hash()
	if first.String() != second.String() {
		t.Fatal("Hash() returned different hashes for the same token value")
	}
}

func TestToken_Hash_differsForDifferentTokenValues(t *testing.T) {
	hashA := eventshare.NewToken("token-a").Hash()
	hashB := eventshare.NewToken("token-b").Hash()
	if hashA.String() == hashB.String() {
		t.Fatal("Hash() returned the same hash for different token values")
	}
}

func TestToken_Hash_returnsNonEmptyHash(t *testing.T) {
	hash := eventshare.NewToken("some-token").Hash()
	if hash.String() == "" {
		t.Fatal("Hash() returned an empty TokenHash")
	}
}

func TestNewTokenHash_roundTripsTheGivenValue(t *testing.T) {
	hash := eventshare.NewTokenHash("stored-hash-value")
	if hash.String() != "stored-hash-value" {
		t.Errorf("String() = %q, want %q", hash.String(), "stored-hash-value")
	}
}

func TestTokenHash_IsZero_trueForZeroValue(t *testing.T) {
	var hash eventshare.TokenHash
	if !hash.IsZero() {
		t.Error("IsZero() = false, want true for the zero value")
	}
}

func TestTokenHash_IsZero_falseForNonEmptyValue(t *testing.T) {
	hash := eventshare.NewTokenHash("stored-hash-value")
	if hash.IsZero() {
		t.Error("IsZero() = true, want false for a non-empty hash")
	}
}
