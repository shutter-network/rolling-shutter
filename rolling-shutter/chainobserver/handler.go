package chainobserver

import (
	"context"
	"reflect"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

var ErrNoHandler = errors.New("EventType has no Handler")

func (c *ChainObserver) handleEvent(
	ctx context.Context, tx pgx.Tx, event any,
) error {
	eventType := reflect.TypeOf(event)
	ev, ok := c.eventTypes[eventType]
	if !ok {
		log.Info().Str("event-type", reflect.TypeOf(event).String()).Interface("event", event).
			Msg("ignoring unknown event")
		return nil
	}
	if ev.Handler == nil {
		return errors.Wrapf(ErrNoHandler, "`%s`: ", ev.Name)
	}

	err := ev.Handler(ctx, tx, event)
	if err != nil {
		return errors.Wrapf(err, "error during `%s` handler invocation", ev.Name)
	}
	return nil
}
