package trace

import (
	"context"
	"errors"
	"testing"

	"go.opentelemetry.io/otel/attribute"
	oteltrace "go.opentelemetry.io/otel/trace"
	"gotest.tools/assert"
)

func TestNoopTracing(t *testing.T) {
	// not running any tracing Run routine should only result
	// in noop's and dealing with nilled TraceContext nil-pointer
	ctx := context.Background()
	assert.Assert(t, !IsEnabled())

	// even though tracing is not enabled and set up
	// starting a span should not result in any panic
	// and still retrieve some noop values
	var (
		sctx        context.Context
		span        oteltrace.Span
		reportError ErrorWrapper
	)
	sctx, span, reportError = StartSpan(ctx)
	assert.Assert(t, !span.IsRecording())
	// span is still able to receive function calls,
	// even if these are noops
	span.SetAttributes(attribute.Bool("foo", true))
	err := errors.New("fake error")
	wrappedErr := reportError(err)
	assert.Equal(t, err, wrappedErr)
	assert.NilError(t, sctx.Err())

	span.End()
}

func TestEnabledTracing(t *testing.T) {
	ctx, err := SetupTestTracing(t)
	assert.NilError(t, err)
	assert.Assert(t, IsEnabled())

	var (
		sctx        context.Context
		span        oteltrace.Span
		reportError ErrorWrapper
	)
	sctx, span, reportError = StartSpan(ctx)
	assert.Assert(t, span.IsRecording())
	// NOTE: Ideally we would look into the span,
	//  and assert some things about the attributes etc.
	// But the span struct we have access to here is only
	// a setter and no getter.
	// For reading values we would probably have to hook
	// into the exporters, and that is just not worth it
	// currently.
	span.SetAttributes(attribute.Bool("foo", true))
	err = errors.New("fake error")
	wrappedErr := reportError(err)
	assert.Equal(t, err, wrappedErr)
	assert.NilError(t, sctx.Err())
	span.End()
}
