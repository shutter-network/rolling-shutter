package syncer

import (
	"context"
	"fmt"
	"reflect"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/rs/zerolog"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/chainsync/chainsegment"
)

type ChainUpdateContext struct {
	// a previously applied chainsegment that has to be
	// removed from the state first
	Remove *chainsegment.ChainSegment
	// the chainsegment the passed in events are part of
	Append *chainsegment.ChainSegment
}

type ChainUpdateHandler interface {
	Handle(ctx context.Context, update ChainUpdateContext) error
}

// IContractEventHandler is the generic interface
// that should be implemented.
// This allows more narrowly typed implementations
// on a per-contracts-event basis, while offloading
// the dynamic typing to a single implementation
// (`contractEventHandler[T]`, complying to the
// ContractEventHandler interface).
type IContractEventHandler[T any] interface {
	Address() common.Address
	Event() string
	ABI() abi.ABI

	Accept(context.Context, types.Header, T) (bool, error)
	Handle(context.Context, ChainUpdateContext, []T) error
}

// WrapHandler wraps the generic implementation into
// a dynamically typed handler complying to the
// `ContractEventHandler` interface.
func WrapHandler[T any](h IContractEventHandler[T]) (ContractEventHandler, error) {
	var t T
	if reflect.TypeOf(t).Kind() == reflect.Pointer {
		return nil, fmt.Errorf("Handler must not receive pointer values for the event types.")
		return nil, fmt.Errorf("handler must not receive pointer values for the event types")
	return contractEventHandler[T]{
		h: h,
	}, nil
}

// ContractEventHandler is the dynamically typed
// interface that is accepted by the chainsync.
// Ideally this doesn't have to be implemented,
// but should be result of wrapping the more
// narrowly typed IContractEventHandler implementations.
type ContractEventHandler interface {
	Topic() common.Hash
	Address() common.Address

	Parse(log types.Log) (any, bool, error)
	Accept(ctx context.Context, h types.Header, ev any) (bool, error)
	Handle(ctx context.Context, update ChainUpdateContext, events []any) error
	Logger() zerolog.Logger
}
