package channel_test

import (
	"context"
	"sort"
	"testing"
	"time"

	"gotest.tools/assert"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/channel"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/service"
)

func putInts(ctx context.Context, ch chan<- int, start, end, step int) {
	for i := start; i < end; i += step {
		select {
		case ch <- i:
			continue
		case <-ctx.Done():
			return
		}
	}
	close(ch)
}

func TestFanIn(t *testing.T) {
	channels := []<-chan int{}
	services := []service.Service{}
	numFns := 3
	end := 100

	for i := 0; i < numFns; i++ {
		ch := make(chan int)
		start := i
		end := end
		step := numFns
		channels = append(channels, ch)
		fnService := service.Function{
			Func: func(ctx context.Context, _ service.Runner) error {
				putInts(ctx, ch, start, end, step)
				return nil
			},
		}
		services = append(services, fnService)
	}
	fanIn := channel.NewFanInService(channels...)
	services = append(services, fanIn)

	result := []int{}
	fanConsumer := service.Function{
		Func: func(ctx context.Context, _ service.Runner) error {
			for val := range fanIn.C {
				result = append(result, val)
			}
			return nil
		},
	}
	services = append(services, fanConsumer)

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	t.Cleanup(cancel)
	// run and wait.
	// this will unblock when the work is done,
	// since the consumer channel is assumed to be closed
	// when all inBound channels are closed
	err := service.Run(ctx, services...)
	assert.NilError(t, err)

	// sort array
	sort.Ints(result)
	expected := []int{}
	for i := 0; i < end; i++ {
		expected = append(expected, i)
	}
	assert.DeepEqual(t, result, expected)
}
