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
	Client  client.EthereumClient
	Log     log.Logger
	Handler event.BlockHandler
	// Handler to be manually triggered
	// to handle their handler function
	// before the own Handler is called:
	SyncedHandler []ManualFilterHandler

	newLatestHeadCh chan *types.Header
}

func (s *UnsafeHeadSyncer) Start(ctx context.Context, runner service.Runner) error {
	s.Log.Info("unsafe head syncer started")
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

func parseLogs() error {
	return nil
}

func (s *UnsafeHeadSyncer) watchLatestUnsafeHead(ctx context.Context) error {
	for {
		select {
		case newHeader, ok := <-s.newLatestHeadCh:
			if !ok {
				return nil
			}
			// TODO: check bloom filter for topic of all
			// synced handlers and only call them if
			// the bloomfilter retrieves something.

			blockNum := number.BigToBlockNumber(newHeader.Number)
			for _, h := range s.SyncedHandler {
				// NOTE: this has to be blocking!
				// So whenever this returns, it is expected
				// that the handlers Handle function
				// has been called and it returned.
				err := h.QueryAndHandle(ctx, blockNum.Uint64())
				if err != nil {
					// XXX: return or log?
					// return err
					s.Log.Error("synced handler call errored, skipping", "error", err)
				}
			}
			ev := &event.LatestBlock{
				Number:    blockNum,
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
