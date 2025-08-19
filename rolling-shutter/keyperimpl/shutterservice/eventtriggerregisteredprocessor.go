package shutterservice

import (
	"bytes"
	"context"
	"math"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	triggerRegistryBindings "github.com/shutter-network/contracts/v2/bindings/shuttereventtriggerregistry"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyperimpl/shutterservice/database"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/shdb"
)

// EventTriggerRegisteredEventProcessor implements the EventProcessor interface for EventTriggerRegistered events
// in the ShutterRegistry contract.
type EventTriggerRegisteredEventProcessor struct {
	Contract *triggerRegistryBindings.Shuttereventtriggerregistry
	DBPool   *pgxpool.Pool
}

func NewEventTriggerRegisteredEventProcessor(
	contract *triggerRegistryBindings.Shuttereventtriggerregistry,
	dbPool *pgxpool.Pool,
) *EventTriggerRegisteredEventProcessor {
	return &EventTriggerRegisteredEventProcessor{
		Contract: contract,
		DBPool:   dbPool,
	}
}

func (p *EventTriggerRegisteredEventProcessor) GetProcessorName() string {
	return "event_trigger_registered"
}

func (p *EventTriggerRegisteredEventProcessor) FetchEvents(ctx context.Context, start, end uint64) ([]Event, error) {
	opts := bind.FilterOpts{
		Start:   start,
		End:     &end,
		Context: ctx,
	}
	it, err := p.Contract.FilterEventTriggerRegistered(&opts, []uint64{})
	if err != nil {
		return nil, errors.Wrap(err, "failed to query identity registered events")
	}

	var events []Event
	for it.Next() {
		events = append(events, it.Event)
	}
	if it.Error() != nil {
		return nil, errors.Wrap(it.Error(), "failed to iterate identity registered events")
	}
	return events, nil
}

func (p *EventTriggerRegisteredEventProcessor) ProcessEvents(ctx context.Context, tx pgx.Tx, events []Event) error {
	queries := database.New(tx)
	for _, event := range events {
		registryEvent := event.(*triggerRegistryBindings.ShuttereventtriggerregistryEventTriggerRegistered)
		evLog := log.With().
			Uint64("block-number", registryEvent.Raw.BlockNumber).
			Hex("block-hash", registryEvent.Raw.BlockHash.Bytes()).
			Uint("tx-index", registryEvent.Raw.TxIndex).
			Uint("log-index", registryEvent.Raw.Index).
			Uint64("eon", registryEvent.Eon).
			Hex("identity-prefix", registryEvent.IdentityPrefix[:]).
			Str("sender", registryEvent.Sender.Hex()).
			Hex("definition", registryEvent.TriggerDefinition).
			Uint64("ttl", registryEvent.Ttl).
			Logger()

		if registryEvent.Eon > math.MaxInt64 {
			evLog.Info().Msg("skipping event trigger registered event with Eon > math.MaxInt64")
			continue
		}
		if registryEvent.Ttl > math.MaxInt64 {
			evLog.Info().Msg("skipping event trigger registered event with TTL > math.MaxInt64")
			continue
		}

		triggerDefinition := EventTriggerDefinition{}
		err := triggerDefinition.UnmarshalBytes(registryEvent.TriggerDefinition)
		if err != nil {
			evLog.Info().Err(err).Msg("skipping invalid trigger definition")
			continue
		}

		_, err = queries.InsertEventTriggerRegisteredEvent(ctx, database.InsertEventTriggerRegisteredEventParams{
			BlockNumber:    int64(registryEvent.Raw.BlockNumber),
			BlockHash:      registryEvent.Raw.BlockHash[:],
			TxIndex:        int64(registryEvent.Raw.TxIndex),
			LogIndex:       int64(registryEvent.Raw.Index),
			Eon:            int64(registryEvent.Eon),
			IdentityPrefix: registryEvent.IdentityPrefix[:],
			Sender:         shdb.EncodeAddress(registryEvent.Sender),
			Definition:     registryEvent.TriggerDefinition,
			Ttl:            int64(registryEvent.Ttl),
			Identity:       computeEventTriggerIdentity(registryEvent),
		})
		if err != nil {
			return errors.Wrap(err, "failed to insert event trigger registered event into db")
		}
		evLog.Info().Msg("processed event trigger registered event")
	}
	return nil
}

func (p *EventTriggerRegisteredEventProcessor) RollbackEvents(ctx context.Context, tx pgx.Tx, toBlock int64) error {
	queries := database.New(tx)
	err := queries.DeleteEventTriggerRegisteredEventsFromBlockNumber(ctx, toBlock+1)
	if err != nil {
		return errors.Wrap(err, "failed to delete event trigger registered events during rollback")
	}
	return nil
}

func computeEventTriggerIdentity(event *triggerRegistryBindings.ShuttereventtriggerregistryEventTriggerRegistered) []byte {
	var buf bytes.Buffer
	buf.Write(event.IdentityPrefix[:])
	buf.Write(event.Sender.Bytes())
	return crypto.Keccak256(buf.Bytes())
}
