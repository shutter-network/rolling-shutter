package syncer

import (
	"context"
	"reflect"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type contractEventHandler[T any] struct {
	h IContractEventHandler[T]
}

func (gh contractEventHandler[T]) Address() common.Address {
	return gh.h.Address()
}

func (gh contractEventHandler[T]) Topic() common.Hash {
	return gh.h.ABI().Events[gh.h.Event()].ID
}

func (gh contractEventHandler[T]) Parse(logger types.Log) (any, error) {
	var event T

	if err := UnpackLog(gh.h.ABI(), &event, gh.h.Event(), logger); err != nil {
		return nil, err
	}
	// Set the log to the Raw field
	f := reflect.ValueOf(&event).Elem().FieldByName("Raw")
	if f.CanSet() {
		f.Set(reflect.ValueOf(logger))
	}
	return event, nil
}

func (gh contractEventHandler[T]) Accept(ctx context.Context, h types.Header, ev any) (bool, error) {
	switch t := ev.(type) {
	case T:
		return gh.h.Accept(ctx, h, t)
	default:
		return false, nil
	}
}

func (gh contractEventHandler[T]) Handle(ctx context.Context, update ChainUpdateContext, events []any) error {
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
	return gh.h.Handle(ctx, update, tList)
}

func (gh contractEventHandler[T]) Logger() zerolog.Logger {
	return log.With().
		Str("contract-event-handler", gh.h.Event()).
		Str("contract-address", gh.Address().String()).
		Logger()
}
