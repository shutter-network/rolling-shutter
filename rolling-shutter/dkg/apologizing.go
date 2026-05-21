package dkg

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/shutter-network/shutter/shlib/puredkg"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/contract"
	corekeyperdb "github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/database"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/txsender"
)

// startApologizing is the per-block reactor for the Apologizing phase. It
// enqueues an apology tx iff someone accused us and our polynomial is still
// in memory. Presence of our own apology row for `(k, r)` is the idempotency
// marker.
func (m *Manager) startApologizing(ctx context.Context, dkgAddr common.Address, keyperConfigIndex, retryCounter int64) error {
	return m.cfg.DBPool.BeginFunc(ctx, func(tx pgx.Tx) error {
		pure, _, ownIndex, isMember, err := m.buildPureDKG(ctx, tx, keyperConfigIndex, retryCounter)
		if err != nil {
			return err
		}
		if !isMember {
			return nil
		}

		queries := corekeyperdb.New(tx)
		existing, err := queries.GetDKGApologies(ctx, corekeyperdb.GetDKGApologiesParams{
			KeyperConfigIndex: keyperConfigIndex,
			RetryCounter:      retryCounter,
		})
		if err != nil {
			return errors.Wrap(err, "load own apology check")
		}
		for _, ap := range existing {
			if uint64(ap.ApologizerIndex) == ownIndex {
				return nil
			}
		}

		if pure.Phase != puredkg.Accusing {
			log.Debug().
				Int64("keyper-config-index", keyperConfigIndex).
				Int64("retry-counter", retryCounter).
				Str("phase", pure.Phase.String()).
				Msg("skipping apologies: puredkg phase not Accusing")
			return nil
		}

		apologies := pure.StartPhase3Apologizing()
		if len(apologies) == 0 {
			return nil
		}

		accuserIndices := make([]uint64, 0, len(apologies))
		polyEvalData := make([][]byte, 0, len(apologies))
		for _, ap := range apologies {
			evalBytes := ap.Eval.Bytes()
			accuserIndices = append(accuserIndices, ap.Accuser)
			polyEvalData = append(polyEvalData, evalBytes)
			if err := queries.InsertDKGApology(ctx, corekeyperdb.InsertDKGApologyParams{
				KeyperConfigIndex: keyperConfigIndex,
				RetryCounter:      retryCounter,
				ApologizerIndex:   int64(ownIndex),
				AccuserIndex:      int64(ap.Accuser),
				PolyEval:          evalBytes,
			}); err != nil {
				return errors.Wrap(err, "store own apology row")
			}
		}

		abi, err := contract.DKGContractMetaData.GetAbi()
		if err != nil {
			return errors.Wrap(err, "load DKG contract ABI")
		}
		data, err := abi.Pack(
			"submitApology",
			uint64(keyperConfigIndex),
			uint64(retryCounter),
			ownIndex,
			accuserIndices,
			polyEvalData,
		)
		if err != nil {
			return errors.Wrap(err, "pack submitApology calldata")
		}
		label := fmt.Sprintf("submitApology ksi=%d retry=%d", keyperConfigIndex, retryCounter)
		outboxID, err := txsender.EnqueueTx(ctx, tx, dkgAddr, data, nil, label)
		if err != nil {
			return errors.Wrap(err, "enqueue submitApology tx")
		}
		log.Info().
			Int64("keyper-config-index", keyperConfigIndex).
			Int64("retry-counter", retryCounter).
			Int("count", len(accuserIndices)).
			Int64("tx-outbox-id", outboxID).
			Msg("enqueued DKG apologies")
		return nil
	})
}
