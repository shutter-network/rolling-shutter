package channel

import (
	"context"
	"sync"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/service"
)

func NewFanOutService[T any](in <-chan T, out ...chan T) *FanOutService[T] {
	return &FanOutService[T]{
		inBound:   in,
		outBounds: out,
	}
}

// FanOutService takes input from one channel and will distribute
// it across multiple downstream out channels.
// This will fan-out the values, but incoming values will only
// be forwarded to ONE outbound channel.
// So this is more reminiscent of a load balancer,
// NOT a subscription service.
type FanOutService[T any] struct {
	inBound   <-chan T
	outBounds []chan T
}

func (p *FanOutService[T]) Start(ctx context.Context, runner service.Runner) error {
	wg := new(sync.WaitGroup)
	for _, outBound := range p.outBounds {
		wg.Add(1)
		runner.Go(func() error {
			defer wg.Done()
			in := p.inBound
			out := outBound
			return Forward(ctx, in, out)
		})
	}
	runner.Go(func() error {
		wg.Wait()
		for _, outBound := range p.outBounds {
			close(outBound)
		}
		return nil
	})
	return nil
}
