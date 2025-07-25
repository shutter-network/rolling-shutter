package syncer

import (
	"context"
	"errors"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/chainsync/client"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/chainsync/event"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/encodeable/number"
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
	if err != nil {
		return err
	}
	runner.Go(func() error {
		err := s.watchLatestUnsafeHead(ctx, subs.Err())
		if err != nil {
			s.Log.Error("error watching latest unsafe head", err.Error())
		}
		subs.Unsubscribe()
		return err
	})
	return nil
}

func (s *UnsafeHeadSyncer) watchLatestUnsafeHead(ctx context.Context, subsErr <-chan error) error {
	for {
		select {
		case newHeader, ok := <-s.newLatestHeadCh:
			if !ok {
				return nil
			}
			ev := &event.LatestBlock{
				Number:    number.BigToBlockNumber(newHeader.Number),
				BlockHash: newHeader.Hash(),
				Header:    newHeader,
			}
			err := s.Handler(ctx, ev)
			if err != nil {
				s.Log.Error(
					"handler for `NewLatestBlock` errored",
					"error",
					err.Error(),
				)
			}
		case err := <-subsErr:
			if err != nil {
				s.Log.Error("subscription error for watchLatestUnsafeHead", err.Error())
				return err
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
