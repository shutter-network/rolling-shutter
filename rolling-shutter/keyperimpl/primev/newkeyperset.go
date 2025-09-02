package primev

import (
	"context"
	"math"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	obskeyper "github.com/shutter-network/rolling-shutter/rolling-shutter/chainobserver/db/keyper"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley"
	syncevent "github.com/shutter-network/rolling-shutter/rolling-shutter/medley/chainsync/event"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/shdb"
)

func (k *Keyper) processNewKeyperSet(ctx context.Context, ev *syncevent.KeyperSet) error {
	isMember := false
	for _, m := range ev.Members {
		if m.Cmp(k.config.GetAddress()) == 0 {
			isMember = true
			break
		}
	}

	log.Info().
		Uint64("activation-block", ev.ActivationBlock).
		Uint64("eon", ev.Eon).
		Int("num-members", len(ev.Members)).
		Uint64("threshold", ev.Threshold).
		Bool("is-member", isMember).
		Msg("new keyper set added")

	return k.dbpool.BeginFunc(ctx, func(tx pgx.Tx) error {
		obskeyperdb := obskeyper.New(tx)

		keyperConfigIndex, err := medley.Uint64ToInt64Safe(ev.Eon)
		if err != nil {
			return errors.Wrap(err, ErrParseKeyperSet.Error())
		}
		activationBlockNumber, err := medley.Uint64ToInt64Safe(ev.ActivationBlock)
		if err != nil {
			return errors.Wrap(err, ErrParseKeyperSet.Error())
		}
		threshold, err := medley.Uint64ToInt64Safe(ev.Threshold)
		if err != nil {
			return errors.Wrap(err, ErrParseKeyperSet.Error())
		}

		if threshold > math.MaxInt32 {
			return errors.Errorf("threshold %d overflows int32", threshold)
		}

		return obskeyperdb.InsertKeyperSet(ctx, obskeyper.InsertKeyperSetParams{
			KeyperConfigIndex:     keyperConfigIndex,
			ActivationBlockNumber: activationBlockNumber,
			Keypers:               shdb.EncodeAddresses(ev.Members),
			Threshold:             int32(threshold),
		})
	})
}
