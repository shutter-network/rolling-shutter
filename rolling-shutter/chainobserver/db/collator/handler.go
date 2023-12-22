package database

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
	CollatorContract *contract.AddrsSeq
}

func (h *Handler) PutDB(
	ctx context.Context, tx pgx.Tx, event contract.CollatorConfigsListNewConfig,
) error {
	addrs, err := chainobserver.RetryGetAddrs(ctx, h.CollatorContract, event.CollatorSetIndex)
	if err != nil {
		return err
	}
	log.Info().
		Uint64("block-number", event.Raw.BlockNumber).
		Uint64("collator-config-index", event.CollatorConfigIndex).
		Uint64("activation-block-number", event.ActivationBlockNumber).
		Msg("handling NewConfig event from collator config contract")
	if event.ActivationBlockNumber > math.MaxInt64 {
		return errors.Errorf(
			"activation block number %d from config contract would overflow int64",
			event.ActivationBlockNumber,
		)
	}
	if len(addrs) > 1 {
		return errors.Errorf("got multiple collators from collator addrs set contract: %s", addrs)
	} else if len(addrs) == 1 {
		db := New(tx)
		err := db.InsertChainCollator(ctx, InsertChainCollatorParams{
			ActivationBlockNumber: int64(event.ActivationBlockNumber),
			Collator:              shdb.EncodeAddress(addrs[0]),
		})
		if err != nil {
			return errors.Wrapf(err, "failed to insert collator into db")
		}
	}
	return nil
}
