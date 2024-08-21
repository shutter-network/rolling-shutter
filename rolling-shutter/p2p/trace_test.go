package p2p

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp/cmpopts"
	"gotest.tools/assert"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2pmsg"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/trace"
)

func TestNilTraceContext(t *testing.T) {
	var tc *p2pmsg.TraceContext
	ctx := context.Background()
	// not running any tracing Run routine should only result
	// in noop's and dealing with nilled TraceContext nil-pointer
	assert.Assert(t, !trace.IsEnabled())

	assert.Assert(t, tc == nil)
	InjectTraceContext(ctx, tc)
	assert.Assert(t, tc == nil)

	_, err := ExtractTraceContext(ctx, tc)
	assert.Assert(t, tc == nil)
	assert.ErrorContains(t, err, "carrier is nil")
}

func TestTraceContextIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx, err := trace.SetupTestTracing(t)
	assert.NilError(t, err)

	sctx, span, _ := trace.StartSpan(ctx)
	defer span.End()
	assert.Assert(t, span.IsRecording())
	assert.Assert(t, trace.IsEnabled())

	// get the trace context from the context with the
	// started span and fill the message
	tc := &p2pmsg.TraceContext{}
	InjectTraceContext(sctx, tc)
	// make sure the message is filled with an actual value
	assert.Assert(t, len(tc.GetTraceId()) == 16)
	assert.Assert(t, len(tc.GetSpanId()) == 8)
	assert.Assert(t, len(tc.GetTraceFlags()) == 1)

	// extract the trace context from the message
	// and fill a new context with a new remote
	// span context
	bctx, err := ExtractTraceContext(context.Background(), tc)
	assert.NilError(t, err)

	// now try to get the trace context again,
	// fill a message and compare with the original message
	ntc := &p2pmsg.TraceContext{}
	InjectTraceContext(bctx, ntc)

	assert.DeepEqual(t, tc, ntc, cmpopts.IgnoreUnexported(p2pmsg.TraceContext{}))
}
