//go:build integration

package handler_test

import (
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
)

var testTracerProvider trace.TracerProvider = noop.NewTracerProvider()
