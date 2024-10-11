package synchandler

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/hashicorp/go-multierror"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/chainsync/syncer"
)

var _ syncer.ChainUpdateHandler = &DecryptionChainUpdateHandler{}

type DecryptionFunction = func(context.Context, *types.Header) error

func NewDecryptionChainUpdateHandler(fn DecryptionFunction) *DecryptionChainUpdateHandler {
	return &DecryptionChainUpdateHandler{
		decrypt: fn,
	}
}

type DecryptionChainUpdateHandler struct {
	decrypt DecryptionFunction
}

func (cu *DecryptionChainUpdateHandler) Handle(
	ctx context.Context,
	update syncer.ChainUpdateContext,
) (result error) {
	// in case of a reorg (non-nil update.Remove segment) we can't roll back any
	// changes, since the keys have been release already publicly.
	if update.Append != nil {
		for _, header := range update.Append.Get() {
			// We can call the decrypt function with all updated headers,
			// even if this was a reorg.
			// This is because the downstream function is expected to keep track of
			// what slots have already been sent out and decide on itself wether
			// to re-release keys.
			err := cu.decrypt(ctx, header)
			if err != nil {
				result = multierror.Append(result,
					fmt.Errorf("failed to decrypt for block %s (num=%d): %w",
						header.Hash().String(),
						header.Number.Uint64(),
						err))
			}
		}
	}
	return nil
}
