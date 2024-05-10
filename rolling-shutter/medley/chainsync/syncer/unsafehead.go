package syncer

import (
	"context"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
	"github.com/pkg/errors"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/chainsync/client"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/chainsync/event"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/encodeable/number"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/retry"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/service"
)

type UnsafeHeadSyncer struct {
	Client  client.EthereumClient
	Log     log.Logger
	Handler event.BlockHandler
	// Handler to be manually triggered
	// to handle their handler function
	// before the own Handler is called:
	SyncedHandler      []ManualFilterHandler
	SyncStartBlock     *number.BlockNumber
	FetchActiveAtStart bool

	newLatestHeadCh chan *types.Header

	syncHead      *types.Header
	nextSyncBlock *big.Int
	latestHead    *types.Header
	headerCache   map[uint64]*types.Header
}

func (s *UnsafeHeadSyncer) Start(ctx context.Context, runner service.Runner) error {
	s.Log.Info("unsafe head syncer started")
	if s.Handler == nil {
		return errors.New("no handler registered")
	}

	s.headerCache = map[uint64]*types.Header{}
	s.newLatestHeadCh = make(chan *types.Header, 1)
	_, err := retry.FunctionCall(
		ctx,
		func(ctx context.Context) (bool, error) {
			err := s.fetchInitialHeaders(ctx)
			return err == nil, err
		},
	)
	if err != nil {
		return errors.Wrap(err, "fetch initial latest header and sync start header")
	}
	s.SyncStartBlock = &number.BlockNumber{Int: s.syncHead.Number}
	if s.FetchActiveAtStart {
		for _, h := range s.SyncedHandler {
			err := h.HandleVirtualEvent(ctx, s.SyncStartBlock)
			if err != nil {
				s.Log.Error("synced handler call errored, skipping", "error", err)
			}
		}
	}
	err = s.handle(ctx, s.syncHead, false)
	if err != nil {
		return errors.Wrap(err, "handle initial sync block")
	}
	s.nextSyncBlock = new(big.Int).Add(s.syncHead.Number, big.NewInt(1))

	subs, err := s.Client.SubscribeNewHead(ctx, s.newLatestHeadCh)
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

func (s *UnsafeHeadSyncer) fetchInitialHeaders(ctx context.Context) error {
	latest, err := s.Client.HeaderByNumber(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "fetch latest header")
	}
	s.latestHead = latest
	if s.SyncStartBlock.IsLatest() {
		s.syncHead = latest
		return nil
	}

	start, err := s.Client.HeaderByNumber(ctx, s.SyncStartBlock.Int)
	if err != nil {
		return errors.Wrap(err, "fetch sync start header")
	}
	s.syncHead = start
	return nil
}

func parseLogs() error {
	return nil
}

func (s *UnsafeHeadSyncer) handle(ctx context.Context, newHeader *types.Header, reorg bool) error {
	blockNum := number.BigToBlockNumber(newHeader.Number)
	if reorg {
		prevNum := new(big.Int).Sub(newHeader.Number, big.NewInt(1))
		prevBlockNum := number.BigToBlockNumber(prevNum)
		// Re-emit the previous block, to pre-emptively signal an
		// incoming reorg. Like this a client is able to e.g.
		// rewind changes first before processing the new
		// events of the reorg
		ev := &event.LatestBlock{
			Number:    prevBlockNum,
			BlockHash: newHeader.ParentHash,
			Header:    newHeader,
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
	return nil
}

func (s *UnsafeHeadSyncer) fetchHeader(ctx context.Context, num *big.Int) (*types.Header, error) {
	h, ok := s.headerCache[num.Uint64()]
	if ok {
		return h, nil
	}
	return s.Client.HeaderByNumber(ctx, num)
}

func (s *UnsafeHeadSyncer) reset(ctx context.Context) error {
	// this means the latest head was reset - check wether that
	// concerns the current sync status
	switch s.latestHead.Number.Cmp(s.syncHead.Number) {
	case 1:
		// we didn't catch up to the reorg position, so it's safe to ignore
		// TODO: delete the forward caches
		return nil
	case 0:
		// we already processed the head we re-orged to
		if s.latestHead.Hash().Cmp(s.syncHead.Hash()) == 0 {
			return nil
		}
	}
	// definite reorg
	if err := s.handle(ctx, s.latestHead, true); err != nil {
		return err
	}
	s.syncHead = s.latestHead
	s.nextSyncBlock = new(big.Int).Add(s.latestHead.Number, big.NewInt(1))
	return nil
}

func (s *UnsafeHeadSyncer) sync(ctx context.Context) (syncing bool, delay time.Duration, err error) {
	var newHead *types.Header
	s.Log.Info("syncing chain")

	delta := new(big.Int).Sub(s.latestHead.Number, s.nextSyncBlock)
	switch delta.Cmp(big.NewInt(0)) {
	case 1:
		// positive delta, we are still catching up
		newHead, err = s.fetchHeader(ctx, s.nextSyncBlock)
		syncing = true
		delay = 1 * time.Second
	case 0:
		// next sync is latest head
		// use the latest head
		newHead = s.latestHead
		syncing = false
	case -1:
		if delta.Cmp(big.NewInt(-1)) != 0 {
			// next sync is more than one block further in the future.
			// this could mean a reorg, but this should have been called before
			// and it shouldn't come to this here
			return false, delay, errors.New("unexpected reorg condition in sync")
		}
		// reorgs are handled outside, at the place new latest-head
		// information arrives.
		// next sync is 1 into the future.
		// this means we called sync but are still waiting for the next latest head.
		return false, delay, err
	}
	if err != nil {
		return true, delay, err
	}

	if handleErr := s.handle(ctx, newHead, false); handleErr != nil {
		return true, delay, handleErr
	}
	s.syncHead = newHead
	delete(s.headerCache, newHead.Number.Uint64())
	s.nextSyncBlock = new(big.Int).Add(newHead.Number, big.NewInt(1))

	s.Log.Info("chain sync",
		"synced-head-num", s.syncHead.Number.Uint64(),
		"synced-head-hash", s.syncHead.Hash(),
		"latest-head-num", s.latestHead.Number.Uint64(),
		"latest-head-hash", s.latestHead.Hash(),
	)
	return syncing, delay, err
}

func (s *UnsafeHeadSyncer) watchLatestUnsafeHead(ctx context.Context) error { //nolint: gocyclo
	t := time.NewTimer(0)
	sync := t.C
work:
	for {
		select {
		case <-ctx.Done():
			if sync != nil && !t.Stop() {
				select {
				case <-t.C:
				// TODO: the non-blocking select is here as a
				// precaution against an already emptied channel.
				// This should not be necessary
				default:
				}
			}
			return ctx.Err()
		case <-sync:
			syncing, delay, err := s.sync(ctx)
			if err != nil {
				s.Log.Error("error during unsafe head sync", "error", err)
			}
			if !syncing {
				s.Log.Info("stop syncing from sync")
				sync = nil
				continue work
			}
			t.Reset(delay)
		case newHeader, ok := <-s.newLatestHeadCh:
			if !ok {
				s.Log.Info("latest head stream closed, exiting handler loop")
				return nil
			}
			s.Log.Info("new latest head from l2 ws-stream", "block-number", newHeader.Number.Uint64())
			if newHeader.Number.Cmp(s.latestHead.Number) <= 0 {
				// reorg
				s.Log.Info("new latest head is re-orging",
					"old-block-number", s.latestHead.Number.Uint64(),
					"new-block-number", newHeader.Number.Uint64(),
				)

				s.latestHead = newHeader
				if err := s.reset(ctx); err != nil {
					s.Log.Error("error resetting reorg", "error", err)
				}
				continue work
			}
			s.headerCache[newHeader.Number.Uint64()] = newHeader
			s.latestHead = newHeader

			if sync != nil && !t.Stop() {
				// only if sync is still actively waiting,
				// we need to drain the timer and reset it
				select {
				case <-t.C:
				// TODO: the non-blocking select is here as a
				// precaution against an already emptied channel.
				// This should not be necessary
				default:
				}
			}
			t.Reset(0)
			s.Log.Info("start syncing from latest head stream")
			sync = t.C
		}
	}
}
