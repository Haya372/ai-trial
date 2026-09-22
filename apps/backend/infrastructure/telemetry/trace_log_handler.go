package telemetry

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/otel/trace"
)

// TraceLogHandler wraps a slog.Handler and injects trace_id and span_id
// attributes when an active span is present in the context.
// This enables correlation between logs and distributed traces.
type TraceLogHandler struct {
	inner slog.Handler
}

// NewTraceLogHandler returns a slog.Handler that injects trace context
// into every log record that has an active span in its context.
func NewTraceLogHandler(inner slog.Handler) *TraceLogHandler {
	return &TraceLogHandler{inner: inner}
}

func (h *TraceLogHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.inner.Enabled(ctx, level)
}

func (h *TraceLogHandler) Handle(ctx context.Context, r slog.Record) error {
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().IsValid() {
		sc := span.SpanContext()
		r.AddAttrs(
			slog.String("trace_id", sc.TraceID().String()),
			slog.String("span_id", sc.SpanID().String()),
		)
	}
	return h.inner.Handle(ctx, r)
}

func (h *TraceLogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &TraceLogHandler{inner: h.inner.WithAttrs(attrs)}
}

func (h *TraceLogHandler) WithGroup(name string) slog.Handler {
	return &TraceLogHandler{inner: h.inner.WithGroup(name)}
}
