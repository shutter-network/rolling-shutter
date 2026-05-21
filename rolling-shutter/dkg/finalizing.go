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

// maybeFinalize is the per-block reactor for the Finalizing phase. It
// enqueues a `submitSuccessVote` tx with the locally-computed eon public key.
//
// Idempotency uses `ExistsDKGSentAction` keyed on (keyperConfigIndex,
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
	keyperConfigIndex, retryCounter int64,
	pure *puredkg.PureDKG,
	ownIndex uint64,
) error {
	queries := corekeyperdb.New(tx)
	alreadySent, err := queries.ExistsDKGSentAction(ctx, corekeyperdb.ExistsDKGSentActionParams{
		KeyperConfigIndex: keyperConfigIndex,
		RetryCounter:      retryCounter,
		Action:            ActionFinalizing,
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
			Int64("keyper-config-index", keyperConfigIndex).
			Int64("retry-counter", retryCounter).
			Msg("cannot compute DKG result; skipping success vote")
		return nil
	}

	eonPubKeyBytes, err := result.PublicKey.GobEncode()
	if err != nil {
		return errors.Wrap(err, "encode eon public key")
	}

	abi, err := contract.DKGContractMetaData.GetAbi()
	if err != nil {
		return errors.Wrap(err, "load DKG contract ABI")
	}
	data, err := abi.Pack(
		"submitSuccessVote",
		uint64(keyperConfigIndex),
		uint64(retryCounter),
		ownIndex,
		eonPubKeyBytes,
	)
	if err != nil {
		return errors.Wrap(err, "pack submitSuccessVote calldata")
	}
	label := fmt.Sprintf("submitSuccessVote ksi=%d retry=%d", keyperConfigIndex, retryCounter)
	outboxID, err := txsender.EnqueueTx(ctx, tx, dkgAddr, data, nil, label)
	if err != nil {
		return errors.Wrap(err, "enqueue submitSuccessVote tx")
	}
	if err := queries.InsertDKGSentAction(ctx, corekeyperdb.InsertDKGSentActionParams{
		KeyperConfigIndex: keyperConfigIndex,
		RetryCounter:      retryCounter,
		Action:            ActionFinalizing,
		OutboxID:          outboxID,
	}); err != nil {
		return errors.Wrap(err, "store finalizing sent action marker")
	}
	log.Info().
		Int64("keyper-config-index", keyperConfigIndex).
		Int64("retry-counter", retryCounter).
		Int64("tx-outbox-id", outboxID).
		Msg("enqueued DKG success vote")
	return nil
}
