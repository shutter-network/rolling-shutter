package trace

import (
	"context"
	"errors"
	"testing"
	"time"

	"golang.org/x/sync/errgroup"
	"gotest.tools/assert"
)

func SetupTestTracing(t *testing.T) (context.Context, error) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	eg, ectx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		client := NoopTraceClient{}
		mExporter := NoopMetricsExporter{}
		return Run(ectx, client, mExporter, "test", "0.0.1")
	})

	tick := time.NewTicker(10 * time.Millisecond)
	t.Cleanup(func() {
		tick.Stop()
		cancel()
		err := eg.Wait()
		assert.NilError(t, err)
	})

	// HACK:this is a little bit ugly:
	// we have to poll the global until it is set
	for !IsEnabled() {
		select {
		case <-tick.C:
			continue
		case <-ctx.Done():
			t.Fail()
			return ctx, errors.New("setting up tracing timed out")
		}
	}
	return ctx, nil
}
