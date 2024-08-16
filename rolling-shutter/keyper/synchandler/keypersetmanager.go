package synchandler

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/chainsync/syncer"
)

var _ syncer.IContractEventHandler[] = &KeyperSetManager{}

type KeyperSetManager struct {
	Address common.Address
}

func (kb *KeyperSetManager) Topic() common.Hash {
	return common.Hash{}
}

func (kb *KeyperSetManager) Address() common.Address {
	return kb.address
}

func (kb *KeyperSetManager) Parse(log types.Log) (any, bool, error) {
	return struct{}{}, false, nil
}

func (kb *KeyperSetManager) Accept(ctx context.Context, h types.Header, ev any) (bool, error) {
	return false, nil

}
func (kb *KeyperSetManager) Handle(ctx context.Context, qCtx QueryContext, events []any) error {
	return nil
}
