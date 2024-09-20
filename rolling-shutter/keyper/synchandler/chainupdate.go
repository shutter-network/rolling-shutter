package synchandler

import (
	"context"

	"github.com/ethereum/go-ethereum/log"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/chainsync/syncer"
)

var _ syncer.ChainUpdateHandler = &ChainUpdate{}

// TODO: implement the call to the operateshuttermint here?
func NewChainUpdate(log log.Logger) *ChainUpdate {
	return &ChainUpdate{}
}

type ChainUpdate struct {
}

func (kb *ChainUpdate) Handle(
	ctx context.Context,
	update syncer.ChainUpdateContext,
) error {
	return nil
}
