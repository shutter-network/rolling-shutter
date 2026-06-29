package shutterservice

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
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

// FilterCriteria walks the active registered triggers and unions their
// contract addresses and topic[0] hashes. If any active trigger leaves
// topic[0] unconstrained, Topics0 is returned empty so the combined query
// fetches every log emitted by the listed addresses.
func (tp *TriggerProcessor) FilterCriteria(ctx context.Context, start, _ uint64) (FilterCriteria, error) {
	triggers, err := tp.activeTriggers(ctx, start)
	if err != nil {
		return FilterCriteria{}, err
	}

	addresses := map[common.Address]struct{}{}
	topics := map[common.Hash]struct{}{}
	anyTopic := false

	for _, t := range triggers {
		q, err := t.def.ToFilterQuery()
		if err != nil {
			// Should not occur for validated triggers, but skip rather than
			// poison the whole criteria union.
			log.Error().Err(err).Msg("failed to build filter query for active trigger; skipping")
			continue
		}
		for _, a := range q.Addresses {
			addresses[a] = struct{}{}
		}
		if len(q.Topics) == 0 || len(q.Topics[0]) == 0 {
			anyTopic = true
			continue
		}
		for _, h := range q.Topics[0] {
			topics[h] = struct{}{}
		}
	}

	crit := FilterCriteria{
		Addresses: make([]common.Address, 0, len(addresses)),
	}
	for a := range addresses {
		crit.Addresses = append(crit.Addresses, a)
	}
	if !anyTopic {
		crit.Topics0 = make([]common.Hash, 0, len(topics))
		for h := range topics {
			crit.Topics0 = append(crit.Topics0, h)
		}
	}
	return crit, nil
}

// ParseEvents inspects each log against every active trigger and emits a
// TriggerEvent for each (log, trigger) pair that matches.
func (tp *TriggerProcessor) ParseEvents(ctx context.Context, start, _ uint64, logs []types.Log) ([]Event, error) {
	if len(logs) == 0 {
		return nil, nil
	}
	triggers, err := tp.activeTriggers(ctx, start)
	if err != nil {
		return nil, err
	}

	var events []Event
	for _, t := range triggers {
		triggerLog := log.With().
			Int64("block-number", t.event.BlockNumber).
			Hex("block-hash", t.event.BlockHash).
			Int64("tx-index", t.event.TxIndex).
			Int64("log-index", t.event.LogIndex).
			Hex("identity-prefix", t.event.IdentityPrefix).
			Str("sender", t.event.Sender).
			Hex("definition", t.event.Definition).
			Int64("expiration-block-number", t.event.ExpirationBlockNumber).
			Logger()

		for i := range logs {
			eventLog := logs[i]
			if eventLog.BlockNumber > uint64(t.event.ExpirationBlockNumber) {
				continue
			}
			match, err := t.def.Match(&eventLog)
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
				EventTriggerRegisteredEvent: t.event,
			})
		}
	}
	return events, nil
}

// activeTrigger pairs a stored EventTriggerRegisteredEvent row with its parsed
// definition.
type activeTrigger struct {
	event database.EventTriggerRegisteredEvent
	def   EventTriggerDefinition
}

func (tp *TriggerProcessor) activeTriggers(ctx context.Context, start uint64) ([]activeTrigger, error) {
	queries := database.New(tp.DBPool)
	rows, err := queries.GetActiveEventTriggerRegisteredEvents(ctx, int64(start))
	if err != nil {
		return nil, errors.Wrap(err, "failed to get event trigger registered events")
	}
	active := make([]activeTrigger, 0, len(rows))
	for _, row := range rows {
		def := EventTriggerDefinition{}
		if err := def.UnmarshalBytes(row.Definition); err != nil {
			log.Error().Err(err).
				Int64("block-number", row.BlockNumber).
				Hex("definition", row.Definition).
				Msg("ignoring invalid trigger definition in database")
			continue
		}
		active = append(active, activeTrigger{event: row, def: def})
	}
	return active, nil
}

func (tp *TriggerProcessor) ProcessEvents(ctx context.Context, tx pgx.Tx, events []Event) error {
	queries := database.New(tx)
	for _, untypedEvent := range events {
		event := untypedEvent.(*TriggerEvent)
		err := queries.InsertFiredTrigger(ctx, database.InsertFiredTriggerParams{
			Eon:            event.EventTriggerRegisteredEvent.Eon,
			Identity:       event.EventTriggerRegisteredEvent.Identity,
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
