package telemetry

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	sdkresource "go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

// Config carries the telemetry initialisation options.
type Config struct {
	// ServiceName labels all spans and metrics emitted by this process.
	ServiceName string
	// OTLPEndpoint is the HTTP target for the OTLP trace exporter, following
	// the OTEL_EXPORTER_OTLP_ENDPOINT convention: either a full URL
	// (scheme+host+port, e.g. "http://otel-collector:4318") or a bare
	// "host:port". When empty a human-readable stdout exporter is used
	// instead, which is suitable for local development.
	OTLPEndpoint string
}

// Providers bundles the initialised OpenTelemetry providers and the
// ManualReader that backs the /metrics HTTP handler.
type Providers struct {
	Tracer trace.TracerProvider
	Meter  metric.MeterProvider
	// Reader is the ManualReader shared with NewMetricsHandler so the
	// /metrics endpoint can pull the latest metric values on demand.
	Reader   *sdkmetric.ManualReader
	Shutdown func(context.Context) error
}

// Init initialises the OpenTelemetry TracerProvider and MeterProvider
// for the given service configuration.
//
// Call Providers.Shutdown on application exit to flush pending telemetry.
func Init(ctx context.Context, cfg Config) (Providers, error) {
	res, err := sdkresource.New(ctx,
		sdkresource.WithAttributes(
			semconv.ServiceName(cfg.ServiceName),
		),
	)
	if err != nil {
		return Providers{}, fmt.Errorf("create resource: %w", err)
	}

	tp, err := newTracerProvider(ctx, cfg, res)
	if err != nil {
		return Providers{}, fmt.Errorf("create tracer provider: %w", err)
	}

	reader := sdkmetric.NewManualReader()
	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(reader),
		sdkmetric.WithResource(res),
	)

	shutdown := func(ctx context.Context) error {
		if err := tp.Shutdown(ctx); err != nil {
			return fmt.Errorf("shutdown tracer provider: %w", err)
		}
		if err := mp.Shutdown(ctx); err != nil {
			return fmt.Errorf("shutdown meter provider: %w", err)
		}
		return nil
	}

	return Providers{
		Tracer:   tp,
		Meter:    mp,
		Reader:   reader,
		Shutdown: shutdown,
	}, nil
}

func newTracerProvider(ctx context.Context, cfg Config, res *sdkresource.Resource) (*sdktrace.TracerProvider, error) {
	exp, err := NewSpanExporter(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("create span exporter: %w", err)
	}

	return sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(res),
	), nil
}

// NewSpanExporter returns the trace exporter for the given configuration.
// When cfg.OTLPEndpoint is empty, a human-readable stdout exporter is used,
// which is suitable for local development. Otherwise spans are shipped to
// an OpenTelemetry Collector via OTLP/HTTP.
func NewSpanExporter(ctx context.Context, cfg Config) (sdktrace.SpanExporter, error) {
	if cfg.OTLPEndpoint == "" {
		return stdouttrace.New(stdouttrace.WithPrettyPrint())
	}
	// otlptracehttp.WithEndpoint expects a bare "host:port" and rejects a
	// scheme; WithEndpointURL expects a full URL but, unlike WithEndpoint,
	// does not default the path to /v1/traces. Route each form to the
	// matching option so both the bare and the OTEL_EXPORTER_OTLP_ENDPOINT
	// convention's full-URL form work.
	if strings.Contains(cfg.OTLPEndpoint, "://") {
		return otlptracehttp.New(ctx,
			otlptracehttp.WithEndpointURL(normalizeOTLPEndpointURL(cfg.OTLPEndpoint)),
			otlptracehttp.WithInsecure(),
		)
	}
	return otlptracehttp.New(ctx,
		otlptracehttp.WithEndpoint(cfg.OTLPEndpoint),
		otlptracehttp.WithInsecure(),
	)
}

// normalizeOTLPEndpointURL appends the default traces path (/v1/traces) to
// a scheme-qualified endpoint URL that has no path of its own, matching the
// behavior operators expect from the OTEL_EXPORTER_OTLP_ENDPOINT convention.
// An explicit path, or a URL that fails to parse, is returned unchanged.
func normalizeOTLPEndpointURL(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil || (u.Path != "" && u.Path != "/") {
		return rawURL
	}
	u.Path = "/v1/traces"
	return u.String()
}
