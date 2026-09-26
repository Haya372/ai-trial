// package telemetry (not telemetry_test): normalizeOTLPEndpointURL is
// unexported plumbing with no public API surface worth exporting solely
// for tests.
//
//nolint:testpackage
package telemetry

import "testing"

func TestNormalizeOTLPEndpointURL_appendsDefaultTracesPath_whenPathMissing(t *testing.T) {
	got := normalizeOTLPEndpointURL("http://otel-collector:4318")
	want := "http://otel-collector:4318/v1/traces"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestNormalizeOTLPEndpointURL_keepsExplicitPath(t *testing.T) {
	got := normalizeOTLPEndpointURL("http://otel-collector:4318/custom/path")
	want := "http://otel-collector:4318/custom/path"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestNormalizeOTLPEndpointURL_returnsInputUnchanged_whenUnparsable(t *testing.T) {
	got := normalizeOTLPEndpointURL("://not a url")
	want := "://not a url"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
