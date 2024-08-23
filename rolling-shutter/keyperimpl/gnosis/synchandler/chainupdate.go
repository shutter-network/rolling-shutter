package synchandler

import (
	"context"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/chainsync/syncer"
)

var _ syncer.ChainUpdateHandler = &ChainUpdate{}

type DecryptionFunction = func(context.Context, *types.Header) error

func NewChainUpdate(fn DecryptionFunction) *ChainUpdate {
	return &ChainUpdate{
		decrypt: fn,
	}
}

type ChainUpdate struct {
	log     log.Logger
	decrypt DecryptionFunction
}

func (cu *ChainUpdate) Log(msg string, ctx ...any) {
	cu.log.Info(msg, ctx)
}

func (cu *ChainUpdate) Handle(
	ctx context.Context,
	qCtx syncer.QueryContext,
) error {
	// TODO: here there was the sequencerSyncer and validatorSyncer before...
	// Make sure that they are called before this as well!
	// (event handler should be called before the chain-update handler)

	if qCtx.Update != nil {
		for _, header := range qCtx.Update.Get() {
			// Call the decrypt function with all updated headers.
			// The downstream function is expected to keep track of
			// what slots have already been sent out.
			// We could also calculate that by comparing the QueryContext.Update
			// with the QueryContext.Remove by blocknumber and only passing
			// in the non-reorged blocks.
			err := cu.decrypt(ctx, header)
			if err != nil {
				// TODO: log, or return with a multierr?
			}
		}
	}
	return nil
}
