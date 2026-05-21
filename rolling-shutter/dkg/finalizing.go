package dkg

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/shutter-network/shutter/shlib/puredkg"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/contract"
	corekeyperdb "github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/database"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/shdb"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/txsender"
)

// maybeFinalize is the per-block reactor for the Finalizing phase. It
// enqueues a `submitSuccessVote` tx with the locally-computed eon public key.
//
// Idempotency: a successful `dkg_result` row already exists either because
// the local finalization wrote it last time around, or because a chain-side
// `SuccessEvent` arrived and was handled by the host keyper. The
// `HandleBlock` loop also short-circuits the entire eon when this row is
// present; the check here defends against direct invocations.
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
	alreadySucceeded, err := queries.ExistsDKGResultSuccess(ctx, keyperConfigIndex)
	if err != nil {
		return errors.Wrap(err, "check existing dkg_result")
	}
	if alreadySucceeded {
		return nil
	}

	// buildPureDKG returns the puredkg at Phase=Apologizing with all
	// stored apologies applied. `ComputeResult` requires phase
	// `>= Finalized`; set it directly rather than calling `Finalize`
	// because we want to skip the trivial `setPhase` invariant check.
	pure.Phase = puredkg.Finalized
	result, err := pure.ComputeResult()
	if err != nil {
		log.Warn().Err(err).
			Int64("keyper-config-index", keyperConfigIndex).
			Int64("retry-counter", retryCounter).
			Msg("cannot compute DKG result; skipping success vote")
		return nil
	}

	// Persist the local DKG result so downstream consumers (key-share
	// signing) can decode our `puredkg.Result` once the success event
	// arrives. This is the per-eon marker that closes the idempotency
	// loop: the next invocation's `ExistsDKGResultSuccess` check sees
	// the row and skips.
	pureBytes, err := shdb.EncodePureDKGResult(&result)
	if err != nil {
		return errors.Wrap(err, "encode pure DKG result")
	}
	if err := queries.InsertDKGResult(ctx, corekeyperdb.InsertDKGResultParams{
		Eon:        keyperConfigIndex,
		Success:    true,
		Error:      sql.NullString{},
		PureResult: pureBytes,
	}); err != nil {
		return errors.Wrap(err, "insert dkg_result success row")
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
	log.Info().
		Int64("keyper-config-index", keyperConfigIndex).
		Int64("retry-counter", retryCounter).
		Int64("tx-outbox-id", outboxID).
		Msg("enqueued DKG success vote")
	return nil
}
