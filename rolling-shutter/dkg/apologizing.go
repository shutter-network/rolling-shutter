package dkg

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/shutter-network/contracts/v2/bindings/dkgcontract"

	"github.com/shutter-network/shutter/shlib/puredkg"

	corekeyperdb "github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/database"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/txsender"
)

// maybeApologize is the per-block reactor for the Apologizing phase. It
// enqueues an apology tx when our locally-rebuilt puredkg has at least one
// accusation against us. Presence of a `dkg_sent_actions` row for
// `(k, r, "apologizing")` is the idempotency marker.
//
// Our own apology rows are NOT written to the shared `dkg_apologies` table
// — the chain syncer is the sole writer to it, so every keyper has the
// same view at each block height. Replay in `buildPureDKG` picks up our
// own apologies only after the on-chain submitApology event is indexed.
//
// The caller passes a `pure` returned by `buildPureDKG(PhaseApologizing)` —
// Phase=Accusing with commitments + evals + accusations replayed. We call
// StartPhase3Apologizing here, which advances the phase and emits apology
// messages with the polynomial loaded from `dkg_initial_states`.
func (m *Manager) maybeApologize(
	ctx context.Context,
	tx pgx.Tx,
	dkgAddr common.Address,
	keyperSetIndex, retryCounter int64,
	pure *puredkg.PureDKG,
	keypers []common.Address,
	ownIndex uint64,
) error {
	queries := corekeyperdb.New(tx)
	alreadySent, err := queries.ExistsDKGSentAction(ctx, corekeyperdb.ExistsDKGSentActionParams{
		KeyperSetIndex: keyperSetIndex,
		RetryCounter:   retryCounter,
		Action:         ActionApologizing,
	})
	if err != nil {
		return errors.Wrap(err, "check dkg sent action existence")
	}
	if alreadySent {
		return nil
	}

	apologies := pure.StartPhase3Apologizing()
	if len(apologies) == 0 {
		log.Info().
			Int64("keyper-set-index", keyperSetIndex).
			Int64("retry-counter", retryCounter).
			Msg("Not sending apologies, nobody accused us")
		// Mark the phase as resolved with a NULL tx_outbox_id row so the
		// log line above runs at most once per DKG Instance. The set of
		// on-chain accusations against us is fixed before the Apologizing
		// phase begins, so re-evaluating on every block is redundant work.
		if err := queries.InsertDKGSentAction(ctx, corekeyperdb.InsertDKGSentActionParams{
			KeyperSetIndex: keyperSetIndex,
			RetryCounter:   retryCounter,
			Action:         ActionApologizing,
			TxOutboxID:     sql.NullInt64{},
		}); err != nil {
			return errors.Wrap(err, "store apologizing sent action marker (no apologies)")
		}
		return nil
	}

	accuserIndices := make([]uint64, 0, len(apologies))
	accuserDescriptions := make([]string, 0, len(apologies))
	polyEvalData := make([][]byte, 0, len(apologies))
	for _, ap := range apologies {
		evalBytes := ap.Eval.Bytes()
		accuserIndices = append(accuserIndices, ap.Accuser)
		accuserDescriptions = append(accuserDescriptions, fmt.Sprintf("%d (%s)", ap.Accuser, keypers[ap.Accuser].Hex()))
		polyEvalData = append(polyEvalData, evalBytes)
	}

	abi, err := dkgcontract.DkgcontractMetaData.GetAbi()
	if err != nil {
		return errors.Wrap(err, "load DKG contract ABI")
	}
	data, err := abi.Pack(
		"submitApology",
		uint64(keyperSetIndex),
		uint64(retryCounter),
		ownIndex,
		accuserIndices,
		polyEvalData,
	)
	if err != nil {
		return errors.Wrap(err, "pack submitApology calldata")
	}
	label := fmt.Sprintf("submitApology ksi=%d retry=%d", keyperSetIndex, retryCounter)
	outboxID, err := txsender.EnqueueTx(ctx, tx, dkgAddr, data, nil, label)
	if err != nil {
		return errors.Wrap(err, "enqueue submitApology tx")
	}
	if err := queries.InsertDKGSentAction(ctx, corekeyperdb.InsertDKGSentActionParams{
		KeyperSetIndex: keyperSetIndex,
		RetryCounter:   retryCounter,
		Action:         ActionApologizing,
		TxOutboxID:     sql.NullInt64{Int64: outboxID, Valid: true},
	}); err != nil {
		return errors.Wrap(err, "store apologizing sent action marker")
	}
	log.Info().
		Int64("keyper-set-index", keyperSetIndex).
		Int64("retry-counter", retryCounter).
		Int("count", len(accuserIndices)).
		Strs("accusers", accuserDescriptions).
		Msg("submitting DKG apologies")
	return nil
}
