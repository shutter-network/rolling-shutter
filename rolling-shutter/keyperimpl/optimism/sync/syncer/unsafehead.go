package syncer

import (
	"context"
	"errors"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyperimpl/optimism/sync/client"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyperimpl/optimism/sync/event"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/service"
)

type UnsafeHeadSyncer struct {
	Client  client.Client
	Log     log.Logger
	Handler event.BlockHandler

	newLatestHeadCh chan *types.Header
}

func (s *UnsafeHeadSyncer) Start(ctx context.Context, runner service.Runner) error {
	if s.Handler == nil {
		return errors.New("no handler registered")
	}
	s.newLatestHeadCh = make(chan *types.Header, 1)

	subs, err := s.Client.SubscribeNewHead(ctx, s.newLatestHeadCh)
	// FIXME: what to do on subs.Error()
	if err != nil {
		return err
	}
	runner.Defer(subs.Unsubscribe)
	runner.Defer(func() {
		close(s.newLatestHeadCh)
	})
	runner.Go(func() error {
		return s.watchLatestUnsafeHead(ctx)
	})
	return nil
}

func (s *UnsafeHeadSyncer) watchLatestUnsafeHead(ctx context.Context) error {
	for {
		select {
		case newHeader, ok := <-s.newLatestHeadCh:
			if !ok {
				return nil
			}
			ev := &event.LatestBlock{
				Number:    newHeader.Number,
				BlockHash: newHeader.Hash(),
			}
			err := s.Handler(ctx, ev)
			if err != nil {
				s.Log.Error(
					"handler for `NewLatestBlock` errored",
					"error",
					err.Error(),
				)
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
