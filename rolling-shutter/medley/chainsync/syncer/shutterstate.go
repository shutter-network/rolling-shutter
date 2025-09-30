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

type ShutterStateSyncer struct {
	Client     client.Client
	Contract   *bindings.KeyperSetManager
	StartBlock *number.BlockNumber
	Log        log.Logger
	Handler    event.ShutterStateHandler

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

func (s *ShutterStateSyncer) Start(ctx context.Context, runner service.Runner) error {
	if s.Handler == nil {
		return errors.New("no handler registered")
	}
	watchOpts := &bind.WatchOpts{
		Start:   s.StartBlock.ToUInt64Ptr(), // nil means latest
		Context: ctx,
	}
	s.pausedCh = make(chan *bindings.KeyperSetManagerPaused)
	subs, err := s.Contract.WatchPaused(watchOpts, s.pausedCh)
	if err != nil {
		return err
	}

	s.unpausedCh = make(chan *bindings.KeyperSetManagerUnpaused)
	subsUnpaused, err := s.Contract.WatchUnpaused(watchOpts, s.unpausedCh)
	if err != nil {
		return err
	}

	runner.Go(func() error {
		err := s.watchPaused(ctx, subs.Err(), subsUnpaused.Err())
		if err != nil {
			s.Log.Error("error watching paused", err.Error())
		}
		subs.Unsubscribe()
		subsUnpaused.Unsubscribe()
		return err
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

func (s *ShutterStateSyncer) watchPaused(
	ctx context.Context,
	subsErr <-chan error,
	subsErrUnpaused <-chan error,
) error {
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
		case err := <-subsErr:
			if err != nil {
				s.Log.Error("subscription error for watchPaused", err.Error())
				return err
			}
		case err := <-subsErrUnpaused:
			if err != nil {
				s.Log.Error("subscription error for watchUnpaused", err.Error())
				return err
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
