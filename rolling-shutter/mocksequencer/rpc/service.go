package rpc

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
	txtypes "github.com/shutter-network/txtypes/types"
)

type Sequencer interface {
	RLock()
	RUnlock()
	Active() bool

	SetCollator(common.Address, uint64)
	BatchIndex(context.Context) (uint64, error)
	SubmitBatch(context.Context, *txtypes.Transaction) (string, error)
}

type RPCService interface {
	Name() string
	InjectProcessor(Sequencer)
}
