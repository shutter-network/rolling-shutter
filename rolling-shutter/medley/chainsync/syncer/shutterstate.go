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
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/service"
)

var _ ManualFilterHandler = &ShutterStateSyncer{}

type ShutterStateSyncer struct {
	Client                  client.EthereumClient
	Contract                *bindings.KeyperSetManager
	StartBlock              *number.BlockNumber
	Log                     log.Logger
	Handler                 event.ShutterStateHandler
	FetchActiveAtStartBlock bool
	DisableEventWatcher     bool

	pausedCh   chan *bindings.KeyperSetManagerPaused
	unpausedCh chan *bindings.KeyperSetManagerUnpaused
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
		select {
		case s.pausedCh <- iterPaused.Event:
		case <-ctx.Done():
			return ctx.Err()
		}
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
		select {
		case s.unpausedCh <- iterUnpaused.Event:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	if err := iterUnpaused.Error(); err != nil {
		return errors.Wrap(err, "filter iterator error")
	}
	return nil
}

func (s *ShutterStateSyncer) Start(ctx context.Context, runner service.Runner) error {
	if s.Handler == nil {
		return errors.New("no handler registered")
	}
	// the latest block still has to be fixed.
	// otherwise we could skip some block events
	// between the initial poll and the subscription.
	if s.StartBlock.IsLatest() {
		latest, err := s.Client.BlockNumber(ctx)
		if err != nil {
			return err
		}
		s.StartBlock.SetUint64(latest)
	}

	opts := &bind.WatchOpts{
		Start:   s.StartBlock.ToUInt64Ptr(),
		Context: ctx,
	}
	s.pausedCh = make(chan *bindings.KeyperSetManagerPaused)
	runner.Defer(func() {
		close(s.pausedCh)
	})
	s.unpausedCh = make(chan *bindings.KeyperSetManagerUnpaused)
	runner.Defer(func() {
		close(s.unpausedCh)
	})

	if !s.DisableEventWatcher {
		subs, err := s.Contract.WatchPaused(opts, s.pausedCh)
		// FIXME: what to do on subs.Error()
		if err != nil {
			return err
		}
		runner.Defer(subs.Unsubscribe)
		subs, err = s.Contract.WatchUnpaused(opts, s.unpausedCh)
		// FIXME: what to do on subs.Error()
		if err != nil {
			return err
		}
		runner.Defer(subs.Unsubscribe)
	}

	runner.Go(func() error {
		return s.watchPaused(ctx)
	})
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

func (s *ShutterStateSyncer) watchPaused(ctx context.Context) error {
	// query the initial state
	// and construct a "virtual"
	// event
	opts := &bind.CallOpts{
		BlockNumber: s.StartBlock.Int,
		Context:     nil,
	}

	if s.FetchActiveAtStartBlock {
		stateAtStartBlock, err := s.GetShutterState(ctx, opts)
		if err != nil {
			// XXX: this will fail everything, do we want that?
			return err
		}
		s.handle(ctx, stateAtStartBlock)
	}
	for {
		select {
		case unpaused, ok := <-s.unpausedCh:
			if !ok {
				return nil
			}
			block := unpaused.Raw.BlockNumber
			ev := &event.ShutterState{
				Active:        true,
				AtBlockNumber: number.NewBlockNumber(&block),
			}
			s.handle(ctx, ev)
		case paused, ok := <-s.pausedCh:
			if !ok {
				return nil
			}
			block := paused.Raw.BlockNumber
			ev := &event.ShutterState{
				Active:        false,
				AtBlockNumber: number.NewBlockNumber(&block),
			}
			s.handle(ctx, ev)
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
