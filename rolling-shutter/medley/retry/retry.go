package retry

import (
	"context"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/introspection"
)

func multDuration(a time.Duration, m float64) time.Duration {
	return time.Duration(float64(a.Nanoseconds()) * m)
}

type retrier struct {
	clock           clock.Clock
	numRetries      int
	infiniteRetries bool
	interval        time.Duration
	maxInterval     time.Duration
	cancelingErrors []error
	multiplier      float64
	zlogContext     zerolog.Context
}

func newRetrier() *retrier {
	return &retrier{
		clock:           clock.New(),
		numRetries:      3,
		interval:        2 * time.Second,
		maxInterval:     60 * time.Second,
		multiplier:      1.,
		cancelingErrors: []error{},
		zlogContext:     log.With().CallerWithSkipFrameCount(3),
	}
}

func (r *retrier) applyOptions(opts []Option) error {
	for _, opt := range opts {
		err := opt(r)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *retrier) iterator(next <-chan time.Time) <-chan time.Time {
	iter := make(chan time.Time, 1)
	go func() {
		defer close(iter)
		interval := r.interval
		// emit the time once, this is the initial action
		// (e.g. the initial function call)
		iter <- r.clock.Now()
		i := 0
		// next receives information when the last
		// action (e.g. function call) was executed
		for lastExecutionFinished := range next {
			if i >= r.numRetries && !r.infiniteRetries {
				return
			}
			nextTick := r.clock.Until(lastExecutionFinished.Add(interval))
			// emit the time, this is a consecutive retry event
			iter <- <-r.clock.After(nextTick)
			i++
			if float64(interval) >= float64(r.maxInterval)/r.multiplier {
				interval = r.maxInterval
			} else {
				interval = multDuration(interval, r.multiplier)
			}
		}
	}()
	return iter
}

type (
	Option                   func(*retrier) error
	RetriableFunction[T any] func(ctx context.Context) (T, error)
)

// FunctionCall calls the given function multiple times until it doesn't return an error
// or one of any optional, user-defined specific errors is returned.
func FunctionCall[T any](ctx context.Context, fn RetriableFunction[T], opts ...Option) (T, error) {
	var err error
	var null T

	retrier := newRetrier()
	if err := retrier.applyOptions(opts); err != nil {
		return null, errors.Wrap(err, "apply options")
	}
	funcName := introspection.GetFuncName(4)
	retrier.zlogContext = retrier.zlogContext.Str("funcName", funcName)
	logger := retrier.zlogContext.Logger()
	next := make(chan time.Time, 1)
	defer close(next)

	retry := retrier.iterator(next)

	callCount := 0

	for {
		select {
		case _, ok := <-retry:
			if !ok {
				logger.Debug().
					Err(err).
					Int("count", callCount).
					Msg("retry limit reached")
				return null, err
			}
			var result T
			start := retrier.clock.Now()
			result, err = fn(ctx)
			stopped := retrier.clock.Now()
			callCount++
			var logev *zerolog.Event
			if callCount == 1 && err == nil {
				logev = logger.Trace()
			} else {
				logev = logger.Debug()
			}

			logev.
				Err(err).
				TimeDiff("duration", time.Now(), start).
				Int("count", callCount).
				Msg("called retriable function")
			if err == nil {
				return result, nil
			}
			if ctx.Err() != nil {
				return null, ctx.Err()
			}
			for _, cErr := range retrier.cancelingErrors {
				if err == cErr {
					logger.Debug().Err(err).Msg("request errored")
					return null, err
				}
			}
			next <- stopped
		case <-ctx.Done():
			return null, ctx.Err()
		}
	}
}
