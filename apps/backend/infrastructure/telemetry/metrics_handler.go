package telemetry

import (
	"fmt"
	"net/http"
	"strings"

	"go.opentelemetry.io/otel/attribute"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
)

// NewMetricsHandler returns an http.Handler that exposes collected metrics
// in Prometheus text exposition format (version 0.0.4).
//
// This provides equivalent functionality to the official OTel Prometheus
// exporter (go.opentelemetry.io/otel/exporters/prometheus) without that
// external dependency: it calls ManualReader.Collect on every request and
// writes the result in the standard Prometheus text format.
func NewMetricsHandler(reader *sdkmetric.ManualReader) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var rm metricdata.ResourceMetrics
		if err := reader.Collect(r.Context(), &rm); err != nil {
			http.Error(w, "failed to collect metrics: "+err.Error(), http.StatusInternalServerError)
			return
		}

		var b strings.Builder
		for _, sm := range rm.ScopeMetrics {
			for _, m := range sm.Metrics {
				writeMetric(&b, m)
			}
		}

		w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
		if _, err := w.Write([]byte(b.String())); err != nil {
			return
		}
	})
}

func writeMetric(b *strings.Builder, m metricdata.Metrics) {
	if m.Description != "" {
		_, _ = fmt.Fprintf(b, "# HELP %s %s\n", m.Name, m.Description)
	}

	switch d := m.Data.(type) {
	case metricdata.Sum[int64]:
		writeCounter(b, m.Name, d.DataPoints)
	case metricdata.Sum[float64]:
		writeCounterFloat(b, m.Name, d.DataPoints)
	case metricdata.Gauge[int64]:
		writeGauge(b, m.Name, d.DataPoints)
	case metricdata.Gauge[float64]:
		writeGaugeFloat(b, m.Name, d.DataPoints)
	case metricdata.Histogram[float64]:
		writeHistogram(b, m.Name, d.DataPoints)
	}
}

func writeCounter(b *strings.Builder, name string, points []metricdata.DataPoint[int64]) {
	_, _ = fmt.Fprintf(b, "# TYPE %s counter\n", name)
	for _, dp := range points {
		_, _ = fmt.Fprintf(b, "%s%s %d\n", name, labelsStr(dp.Attributes), dp.Value)
	}
}

func writeCounterFloat(b *strings.Builder, name string, points []metricdata.DataPoint[float64]) {
	_, _ = fmt.Fprintf(b, "# TYPE %s counter\n", name)
	for _, dp := range points {
		_, _ = fmt.Fprintf(b, "%s%s %g\n", name, labelsStr(dp.Attributes), dp.Value)
	}
}

func writeGauge(b *strings.Builder, name string, points []metricdata.DataPoint[int64]) {
	_, _ = fmt.Fprintf(b, "# TYPE %s gauge\n", name)
	for _, dp := range points {
		_, _ = fmt.Fprintf(b, "%s%s %d\n", name, labelsStr(dp.Attributes), dp.Value)
	}
}

func writeGaugeFloat(b *strings.Builder, name string, points []metricdata.DataPoint[float64]) {
	_, _ = fmt.Fprintf(b, "# TYPE %s gauge\n", name)
	for _, dp := range points {
		_, _ = fmt.Fprintf(b, "%s%s %g\n", name, labelsStr(dp.Attributes), dp.Value)
	}
}

func writeHistogram(b *strings.Builder, name string, points []metricdata.HistogramDataPoint[float64]) {
	_, _ = fmt.Fprintf(b, "# TYPE %s histogram\n", name)
	for _, dp := range points {
		lbls := labelsStr(dp.Attributes)
		for i, bound := range dp.Bounds {
			le := lblsWithLE(dp.Attributes, fmt.Sprintf("%g", bound))
			_, _ = fmt.Fprintf(b, "%s_bucket%s %d\n", name, le, dp.BucketCounts[i])
		}
		_, _ = fmt.Fprintf(b, "%s_bucket{le=\"+Inf\"%s} %d\n", name, labelsInner(dp.Attributes), dp.Count)
		_, _ = fmt.Fprintf(b, "%s_sum%s %g\n", name, lbls, dp.Sum)
		_, _ = fmt.Fprintf(b, "%s_count%s %d\n", name, lbls, dp.Count)
	}
}

func labelsStr(attrs attribute.Set) string {
	if attrs.Len() == 0 {
		return ""
	}
	return "{" + labelsInner(attrs) + "}"
}

func labelsInner(attrs attribute.Set) string {
	if attrs.Len() == 0 {
		return ""
	}
	var parts []string
	iter := attrs.Iter()
	for iter.Next() {
		kv := iter.Attribute()
		parts = append(parts, fmt.Sprintf("%s=%q", kv.Key, kv.Value.AsString()))
	}
	return strings.Join(parts, ",")
}

func lblsWithLE(attrs attribute.Set, le string) string {
	inner := labelsInner(attrs)
	if inner == "" {
		return fmt.Sprintf("{le=%q}", le)
	}
	return fmt.Sprintf("{le=%q,%s}", le, inner)
}
