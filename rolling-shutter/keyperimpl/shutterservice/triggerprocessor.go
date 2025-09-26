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
		triggerLog := log.With().
			Int64("block-number", triggerRegisteredEvent.BlockNumber).
			Hex("block-hash", triggerRegisteredEvent.BlockHash).
			Int64("tx-index", triggerRegisteredEvent.TxIndex).
			Int64("log-index", triggerRegisteredEvent.LogIndex).
			Hex("identity-prefix", triggerRegisteredEvent.IdentityPrefix).
			Str("sender", triggerRegisteredEvent.Sender).
			Hex("definition", triggerRegisteredEvent.Definition).
			Int64("expiration-block-number", triggerRegisteredEvent.ExpirationBlockNumber).
			Logger()

		trigger := EventTriggerDefinition{}
		err := trigger.UnmarshalBytes(triggerRegisteredEvent.Definition)
		if err != nil {
			// This is not supposed to happen as only valid triggers are inserted into the database.
			triggerLog.Error().Err(err).Msg("ignoring invalid trigger definition in database")
			continue
		}

		filterQuery, err := trigger.ToFilterQuery()
		if err != nil {
			// This is not supposed to happen as only valid triggers are inserted into the database
			// and valid triggers should always have a valid filter query.
			triggerLog.Error().Err(err).Msg("failed to create filter query for trigger")
			continue
		}
		filterQuery.FromBlock = new(big.Int).SetUint64(start)
		filterQuery.ToBlock = new(big.Int).SetUint64(end)

		logs, err := tp.ExecutionClient.FilterLogs(ctx, filterQuery)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to filter logs for event trigger")
		}

		for _, eventLog := range logs {
			// Check that the trigger has not expired at the time of the event.
			if eventLog.BlockNumber > uint64(triggerRegisteredEvent.ExpirationBlockNumber) {
				continue
			}
			match, err := trigger.Match(&eventLog)
			if err != nil {
				triggerLog.Error().Err(err).Msg("failed to match trigger with event log")
				continue
			}
			if !match {
				triggerLog.Debug().
					Str("log", fmt.Sprintf("%+v", eventLog)).
					Msg("skipping log that matched filter but not additional predicates")
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
			Eon:            event.EventTriggerRegisteredEvent.Eon,
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
		log.Info().
			Int64("trigger-registered-block-number", event.EventTriggerRegisteredEvent.BlockNumber).
			Hex("trigger-registered-block-hash", event.EventTriggerRegisteredEvent.BlockHash).
			Int64("trigger-registered-tx-index", event.EventTriggerRegisteredEvent.TxIndex).
			Int64("trigger-registered-log-index", event.EventTriggerRegisteredEvent.LogIndex).
			Uint64("event-block-number", event.Log.BlockNumber).
			Hex("event-block-hash", event.Log.BlockHash.Bytes()).
			Uint("event-tx-index", event.Log.TxIndex).
			Uint("event-log-index", event.Log.Index).
			Msg("processed fired trigger event")
	}

	return nil
}

func (tp *TriggerProcessor) RollbackEvents(ctx context.Context, tx pgx.Tx, toBlock int64) error {
	queries := database.New(tx)
	err := queries.DeleteFiredTriggersFromBlockNumber(ctx, toBlock+1)
	if err != nil {
		return fmt.Errorf("failed to delete fired triggers from block number: %w", err)
	}
	return nil
}
