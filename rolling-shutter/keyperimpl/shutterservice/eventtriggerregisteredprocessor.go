package shutterservice

import (
	"bytes"
	"context"
	"math"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	triggerRegistryV1Bindings "github.com/shutter-network/contracts/v2/bindings/shuttereventtriggerregistryv1"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyperimpl/shutterservice/database"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/shdb"
)

// EventTriggerRegisteredEventProcessor implements the EventProcessor interface for EventTriggerRegistered events
// in the ShutterRegistry contract.
type EventTriggerRegisteredEventProcessor struct {
	Contract *triggerRegistryV1Bindings.Shuttereventtriggerregistryv1
	Address  common.Address
	DBPool   *pgxpool.Pool

	topic common.Hash
}

func NewEventTriggerRegisteredEventProcessor(
	contract *triggerRegistryV1Bindings.Shuttereventtriggerregistryv1,
	address common.Address,
	dbPool *pgxpool.Pool,
) (*EventTriggerRegisteredEventProcessor, error) {
	parsedABI, err := triggerRegistryV1Bindings.Shuttereventtriggerregistryv1MetaData.GetAbi()
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse trigger registry ABI")
	}
	ev, ok := parsedABI.Events["EventTriggerRegistered"]
	if !ok {
		return nil, errors.New("EventTriggerRegistered event not present in trigger registry ABI")
	}
	return &EventTriggerRegisteredEventProcessor{
		Contract: contract,
		Address:  address,
		DBPool:   dbPool,
		topic:    ev.ID,
	}, nil
}

func (p *EventTriggerRegisteredEventProcessor) GetProcessorName() string {
	return "event_trigger_registered"
}

func (p *EventTriggerRegisteredEventProcessor) FilterCriteria(_ context.Context, _, _ uint64) (FilterCriteria, error) {
	return FilterCriteria{
		Addresses: []common.Address{p.Address},
		Topics0:   []common.Hash{p.topic},
	}, nil
}

func (p *EventTriggerRegisteredEventProcessor) ParseEvents(_ context.Context, _, _ uint64, logs []types.Log) ([]Event, error) {
	events := make([]Event, 0, len(logs))
	for i := range logs {
		l := logs[i]
		ev, err := p.Contract.ParseEventTriggerRegistered(l)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse EventTriggerRegistered log")
		}
		events = append(events, ev)
	}
	return events, nil
}

func (p *EventTriggerRegisteredEventProcessor) ProcessEvents(ctx context.Context, tx pgx.Tx, events []Event) error {
	queries := database.New(tx)
	for _, event := range events {
		registryEvent := event.(*triggerRegistryV1Bindings.Shuttereventtriggerregistryv1EventTriggerRegistered)
		evLog := log.With().
			Uint64("block-number", registryEvent.Raw.BlockNumber).
			Hex("block-hash", registryEvent.Raw.BlockHash.Bytes()).
			Uint("tx-index", registryEvent.Raw.TxIndex).
			Uint("log-index", registryEvent.Raw.Index).
			Uint64("eon", registryEvent.Eon).
			Hex("identity-prefix", registryEvent.IdentityPrefix[:]).
			Str("sender", registryEvent.Sender.Hex()).
			Hex("definition", registryEvent.TriggerDefinition).
			Uint64("expirationBlockNumber", registryEvent.ExpirationBlockNumber).
			Logger()

		if registryEvent.Eon > math.MaxInt64 {
			evLog.Info().Msg("skipping event trigger registered event with Eon > math.MaxInt64")
			continue
		}
		if registryEvent.ExpirationBlockNumber > math.MaxInt64 {
			evLog.Info().Msg("skipping event trigger registered event with ExpirationBlockNumber > math.MaxInt64")
			continue
		}

		triggerDefinition := EventTriggerDefinition{}
		err := triggerDefinition.UnmarshalBytes(registryEvent.TriggerDefinition)
		if err != nil {
			evLog.Info().Err(err).Msg("skipping invalid trigger definition")
			continue
		}

		_, err = queries.InsertEventTriggerRegisteredEvent(ctx, database.InsertEventTriggerRegisteredEventParams{
			BlockNumber:           int64(registryEvent.Raw.BlockNumber),
			BlockHash:             registryEvent.Raw.BlockHash[:],
			TxIndex:               int64(registryEvent.Raw.TxIndex),
			LogIndex:              int64(registryEvent.Raw.Index),
			Eon:                   int64(registryEvent.Eon),
			IdentityPrefix:        registryEvent.IdentityPrefix[:],
			Sender:                shdb.EncodeAddress(registryEvent.Sender),
			Definition:            registryEvent.TriggerDefinition,
			ExpirationBlockNumber: int64(registryEvent.ExpirationBlockNumber),
			Identity:              computeEventTriggerIdentity(registryEvent),
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

func computeEventTriggerIdentity(event *triggerRegistryV1Bindings.Shuttereventtriggerregistryv1EventTriggerRegistered) []byte {
	var buf bytes.Buffer
	buf.Write(event.IdentityPrefix[:])
	buf.Write(event.Sender.Bytes())
	buf.Write(event.TriggerDefinition)
	return crypto.Keccak256(buf.Bytes())
}
