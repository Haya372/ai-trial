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

	"github.com/google/uuid"
)

func TestToPgUUID_wrapsIDAsValid(t *testing.T) {
	id := uuid.New()

	got := toPgUUID(id)

	if !got.Valid {
		t.Fatal("expected Valid=true")
	}
	if got.Bytes != id {
		t.Errorf("expected Bytes=%v, got %v", id, got.Bytes)
	}
}

func TestCheckNotFound_reportsNotNotFound_whenErrNil(t *testing.T) {
	notFound, wasNoRows, err := checkNotFound(nil, errBoom)

	if notFound {
		t.Error("expected notFound=false for a nil error")
	}
	if wasNoRows {
		t.Error("expected wasNoRows=false for a nil error")
	}
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}

func TestCheckNotFound_mapsToNotFoundErr_whenErrNoRows(t *testing.T) {
	notFound, wasNoRows, err := checkNotFound(pgx.ErrNoRows, errBoom)

	if !notFound {
		t.Error("expected notFound=true for pgx.ErrNoRows")
	}
	if !wasNoRows {
		t.Error("expected wasNoRows=true for pgx.ErrNoRows")
	}
	if !errors.Is(err, errBoom) {
		t.Errorf("expected mapped error %v, got %v", errBoom, err)
	}
}

func TestCheckNotFound_allowsNilNotFoundErr(t *testing.T) {
	notFound, wasNoRows, err := checkNotFound(pgx.ErrNoRows, nil)

	if !notFound {
		t.Error("expected notFound=true for pgx.ErrNoRows")
	}
	if !wasNoRows {
		t.Error("expected wasNoRows=true for pgx.ErrNoRows")
	}
	if err != nil {
		t.Errorf("expected nil mapped error, got %v", err)
	}
}

func TestCheckNotFound_passesThroughOtherErrors_unchanged(t *testing.T) {
	notFound, wasNoRows, err := checkNotFound(errBoom, errNotFoundTester)

	if !notFound {
		t.Error("expected notFound=true for a non-NoRows error")
	}
	if wasNoRows {
		t.Error("expected wasNoRows=false for a non-NoRows error")
	}
	if !errors.Is(err, errBoom) {
		t.Errorf("expected the original error %v unchanged, got %v", errBoom, err)
	}
}

var (
	errBoom           = errors.New("boom")
	errNotFoundTester = errors.New("not found sentinel")
)

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
