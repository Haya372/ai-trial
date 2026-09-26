package telemetry_test

import (
	"context"
	"log/slog"
	"testing"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"

	"github.com/Haya372/ai-trial/backend/infrastructure/telemetry"
)

// captureHandler records slog.Record entries for inspection.
type captureHandler struct {
	records []slog.Record
}

func (h *captureHandler) Enabled(_ context.Context, _ slog.Level) bool { return true }
func (h *captureHandler) Handle(_ context.Context, r slog.Record) error {
	h.records = append(h.records, r)
	return nil
}
func (h *captureHandler) WithAttrs(_ []slog.Attr) slog.Handler { return h }
func (h *captureHandler) WithGroup(_ string) slog.Handler      { return h }

func recordAttrs(r slog.Record) map[string]string {
	m := make(map[string]string)
	r.Attrs(func(a slog.Attr) bool {
		m[a.Key] = a.Value.String()
		return true
	})
	return m
}

func TestTraceLogHandler_addsTraceAndSpanID_whenSpanActive(t *testing.T) {
	inner := &captureHandler{}
	handler := telemetry.NewTraceLogHandler(inner)
	logger := slog.New(handler)

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSyncer(tracetest.NewNoopExporter()),
	)
	tracer := tp.Tracer("test")
	ctx, span := tracer.Start(context.Background(), "test.op")
	defer span.End()

	logger.InfoContext(ctx, "test message")

	if len(inner.records) != 1 {
		t.Fatalf("expected 1 log record, got %d", len(inner.records))
	}
	attrs := recordAttrs(inner.records[0])

	traceID := span.SpanContext().TraceID().String()
	spanID := span.SpanContext().SpanID().String()

	if got := attrs["trace_id"]; got != traceID {
		t.Errorf("trace_id: expected %q, got %q", traceID, got)
	}
	if got := attrs["span_id"]; got != spanID {
		t.Errorf("span_id: expected %q, got %q", spanID, got)
	}
}

func TestTraceLogHandler_noExtraFields_whenNoActiveSpan(t *testing.T) {
	inner := &captureHandler{}
	handler := telemetry.NewTraceLogHandler(inner)
	logger := slog.New(handler)

	logger.InfoContext(context.Background(), "test message")

	if len(inner.records) != 1 {
		t.Fatalf("expected 1 log record, got %d", len(inner.records))
	}
	attrs := recordAttrs(inner.records[0])
	if _, ok := attrs["trace_id"]; ok {
		t.Error("trace_id should not be present when no span is active")
	}
	if _, ok := attrs["span_id"]; ok {
		t.Error("span_id should not be present when no span is active")
	}
}

func TestTraceLogHandler_delegatesEnabled(t *testing.T) {
	inner := &captureHandler{}
	handler := telemetry.NewTraceLogHandler(inner)

	if !handler.Enabled(context.Background(), slog.LevelInfo) {
		t.Error("Enabled should delegate to inner handler")
	}
}
