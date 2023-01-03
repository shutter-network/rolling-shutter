package retry

import (
	"fmt"
	"runtime"
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
	return func(r *retrier) {
		r.numRetries = n
		if n == -1 {
			r.infiniteRetries = true
		}
	}
}

func MaxInterval(t time.Duration) Option {
	return func(r *retrier) {
		r.maxInterval = t
	}
}

func Interval(t time.Duration) Option {
	return func(r *retrier) {
		r.interval = t
	}
}

func StopOnErrors(e ...error) Option {
	return func(r *retrier) {
		r.cancelingErrors = e
	}
}

func ExponentialBackoff() Option {
	return func(r *retrier) {
		// for now just use a fixed value
		r.multiplier = 1.5
	}
}

func LogCaptureStackFrameContext() Option {
	pc := make([]uintptr, 10)
	// capture the caller of the option function
	runtime.Callers(2, pc)
	frms := runtime.CallersFrames(pc)
	frm, _ := frms.Next()
	frmCtx := fmt.Sprintf("%s:%d %s", frm.File, frm.Line, frm.Function)
	return func(r *retrier) {
		r.executionContext = frmCtx
	}
}

func LogIdentifier(s string) Option {
	id := uuid.NewString()
	return func(r *retrier) {
		r.identifier = id + ":" + s
	}
}

// UseClock injects a different `clock.Clock`
// implementation than the default `time`
// wrapper. Mainly used for mocking.
func UseClock(c clock.Clock) Option {
	return func(r *retrier) {
		r.clock = c
	}
}
