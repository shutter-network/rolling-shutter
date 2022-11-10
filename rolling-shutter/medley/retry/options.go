package retry

import (
	"fmt"
	"runtime"
	"time"

	"github.com/google/uuid"
)

func NumberOfRetries(n int) Option {
	return func(r *retrier) {
		r.numRetries = n
	}
}

func MaxInterval(n int) Option {
	return func(r *retrier) {
		r.numRetries = n
	}
}

func Interval(t time.Duration) Option {
	return func(r *retrier) {
		r.interval = t
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
