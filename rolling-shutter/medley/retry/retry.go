package retry

import (
	"context"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type retrier struct {
	numRetries       int
	interval         time.Duration
	maxInterval      time.Duration
	multiplier       float64
	executionContext string
	identifier       string
}

func defaultOptions() Option {
	return func(r *retrier) {
		r.numRetries = 3
		r.interval = 2 * time.Second
		r.maxInterval = 60 * time.Second
		r.multiplier = 1.
	}
}

func (r *retrier) option(opts []Option) {
	for _, opt := range opts {
		opt(r)
	}
}

func (r *retrier) logWithContext(e *zerolog.Event) *zerolog.Event {
	if r.identifier != "" {
		e = e.Str("id", r.identifier)
	}

	if r.identifier != "" {
		e = e.Str("context", r.executionContext)
	}
	return e
}

func (r *retrier) logError(err error, msg string) {
	e := log.Error().Err(err)
	r.logWithContext(e).Msg(msg)
}

func (r *retrier) iterator(next <-chan time.Time) <-chan time.Time {
	rc := make(chan time.Time, 1)
	go func() {
		i := 0
		interval := r.interval
		// emit the time once, this is the initial event
		// (e.g. the initial function call)
		rc <- time.Now()
		for range next {
			if i >= r.numRetries {
				return
			}
			// emit the time, this is a consecutive retry event
			rc <- <-time.After(interval)
			i++
			if float64(interval) >= float64(r.maxInterval)/r.multiplier {
				interval = r.maxInterval
			} else {
				interval = time.Duration(float64(interval) * r.multiplier)
			}
		}
	}()
	return rc
}

type (
	Option                   func(*retrier)
	RetriableFunction[T any] func(ctx context.Context) (T, error)
)

// FunctionCall calls the given function multiple times until it doesn't return an error.
func FunctionCall[T any](ctx context.Context, fn RetriableFunction[T], opts ...Option) (T, error) {
	retrier := &retrier{}
	opts = append([]Option{defaultOptions()}, opts...)
	retrier.option(opts)

	next := make(chan time.Time, 1)
	defer close(next)

	retry := retrier.iterator(next)

	var err error
	var null T

	for {
		select {
		case _, ok := <-retry:
			if !ok {
				retrier.logError(err, "request errored, retry limit reached")
				return null, err
			}
			var result T
			start := time.Now()
			result, err = fn(ctx)
			retrier.logWithContext(
				log.Debug().TimeDiff("took", time.Now(), start),
			).Msg("called retriable function")
			if err == nil {
				return result, nil
			}
			// don't retry when the context was canceled during function execution
			if ctx.Err() != nil {
				return null, ctx.Err()
			}
			retrier.logError(err, "request errored, retrying")
			next <- time.Now()
		case <-ctx.Done():
			return null, ctx.Err()
		}
	}
}
