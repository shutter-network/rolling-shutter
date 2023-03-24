package trace

import (
	"context"
	"fmt"
	"time"

	"github.com/tendermint/tendermint/libs/sync"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	metricGlb "go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/sdk/metric"
	resourcesdk "go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	oteltrace "go.opentelemetry.io/otel/trace"
	"golang.org/x/sync/errgroup"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/introspection"
)

const name = "p2p"

var enabled sync.AtomicBool

type ErrorWrapper func(error) error

// StartSpan starts a new opentelemtry span and injects it in the returned context.
// The span will have caller-info that is extracted from the stack-trace attached
// to it's attributes and the span name.
// For most observability applications this will be sufficient, but addititional
// attributes can be attached to the span (see the oteltrace.Span interface).
// The function can also be called when tracing is not enabled, since all functions
// make use of a NOOP-object, when the global otel tracer is not set.
//
// StartSpan returns 3 values:
// The returned context `nctx` has the information about the started attached
// to it, so this context has to be used to correctly ensure downstream child-span
// creation and to propagate the trace-context between cross-cutting boundaries.
// The returned `span` has to be closed by calling `span.End()` on it, otherwise
// it can leak memory!
// If additional attributes etc. should be attached to the span, this can still be done
// on the returned span instance as described by the `oteltrace.Span` interface.
// The returned ErrorWrapper `errWrap` can be used to report failure of the span
// by simply using it as a passtrhough functio for the error. This is not required,
// but is helpful for observability.
func StartSpan(ctx context.Context) (nctx context.Context, span oteltrace.Span, errWrap ErrorWrapper) {
	callerInfo := introspection.GetCallerInfo(4)
	nctx, span = otel.Tracer(callerInfo.Library).Start(
		ctx,
		fmt.Sprintf("process %s", callerInfo.Function),
		oteltrace.WithAttributes(
			attribute.String("function", callerInfo.Function),
			attribute.String("package", callerInfo.Package),
			attribute.String("module", callerInfo.Module),
			attribute.String("file", callerInfo.FileLocation),
		))

	errWrap = func(err error) error {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}
	return nctx, span, errWrap
}

// IsEnabled can be used for querying wether tracing is enabled.
// Internally it uses a threadsafe boolean.
//
// In general, we should be careful with this approach - the
// value is expected to not be constantly set during application
// runtime, but should rather only be set/unset in the very outer
// entrypoint layers. This is a quick measure to avoid passing
// a configuration option through the whole callstack
// in all relevant parts of the application.
func IsEnabled() bool {
	return enabled.IsSet()
}

func SetEnabled() {
	enabled.Set()
}

func SetDisabled() {
	enabled.UnSet()
}

func Run(
	ctx context.Context,
	traceClient otlptrace.Client,
	metricExporter metric.Exporter,
	serviceName,
	serviceVersion string,
) error {
	traceExporter, err := otlptrace.New(ctx, traceClient)
	if err != nil {
		return err
	}

	res, err := resourcesdk.Merge(
		resourcesdk.Default(),
		resourcesdk.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(serviceName),
			semconv.ServiceVersion(serviceVersion),
		),
	)
	if err != nil {
		return err
	}
	tp := tracesdk.NewTracerProvider(
		tracesdk.WithBatcher(traceExporter),
		tracesdk.WithResource(res),
	)

	meterProvider := metric.NewMeterProvider(
		metric.WithReader(metric.NewPeriodicReader(metricExporter)),
		metric.WithResource(res),
	)

	// this will set the global tracer provider.
	// when this is not set, the global getters
	// will return a noop-trace provider, that
	// only spawns noop traces.
	otel.SetTracerProvider(tp)
	metricGlb.SetMeterProvider(meterProvider)

	SetEnabled()
	defer SetDisabled()

	<-ctx.Done()
	// cleanup routines

	errorgroup, _ := errgroup.WithContext(context.Background())
	errorgroup.Go(func() error {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		return tp.Shutdown(shutdownCtx)
	})
	errorgroup.Go(func() error {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		return meterProvider.Shutdown(shutdownCtx)
	})
	return errorgroup.Wait()
}
