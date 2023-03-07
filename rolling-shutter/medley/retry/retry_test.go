package retry

import (
	"context"
	"errors"
	"math"
	"testing"
	"time"

	"github.com/benbjohnson/clock"
	"gotest.tools/assert"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/comparer"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/testlog"
)

func init() {
	testlog.Setup()
}

var (
	errDefault          = errors.New("retry_test: default error")
	errAtSecondCall     = errors.New("retry_test: error at 2nd call")
	nullTime            time.Time
	baseInterval        = 50 * time.Millisecond
	functionRuntime     = 500 * time.Microsecond
	defaultTestDeadline = 1 * time.Second
)

func TestMultDuration(t *testing.T) {
	assert.Equal(t, multDuration(time.Millisecond, 0), 0*time.Millisecond)
	assert.Equal(t, multDuration(time.Millisecond, 2), 2*time.Millisecond)
	assert.Equal(t, multDuration(time.Millisecond, math.Pow(1.5, 3)), 3375*time.Microsecond)
}

type testFlags struct {
	name          string
	opts          []Option
	expectedErr   error
	expectedTimes []time.Duration
	deadline      time.Duration
}

var testFlagTable = []testFlags{
	{
		"test 5 retries with constant retry time",
		[]Option{
			Interval(baseInterval),
			NumberOfRetries(5),
		},
		errDefault,
		[]time.Duration{
			baseInterval + functionRuntime,
			baseInterval + functionRuntime,
			baseInterval + functionRuntime,
			baseInterval + functionRuntime,
			baseInterval + functionRuntime,
		},
		defaultTestDeadline,
	},
	{
		"test 5 retries with exponential backoff time",
		[]Option{
			Interval(baseInterval),
			NumberOfRetries(5),
			ExponentialBackoff(),
		},
		errDefault,
		[]time.Duration{
			baseInterval + functionRuntime,
			multDuration(baseInterval, 1.5) + functionRuntime,
			multDuration(baseInterval, math.Pow(1.5, 2)) + functionRuntime,
			multDuration(baseInterval, math.Pow(1.5, 3)) + functionRuntime,
			multDuration(baseInterval, math.Pow(1.5, 4)) + functionRuntime,
		},
		defaultTestDeadline,
	},
	{
		"test max interval on exponential backoff",
		[]Option{
			Interval(baseInterval),
			NumberOfRetries(5),
			MaxInterval(
				multDuration(baseInterval, math.Pow(1.5, 2)),
			),
			ExponentialBackoff(),
		},
		errDefault,
		[]time.Duration{
			baseInterval + functionRuntime,
			multDuration(baseInterval, 1.5) + functionRuntime,
			multDuration(baseInterval, math.Pow(1.5, 2)) + functionRuntime,
			// now the max interval kicks in,
			// stopping the exponential backoff from here
			multDuration(baseInterval, math.Pow(1.5, 2)) + functionRuntime,
			multDuration(baseInterval, math.Pow(1.5, 2)) + functionRuntime,
		},
		defaultTestDeadline,
	},
	{
		"test stop on specific error",
		[]Option{
			Interval(baseInterval),
			NumberOfRetries(2),
			StopOnErrors(errAtSecondCall),
		},
		// error should get passed through as well
		errAtSecondCall,
		// only 1 retry
		[]time.Duration{
			baseInterval + functionRuntime,
		},
		defaultTestDeadline,
	},
	{
		"test run infinitely until ctx cancel",
		[]Option{
			Interval(baseInterval),
			// -1 means infinite retries
			NumberOfRetries(-1),
		},
		// we expect the DeadlineExceeded error here
		// instead of the error raised by the function
		context.DeadlineExceeded,
		nil, // don't test the times here
		10 * baseInterval,
	},
	{
		"smoketest logging options",
		// just test that the logging options
		// don't cause runtime errors
		[]Option{
			Interval(baseInterval),
			LogIdentifier("test_retry"),
			NumberOfRetries(2),
		},
		errDefault,
		nil, // don't test the times here
		defaultTestDeadline,
	},
}

func TestRetryFunctionCall(t *testing.T) {
	for _, flags := range testFlagTable {
		t.Run(flags.name, func(t *testing.T) {
			testRetryFunctionCall(t, flags)
		})
	}
}

func testRetryFunctionCall(t *testing.T, flags testFlags) { //nolint:thelper
	var lastCall time.Time
	mockClock := clock.NewMock()
	callCount := 0
	calledRelativeTimes := make([]time.Duration, 0)
	fn := func(ctx context.Context) (struct{}, error) {
		called := mockClock.Now()
		callCount++

		// simulate some work
		mockClock.Sleep(functionRuntime)

		if !lastCall.Equal(nullTime) {
			delta := called.Sub(lastCall)
			calledRelativeTimes = append(calledRelativeTimes, delta)
		}
		lastCall = called

		if callCount == 2 {
			// only call this at the second call,
			// this is useful to test
			// StopOnErrors() option
			return struct{}{}, errAtSecondCall
		}
		return struct{}{}, errDefault
	}
	// This uses the mock clock virtual timeout
	ctx, cancel := mockClock.WithTimeout(context.Background(), flags.deadline)
	defer cancel()

	opts := flags.opts
	opts = append(opts, UseClock(mockClock))

	go func(ctx context.Context) {
		// we need to progress the mock clock in a
		// separate routine
		for {
			select {
			case <-ctx.Done():
				return
			default:
				mockClock.Add(10 * time.Millisecond)
			}
		}
	}(ctx)

	_, err := FunctionCall(ctx, fn, opts...)
	if flags.expectedErr != nil {
		assert.Error(t, err, flags.expectedErr.Error())
	} else {
		assert.NilError(t, err)
	}
	if flags.expectedTimes != nil {
		assert.DeepEqual(
			t,
			flags.expectedTimes,
			calledRelativeTimes,
			comparer.DurationComparerStrict,
		)
	}
}
