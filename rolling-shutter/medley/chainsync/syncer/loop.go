package syncer

import (
	"context"
	"errors"
	"fmt"
	"math/big"

	"github.com/rs/zerolog/log"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/chainsync/chainsegment"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/errs"
)

var errUint64Overflow = errors.New("uint64 overflow in conversion from math.Big")

func (f *Fetcher) triggerHandlerProcessing() {
	// nonblocking setter of updates,
	// in case this has already been triggered
	select {
	case f.processingTrig <- struct{}{}:
	default:
	}
}

// success will be True, when we successfully applied the updated chain segment
// to the old chain. If there remains a gap, this has to return false.
func (f *Fetcher) handlerSync(ctx context.Context) (success bool, err error) {
	var syncedChain, removedSegment, updatedSegment *chainsegment.ChainSegment

	syncedChain, err = f.chainCache.Get(ctx)
	if f.chainUpdate == nil {
		// nothing to update
		success = true
		return success, err
	}
	if errors.Is(err, ErrEmpy) {
		// no chain-cache present yet,
		// just set the chain-update without
		// checking for reorgs
		// FIXME: here we could incorporate a starting block
		// option, fetch the starting block, and set this in
		// the cache and sync from there.
		removedSegment = nil
		updatedSegment = f.chainUpdate
		log.Trace().Msg("internal chain cache empty, setting updated chain segment")
	} else if err != nil {
		//TODO: what to do on db error?
		success = false
		return success, err
	} else {
		if new(big.Int).Add(syncedChain.Latest().Number, big.NewInt(1)).
			Cmp(f.chainUpdate.Earliest().Number) == -1 {
			//FIXME: overflow
			diff := new(big.Int).Sub(f.chainUpdate.Earliest().Number, syncedChain.Latest().Number).Uint64()
			queryBlocks := MaxRequestBlockRange
			// cap the extend range at the diff to the update to not overshoot
			if diff < uint64(queryBlocks) {

			diffBig := new(big.Int).Sub(f.chainUpdate.Earliest().Number, syncedChain.Latest().Number)
			if !diffBig.IsUint64() {
				success = false
				return success, fmt.Errorf("chain-update difference too big: %w", errUint64Overflow)
			}
			diff := diffBig.Uint64()

			// we are not synced to the chain-update
			// so first construct an update to the right of the synced chain
			log.Trace().
				Uint64("synced-latest-blocknum", syncedChain.Latest().Number.Uint64()).
				Uint64("update-earliest-blocknum", f.chainUpdate.Earliest().Number.Uint64()).
				Int("num-query-blocks", queryBlocks).
				Msg("chain update ahead of synced chain, fetching gap blocks")
			updatedSegment, err = syncedChain.NewSegmentRight(ctx, f.ethClient, queryBlocks)
			if errors.Is(err, chainsegment.ErrReorg) {
				// this means we reorged the old chain segment.
				f.chainUpdate.ExtendLeft(ctx, f.ethClient, queryBlocks)
				f.chainUpdate.ExtendLeft(ctx, f.client, queryBlocks)
				err = nil
				f.chainUpdate, err = f.chainUpdate.ExtendLeft(ctx, f.ethClient, queryBlocks)
				if err != nil {
					err = fmt.Errorf("error while querying older blocks from reorg update: %w", err)
				}
				success = false
				return success, err
			}
			removedSegment = nil
			success = false
			result, errr := syncedChain.UpdateLatest(ctx, f.ethClient, f.chainUpdate)
			result, errr := syncedChain.UpdateLatest(ctx, f.client, f.chainUpdate)
			if errr != nil {
				// TODO: for ErrUpdateBlockTooFarInPast this should shut down the
				// client? Since we can't really play back a reorg that reaches too far in the past.
				// Now as long as the chain-cache eviction policy is not agressive (easily doable)
				log.Error().Err(err).Msg("error updating chain")
				err = updateErr
				// Now as long as the chain-cache eviction policy is not aggressive (easily doable)
			removedSegment = result.RemovedSegment
			updatedSegment = result.UpdatedSegment
			// we will process the whole segment of the chain update
			f.chainUpdate = nil
			success = true
		}
	}

	update := ChainUpdateContext{
		Remove: removedSegment,
		Append: updatedSegment,
	}
	if update.Append == nil {
		return success, err
	}

	// blocking call, until all handlers are done processing the
	// new chain segment
	err = f.FetchAndHandle(ctx, update)
	if err != nil {
		return false, err
	}
	err = f.chainCache.Update(ctx, update)
	if err != nil {
		return false, err
	}
	return success, err
}

func (f *Fetcher) loop(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case newHeader, ok := <-f.inChan:
				log.Debug().Msg("latest head stream closed, exiting handler loop")
				f.log.Info("latest head stream closed, exiting handler loop")
				return nil
			log.Debug().Uint64("block-number", newHeader.Number.Uint64()).Msg("new latest head from l2 ws-stream")
			f.log.Debug("new latest head from l2 ws-stream", "block-number", newHeader.Number.Uint64())
			newSegment := chainsegment.NewChainSegment(newHeader)
			if f.chainUpdate != nil {
				// apply the updates to the chain-update buffer that hasn't been processed
				result, err := f.chainUpdate.Copy().UpdateLatest(ctx, f.ethClient, newSegment)
				result, err := f.chainUpdate.Copy().UpdateLatest(ctx, f.client, newSegment)
				fullUpdated := result.FullSegment
				removed := result.RemovedSegment
				if err != nil {
					if errors.Is(err, chainsegment.ErrUpdateBlockTooFarInPast) {
						// reorg beyond the chain-update segment, just set the new header
						removed = f.chainUpdate
						fullUpdated = newSegment
					}
					if errors.Is(err, errs.ErrCritical) {
						return err
					log.Error().Err(err).Msg("error updating chain segment")
					f.log.Error("error updating chain segment", "error", err)
				}
					log.Info().Uint64("block-number", newHeader.Number.Uint64()).Msg("received a new reorg block")
					f.log.Info("received a new reorg block", "block-number", newHeader.Number.Uint64())
				}
				f.chainUpdate = fullUpdated
			} else {
				f.chainUpdate = newSegment
			}
			f.triggerHandlerProcessing()

			f.log.Trace("fetcher loop: received handler sync trigger")
			success, err := f.handlerSync(ctx)
			if err != nil {
				if errors.Is(err, errs.ErrCritical) {
					return err
				log.Error().Err(err).Msg("error during handler-sync")
				f.log.Error("error during handler-sync", "error", err)
			}
			if !success {
				// keep processing the handler without waiting for updates
				f.triggerHandlerProcessing()
			}
		}
	}
}
