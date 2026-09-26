package telemetry_test

import (
	"context"
	"testing"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"

	"github.com/Haya372/ai-trial/backend/infrastructure/telemetry"
)

const testServiceName = "test-service"

func TestInitTelemetry_returnsProviders(t *testing.T) {
	cfg := telemetry.Config{ServiceName: testServiceName}
	p, err := telemetry.Init(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	defer p.Shutdown(context.Background())

	if p.Tracer == nil {
		t.Error("TracerProvider is nil")
	}
	if p.Meter == nil {
		t.Error("MeterProvider is nil")
	}
	if p.Reader == nil {
		t.Error("ManualReader is nil")
	}
}

func TestInitTelemetry_providersAreUsable(t *testing.T) {
	cfg := telemetry.Config{ServiceName: testServiceName}
	p, err := telemetry.Init(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	defer p.Shutdown(context.Background())

	ctx, span := p.Tracer.Tracer("test").Start(context.Background(), "test.op")
	if !span.SpanContext().IsValid() {
		t.Error("span context should be valid")
	}
	span.End()
	_ = ctx

	counter, err := p.Meter.Meter("test").Int64Counter("test.counter")
	if err != nil {
		t.Fatalf("failed to create counter: %v", err)
	}
	counter.Add(context.Background(), 1)
}

func TestNewSpanExporter_usesStdout_whenOTLPEndpointEmpty(t *testing.T) {
	exp, err := telemetry.NewSpanExporter(context.Background(), telemetry.Config{})
	if err != nil {
		t.Fatalf("NewSpanExporter failed: %v", err)
	}
	defer exp.Shutdown(context.Background())

	if _, ok := exp.(*stdouttrace.Exporter); !ok {
		t.Errorf("expected *stdouttrace.Exporter, got %T", exp)
	}
}

func TestNewSpanExporter_usesOTLP_whenOTLPEndpointSet(t *testing.T) {
	exp, err := telemetry.NewSpanExporter(context.Background(), telemetry.Config{OTLPEndpoint: "localhost:4318"})
	if err != nil {
		t.Fatalf("NewSpanExporter failed: %v", err)
	}
	defer exp.Shutdown(context.Background())

	if _, ok := exp.(*otlptrace.Exporter); !ok {
		t.Errorf("expected *otlptrace.Exporter, got %T", exp)
	}
}

func TestNewSpanExporter_usesOTLP_whenOTLPEndpointIsFullURL(t *testing.T) {
	exp, err := telemetry.NewSpanExporter(context.Background(), telemetry.Config{OTLPEndpoint: "http://localhost:4318"})
	if err != nil {
		t.Fatalf("NewSpanExporter failed: %v", err)
	}
	defer exp.Shutdown(context.Background())

	if _, ok := exp.(*otlptrace.Exporter); !ok {
		t.Errorf("expected *otlptrace.Exporter, got %T", exp)
	}
}
