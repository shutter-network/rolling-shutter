package synchandler

import (
	"context"
	"math"

	"github.com/ethereum/go-ethereum/log"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	bindings "github.com/shutter-network/gnosh-contracts/gnoshcontracts/sequencer"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyperimpl/gnosis/database"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/chainsync/syncer"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/shdb"
)

func init() {
	var err error
	SequencerContractABI, err = bindings.SequencerMetaData.GetAbi()
	if err != nil {
		panic(err)
	}
}

var SequencerContractABI *abi.ABI

func NewSequencerTransactionSubmitted(dbPool *pgxpool.Pool, address common.Address) (syncer.ContractEventHandler, error) {
	return syncer.WrapHandler(
		&SequencerTransactionSubmitted{
			evABI:   SequencerContractABI,
			address: address,
			dbPool:  dbPool,
		})
}

type SequencerTransactionSubmitted struct {
	log     log.Logger
	evABI   *abi.ABI
	address common.Address

	dbPool *pgxpool.Pool
}

func (sts *SequencerTransactionSubmitted) Address() common.Address {
	return sts.address
}

func (kb *SequencerTransactionSubmitted) Log(msg string, ctx ...any) {
	kb.log.Info(msg, ctx)
}

func (_ *SequencerTransactionSubmitted) Event() string {
	return "TransactionSubmitted"
}

func (sts *SequencerTransactionSubmitted) ABI() abi.ABI {
	return *sts.evABI
}

func (sts *SequencerTransactionSubmitted) Accept(
	ctx context.Context,
	header types.Header,
	ev bindings.SequencerTransactionSubmitted,
) (bool, error) {
	return true, nil
}
func (sts *SequencerTransactionSubmitted) Handle(
	ctx context.Context,
	qCtx syncer.QueryContext,
	events []bindings.SequencerTransactionSubmitted,
) error {
	err := sts.dbPool.BeginFunc(ctx, func(tx pgx.Tx) error {
		db := database.New(tx)
		if qCtx.Remove != nil {
			for _, header := range qCtx.Remove.Get() {
				if err := db.DeleteTransactionSubmittedEventsFromBlockHash(ctx, header.Hash().Bytes()); err != nil {
					return errors.Wrap(err, "failed to delete transaction submitted events from db")
				}
			}
			// log.Info().
			// 	Int("depth", numReorgedBlocks).
			// 	Int64("previous-synced-until", syncStatus.BlockNumber).
			// 	Int64("new-synced-until", newSyncedUntilBlockNumber).
			// 	Msg("sync status reset due to reorg")
		}
		for _, event := range sts.filterEvents(events) {
			_, err := db.InsertTransactionSubmittedEvent(ctx, database.InsertTransactionSubmittedEventParams{
				Index:          int64(event.TxIndex),
				BlockNumber:    int64(event.Raw.BlockNumber),
				BlockHash:      event.Raw.BlockHash[:],
				TxIndex:        int64(event.Raw.TxIndex),
				LogIndex:       int64(event.Raw.Index),
				Eon:            int64(event.Eon),
				IdentityPrefix: event.IdentityPrefix[:],
				Sender:         shdb.EncodeAddress(event.Sender),
				GasLimit:       event.GasLimit.Int64(),
			})
			if err != nil {
				return errors.Wrap(err, "failed to insert transaction submitted event into db")
			}
			// metricsLatestTxSubmittedEventIndex.WithLabelValues(fmt.Sprint(event.Eon)).Set(float64(event.TxIndex))
			// 	log.Debug().
			// 		Uint64("index", event.TxIndex).
			// 		Uint64("block", event.Raw.BlockNumber).
			// 		Uint64("eon", event.Eon).
			// 		Hex("identityPrefix", event.IdentityPrefix[:]).
			// 		Hex("sender", event.Sender.Bytes()).
			// 		Uint64("gasLimit", event.GasLimit.Uint64()).
			// 		Msg("synced new transaction submitted event")
			// }
		}

		return nil
	})
	// log.Info().
	// 	Uint64("start-block", start).
	// 	Uint64("end-block", end).
	// 	Int("num-inserted-events", len(filteredEvents)).
	// 	Int("num-discarded-events", len(events)-len(filteredEvents)).
	// 	Msg("synced sequencer contract")
	// metricsTxSubmittedEventsSyncedUntil.Set(float64(end))
	return err
}

func (sts *SequencerTransactionSubmitted) filterEvents(
	events []bindings.SequencerTransactionSubmitted,
) []bindings.SequencerTransactionSubmitted {
	filteredEvents := []bindings.SequencerTransactionSubmitted{}
	for _, event := range events {
		if event.Eon > math.MaxInt64 ||
			!event.GasLimit.IsInt64() {
			// log.Debug().
			// 	Uint64("eon", event.Eon).
			// 	Uint64("block-number", event.Raw.BlockNumber).
			// 	Str("block-hash", event.Raw.BlockHash.Hex()).
			// 	Uint("tx-index", event.Raw.TxIndex).
			// 	Uint("log-index", event.Raw.Index).
			// 	Msg("ignoring transaction submitted event with high eon")
			continue
		}
		filteredEvents = append(filteredEvents, event)
	}
	return filteredEvents
}
