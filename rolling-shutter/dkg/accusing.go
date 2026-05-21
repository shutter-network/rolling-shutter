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

// maybeAccuse is the per-block reactor for the Accusing phase. It enqueues a
// `submitAccusation` row when our locally-rebuilt puredkg (loaded from
// `dkg_initial_states` and replayed with received dealings) yields at least
// one accusation. The function is safely re-invokable: presence of a
// `dkg_sent_actions` row for `(k, r, "accusing")` is the idempotency marker.
//
// Our own accusation rows are NOT written to the shared `dkg_accusations`
// table — the chain syncer is the sole writer to it, so every keyper has
// the same view at each block height. Replay in `buildPureDKG` picks up
// our own accusations only after the on-chain submitAccusation event is
// indexed.
//
// The caller passes a `pure` returned by `buildPureDKG(PhaseAccusing)` —
// Phase=Dealing with commitments + evals applied; here we call
// StartPhase2Accusing, which advances the phase and emits accusations for
// dealers whose PolyEval is missing or fails verification.
func (m *Manager) maybeAccuse(
	ctx context.Context,
	tx pgx.Tx,
	dkgAddr common.Address,
	keyperConfigIndex, retryCounter int64,
	pure *puredkg.PureDKG,
	ownIndex uint64,
) error {
	queries := corekeyperdb.New(tx)
	alreadySent, err := queries.ExistsDKGSentAction(ctx, corekeyperdb.ExistsDKGSentActionParams{
		KeyperConfigIndex: keyperConfigIndex,
		RetryCounter:      retryCounter,
		Action:            ActionAccusing,
	})
	if err != nil {
		return errors.Wrap(err, "check dkg sent action existence")
	}
	if alreadySent {
		return nil
	}

	accusations := pure.StartPhase2Accusing()
	if len(accusations) == 0 {
		log.Debug().
			Int64("keyper-config-index", keyperConfigIndex).
			Int64("retry-counter", retryCounter).
			Msg("no DKG accusations to submit: all dealers honest")
		return nil
	}

	accusedIndices := make([]uint64, 0, len(accusations))
	for _, a := range accusations {
		accusedIndices = append(accusedIndices, a.Accused)
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
	label := fmt.Sprintf("submitAccusation ksi=%d retry=%d", keyperConfigIndex, retryCounter)
	outboxID, err := txsender.EnqueueTx(ctx, tx, dkgAddr, data, nil, label)
	if err != nil {
		return errors.Wrap(err, "enqueue submitAccusation tx")
	}
	if err := queries.InsertDKGSentAction(ctx, corekeyperdb.InsertDKGSentActionParams{
		KeyperConfigIndex: keyperConfigIndex,
		RetryCounter:      retryCounter,
		Action:            ActionAccusing,
		OutboxID:          outboxID,
	}); err != nil {
		return errors.Wrap(err, "store accusing sent action marker")
	}
	log.Info().
		Int64("keyper-config-index", keyperConfigIndex).
		Int64("retry-counter", retryCounter).
		Int("count", len(accusedIndices)).
		Int64("tx-outbox-id", outboxID).
		Msg("enqueued DKG accusations")
	return nil
}
