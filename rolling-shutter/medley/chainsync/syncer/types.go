package syncer

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/chainsync/chainsegment"
)

var ErrCritical = errors.New("critical error, signalling shutdown")

type QueryContext struct {
	// a previously applied chainsegment that has to be
	// removed from the state first
	Remove *chainsegment.ChainSegment
	// the chainsegment the passed in events are part of
	Update *chainsegment.ChainSegment
}

type ChainUpdateHandler interface {
	Handle(ctx context.Context, qCtx QueryContext) error
}

type ContractEventHandler interface {
	Topic() common.Hash
	Address() common.Address

	Parse(log types.Log) (any, bool, error)
	Accept(ctx context.Context, h types.Header, ev any) (bool, error)
	Handle(ctx context.Context, qCtx QueryContext, events []any) error
}

// IContractEventHandler is the generic interface
// that should be implemented.
type IContractEventHandler[T any] interface {
	Address() common.Address
	Event() string
	ABI() abi.ABI

	Accept(context.Context, types.Header, T) (bool, error)
	Handle(context.Context, QueryContext, []T) error
}

// WrapHandler wraps the generic implementation into
// a dynamically typed handler complying to the
// `ContractEventHandler` interface.
func WrapHandler[T any](h IContractEventHandler[T]) (ContractEventHandler, error) {
	var t T
	if reflect.TypeOf(t).Kind() == reflect.Pointer {
		return nil, fmt.Errorf("Handler must not receive pointer values for the event types.")
	}
	return contractEventHandler[T]{
		h: h,
	}, nil
}
