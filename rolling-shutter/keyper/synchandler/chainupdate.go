package synchandler

import (
	"context"

	"github.com/ethereum/go-ethereum/log"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/chainsync/syncer"
)

var _ syncer.ChainUpdateHandler = &ChainUpdate{}

// TODO: implement the call to the operateshuttermint here?
func NewChainUpdate(log log.Logger) *ChainUpdate {
	return &ChainUpdate{
		log: log,
	}
}

type ChainUpdate struct {
	log log.Logger
}

func (kb *ChainUpdate) Log(msg string, ctx ...any) {
	kb.log.Info(msg, ctx)
}

func (kb *ChainUpdate) Handle(
	ctx context.Context,
	qCtx syncer.QueryContext,
) error {
	return nil
}
