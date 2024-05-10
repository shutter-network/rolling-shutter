package syncer

import (
	"context"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/log"
	"github.com/pkg/errors"
	"github.com/shutter-network/shop-contracts/bindings"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/chainsync/client"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/chainsync/event"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/encodeable/number"
)

var _ ManualFilterHandler = &ShutterStateSyncer{}

type ShutterStateSyncer struct {
	Client   client.EthereumClient
	Contract *bindings.KeyperSetManager
	Log      log.Logger
	Handler  event.ShutterStateHandler
}

func (s *ShutterStateSyncer) GetShutterState(ctx context.Context, opts *bind.CallOpts) (*event.ShutterState, error) {
	opts, _, err := fixCallOpts(ctx, s.Client, opts)
	if err != nil {
		return nil, err
	}
	isPaused, err := s.Contract.Paused(opts)
	if err != nil {
		return nil, errors.Wrap(err, "query paused state")
	}
	return &event.ShutterState{
		Active:        !isPaused,
		AtBlockNumber: number.BigToBlockNumber(opts.BlockNumber),
	}, nil
}

func (s *ShutterStateSyncer) QueryAndHandle(ctx context.Context, block uint64) error {
	opts := &bind.FilterOpts{
		Start:   block,
		End:     &block,
		Context: ctx,
	}
	iterPaused, err := s.Contract.FilterPaused(opts)
	if err != nil {
		return err
	}
	defer iterPaused.Close()

	for iterPaused.Next() {
		block := iterPaused.Event.Raw.BlockNumber
		ev := &event.ShutterState{
			Active:        false,
			AtBlockNumber: number.NewBlockNumber(&block),
		}
		s.handle(ctx, ev)
	}
	if err := iterPaused.Error(); err != nil {
		return errors.Wrap(err, "filter iterator error")
	}

	iterUnpaused, err := s.Contract.FilterUnpaused(opts)
	if err != nil {
		return err
	}
	defer iterUnpaused.Close()

	for iterUnpaused.Next() {
		block := iterUnpaused.Event.Raw.BlockNumber
		ev := &event.ShutterState{
			Active:        true,
			AtBlockNumber: number.NewBlockNumber(&block),
		}
		s.handle(ctx, ev)
	}
	if err := iterUnpaused.Error(); err != nil {
		return errors.Wrap(err, "filter iterator error")
	}
	return nil
}

func (s *ShutterStateSyncer) HandleVirtualEvent(ctx context.Context, block *number.BlockNumber) error {
	// query the initial state and re-construct a "virtual" event from the contract state
	opts := &bind.CallOpts{
		BlockNumber: block.Int,
		Context:     ctx,
	}
	stateAtBlock, err := s.GetShutterState(ctx, opts)
	if err != nil {
		return err
	}
	s.handle(ctx, stateAtBlock)
	return nil
}

func (s *ShutterStateSyncer) handle(ctx context.Context, ev *event.ShutterState) {
	err := s.Handler(ctx, ev)
	if err != nil {
		s.Log.Error(
			"handler for `NewShutterState` errored",
			"error",
			err.Error(),
		)
	}
}
