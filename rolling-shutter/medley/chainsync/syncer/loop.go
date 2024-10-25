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
func (f *Fetcher) handlerSync(ctx context.Context) (success bool, err error) { //nolint: funlen
	var syncedChain, removedSegment, updatedSegment *chainsegment.ChainSegment

	syncedChain, err = f.chainCache.Get(ctx)
	if f.chainUpdate == nil {
		// nothing to update
		success = true
		return success, err
	}
	if errors.Is(err, ErrEmpy) {
		// no chain-cache present yet, just set the chain-update without
		// checking for reorgs
		removedSegment = nil
		updatedSegment = f.chainUpdate
		log.Trace().Msg("internal chain cache empty, setting updated chain segment")
	} else if err != nil {
		success = false
		return success, err
	} else {
		if new(big.Int).Add(syncedChain.Latest().Number, big.NewInt(1)).
			Cmp(f.chainUpdate.Earliest().Number) == -1 {
			diffBig := new(big.Int).Sub(f.chainUpdate.Earliest().Number, syncedChain.Latest().Number)
			if !diffBig.IsUint64() {
				success = false
				return success, fmt.Errorf("chain-update difference too big: %w", errUint64Overflow)
			}
			diff := diffBig.Uint64()
			queryBlocks := MaxRequestBlockRange
			// cap the extend range at the diff to the update to not overshoot
			if diff < uint64(queryBlocks) {
				queryBlocks = int(diff)
			}

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
				// extend the chain update to the left in chunks and try again
				var extendErr error
				f.chainUpdate, extendErr = f.chainUpdate.ExtendLeft(ctx, f.ethClient, queryBlocks)
				if extendErr != nil {
					err = fmt.Errorf("error while querying older blocks from reorg update: %w", err)
				}
				success = false
				return success, err
			}
			removedSegment = nil
			success = false
		} else {
			result, updateErr := syncedChain.UpdateLatest(ctx, f.ethClient, f.chainUpdate)
			success = true
			if updateErr != nil {
				log.Error().Err(err).Msg("error updating chain with latest segment")
				if errors.Is(err, chainsegment.ErrUpdateBlockTooFarInPast) {
					// TODO: what should we do on 'ErrUpdateBlockTooFarInPast'?
					// We can't provide handler calls with the same accuracy of
					// information on the potentially "removed" chain-segment,
					// since our chain-cache does not have the full old chain segment
					// in it's storage anymore, and especially the block-hash
					// of the reorged away chain is not present anymore.
					// The client should probably panic with a critical log error.
					// In general this is very unlikely when the chain-cache capacity is
					// larger than the most unlikely, still realistic reorg-size.
					// However the described condition might currently occur during
					// initial syncing, when the block-cache is not filled to capacity
					// yet.
					log.Warn().Err(err).Msg("received a reorg that pre-dates the internal chain-cache." +
						" ignoring chain-update for now, but this condition might be irrecoverable")
				}
				err = updateErr
			}
			removedSegment = result.RemovedSegment
			updatedSegment = result.UpdatedSegment
			// we will process the whole segment of the chain update
			f.chainUpdate = nil
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
			if !ok {
				log.Debug().Msg("latest head stream closed, exiting handler loop")
				return nil
			}
			log.Debug().Uint64("block-number", newHeader.Number.Uint64()).Msg("new latest head from l2 ws-stream")
			newSegment := chainsegment.NewChainSegment(newHeader)
			if f.chainUpdate != nil {
				// apply the updates to the chain-update buffer that hasn't been processed
				// yet by the handlers
				result, err := f.chainUpdate.Copy().UpdateLatest(ctx, f.ethClient, newSegment)
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
					}
					log.Error().Err(err).Msg("error updating chain segment")
				}
				if removed != nil {
					log.Info().Uint64("block-number", newHeader.Number.Uint64()).Msg("received a new reorg block")
				}
				f.chainUpdate = fullUpdated
			} else {
				f.chainUpdate = newSegment
			}
			f.triggerHandlerProcessing()

		case <-f.processingTrig:
			success, err := f.handlerSync(ctx)
			if err != nil {
				if errors.Is(err, errs.ErrCritical) {
					return err
				}
				log.Error().Err(err).Msg("error during handler-sync")
			}
			if !success {
				// keep processing the handler without waiting for updates
				f.triggerHandlerProcessing()
			}
		}
	}
}
