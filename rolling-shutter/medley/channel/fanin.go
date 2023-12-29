package channel

import (
	"context"
	"sync"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/service"
)

func NewFanInService[T any](in ...<-chan T) *FanInService[T] {
	ch := make(chan T)
	return &FanInService[T]{
		inBounds: in,
		outBound: ch,
		C:        ch,
	}
}

type FanInService[T any] struct {
	inBounds []<-chan T
	outBound chan T
	C        <-chan T
}

func (p *FanInService[T]) Start(ctx context.Context, runner service.Runner) error {
	wg := new(sync.WaitGroup)

	for _, inBound := range p.inBounds {
		wg.Add(1)
		// avoid capture of iteration variables
		in := inBound
		out := p.outBound
		runner.Go(func() error {
			defer wg.Done()
			return Forward(ctx, in, out)
		})
	}
	runner.Go(func() error {
		wg.Wait()
		close(p.outBound)
		return nil
	})
	return nil
}
