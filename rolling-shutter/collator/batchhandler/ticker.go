package batchhandler

import (
	"time"

	"github.com/rs/zerolog/log"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley"
)

type EpochTicker struct {
	ticker *time.Ticker
	broker *medley.Broker[time.Time]
	stop   chan struct{}
}

func StartEpochTicker(epochDuration time.Duration) *EpochTicker {
	return &EpochTicker{
		ticker: time.NewTicker(epochDuration),
		broker: medley.NewBroker[time.Time](true),
		stop:   make(chan struct{}),
	}
}

func (tick *EpochTicker) Unsubscribe(channel chan time.Time) {
	tick.broker.Unsubscribe(channel)
}

func (tick *EpochTicker) Subscribe() chan time.Time {
	channel := make(chan time.Time)
	tick.broker.Subscribe(channel)
	return channel
}

func (tick *EpochTicker) Run() {
	// FIXME there seems to be one problem:
	// since consumers subscribe to the ticker
	// with a non-blocking send,
	// it is possible to skip a tick.
	// unlike in the actual ticker,
	// the consumer now has to wait until the
	// next tick is reached.
	// In the ticker, a slow consumer will
	// receive immediately when the time was reached
	go tick.broker.Start()
	for {
		select {
		case val := <-tick.ticker.C:
			log.Debug().Str("time", val.String()).Msg("tick broadcast")
			tick.broker.Publish(val)
		case <-tick.stop:
			log.Debug().Msg("ticker received stop signal")
			tick.broker.Stop()
			return
		}
	}
}

func (tick *EpochTicker) Stop() {
	close(tick.stop)
}
