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
	"github.com/shutter-network/rolling-shutter/rolling-shutter/shdb"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/txsender"
)

// maybeDeal is the per-block reactor for the Dealing phase. Idempotency is
// keyed on the presence of a `dkg_sent_actions` row for
// `(keyperConfigIndex, retryCounter, "dealing")`; if one exists the call is a
// no-op. Otherwise we sample a polynomial via `puredkg.StartPhase1Dealing`,
// persist the resulting puredkg state (so accusations/apologies can be
// produced on later blocks even after a process restart), persist our own
// commitment + per-receiver evals, enqueue a `submitDealing` row in
// `tx_outbox` for `TxSender` to sign and submit, and insert the
// `dkg_sent_actions` marker atomically with the outbox row.
//
// The caller owns the (write) transaction and is responsible for committing
// or rolling back. `pure`, `keypers`, and `ownIndex` come from a prior read
// transaction's `buildPureDKG` call at the `handleEon` level.
func (m *Manager) maybeDeal(
	ctx context.Context,
	tx pgx.Tx,
	dkgAddr common.Address,
	keyperConfigIndex, retryCounter int64,
	pure *puredkg.PureDKG,
	keypers []common.Address,
	ownIndex uint64,
) error {
	queries := corekeyperdb.New(tx)
	alreadySent, err := queries.ExistsDKGSentAction(ctx, corekeyperdb.ExistsDKGSentActionParams{
		KeyperConfigIndex: keyperConfigIndex,
		RetryCounter:      retryCounter,
		Action:            ActionDealing,
	})
	if err != nil {
		return errors.Wrap(err, "check dkg sent action existence")
	}
	if alreadySent {
		return nil
	}

	commitmentMsg, polyEvalMsgs, err := pure.StartPhase1Dealing()
	if err != nil {
		return errors.Wrap(err, "puredkg StartPhase1Dealing")
	}

	// Persist the initial puredkg state — Phase=Dealing, polynomial set,
	// `Evals[ownIndex]` populated from the self-eval — before writing any
	// downstream rows. Subsequent blocks (and post-restart invocations)
	// can rebuild from this blob to drive Accusing/Apologizing.
	pureBytes, err := shdb.EncodePureDKG(pure)
	if err != nil {
		return errors.Wrap(err, "encode initial puredkg state")
	}
	if err := queries.InsertDKGInitialState(ctx, corekeyperdb.InsertDKGInitialStateParams{
		KeyperConfigIndex: keyperConfigIndex,
		RetryCounter:      retryCounter,
		PuredkgBytes:      pureBytes,
	}); err != nil {
		return errors.Wrap(err, "store initial puredkg state")
	}

	commitmentBytes := commitmentMsg.Gammas.Marshal()
	if err := queries.InsertDKGPolyCommitment(ctx, corekeyperdb.InsertDKGPolyCommitmentParams{
		KeyperConfigIndex: keyperConfigIndex,
		RetryCounter:      retryCounter,
		KeyperIndex:       int64(ownIndex),
		Commitment:        commitmentBytes,
	}); err != nil {
		return errors.Wrap(err, "store own poly commitment")
	}

	receivers := ReceiverIndicesForSender(uint64(len(keypers)), ownIndex)
	encryptedEvals := make([][]byte, 0, len(receivers))
	for _, recvIdx := range receivers {
		var evalMsg *puredkg.PolyEvalMsg
		for i := range polyEvalMsgs {
			if polyEvalMsgs[i].Receiver == recvIdx {
				evalMsg = &polyEvalMsgs[i]
				break
			}
		}
		if evalMsg == nil {
			return errors.Errorf("no poly eval for receiver %d", recvIdx)
		}
		recvAddr := keypers[recvIdx]
		ciphertext, err := m.encryptPolyEvalFor(ctx, queries, recvAddr, evalMsg.Eval)
		if err != nil {
			return errors.Wrapf(err, "encrypt poly eval for receiver %d (%s)", recvIdx, recvAddr.Hex())
		}
		encryptedEvals = append(encryptedEvals, ciphertext)

		if err := queries.InsertDKGPolyEval(ctx, corekeyperdb.InsertDKGPolyEvalParams{
			KeyperConfigIndex: keyperConfigIndex,
			RetryCounter:      retryCounter,
			SenderIndex:       int64(ownIndex),
			ReceiverIndex:     int64(recvIdx),
			EncryptedEval:     ciphertext,
		}); err != nil {
			return errors.Wrap(err, "store own poly eval row")
		}
	}

	// Also persist the self-eval. puredkg's `StartPhase1Dealing` consumes
	// the self-message in-memory (setting `pure.Evals[ownIndex]`); without
	// a DB-backed copy the replay path leaves that slot nil and we cannot
	// later compute the result on a per-block reactor cycle. We encrypt
	// to ourselves so `decryptPolyEval` recovers it transparently during
	// replay; the on-chain `submitDealing` call does NOT include this row
	// (it is excluded by `ReceiverIndicesForSender`).
	selfEval := pure.Evals[ownIndex]
	if selfEval == nil {
		return errors.Errorf("puredkg did not produce a self-eval after StartPhase1Dealing")
	}
	selfCiphertext, err := m.encryptPolyEvalFor(ctx, queries, m.cfg.OwnAddress, selfEval)
	if err != nil {
		return errors.Wrap(err, "encrypt self poly eval")
	}
	if err := queries.InsertDKGPolyEval(ctx, corekeyperdb.InsertDKGPolyEvalParams{
		KeyperConfigIndex: keyperConfigIndex,
		RetryCounter:      retryCounter,
		SenderIndex:       int64(ownIndex),
		ReceiverIndex:     int64(ownIndex),
		EncryptedEval:     selfCiphertext,
	}); err != nil {
		return errors.Wrap(err, "store self poly eval row")
	}

	abi, err := contract.DKGContractMetaData.GetAbi()
	if err != nil {
		return errors.Wrap(err, "load DKG contract ABI")
	}
	data, err := abi.Pack(
		"submitDealing",
		uint64(keyperConfigIndex),
		uint64(retryCounter),
		ownIndex,
		commitmentBytes,
		encryptedEvals,
	)
	if err != nil {
		return errors.Wrap(err, "pack submitDealing calldata")
	}
	label := fmt.Sprintf("submitDealing ksi=%d retry=%d", keyperConfigIndex, retryCounter)
	outboxID, err := txsender.EnqueueTx(ctx, tx, dkgAddr, data, nil, label)
	if err != nil {
		return errors.Wrap(err, "enqueue submitDealing tx")
	}
	if err := queries.InsertDKGSentAction(ctx, corekeyperdb.InsertDKGSentActionParams{
		KeyperConfigIndex: keyperConfigIndex,
		RetryCounter:      retryCounter,
		Action:            ActionDealing,
		OutboxID:          outboxID,
	}); err != nil {
		return errors.Wrap(err, "store dealing sent action marker")
	}
	log.Info().
		Int64("keyper-config-index", keyperConfigIndex).
		Int64("retry-counter", retryCounter).
		Uint64("keyper-index", ownIndex).
		Int64("tx-outbox-id", outboxID).
		Msg("enqueued DKG dealing")
	return nil
}
