package trace

import (
	"context"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/aggregation"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	tracepb "go.opentelemetry.io/proto/otlp/trace/v1"
)

type (
	NoopTraceClient     struct{}
	NoopMetricsExporter struct{}
)

var (
	_ metric.Exporter  = NoopMetricsExporter{}
	_ otlptrace.Client = NoopTraceClient{}
)

func (NoopTraceClient) Start(context.Context) error {
	return nil
}

func (NoopTraceClient) Stop(context.Context) error {
	return nil
}

func (NoopTraceClient) UploadTraces(context.Context, []*tracepb.ResourceSpans) error {
	return nil
}

func (NoopMetricsExporter) Temporality(metric.InstrumentKind) metricdata.Temporality {
	return metricdata.DeltaTemporality
}

func (NoopMetricsExporter) Aggregation(metric.InstrumentKind) aggregation.Aggregation {
	return aggregation.Drop{}
}

func (NoopMetricsExporter) Export(context.Context, metricdata.ResourceMetrics) error {
	return nil
}

func (NoopMetricsExporter) ForceFlush(context.Context) error {
	return nil
}

func (NoopMetricsExporter) Shutdown(context.Context) error {
	return nil
}
