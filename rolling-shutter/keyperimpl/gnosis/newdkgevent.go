package gnosis

import (
	"context"
	"database/sql"
	"math/big"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/shutter-network/shutter/shlib/puredkg"
	"github.com/shutter-network/shutter/shlib/shcrypto"

	obskeyper "github.com/shutter-network/rolling-shutter/rolling-shutter/chainobserver/db/keyper"
	corekeyperdb "github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/database"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley"
	syncevent "github.com/shutter-network/rolling-shutter/rolling-shutter/medley/chainsync/event"
)

// processNewDKGEvent stores a DKG Contract event in the appropriate per-type
// table and writes a `dkg_result` row when the event is a success notification.
// Events for a `keyper_config_index` whose DKG already succeeded locally are
// silently ignored so re-emitted or retried-and-superseded messages do not
// pollute the message tables.
func (kpr *Keyper) processNewDKGEvent(ctx context.Context, ev *syncevent.DKGEvent) error {
	keyperConfigIndex, err := medley.Uint64ToInt64Safe(ev.KeyperSetIndex)
	if err != nil {
		return errors.Wrap(err, "convert keyper set index")
	}
	retryCounter, err := medley.Uint64ToInt64Safe(ev.RetryCounter)
	if err != nil {
		return errors.Wrap(err, "convert retry counter")
	}

	return kpr.dbpool.BeginFunc(ctx, func(tx pgx.Tx) error {
		queries := corekeyperdb.New(tx)
		obsQueries := obskeyper.New(tx)

		// Skip events for DKG instances we have already recorded as successful.
		// Once a keyper set has produced a usable eon key, late-arriving
		// dealings/accusations from retries we no longer participate in are not
		// interesting and would just bloat the message tables.
		exists, err := queries.ExistsDKGResultSuccess(ctx, keyperConfigIndex)
		if err != nil {
			return errors.Wrap(err, "check existing dkg_result success")
		}
		if exists && ev.Kind != syncevent.DKGEventKindSuccess {
			log.Debug().
				Uint64("keyper-set-index", ev.KeyperSetIndex).
				Uint64("retry-counter", ev.RetryCounter).
				Int("kind", int(ev.Kind)).
				Msg("ignoring DKG event for already-succeeded keyper set")
			return nil
		}

		switch ev.Kind {
		case syncevent.DKGEventKindDealing:
			if err := storeDealing(ctx, queries, obsQueries, ev, keyperConfigIndex, retryCounter); err != nil {
				return err
			}
			return kpr.applyDealingToInstance(ctx, tx, ev, keyperConfigIndex, retryCounter)
		case syncevent.DKGEventKindAccusation:
			if err := storeAccusation(ctx, queries, ev, keyperConfigIndex, retryCounter); err != nil {
				return err
			}
			return kpr.applyAccusationToInstance(ctx, tx, ev, keyperConfigIndex, retryCounter)
		case syncevent.DKGEventKindApology:
			if err := storeApology(ctx, queries, ev, keyperConfigIndex, retryCounter); err != nil {
				return err
			}
			return kpr.applyApologyToInstance(ctx, tx, ev, keyperConfigIndex, retryCounter)
		case syncevent.DKGEventKindSuccessVote:
			// Success votes are not stored in their own table; the aggregate
			// outcome arrives as a separate DKGSucceeded event.
			log.Debug().
				Uint64("keyper-set-index", ev.KeyperSetIndex).
				Uint64("retry-counter", ev.RetryCounter).
				Uint64("voter", ev.KeyperIndex).
				Msg("observed DKG success vote")
			return nil
		case syncevent.DKGEventKindSuccess:
			return storeDKGSuccess(ctx, queries, ev, keyperConfigIndex)
		default:
			return errors.Errorf("unknown DKG event kind %d", ev.Kind)
		}
	})
}

// applyDealingToInstance forwards an on-chain dealing to our cached puredkg
// state. The DB store has already happened in the same transaction; this
// keeps the in-memory state in sync so the next phase boundary can act on
// up-to-date information.
func (kpr *Keyper) applyDealingToInstance(
	ctx context.Context,
	tx pgx.Tx,
	ev *syncevent.DKGEvent,
	keyperConfigIndex, retryCounter int64,
) error {
	inst, err := kpr.loadOrBuildInstance(ctx, tx, keyperConfigIndex, retryCounter)
	if err != nil || inst == nil {
		return err
	}
	// Skip our own dealing — we already applied it to puredkg via the
	// `StartPhase1Dealing` path.
	if ev.KeyperIndex == inst.ownIndex {
		return nil
	}
	gammas := &shcrypto.Gammas{}
	if err := gammas.Unmarshal(ev.Commitment); err != nil {
		log.Warn().Err(err).
			Uint64("sender", ev.KeyperIndex).
			Msg("ignoring undecodable commitment from chain")
		return nil
	}
	if applyErr := inst.pure.HandlePolyCommitmentMsg(puredkg.PolyCommitmentMsg{
		Eon:    uint64(keyperConfigIndex),
		Sender: ev.KeyperIndex,
		Gammas: gammas,
	}); applyErr != nil {
		log.Debug().Err(applyErr).
			Uint64("sender", ev.KeyperIndex).
			Msg("ignoring duplicate/late commitment")
	}

	evals := ev.PolyEvals
	receivers := ReceiverIndicesForSender(uint64(len(inst.keypers)), ev.KeyperIndex)
	if len(evals) != len(receivers) {
		log.Warn().
			Int("evals", len(evals)).
			Int("receivers", len(receivers)).
			Msg("poly eval blob size mismatch")
		return nil
	}
	for i, recvIdx := range receivers {
		if recvIdx != inst.ownIndex {
			continue
		}
		eval, decErr := kpr.decryptPolyEval(evals[i])
		if decErr != nil {
			log.Debug().Err(decErr).
				Uint64("sender", ev.KeyperIndex).
				Msg("could not decrypt poly eval addressed to me")
			return nil
		}
		if applyErr := inst.pure.HandlePolyEvalMsg(puredkg.PolyEvalMsg{
			Eon:      uint64(keyperConfigIndex),
			Sender:   ev.KeyperIndex,
			Receiver: inst.ownIndex,
			Eval:     eval,
		}); applyErr != nil {
			log.Debug().Err(applyErr).
				Uint64("sender", ev.KeyperIndex).
				Msg("ignoring duplicate/late poly eval")
		}
		break
	}
	return nil
}

func (kpr *Keyper) applyAccusationToInstance(
	ctx context.Context,
	tx pgx.Tx,
	ev *syncevent.DKGEvent,
	keyperConfigIndex, retryCounter int64,
) error {
	inst, err := kpr.loadOrBuildInstance(ctx, tx, keyperConfigIndex, retryCounter)
	if err != nil || inst == nil {
		return err
	}
	for _, accused := range ev.AccusedIndices {
		if err := inst.pure.HandleAccusationMsg(puredkg.AccusationMsg{
			Eon:     uint64(keyperConfigIndex),
			Accuser: ev.KeyperIndex,
			Accused: accused,
		}); err != nil {
			log.Debug().Err(err).
				Uint64("accuser", ev.KeyperIndex).
				Uint64("accused", accused).
				Msg("ignoring duplicate/late accusation")
		}
	}
	return nil
}

func (kpr *Keyper) applyApologyToInstance(
	ctx context.Context,
	tx pgx.Tx,
	ev *syncevent.DKGEvent,
	keyperConfigIndex, retryCounter int64,
) error {
	inst, err := kpr.loadOrBuildInstance(ctx, tx, keyperConfigIndex, retryCounter)
	if err != nil || inst == nil {
		return err
	}
	if len(ev.AccuserIndices) != len(ev.PolyEvalData) {
		return nil
	}
	for i, accuser := range ev.AccuserIndices {
		eval := new(big.Int).SetBytes(ev.PolyEvalData[i])
		if err := inst.pure.HandleApologyMsg(puredkg.ApologyMsg{
			Eon:     uint64(keyperConfigIndex),
			Accuser: accuser,
			Accused: ev.KeyperIndex,
			Eval:    eval,
		}); err != nil {
			log.Debug().Err(err).
				Uint64("accuser", accuser).
				Uint64("apologizer", ev.KeyperIndex).
				Msg("ignoring duplicate/late apology")
		}
	}
	return nil
}

func storeDealing(
	ctx context.Context,
	queries *corekeyperdb.Queries,
	obsQueries *obskeyper.Queries,
	ev *syncevent.DKGEvent,
	keyperConfigIndex, retryCounter int64,
) error {
	senderIndex, err := medley.Uint64ToInt64Safe(ev.KeyperIndex)
	if err != nil {
		return errors.Wrap(err, "convert sender index")
	}

	if err := queries.InsertDKGPolyCommitment(ctx, corekeyperdb.InsertDKGPolyCommitmentParams{
		KeyperConfigIndex: keyperConfigIndex,
		RetryCounter:      retryCounter,
		KeyperIndex:       senderIndex,
		Commitment:        ev.Commitment,
	}); err != nil {
		return errors.Wrap(err, "insert dkg_poly_commitment")
	}

	evals := ev.PolyEvals

	keyperSet, err := obsQueries.GetKeyperSetByKeyperConfigIndex(ctx, keyperConfigIndex)
	if err != nil {
		return errors.Wrapf(err, "load keyper set %d for polyEval split", keyperConfigIndex)
	}
	n := uint64(len(keyperSet.Keypers))
	receivers := ReceiverIndicesForSender(n, ev.KeyperIndex)
	if uint64(len(evals)) != uint64(len(receivers)) {
		return errors.Errorf(
			"polyEval count %d does not match expected receiver count %d for keyper set %d",
			len(evals), len(receivers), keyperConfigIndex,
		)
	}

	for i, encryptedEval := range evals {
		receiverIndex, err := medley.Uint64ToInt64Safe(receivers[i])
		if err != nil {
			return errors.Wrap(err, "convert receiver index")
		}
		if err := queries.InsertDKGPolyEval(ctx, corekeyperdb.InsertDKGPolyEvalParams{
			KeyperConfigIndex: keyperConfigIndex,
			RetryCounter:      retryCounter,
			SenderIndex:       senderIndex,
			ReceiverIndex:     receiverIndex,
			EncryptedEval:     encryptedEval,
		}); err != nil {
			return errors.Wrap(err, "insert dkg_poly_eval")
		}
	}
	return nil
}

func storeAccusation(
	ctx context.Context,
	queries *corekeyperdb.Queries,
	ev *syncevent.DKGEvent,
	keyperConfigIndex, retryCounter int64,
) error {
	accuserIndex, err := medley.Uint64ToInt64Safe(ev.KeyperIndex)
	if err != nil {
		return errors.Wrap(err, "convert accuser index")
	}
	for _, accused := range ev.AccusedIndices {
		accusedIndex, err := medley.Uint64ToInt64Safe(accused)
		if err != nil {
			return errors.Wrap(err, "convert accused index")
		}
		if err := queries.InsertDKGAccusation(ctx, corekeyperdb.InsertDKGAccusationParams{
			KeyperConfigIndex: keyperConfigIndex,
			RetryCounter:      retryCounter,
			AccuserIndex:      accuserIndex,
			AccusedIndex:      accusedIndex,
		}); err != nil {
			return errors.Wrap(err, "insert dkg_accusation")
		}
	}
	return nil
}

func storeApology(
	ctx context.Context,
	queries *corekeyperdb.Queries,
	ev *syncevent.DKGEvent,
	keyperConfigIndex, retryCounter int64,
) error {
	if len(ev.AccuserIndices) != len(ev.PolyEvalData) {
		return errors.Errorf(
			"apology accuser count %d does not match polyEval data count %d",
			len(ev.AccuserIndices), len(ev.PolyEvalData),
		)
	}
	apologizerIndex, err := medley.Uint64ToInt64Safe(ev.KeyperIndex)
	if err != nil {
		return errors.Wrap(err, "convert apologizer index")
	}
	for i, accuser := range ev.AccuserIndices {
		accuserIndex, err := medley.Uint64ToInt64Safe(accuser)
		if err != nil {
			return errors.Wrap(err, "convert accuser index")
		}
		if err := queries.InsertDKGApology(ctx, corekeyperdb.InsertDKGApologyParams{
			KeyperConfigIndex: keyperConfigIndex,
			RetryCounter:      retryCounter,
			ApologizerIndex:   apologizerIndex,
			AccuserIndex:      accuserIndex,
			PolyEval:          ev.PolyEvalData[i],
		}); err != nil {
			return errors.Wrap(err, "insert dkg_apology")
		}
	}
	return nil
}

func storeDKGSuccess(
	ctx context.Context,
	queries *corekeyperdb.Queries,
	ev *syncevent.DKGEvent,
	keyperConfigIndex int64,
) error {
	exists, err := queries.ExistsDKGResultSuccess(ctx, keyperConfigIndex)
	if err != nil {
		return errors.Wrap(err, "check existing dkg_result success")
	}
	if exists {
		return nil
	}
	log.Info().
		Uint64("keyper-set-index", ev.KeyperSetIndex).
		Uint64("retry-counter", ev.RetryCounter).
		Msg("recording DKG success")
	return queries.InsertDKGResult(ctx, corekeyperdb.InsertDKGResultParams{
		Eon:        keyperConfigIndex,
		Success:    true,
		Error:      sql.NullString{},
		PureResult: nil,
	})
}
