package gnosis

import (
	"context"
	"database/sql"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/shutter-network/shutter/shlib/puredkg"

	corekeyperdb "github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/database"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/shdb"
)

// phaseParamsForEon returns the DKG phase length and lead length for the
// given eon. Rows populated by `processNewKeyperSet` carry the values read
// from the keyper-set-specific DKG contract; rows from older databases (or
// rows where the lookup failed) carry NULLs, in which case the keyper falls
// back to the startup-time values from the config-supplied DKG contract.
func (kpr *Keyper) phaseParamsForEon(eon corekeyperdb.Eon) (phaseLength, leadLength uint64) {
	if eon.PhaseLength.Valid && eon.LeadLength.Valid {
		return uint64(eon.PhaseLength.Int64), uint64(eon.LeadLength.Int64)
	}
	log.Warn().
		Int64("keyper-config-index", eon.KeyperConfigIndex).
		Uint64("fallback-phase-length", kpr.dkgPhaseLength).
		Uint64("fallback-lead-length", kpr.dkgLeadLength).
		Msg("eons row missing DKG phase params; falling back to config-supplied DKG contract")
	return kpr.dkgPhaseLength, kpr.dkgLeadLength
}

// processDKGBlock advances the DKG participation loop for every eons row the
// keyper is a member of. It is invoked once per Gnosis Chain block from
// `processNewBlock`. The phase boundary is detected by comparing the phase
// at the current block with the phase at block N-1; on a boundary, the
// corresponding `start*` action runs.
func (kpr *Keyper) processDKGBlock(ctx context.Context, blockNumber uint64) error {
	if blockNumber == 0 {
		return nil
	}
	eons, err := corekeyperdb.New(kpr.dbpool).GetAllEons(ctx)
	if err != nil {
		return errors.Wrap(err, "list eons")
	}
	for _, eon := range eons {
		activationBlock, err := medley.Int64ToUint64Safe(eon.ActivationBlockNumber)
		if err != nil {
			return errors.Wrap(err, "convert activation block")
		}
		phaseLength, leadLength := kpr.phaseParamsForEon(eon)
		if phaseLength == 0 {
			continue
		}
		retry := CurrentRetryCounter(activationBlock, leadLength, phaseLength, blockNumber)
		phaseNow := PhaseAt(activationBlock, leadLength, phaseLength, retry, blockNumber)
		phasePrev := PhaseAt(activationBlock, leadLength, phaseLength, retry, blockNumber-1)
		retryPrev := CurrentRetryCounter(activationBlock, leadLength, phaseLength, blockNumber-1)

		// Phase boundary within the current retry, or a retry-boundary
		// (which also surfaces as a Dealing-start for the new retry).
		boundary := phaseNow != phasePrev || retry != retryPrev
		if !boundary {
			continue
		}

		if err := kpr.handlePhaseBoundary(ctx, eon, retry, phaseNow, retryPrev, phasePrev); err != nil {
			log.Error().Err(err).
				Int64("keyper-config-index", eon.KeyperConfigIndex).
				Uint64("retry-counter", retry).
				Str("phase", phaseNow.String()).
				Msg("DKG phase action failed")
		}
	}
	return nil
}

// handlePhaseBoundary dispatches phase-start and retry-rollover actions for a
// single eon. Errors are returned so the caller can log per-eon but failures
// of one eon don't abort progress for the others.
func (kpr *Keyper) handlePhaseBoundary(
	ctx context.Context,
	eon corekeyperdb.Eon,
	retry uint64,
	phaseNow DKGPhase,
	retryPrev uint64,
	phasePrev DKGPhase,
) error {
	keyperConfigIndex := eon.KeyperConfigIndex
	retryInt64 := int64(retry)

	// A retry rollover means the previous retry's Finalizing phase elapsed
	// without an on-chain success — record the failure once.
	if retry != retryPrev {
		if err := kpr.recordDKGFailureIfNotSucceeded(ctx, keyperConfigIndex, int64(retryPrev)); err != nil {
			log.Error().Err(err).
				Int64("keyper-config-index", keyperConfigIndex).
				Uint64("retry-counter", retryPrev).
				Msg("record DKG failure on retry rollover")
		}
	}

	switch phaseNow {
	case PhaseDealing:
		return kpr.startDealing(ctx, keyperConfigIndex, retryInt64)
	case PhaseAccusing:
		return kpr.startAccusing(ctx, keyperConfigIndex, retryInt64)
	case PhaseApologizing:
		return kpr.startApologizing(ctx, keyperConfigIndex, retryInt64)
	case PhaseFinalizing:
		return kpr.startFinalizing(ctx, keyperConfigIndex, retryInt64)
	case PhaseNone:
		// Finalizing-ended boundary within the same retry: write a failure
		// row if the DKG did not succeed.
		if phasePrev == PhaseFinalizing && retry == retryPrev {
			return kpr.recordDKGFailureIfNotSucceeded(ctx, keyperConfigIndex, retryInt64)
		}
		return nil
	}
	return nil
}

// startDealing runs at the Dealing-phase boundary. If we have already stored
// our own commitment row for `(k, r)`, we treat the message as already sent
// and skip the on-chain transaction. Otherwise we drive puredkg through
// `StartPhase1Dealing`, persist our own dealing locally, and submit the
// transaction.
func (kpr *Keyper) startDealing(ctx context.Context, keyperConfigIndex, retryCounter int64) error {
	return kpr.dbpool.BeginFunc(ctx, func(tx pgx.Tx) error {
		pure, keypers, ownIndex, isMember, err := kpr.buildPureDKG(ctx, tx, keyperConfigIndex, retryCounter)
		if err != nil {
			return err
		}
		if !isMember {
			return nil
		}

		queries := corekeyperdb.New(tx)
		commitments, err := queries.GetDKGPolyCommitments(ctx, corekeyperdb.GetDKGPolyCommitmentsParams{
			KeyperConfigIndex: keyperConfigIndex,
			RetryCounter:      retryCounter,
		})
		if err != nil {
			return errors.Wrap(err, "load own dealing check")
		}
		for _, c := range commitments {
			if uint64(c.KeyperIndex) == ownIndex {
				return nil
			}
		}

		if pure.Phase != puredkg.Off {
			// Already advanced past Off — probably a duplicate trigger from
			// the previous-block comparison after a restart. Nothing more
			// to do.
			return nil
		}

		commitmentMsg, polyEvalMsgs, err := pure.StartPhase1Dealing()
		if err != nil {
			return errors.Wrap(err, "puredkg StartPhase1Dealing")
		}

		// Persist our own commitment.
		commitmentBytes := commitmentMsg.Gammas.Marshal()
		if err := queries.InsertDKGPolyCommitment(ctx, corekeyperdb.InsertDKGPolyCommitmentParams{
			KeyperConfigIndex: keyperConfigIndex,
			RetryCounter:      retryCounter,
			KeyperIndex:       int64(ownIndex),
			Commitment:        commitmentBytes,
		}); err != nil {
			return errors.Wrap(err, "store own poly commitment")
		}

		// Encrypt one eval per other keyper, in receiver-index order.
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
			ciphertext, err := kpr.encryptPolyEvalFor(ctx, queries, recvAddr, evalMsg.Eval)
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

		opts, err := kpr.makeTransactOpts(ctx)
		if err != nil {
			return err
		}
		tx2, err := kpr.chainSyncClient.DKGContract.SubmitDealing(
			opts,
			uint64(keyperConfigIndex),
			uint64(retryCounter),
			ownIndex,
			commitmentBytes,
			encryptedEvals,
		)
		if err != nil {
			return errors.Wrap(err, "submit dealing transaction")
		}
		log.Info().
			Int64("keyper-config-index", keyperConfigIndex).
			Int64("retry-counter", retryCounter).
			Uint64("keyper-index", ownIndex).
			Str("tx-hash", tx2.Hash().Hex()).
			Msg("submitted DKG dealing")
		return nil
	})
}

// startAccusing runs at the Accusing-phase boundary. It calls
// `StartPhase2Accusing` to produce the list of accusations and submits a
// single transaction if any are needed and we haven't already accused.
func (kpr *Keyper) startAccusing(ctx context.Context, keyperConfigIndex, retryCounter int64) error {
	return kpr.dbpool.BeginFunc(ctx, func(tx pgx.Tx) error {
		pure, _, ownIndex, isMember, err := kpr.buildPureDKG(ctx, tx, keyperConfigIndex, retryCounter)
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
			// Most often after restart: we never advanced past Off because
			// our polynomial is lost. Skip submitting; the next retry will
			// generate a fresh one.
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

		opts, err := kpr.makeTransactOpts(ctx)
		if err != nil {
			return err
		}
		tx2, err := kpr.chainSyncClient.DKGContract.SubmitAccusation(
			opts,
			uint64(keyperConfigIndex),
			uint64(retryCounter),
			ownIndex,
			accusedIndices,
		)
		if err != nil {
			return errors.Wrap(err, "submit accusation transaction")
		}
		log.Info().
			Int64("keyper-config-index", keyperConfigIndex).
			Int64("retry-counter", retryCounter).
			Int("count", len(accusedIndices)).
			Str("tx-hash", tx2.Hash().Hex()).
			Msg("submitted DKG accusations")
		return nil
	})
}

// startApologizing runs at the Apologizing-phase boundary, submitting an
// apology transaction iff someone accused us and we still hold our polynomial.
func (kpr *Keyper) startApologizing(ctx context.Context, keyperConfigIndex, retryCounter int64) error {
	return kpr.dbpool.BeginFunc(ctx, func(tx pgx.Tx) error {
		pure, _, ownIndex, isMember, err := kpr.buildPureDKG(ctx, tx, keyperConfigIndex, retryCounter)
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

		opts, err := kpr.makeTransactOpts(ctx)
		if err != nil {
			return err
		}
		tx2, err := kpr.chainSyncClient.DKGContract.SubmitApology(
			opts,
			uint64(keyperConfigIndex),
			uint64(retryCounter),
			ownIndex,
			accuserIndices,
			polyEvalData,
		)
		if err != nil {
			return errors.Wrap(err, "submit apology transaction")
		}
		log.Info().
			Int64("keyper-config-index", keyperConfigIndex).
			Int64("retry-counter", retryCounter).
			Int("count", len(accuserIndices)).
			Str("tx-hash", tx2.Hash().Hex()).
			Msg("submitted DKG apologies")
		return nil
	})
}

// startFinalizing runs at the Finalizing-phase boundary, casting the success
// vote with the locally-computed eon public key. If puredkg's local state was
// not advanced past Apologizing (e.g. we restarted in the middle of the DKG
// and dropped our polynomial) we cannot finalize and skip submitting.
func (kpr *Keyper) startFinalizing(ctx context.Context, keyperConfigIndex, retryCounter int64) error {
	return kpr.dbpool.BeginFunc(ctx, func(tx pgx.Tx) error {
		pure, _, ownIndex, isMember, err := kpr.buildPureDKG(ctx, tx, keyperConfigIndex, retryCounter)
		if err != nil {
			return err
		}
		if !isMember {
			return nil
		}

		queries := corekeyperdb.New(tx)
		alreadyVoted, err := kpr.hasVotedOnChain(ctx, uint64(keyperConfigIndex), uint64(retryCounter))
		if err != nil {
			return errors.Wrap(err, "check on-chain hasVoted")
		}
		if alreadyVoted {
			return nil
		}

		if pure.Phase != puredkg.Apologizing {
			log.Debug().
				Int64("keyper-config-index", keyperConfigIndex).
				Int64("retry-counter", retryCounter).
				Str("phase", pure.Phase.String()).
				Msg("skipping success vote: puredkg phase not Apologizing")
			return nil
		}

		pure.Finalize()
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
		// arrives. Insert is idempotent via a check.
		pureBytes, err := shdb.EncodePureDKGResult(&result)
		if err != nil {
			return errors.Wrap(err, "encode pure DKG result")
		}
		exists, err := queries.ExistsDKGResultSuccess(ctx, keyperConfigIndex)
		if err != nil {
			return errors.Wrap(err, "check existing dkg_result")
		}
		if !exists {
			if err := queries.InsertDKGResult(ctx, corekeyperdb.InsertDKGResultParams{
				Eon:        keyperConfigIndex,
				Success:    true,
				Error:      sql.NullString{},
				PureResult: pureBytes,
			}); err != nil {
				return errors.Wrap(err, "insert dkg_result success row")
			}
		}

		eonPubKeyBytes, err := result.PublicKey.GobEncode()
		if err != nil {
			return errors.Wrap(err, "encode eon public key")
		}

		opts, err := kpr.makeTransactOpts(ctx)
		if err != nil {
			return err
		}
		tx2, err := kpr.chainSyncClient.DKGContract.SubmitSuccessVote(
			opts,
			uint64(keyperConfigIndex),
			uint64(retryCounter),
			ownIndex,
			eonPubKeyBytes,
		)
		if err != nil {
			return errors.Wrap(err, "submit success vote transaction")
		}
		log.Info().
			Int64("keyper-config-index", keyperConfigIndex).
			Int64("retry-counter", retryCounter).
			Str("tx-hash", tx2.Hash().Hex()).
			Msg("submitted DKG success vote")
		return nil
	})
}

// recordDKGFailureIfNotSucceeded inserts a failure row in `dkg_result` if no
// row exists yet for the eon. Calls are idempotent: a pre-existing success
// row is preserved, and a pre-existing failure row is left alone.
func (kpr *Keyper) recordDKGFailureIfNotSucceeded(ctx context.Context, keyperConfigIndex, retryCounter int64) error {
	queries := corekeyperdb.New(kpr.dbpool)
	exists, err := queries.ExistsDKGResultSuccess(ctx, keyperConfigIndex)
	if err != nil {
		return errors.Wrap(err, "check existing dkg_result")
	}
	if exists {
		return nil
	}
	// `GetDKGResult` returns pgx.ErrNoRows if no row at all is present.
	_, err = queries.GetDKGResult(ctx, keyperConfigIndex)
	if err == nil {
		// A failure row already exists; leave it.
		return nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return errors.Wrap(err, "look up existing dkg_result")
	}
	log.Warn().
		Int64("keyper-config-index", keyperConfigIndex).
		Int64("retry-counter", retryCounter).
		Msg("recording DKG failure: Finalizing phase elapsed without on-chain success")
	return queries.InsertDKGResult(ctx, corekeyperdb.InsertDKGResultParams{
		Eon:        keyperConfigIndex,
		Success:    false,
		Error:      sql.NullString{String: "Finalizing phase ended without on-chain DKG success", Valid: true},
		PureResult: nil,
	})
}

// hasVotedOnChain queries the DKG contract directly for whether this keyper
// has already cast a success vote. Using on-chain state here avoids tracking
// our own SuccessVoteSubmitted events separately.
func (kpr *Keyper) hasVotedOnChain(ctx context.Context, keyperSetIndex, retryCounter uint64) (bool, error) {
	opts := &bind.CallOpts{Context: ctx}
	return kpr.chainSyncClient.DKGContract.HasVoted(opts, keyperSetIndex, retryCounter, kpr.config.GetAddress())
}

// makeTransactOpts produces signed transaction opts for the keyper's signing
// key, with the context attached so caller cancellation propagates.
func (kpr *Keyper) makeTransactOpts(ctx context.Context) (*bind.TransactOpts, error) {
	chainID, err := kpr.chainSyncClient.ChainID(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "get chain id")
	}
	opts, err := bind.NewKeyedTransactorWithChainID(kpr.config.Gnosis.Node.PrivateKey.Key, chainID)
	if err != nil {
		return nil, errors.Wrap(err, "construct signer transaction opts")
	}
	opts.Context = ctx
	return opts, nil
}
