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

// startFinalizing is the per-block reactor for the Finalizing phase. It
// enqueues a `submitSuccessVote` tx with the locally-computed eon public key.
//
// Idempotency: a successful `dkg_result` row already exists either because
// the local finalization wrote it last time around, or because a chain-side
// `SuccessEvent` arrived and was handled by the host keyper. In either case
// no fresh vote is needed.
//
// The recovery path on every block: rebuild `puredkg` from stored messages
// (including the self-eval row written by `maybeDeal`), bypass puredkg's
// phase machinery by setting `pure.Phase = Finalized` directly, and call
// `ComputeResult`. The result depends only on `Commitments` / `Evals` /
// `Accusations` / `Apologies`, all of which are populated by replay.
func (m *Manager) startFinalizing(ctx context.Context, dkgAddr common.Address, keyperConfigIndex, retryCounter int64) error {
	return m.cfg.DBPool.BeginFunc(ctx, func(tx pgx.Tx) error {
		pure, _, ownIndex, isMember, err := m.buildPureDKG(ctx, tx, keyperConfigIndex, retryCounter)
		if err != nil {
			return err
		}
		if !isMember {
			return nil
		}

		queries := corekeyperdb.New(tx)

		// Idempotency: presence of a `dkg_result` success row means either
		// we have already enqueued our vote in a prior call or the chain
		// has already concluded DKG success via another keyper's vote.
		// In both cases there is nothing to do.
		alreadySucceeded, err := queries.ExistsDKGResultSuccess(ctx, keyperConfigIndex)
		if err != nil {
			return errors.Wrap(err, "check existing dkg_result")
		}
		if alreadySucceeded {
			return nil
		}

		// We never persist puredkg's internal phase, and the rebuild path
		// leaves `pure.Phase` at `Off`. `ComputeResult` requires phase
		// `>= Finalized`; set it directly. The other phase-gated methods
		// (`Finalize`, the `Start*` calls) are not invoked here — instead
		// we rely on the replayed `Commitments` / `Evals` / `Accusations` /
		// `Apologies` state that `ComputeResult` actually reads. The
		// self-eval row written by `maybeDeal` populates
		// `pure.Evals[ownIndex]` during replay.
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
	})
}
