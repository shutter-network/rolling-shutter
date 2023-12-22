package chainobserver

import (
	"context"
	"reflect"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	syncdb "github.com/shutter-network/rolling-shutter/rolling-shutter/chainobserver/db/sync"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/eventsyncer"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/service"
)

const finalityOffset = 3

type ChainObserver struct {
	Client     *ethclient.Client
	dbpool     *pgxpool.Pool
	eventTypes map[reflect.Type]*eventsyncer.EventType
}

func New(client *ethclient.Client, dbpool *pgxpool.Pool) *ChainObserver {
	return &ChainObserver{
		Client:     client,
		dbpool:     dbpool,
		eventTypes: map[reflect.Type]*eventsyncer.EventType{},
	}
}

func (c *ChainObserver) AddListenEvent(ev *eventsyncer.EventType) error {
	if _, ok := c.eventTypes[ev.Type]; ok {
		return errors.Errorf("event %s already registered", ev.Type.Name())
	}
	if ev.Handler == nil {
		return errors.Errorf("event %s has no handler function", ev.Type.Name())
	}
	c.eventTypes[ev.Type] = ev
	return nil
}

func (c *ChainObserver) Start(ctx context.Context, runner service.Runner) error {
	db := syncdb.New(c.dbpool)
	eventSyncProgress, err := db.GetEventSyncProgress(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get last synced event from db")
	}
	var fromBlock, fromLogIndex uint64
	eventTypes := []*eventsyncer.EventType{}
	// first find the min of all event's from-blocks
	processedFirstEventType := false
	for _, event := range c.eventTypes {
		eventTypes = append(eventTypes, event)
		if !processedFirstEventType {
			fromBlock = event.FromBlockNumber
			processedFirstEventType = true
			continue
		}
		if event.FromBlockNumber < fromBlock {
			fromBlock = event.FromBlockNumber
		}
	}
	if !processedFirstEventType {
		// this is a programming error - if we didn't register handlers,
		// we shouldn't start the service in the first place
		return errors.New("no event handlers registered")
	}

	// then check if our saved progress is already later
	progressBlock := uint64(eventSyncProgress.NextBlockNumber)
	if progressBlock > fromBlock {
		fromBlock = progressBlock
		// only use the saved log index when we're using the
		// saved block-number
		fromLogIndex = uint64(eventSyncProgress.NextLogIndex)
	}
	log.Info().Uint64("from-block", fromBlock).Uint64("from-log-index", fromLogIndex).
		Msg("starting event syncing")
	syncer := eventsyncer.New(c.Client, finalityOffset, eventTypes, fromBlock, fromLogIndex)
	handleSyncLoop := func(ctx context.Context) error {
		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				eventSyncUpdate, err := syncer.Next(ctx)
				if err != nil {
					return err
				}
				err = c.handleEventSyncUpdate(ctx, eventSyncUpdate)
				if errors.Is(err, ErrNoHandler) {
					return err
				}
				if errors.Is(err, ErrDBUpdateFail) {
					return err
				}
				if err != nil {
					log.Error().Err(err).Msg("error in handler function, skipping event")
				}
			}
		}
	}
	return runner.StartService(syncer, service.ServiceFn{Fn: handleSyncLoop})
}
