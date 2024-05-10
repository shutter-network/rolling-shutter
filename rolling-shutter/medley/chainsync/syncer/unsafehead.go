package syncer

import (
	"context"
	"errors"
	"math/big"

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
	var currentBlock *big.Int
	for {
		select {
		case newHeader, ok := <-s.newLatestHeadCh:
			if !ok {
				return nil
			}
			blockNum := number.BigToBlockNumber(newHeader.Number)
			if currentBlock != nil {
				switch newHeader.Number.Cmp(currentBlock) {
				case -1, 0:
					prevNum := new(big.Int).Sub(newHeader.Number, big.NewInt(1))
					prevBlockNum := number.BigToBlockNumber(prevNum)
					// Re-emit the previous block, to pre-emptively signal an
					// incoming reorg. Like this a client is able to e.g.
					// rewind changes first before processing the new
					// events of the reorg
					ev := &event.LatestBlock{
						Number:    prevBlockNum,
						BlockHash: newHeader.ParentHash,
					}
					err := s.Handler(ctx, ev)
					if err != nil {
						// XXX: return or log?
						// return err
						s.Log.Error(
							"handler for `NewLatestBlock` errored",
							"error",
							err.Error(),
						)
					}
				case 1:
					// expected
				}
			}

			// TODO: check bloom filter for topic of all
			// synced handlers and only call them if
			// the bloomfilter retrieves something.
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
			currentBlock = newHeader.Number
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
