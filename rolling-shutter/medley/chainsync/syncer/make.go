package syncer

import (
	"context"
	"reflect"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/pkg/errors"
)

// TODO: errors for unpacklog
var todoError = errors.New("todo")

func NewBoundContract(address common.Address, backend bind.ContractBackend, metadata *bind.MetaData) (*bind.BoundContract, error) {
	parsed, err := metadata.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, backend, backend, backend), nil
}

func topics(handler []ContractEventHandler) ([][]common.Hash, error) {
	var query [][]any
	for _, h := range handler {
		// Append the event selector to the query parameters and construct the topic set
		query = append([][]any{{h.Topic()}}, query...)
	}
	topics, err := abi.MakeTopics(query...)
	if err != nil {
		return nil, err
	}
	return topics, nil
}

// UnpackLog unpacks a retrieved log into the provided output structure.
func UnpackLog(a abi.ABI, out interface{}, event string, log types.Log) error {
	// Copy of bind.BoundContract.UnpackLog

	// Anonymous events are not supported.
	if len(log.Topics) == 0 {
		//TODO:
		return todoError
	}
	if log.Topics[0] != a.Events[event].ID {
		//TODO:
		return todoError
	}
	if len(log.Data) > 0 {
		if err := a.UnpackIntoInterface(out, event, log.Data); err != nil {
			return err
		}
	}
	var indexed abi.Arguments
	for _, arg := range a.Events[event].Inputs {
		if arg.Indexed {
			indexed = append(indexed, arg)
		}
	}
	return abi.ParseTopics(out, indexed, log.Topics[1:])
}

type contractEventHandler[T any] struct {
	h IContractEventHandler[T]
}

func (gh contractEventHandler[T]) Address() common.Address {
	return gh.h.Address()
}

func (gh contractEventHandler[T]) Topic() common.Hash {
	return gh.h.ABI().Events[gh.h.Event()].ID
}

func (gh contractEventHandler[T]) Parse(log types.Log) (any, bool, error) {
	var event T

	if err := UnpackLog(gh.h.ABI(), &event, gh.h.Event(), log); err != nil {
		return nil, false, err
	}
	// Set the log to the Raw field
	f := reflect.ValueOf(&event).Elem().FieldByName("Raw")
	if f.CanSet() {
		f.Set(reflect.ValueOf(log))
	}
	return event, true, nil

}
func (gh contractEventHandler[T]) Accept(ctx context.Context, h types.Header, ev any) (bool, error) {
	switch t := ev.(type) {
	case T:
		return gh.h.Accept(ctx, h, t)
	default:
		return false, nil
	}
}

func (gh contractEventHandler[T]) Handle(ctx context.Context, qCtx QueryContext, events []any) error {
	tList := []T{}
	for _, ev := range events {
		switch t := ev.(type) {
		case T:
			tList = append(tList, t)
		default:
		}
	}
	if len(tList) == 0 {
		return nil
	}
	return gh.h.Handle(ctx, qCtx, tList)
}
