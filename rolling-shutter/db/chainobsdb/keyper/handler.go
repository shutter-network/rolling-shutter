package chainobsdb

import (
	"context"
	"math"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/chainobserver"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/contract"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/shdb"
)

type Handler struct {
	KeyperContract *contract.AddrsSeq
}

func (h *Handler) HandleKeypersConfigsListNewConfigEvent(
	ctx context.Context, tx pgx.Tx, event contract.KeypersConfigsListNewConfig,
) error {
	addrs, err := chainobserver.RetryGetAddrs(ctx, h.KeyperContract, event.KeyperSetIndex)
	if err != nil {
		return err
	}
	log.Info().
		Uint64("block-number", event.Raw.BlockNumber).
		Uint64("keyper-config-index", event.KeyperConfigIndex).
		Uint64("activation-block-number", event.ActivationBlockNumber).
		Msg("handling NewConfig event from keypers config contract")

	if event.ActivationBlockNumber > math.MaxInt64 {
		return errors.Errorf(
			"activation block number %d from config contract would overflow int64",
			event.ActivationBlockNumber)
	}
	db := New(tx)
	err = db.InsertKeyperSet(ctx, InsertKeyperSetParams{
		KeyperConfigIndex:     int64(event.KeyperConfigIndex),
		ActivationBlockNumber: int64(event.ActivationBlockNumber),
		Keypers:               shdb.EncodeAddresses(addrs),
		Threshold:             int32(event.Threshold),
	})
	if err != nil {
		return errors.Wrapf(err, "failed to insert keyper set into db")
	}
	return nil
}
