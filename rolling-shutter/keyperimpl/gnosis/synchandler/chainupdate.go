package synchandler

import (
	"context"

	"github.com/ethereum/go-ethereum/core/types"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/chainsync/syncer"
)

var _ syncer.ChainUpdateHandler = &DecryptOnhChainUpdateHandler{}

type DecryptionFunction = func(context.Context, *types.Header) error

func NewDecryptOnChainUpdateHandler(fn DecryptionFunction) *DecryptOnhChainUpdateHandler {
	return &DecryptOnhChainUpdateHandler{
		decrypt: fn,
	}
}

type DecryptOnhChainUpdateHandler struct {
	decrypt DecryptionFunction
}

func (cu *DecryptOnhChainUpdateHandler) Handle(
	ctx context.Context,
	update syncer.ChainUpdateContext,
) error {
	if update.Append != nil {
		for _, header := range update.Append.Get() {
			// Call the decrypt function with all updated headers.
			// The downstream function is expected to keep track of
			// what slots have already been sent out.
			// We could also calculate that by comparing the QueryContext.Update
			// with the QueryContext.Remove by blocknumber and only passing
			// in the non-reorged blocks.
			// TODO: do that instead --^
			err := cu.decrypt(ctx, header)
			if err != nil {
				// TODO: log, or return with a multierr?
			}
		}
	}
	return nil
}
