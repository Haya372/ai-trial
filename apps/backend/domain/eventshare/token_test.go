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
	if token == "" {
		t.Fatal("GenerateToken() returned an empty token")
	}
}

func TestGenerateToken_returnsURLSafeCharactersOnly(t *testing.T) {
	token, err := eventshare.GenerateToken()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, c := range token {
		isURLSafe := (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-' || c == '_'
		if !isURLSafe {
			t.Fatalf("token contains non URL-safe character: %q in %q", c, token)
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
	if token1 == token2 {
		t.Fatalf("expected two calls to GenerateToken() to differ, both were %q", token1)
	}
}

func TestHashToken_isDeterministic(t *testing.T) {
	const token = "same-token"
	first := eventshare.HashToken(token)
	second := eventshare.HashToken(token)
	if first != second {
		t.Fatal("HashToken() returned different hashes for the same input")
	}
}

func TestHashToken_differsForDifferentInput(t *testing.T) {
	if eventshare.HashToken("token-a") == eventshare.HashToken("token-b") {
		t.Fatal("HashToken() returned the same hash for different inputs")
	}
}
