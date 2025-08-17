package shutterservice

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyperimpl/shutterservice/database"
)

// TriggerProcessor implements the EventProcessor interface for processing trigger events.
type TriggerProcessor struct {
	ExecutionClient *ethclient.Client
	DBPool          *pgxpool.Pool
}

type TriggerEvent struct {
	EventTriggerRegisteredEvent database.EventTriggerRegisteredEvent
	Log                         types.Log
}

func NewTriggerProcessor(
	executionClient *ethclient.Client,
	dbPool *pgxpool.Pool,
) *TriggerProcessor {
	return &TriggerProcessor{
		ExecutionClient: executionClient,
		DBPool:          dbPool,
	}
}

func (tp *TriggerProcessor) GetProcessorName() string {
	return "trigger"
}

func (tp *TriggerProcessor) FetchEvents(ctx context.Context, start, end uint64) ([]Event, error) {
	queries := database.New(tp.DBPool)
	// Consider event triggers that have not fired yet and have not expired at the start block.
	// They might have expired at the end block though which will be checked later.
	triggerRegisteredEvents, err := queries.GetActiveEventTriggerRegisteredEvents(ctx, int64(start))
	if err != nil {
		return nil, errors.Wrap(err, "failed to get event trigger registered events")
	}

	var events []Event
	for _, triggerRegisteredEvent := range triggerRegisteredEvents {
		trigger := EventTriggerDefinition{}
		err := trigger.UnmarshalBytes(triggerRegisteredEvent.Definition)
		if err != nil {
			log.Info().Err(err).Int64("block-number", triggerRegisteredEvent.BlockNumber).
				Hex("block-hash", triggerRegisteredEvent.BlockHash).
				Int64("tx-index", triggerRegisteredEvent.TxIndex).
				Int64("log-index", triggerRegisteredEvent.LogIndex).
				Msg("encountered invalid trigger definition, skipping")
			continue
		}

		filterQuery := trigger.ToFilterQuery()
		filterQuery.FromBlock = new(big.Int).SetUint64(start)
		filterQuery.ToBlock = new(big.Int).SetUint64(end)

		logs, err := tp.ExecutionClient.FilterLogs(ctx, filterQuery)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to filter logs for event trigger")
		}

		for _, eventLog := range logs {
			// Check that the trigger has not expired at the time of the event.
			if eventLog.BlockNumber > uint64(triggerRegisteredEvent.BlockNumber+triggerRegisteredEvent.Ttl) {
				continue
			}
			if !trigger.Match(eventLog, false) {
				continue
			}
			events = append(events, &TriggerEvent{
				Log:                         eventLog,
				EventTriggerRegisteredEvent: triggerRegisteredEvent,
			})
		}
	}

	return events, nil
}

func (tp *TriggerProcessor) ProcessEvents(ctx context.Context, tx pgx.Tx, events []Event) error {
	queries := database.New(tx)
	for _, untypedEvent := range events {
		event := untypedEvent.(*TriggerEvent)
		err := queries.InsertFiredTrigger(ctx, database.InsertFiredTriggerParams{
			IdentityPrefix: event.EventTriggerRegisteredEvent.IdentityPrefix,
			Sender:         event.EventTriggerRegisteredEvent.Sender,
			BlockNumber:    int64(event.Log.BlockNumber),
			BlockHash:      event.Log.BlockHash[:],
			TxIndex:        int64(event.Log.TxIndex),
			LogIndex:       int64(event.Log.Index),
		})
		if err != nil {
			return fmt.Errorf("failed to insert fired trigger: %w", err)
		}
	}

	return nil
}

func (tp *TriggerProcessor) RollbackEvents(ctx context.Context, tx pgx.Tx, toBlock int64) error {
	queries := database.New(tx)
	err := queries.DeleteEventTriggerRegisteredEventsFromBlockNumber(ctx, toBlock+1)
	if err != nil {
		return fmt.Errorf("failed to delete event trigger registered events from block number: %w", err)
	}
	return nil
}
