// package repository (not repository_test): startDBSpan/endDBSpan are
// unexported plumbing with no public API surface worth exporting solely
// for tests.
//
//nolint:testpackage
package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

var errBoom = errors.New("boom")

func newTracedRepository(t *testing.T) (*baseRepository, *tracetest.InMemoryExporter) {
	t.Helper()
	exp := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exp))
	t.Cleanup(func() { _ = tp.Shutdown(context.Background()) })

	return &baseRepository{tracer: tp.Tracer("test")}, exp
}

func TestStartDBSpan_recordsNameAndAttributes(t *testing.T) {
	r, exp := newTracedRepository(t)

	_, span := r.startDBSpan(context.Background(), "SELECT", "events")
	span.End()

	spans := exp.GetSpans()
	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}
	if spans[0].Name != "db.query" {
		t.Errorf("expected span name db.query, got %s", spans[0].Name)
	}

	attrs := map[string]string{}
	for _, kv := range spans[0].Attributes {
		attrs[string(kv.Key)] = kv.Value.String()
	}
	if attrs["db.system"] != "postgresql" {
		t.Errorf("expected db.system=postgresql, got %q", attrs["db.system"])
	}
	if attrs["db.operation"] != "SELECT" {
		t.Errorf("expected db.operation=SELECT, got %q", attrs["db.operation"])
	}
	if attrs["db.table"] != "events" {
		t.Errorf("expected db.table=events, got %q", attrs["db.table"])
	}
}

func TestEndDBSpan_setsErrorStatus_whenErrNonNil(t *testing.T) {
	r, exp := newTracedRepository(t)

	_, span := r.startDBSpan(context.Background(), "SELECT", "events")
	endDBSpan(span, errBoom)

	spans := exp.GetSpans()
	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}
	if spans[0].Status.Code != codes.Error {
		t.Errorf("expected status Error, got %v", spans[0].Status.Code)
	}
}

func TestEndDBSpan_leavesStatusUnset_whenErrNil(t *testing.T) {
	r, exp := newTracedRepository(t)

	_, span := r.startDBSpan(context.Background(), "SELECT", "events")
	endDBSpan(span, nil)

	spans := exp.GetSpans()
	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}
	if spans[0].Status.Code == codes.Error {
		t.Error("expected status to not be Error")
	}
}

func TestEndDBSpanNotFound_leavesStatusUnset_forErrNoRows(t *testing.T) {
	r, exp := newTracedRepository(t)

	_, span := r.startDBSpan(context.Background(), "SELECT", "events")
	endDBSpanNotFound(span, pgx.ErrNoRows)

	spans := exp.GetSpans()
	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}
	if spans[0].Status.Code == codes.Error {
		t.Error("pgx.ErrNoRows is an expected outcome and must not mark the span as Error")
	}
}

func TestEndDBSpanNotFound_setsErrorStatus_forOtherErrors(t *testing.T) {
	r, exp := newTracedRepository(t)

	_, span := r.startDBSpan(context.Background(), "SELECT", "events")
	endDBSpanNotFound(span, errBoom)

	spans := exp.GetSpans()
	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}
	if spans[0].Status.Code != codes.Error {
		t.Errorf("expected status Error for a real failure, got %v", spans[0].Status.Code)
	}
}
