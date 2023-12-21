package syncer

import (
	"context"
	"errors"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/log"
	"github.com/shutter-network/shop-contracts/bindings"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyperimpl/optimism/sync/client"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyperimpl/optimism/sync/event"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/service"
)

type ShutterStateSyncer struct {
	Client     client.Client
	Contract   *bindings.KeyperSetManager
	StartBlock *uint64
	Log        log.Logger
	Handler    event.ShutterStateHandler

	pausedCh   chan *bindings.KeyperSetManagerPaused
	unpausedCh chan *bindings.KeyperSetManagerUnpaused
}

func (s *ShutterStateSyncer) GetShutterState(ctx context.Context, opts *bind.CallOpts) (*event.ShutterState, error) {
	if opts == nil {
		opts = &bind.CallOpts{
			Context: ctx,
		}
	}
	isPaused, err := s.Contract.Paused(opts)
	if err != nil {
		return nil, err
	}
	return &event.ShutterState{
		Active: !isPaused,
	}, nil
}

func (s *ShutterStateSyncer) Start(ctx context.Context, runner service.Runner) error {
	if s.Handler == nil {
		return errors.New("no handler registered")
	}
	watchOpts := &bind.WatchOpts{
		Start:   s.StartBlock, // nil means latest
		Context: ctx,
	}
	s.pausedCh = make(chan *bindings.KeyperSetManagerPaused)
	subs, err := s.Contract.WatchPaused(watchOpts, s.pausedCh)
	// FIXME: what to do on subs.Error()
	if err != nil {
		return err
	}
	runner.Defer(subs.Unsubscribe)
	runner.Defer(func() {
		close(s.pausedCh)
	})

	s.unpausedCh = make(chan *bindings.KeyperSetManagerUnpaused)
	subs, err = s.Contract.WatchUnpaused(watchOpts, s.unpausedCh)
	// FIXME: what to do on subs.Error()
	if err != nil {
		return err
	}
	runner.Defer(subs.Unsubscribe)
	runner.Defer(func() {
		close(s.unpausedCh)
	})

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

func (s *ShutterStateSyncer) handle(ctx context.Context, ev *event.ShutterState) bool {
	err := s.Handler(ctx, ev)
	if err != nil {
		s.Log.Error(
			"handler for `NewShutterState` errored",
			"error",
			err.Error(),
		)
		return false
	}
	return true
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
