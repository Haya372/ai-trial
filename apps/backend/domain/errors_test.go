package domain_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/Haya372/ai-trial/backend/domain"
)

func TestNewDomainError_accessors(t *testing.T) {
	err := domain.NewDomainError("SOME_CODE", "something went wrong")

	var de *domain.DomainError
	if !errors.As(err, &de) {
		t.Fatalf("errors.As() failed to extract *domain.DomainError from %v", err)
	}
	if de.Code() != "SOME_CODE" {
		t.Errorf("Code() = %q, want %q", de.Code(), "SOME_CODE")
	}
	if de.Message() != "something went wrong" {
		t.Errorf("Message() = %q, want %q", de.Message(), "something went wrong")
	}
	if err.Error() != "something went wrong" {
		t.Errorf("Error() = %q, want %q", err.Error(), "something went wrong")
	}
}

func TestNewDomainError_identityPreservedThroughWrap(t *testing.T) {
	sentinel := domain.NewDomainError("SENTINEL", "sentinel error")
	wrapped := fmt.Errorf("context: %w", sentinel)

	if !errors.Is(wrapped, sentinel) {
		t.Error("errors.Is() must match the sentinel through wrapping")
	}
}
