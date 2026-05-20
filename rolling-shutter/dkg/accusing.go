package dkg

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/shutter-network/shutter/shlib/puredkg"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/contract"
	corekeyperdb "github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/database"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/txsender"
)

// startAccusing is the per-block reactor for the Accusing phase. It enqueues
// a `submitAccusation` row iff this keyper has a valid in-memory dealing
// state (i.e. our polynomial is still alive in this process) and produces
// at least one accusation. The function is safely re-invokable: presence of
// any own accusation row for `(k, r)` is the idempotency marker.
func (m *Manager) startAccusing(ctx context.Context, dkgAddr common.Address, keyperConfigIndex, retryCounter int64) error {
	return m.cfg.DBPool.BeginFunc(ctx, func(tx pgx.Tx) error {
		pure, _, ownIndex, isMember, err := m.buildPureDKG(ctx, tx, keyperConfigIndex, retryCounter)
		if err != nil {
			return err
		}
		if !isMember {
			return nil
		}

		queries := corekeyperdb.New(tx)
		existing, err := queries.GetDKGAccusations(ctx, corekeyperdb.GetDKGAccusationsParams{
			KeyperConfigIndex: keyperConfigIndex,
			RetryCounter:      retryCounter,
		})
		if err != nil {
			return errors.Wrap(err, "load own accusation check")
		}
		for _, a := range existing {
			if uint64(a.AccuserIndex) == ownIndex {
				return nil
			}
		}

		if pure.Phase != puredkg.Dealing {
			// Per ADR 0003, the polynomial is not persisted. Without it we
			// cannot drive puredkg forward from Off → Accusing within this
			// invocation; skip submitting and the retry will start fresh.
			log.Debug().
				Int64("keyper-config-index", keyperConfigIndex).
				Int64("retry-counter", retryCounter).
				Str("phase", pure.Phase.String()).
				Msg("skipping accusations: puredkg phase not Dealing")
			return nil
		}

		accusations := pure.StartPhase2Accusing()
		if len(accusations) == 0 {
			return nil
		}

		accusedIndices := make([]uint64, 0, len(accusations))
		for _, a := range accusations {
			accusedIndices = append(accusedIndices, a.Accused)
			if err := queries.InsertDKGAccusation(ctx, corekeyperdb.InsertDKGAccusationParams{
				KeyperConfigIndex: keyperConfigIndex,
				RetryCounter:      retryCounter,
				AccuserIndex:      int64(ownIndex),
				AccusedIndex:      int64(a.Accused),
			}); err != nil {
				return errors.Wrap(err, "store own accusation row")
			}
		}

		abi, err := contract.DKGContractMetaData.GetAbi()
		if err != nil {
			return errors.Wrap(err, "load DKG contract ABI")
		}
		data, err := abi.Pack(
			"submitAccusation",
			uint64(keyperConfigIndex),
			uint64(retryCounter),
			ownIndex,
			accusedIndices,
		)
		if err != nil {
			return errors.Wrap(err, "pack submitAccusation calldata")
		}
		outboxID, err := txsender.EnqueueTx(ctx, tx, dkgAddr, data, nil)
		if err != nil {
			return errors.Wrap(err, "enqueue submitAccusation tx")
		}
		log.Info().
			Int64("keyper-config-index", keyperConfigIndex).
			Int64("retry-counter", retryCounter).
			Int("count", len(accusedIndices)).
			Int64("tx-outbox-id", outboxID).
			Msg("enqueued DKG accusations")
		return nil
	})
}
