package retry

import (
	"errors"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/google/uuid"
)

// NumberOfRetries specifies the
// number of retries are conducted
// until the call finally returns.
// `-1` is a special value that results in
// infinite retries.
func NumberOfRetries(n int) Option {
	return func(r *retrier) error {
		r.numRetries = n
		if n == -1 {
			r.infiniteRetries = true
		}
		return nil
	}
}

func MaxInterval(t time.Duration) Option {
	return func(r *retrier) error {
		r.maxInterval = t
		return nil
	}
}

func Interval(t time.Duration) Option {
	return func(r *retrier) error {
		r.interval = t
		return nil
	}
}

func StopOnErrors(e ...error) Option {
	return func(r *retrier) error {
		r.cancelingErrors = e
		return nil
	}
}

func ExponentialBackoff(multiplier *float64) Option {
	return func(r *retrier) error {
		if multiplier == nil {
			r.multiplier = 1.5
			return nil
		}
		r.multiplier = *multiplier
		if r.multiplier <= 1.0 {
			return errors.New("can't use value <=1.0 as exponential multiplier")
		}
		return nil
	}
}

func LogIdentifier(s string) Option {
	id := uuid.NewString()
	return func(r *retrier) error {
		r.zlogContext = r.zlogContext.Str("id", id+":"+s)
		return nil
	}
}

// UseClock injects a different `clock.Clock`
// implementation than the default `time`
// wrapper. Mainly used for mocking.
func UseClock(c clock.Clock) Option {
	return func(r *retrier) error {
		r.clock = c
		return nil
	}
}
