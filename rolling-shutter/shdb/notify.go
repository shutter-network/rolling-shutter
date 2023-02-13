package shdb

import (
	"context"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/rs/zerolog/log"
)

// SlurpNotifications waits for notifications from the postgres database and puts them onto the
// given channel in a loop.
func SlurpNotifications(ctx context.Context, conn *pgx.Conn, chann chan<- *pgconn.Notification) {
	for {
		notification, err := conn.WaitForNotification(ctx)
		select {
		case <-ctx.Done():
			return
		default:
			if err != nil {
				log.Error().Err(err).Msg("error waiting for notification")
				continue
			}
			log.Info().Str("channel", notification.Channel).Msg("database notification received")
			chann <- notification
		}
	}
}

func ExecListenChannels(ctx context.Context, conn *pgx.Conn, channels []string) error {
	for _, dbch := range channels {
		_, err := conn.Exec(ctx, "listen "+dbch)
		if err != nil {
			log.Error().
				Err(err).
				Str("channel", dbch).
				Msg("error listening to channel")
			return err
		}
	}
	return nil
}

type (
	SignalLoopFunc func() error
	SignalFunc     func()
	SignalHandler  func(ctx context.Context) error
)

// NewSignal creates a new 'Signal'. This can be used to asynchronously call the given signal
// handler from another go routine. A signal can be delivered while the signal handler is running,
// but we will not run multiple signal handler calls in parallel. Instead the signal handler will
// be called once for all signals delivered, while the handler was active.
func NewSignal(ctx context.Context, signalName string, handler SignalHandler) (SignalFunc, SignalLoopFunc) {
	logger := log.With().Str("signal", signalName).Logger()
	ch := make(chan struct{}, 1)
	loop := func() error {
		for {
			select {
			case <-ch:
				err := handler(ctx)
				if err != nil {
					logger.Info().Err(err).Msg("error calling handler function")
				}
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}

	signal := func() {
		select {
		case ch <- struct{}{}:
		default:
		}
	}

	return signal, loop
}
