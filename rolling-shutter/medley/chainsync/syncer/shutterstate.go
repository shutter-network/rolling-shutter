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
	Client              client.EthereumClient
	Contract            *bindings.KeyperSetManager
	StartBlock          *number.BlockNumber
	Log                 log.Logger
	Handler             event.ShutterStateHandler
	DisableEventWatcher bool

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
	watchOpts := &bind.WatchOpts{
		Start:   s.StartBlock.ToUInt64Ptr(), // nil means latest
		Context: ctx,
	}
	s.pausedCh = make(chan *bindings.KeyperSetManagerPaused)
	runner.Defer(func() {
		close(s.pausedCh)
	})
	s.unpausedCh = make(chan *bindings.KeyperSetManagerUnpaused)
	if !s.DisableEventWatcher {
	}
	runner.Defer(func() {
		close(s.unpausedCh)
	})

	if !s.DisableEventWatcher {
		subs, err := s.Contract.WatchPaused(watchOpts, s.pausedCh)
		// FIXME: what to do on subs.Error()
		if err != nil {
			return err
		}
		runner.Defer(subs.Unsubscribe)
		subs, err = s.Contract.WatchUnpaused(watchOpts, s.unpausedCh)
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

func (s *ShutterStateSyncer) pollIsActive(ctx context.Context) (bool, error) {
	callOpts := bind.CallOpts{
		Context: ctx,
	}
	paused, err := s.Contract.Paused(&callOpts)
	return !paused, err
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
	isActive, err := s.pollIsActive(ctx)
	if err != nil {
		// XXX: this will fail everything, do we want that?
		return err
	}
	ev := &event.ShutterState{
		Active: isActive,
	}
	s.handle(ctx, ev)
	for {
		select {
		case _, ok := <-s.unpausedCh:
			if !ok {
				return nil
			}
			if isActive {
				s.Log.Error("state mismatch", "got", "actice", "have", "inactive")
			}
			ev := &event.ShutterState{
				Active: true,
			}
			isActive = ev.Active
			s.handle(ctx, ev)
		case _, ok := <-s.pausedCh:
			if !ok {
				return nil
			}
			if isActive {
				s.Log.Error("state mismatch", "got", "inactive", "have", "active")
			}
			ev := &event.ShutterState{
				Active: false,
			}
			isActive = ev.Active
			s.handle(ctx, ev)
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
