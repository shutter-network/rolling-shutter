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

// maybeFinalize is the per-block reactor for the Finalizing phase. It
// enqueues a `submitSuccessVote` tx with the locally-computed eon public key.
//
// Idempotency uses `ExistsDKGSentAction` keyed on (keyperSetIndex,
// retryCounter, ActionFinalizing). Using the retry-scoped sent-actions row
// rather than `ExistsDKGResultSuccess` means a keyper that voted for retry 0
// is not blocked from voting again when retry 1 starts: a different
// retryCounter produces a different row. `dkg_result` is written only when
// the chain emits a DKGSucceeded event (via `Manager.HandleDKGSuccess`), so
// it is never present here.
//
// The caller passes a `pure` returned by `buildPureDKG(PhaseFinalizing)` —
// Phase=Apologizing with all stored apologies applied. We bypass puredkg's
// phase machinery by setting `pure.Phase = Finalized` directly and call
// `ComputeResult`.
func (m *Manager) maybeFinalize(
	ctx context.Context,
	tx pgx.Tx,
	dkgAddr common.Address,
	keyperSetIndex, retryCounter int64,
	pure *puredkg.PureDKG,
	ownIndex uint64,
) error {
	queries := corekeyperdb.New(tx)
	alreadySent, err := queries.ExistsDKGSentAction(ctx, corekeyperdb.ExistsDKGSentActionParams{
		KeyperSetIndex: keyperSetIndex,
		RetryCounter:   retryCounter,
		Action:         ActionFinalizing,
	})
	if err != nil {
		return errors.Wrap(err, "check finalizing sent action")
	}
	if alreadySent {
		return nil
	}

	// buildPureDKG returns the puredkg at Phase=Apologizing with all
	// stored apologies applied. Finalize() advances it to Finalized so
	// ComputeResult can run.
	pure.Finalize()
	result, err := pure.ComputeResult()
	if err != nil {
		log.Warn().Err(err).
			Int64("keyper-set-index", keyperSetIndex).
			Int64("retry-counter", retryCounter).
			Msg("cannot compute DKG result; skipping success vote")
		// Mark the phase as resolved with a NULL tx_outbox_id row so the
		// warning above runs at most once per DKG Instance. The local
		// puredkg state is stable once all on-chain messages have been
		// indexed; re-running ComputeResult on every block is redundant.
		if insertErr := queries.InsertDKGSentAction(ctx, corekeyperdb.InsertDKGSentActionParams{
			KeyperSetIndex: keyperSetIndex,
			RetryCounter:   retryCounter,
			Action:         ActionFinalizing,
			TxOutboxID:     sql.NullInt64{},
		}); insertErr != nil {
			return errors.Wrap(insertErr, "store finalizing sent action marker (compute result failed)")
		}
		return nil
	}

	eonPubKeyBytes, err := result.PublicKey.GobEncode()
	if err != nil {
		return errors.Wrap(err, "encode eon public key")
	}

	abi, err := dkgcontract.DkgcontractMetaData.GetAbi()
	if err != nil {
		return errors.Wrap(err, "load DKG contract ABI")
	}
	data, err := abi.Pack(
		"submitSuccessVote",
		uint64(keyperSetIndex),
		uint64(retryCounter),
		ownIndex,
		eonPubKeyBytes,
	)
	if err != nil {
		return errors.Wrap(err, "pack submitSuccessVote calldata")
	}
	label := fmt.Sprintf("submitSuccessVote ksi=%d retry=%d", keyperSetIndex, retryCounter)
	outboxID, err := txsender.EnqueueTx(ctx, tx, dkgAddr, data, nil, label)
	if err != nil {
		return errors.Wrap(err, "enqueue submitSuccessVote tx")
	}
	if err := queries.InsertDKGSentAction(ctx, corekeyperdb.InsertDKGSentActionParams{
		KeyperSetIndex: keyperSetIndex,
		RetryCounter:   retryCounter,
		Action:         ActionFinalizing,
		TxOutboxID:     sql.NullInt64{Int64: outboxID, Valid: true},
	}); err != nil {
		return errors.Wrap(err, "store finalizing sent action marker")
	}
	log.Info().
		Int64("keyper-set-index", keyperSetIndex).
		Int64("retry-counter", retryCounter).
		Int64("tx-outbox-id", outboxID).
		Msg("enqueued DKG success vote")
	return nil
}
