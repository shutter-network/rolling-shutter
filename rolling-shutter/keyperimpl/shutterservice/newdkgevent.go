package shutterservice

import (
	"context"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	obskeyper "github.com/shutter-network/rolling-shutter/rolling-shutter/chainobserver/db/keyper"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/dkg"
	corekeyperdb "github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/database"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley"
	syncevent "github.com/shutter-network/rolling-shutter/rolling-shutter/medley/chainsync/event"
)

// processNewDKGEvent stores a DKG Contract event in the appropriate per-type
// table and writes a `dkg_result` row when the event is a success notification.
// Events for a `keyper_set_index` whose DKG already succeeded locally are
// silently ignored so re-emitted or retried-and-superseded messages do not
// pollute the message tables.
func (kpr *Keyper) processNewDKGEvent(ctx context.Context, ev syncevent.DKGEvent) error {
	keyperSetIndex, retryCounter := dkgEventKeys(ev)
	keyperSetIndexInt, err := medley.Uint64ToInt64Safe(keyperSetIndex)
	if err != nil {
		return errors.Wrap(err, "convert keyper set index")
	}
	retryCounterInt, err := medley.Uint64ToInt64Safe(retryCounter)
	if err != nil {
		return errors.Wrap(err, "convert retry counter")
	}

	return kpr.dbpool.BeginFunc(ctx, func(tx pgx.Tx) error {
		queries := corekeyperdb.New(tx)
		obsQueries := obskeyper.New(tx)

		exists, err := queries.ExistsDKGResultSuccess(ctx, keyperSetIndexInt)
		if err != nil {
			return errors.Wrap(err, "check existing dkg_result success")
		}
		if _, isSuccess := ev.(*syncevent.SuccessEvent); exists && !isSuccess {
			log.Debug().
				Uint64("keyper-set-index", keyperSetIndex).
				Uint64("retry-counter", retryCounter).
				Msg("ignoring DKG event for already-succeeded keyper set")
			return nil
		}

		switch e := ev.(type) {
		case *syncevent.DealingEvent:
			return storeDealing(ctx, queries, obsQueries, e, keyperSetIndexInt, retryCounterInt)
		case *syncevent.AccusationEvent:
			return storeAccusation(ctx, queries, e, keyperSetIndexInt, retryCounterInt)
		case *syncevent.ApologyEvent:
			return storeApology(ctx, queries, e, keyperSetIndexInt, retryCounterInt)
		case *syncevent.SuccessVoteEvent:
			log.Debug().
				Uint64("keyper-set-index", e.KeyperSetIndex).
				Uint64("retry-counter", e.RetryCounter).
				Uint64("voter", e.KeyperIndex).
				Msg("observed DKG success vote")
			return nil
		case *syncevent.SuccessEvent:
			return kpr.dkgMgr.HandleDKGSuccess(ctx, tx, keyperSetIndexInt, retryCounterInt)
		default:
			return errors.Errorf("unknown DKG event type %T", ev)
		}
	})
}

// dkgEventKeys extracts the (keyperSetIndex, retryCounter) pair common to every
// DKG event variant. SuccessEvent has no retry counter and reports zero.
func dkgEventKeys(ev syncevent.DKGEvent) (keyperSetIndex, retryCounter uint64) {
	switch e := ev.(type) {
	case *syncevent.DealingEvent:
		return e.KeyperSetIndex, e.RetryCounter
	case *syncevent.AccusationEvent:
		return e.KeyperSetIndex, e.RetryCounter
	case *syncevent.ApologyEvent:
		return e.KeyperSetIndex, e.RetryCounter
	case *syncevent.SuccessVoteEvent:
		return e.KeyperSetIndex, e.RetryCounter
	case *syncevent.SuccessEvent:
		return e.KeyperSetIndex, e.RetryCounter
	}
	return 0, 0
}

func storeDealing(
	ctx context.Context,
	queries *corekeyperdb.Queries,
	obsQueries *obskeyper.Queries,
	ev *syncevent.DealingEvent,
	keyperSetIndex, retryCounter int64,
) error {
	senderIndex, err := medley.Uint64ToInt64Safe(ev.KeyperIndex)
	if err != nil {
		return errors.Wrap(err, "convert sender index")
	}

	if err := queries.InsertDKGPolyCommitment(ctx, corekeyperdb.InsertDKGPolyCommitmentParams{
		KeyperSetIndex: keyperSetIndex,
		RetryCounter:      retryCounter,
		KeyperIndex:       senderIndex,
		Commitment:        ev.Commitment,
	}); err != nil {
		return errors.Wrap(err, "insert dkg_poly_commitment")
	}

	evals := ev.PolyEvals

	keyperSet, err := obsQueries.GetKeyperSetByKeyperConfigIndex(ctx, keyperSetIndex)
	if err != nil {
		return errors.Wrapf(err, "load keyper set %d for polyEval split", keyperSetIndex)
	}
	n := uint64(len(keyperSet.Keypers))
	receivers := dkg.ReceiverIndicesForSender(n, ev.KeyperIndex)
	if uint64(len(evals)) != uint64(len(receivers)) {
		return errors.Errorf(
			"polyEval count %d does not match expected receiver count %d for keyper set %d",
			len(evals), len(receivers), keyperSetIndex,
		)
	}

	for i, encryptedEval := range evals {
		receiverIndex, err := medley.Uint64ToInt64Safe(receivers[i])
		if err != nil {
			return errors.Wrap(err, "convert receiver index")
		}
		if err := queries.InsertDKGPolyEval(ctx, corekeyperdb.InsertDKGPolyEvalParams{
			KeyperSetIndex: keyperSetIndex,
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
	ev *syncevent.AccusationEvent,
	keyperSetIndex, retryCounter int64,
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
			KeyperSetIndex: keyperSetIndex,
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
	ev *syncevent.ApologyEvent,
	keyperSetIndex, retryCounter int64,
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
			KeyperSetIndex: keyperSetIndex,
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

