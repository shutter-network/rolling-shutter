package synchandler

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/chainsync/syncer"
)

var _ syncer.ContractEventHandler = &KeyBroadcast{}

type KeyBroadcast struct {
}

func (kb *KeyBroadcast) Topic() common.Hash {
	return common.Hash{}
}

func (kb *KeyBroadcast) Address() common.Address {
	return common.Address{}
}

func (kb *KeyBroadcast) Parse(log types.Log) (any, bool, error) {
	return struct{}{}, false, nil
}

func (kb *KeyBroadcast) Accept(ctx context.Context, h types.Header, ev any) (bool, error) {
	return false, nil

}
func (kb *KeyBroadcast) Handle(ctx context.Context, qCtx QueryContext, events []any) error {
	return nil

}
