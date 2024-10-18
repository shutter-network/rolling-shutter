package chainsegment

import (
	"context"
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/chainsync/client"
)

const MaxNumPollBlocks = 50

var (
	ErrReorg                   = errors.New("detected reorg in updated chain-segment")
	ErrEmpty                   = errors.New("empty chain-segment")
	ErrUpdateBlockTooFarInPast = errors.New("the updated block reaches too far in the past for the chain-segment")
	ErrOverlapTooBig           = errors.New("chain-segment overlap too big")
)

type UpdateLatestResult struct {
	// the full new segment with the reorg applied
	FullSegment *ChainSegment
	// the removed segment that is not part of the new full segment anymore
	// (reorged blocks)
	RemovedSegment *ChainSegment
	// the updated segment of new blocks that were not part of the old chain
	// (new blocks including the replacement blocks from a reorg)
	UpdatedSegment *ChainSegment
}

func capNumPollBlocks(num int) int {
	if num > MaxNumPollBlocks {
		return MaxNumPollBlocks
	} else if num < 1 {
		return 1
	}
	return num
}

type ChainSegment struct {
	chain []*types.Header
}

// assumes reverse order, and that it is
// a connected chain-segment.
func NewChainSegment(chain ...*types.Header) *ChainSegment {
	bc := &ChainSegment{
		chain: chain,
	}
	return bc
}

func (bc *ChainSegment) GetHeaderByHash(h common.Hash) *types.Header {
	// OPTIM: this should be implemented more efficiently
	// with a hash-map
	for _, header := range bc.chain {
		if header.Hash().Cmp(h) == 0 {
			return header
		}
	}
	return nil
}

func (bc *ChainSegment) Len() int {
	return len(bc.chain)
}

func (bc *ChainSegment) Earliest() *types.Header {
	if len(bc.chain) == 0 {
		return nil
	}
	return bc.chain[0]
}

func (bc *ChainSegment) Latest() *types.Header {
	if len(bc.chain) == 0 {
		return nil
	}
	return bc.chain[len(bc.chain)-1]
}

func (bc *ChainSegment) Get() []*types.Header {
	return bc.chain
}

func (bc *ChainSegment) Copy() *ChainSegment {
	return NewChainSegment(bc.chain...)
}

func (bc *ChainSegment) UpdateLatest(ctx context.Context, c client.Sync, update *ChainSegment) (UpdateLatestResult, error) {
	update = update.Copy()
	if bc.Len() == 0 {
		return UpdateLatestResult{}, ErrEmpty
	}

	if bc.Earliest().Number.Cmp(update.Earliest().Number) == 1 {
		// We don't reach so far in the past.
		// This only happens when the cache of used blocks in
		// doesn't reach so far.
		return UpdateLatestResult{}, fmt.Errorf(
			"segment earliest=%d, update earliest=%d: %w",
			bc.Earliest().Number.Int64(), update.Earliest().Number.Int64(),
			ErrUpdateBlockTooFarInPast,
		)
	}
	overlapBig := new(big.Int).Add(
		new(big.Int).Sub(bc.Latest().Number, update.Earliest().Number),
		// both being the same height means one block overlap, so add 1
		big.NewInt(1),
	)
	if !overlapBig.IsInt64() {
		// this should never happen, this would be too large of a gap
		return UpdateLatestResult{}, ErrOverlapTooBig
	}

	overlap := int(overlapBig.Int64())
	if overlap < 0 {
		// overlap is negative, this means we have a gap:
		extendedUpdate, err := update.ExtendLeft(ctx, c, capNumPollBlocks(-overlap))
		if err != nil {
			return UpdateLatestResult{}, fmt.Errorf("failed to extend left gap: %w", err)
		}
		return bc.UpdateLatest(ctx, c, extendedUpdate)
	} else if overlap == 0 {
		if update.Earliest().ParentHash.Cmp(bc.Latest().Hash()) == 0 {
			// the new segment extends the old one perfectly
			return UpdateLatestResult{
				FullSegment:    bc.Copy().AddRight(update),
				RemovedSegment: nil,
				UpdatedSegment: update,
			}, nil
		}
		// the block-numbers align, but the new segment
		// seems to be from a reorg that branches off within the old segment
		_, err := update.ExtendLeft(ctx, c, capNumPollBlocks(bc.Len()))
		if err != nil {
			return UpdateLatestResult{}, fmt.Errorf("failed to extend into reorg: %w", err)
		}
		return bc.UpdateLatest(ctx, c, update)
	}
	// implicit case - overlap > 0:
	removed, updated := bc.GetLatest(overlap).DiffLeftAligned(update)
	full := bc
	if removed != nil {
		full = full.GetEarliest(full.Len() - removed.Len())
	}
	if updated != nil {
		full.AddRight(updated)
	}
	return UpdateLatestResult{
		FullSegment:    full,
		RemovedSegment: removed,
		UpdatedSegment: updated,
	}, nil
}

func (bc *ChainSegment) AddRight(add *ChainSegment) *ChainSegment {
	bc.chain = append(bc.chain, add.chain...)
	return bc
}

func (bc *ChainSegment) DiffLeftAligned(other *ChainSegment) (remove, update *ChainSegment) {
	// 1) assumes both segments start at the same block height (earliest block at index 0 with same blocknum)
	// 2) assumes the other.Len() >= bc.Len()

	// Compare the two and see if we have to reorg based on the hashes
	removed := []*types.Header{}
	updated := []*types.Header{}
	oldChain := bc.Get()
	newChain := other.Get()

	for i := 0; i < len(newChain); i++ {
		var oldHeader *types.Header
		newHeader := newChain[i]
		if len(oldChain) > i {
			oldHeader = oldChain[i]
		}
		if oldHeader == nil {
			updated = append(updated, newHeader)
			// XXX: is just checking the hash sufficient?
			// sanity check also the blocknum + parent hash chain?
		} else if oldHeader.Hash().Cmp(newHeader.Hash()) != 0 {
			removed = append(removed, oldHeader)
			updated = append(updated, newHeader)
		}
	}
	var removedSegment, updatedSegment *ChainSegment
	if len(removed) > 0 {
		removedSegment = NewChainSegment(removed...)
	}
	if len(updated) > 0 {
		updatedSegment = NewChainSegment(updated...)
	}
	return removedSegment, updatedSegment
}

// GetLatest retrieves the "n" latest blocks from this
// ChainSegment.
// If the segment is shorter than n, the whole segment gets returned.
func (bc *ChainSegment) GetLatest(n int) *ChainSegment {
	if n > bc.Len() {
		n = bc.Len()
	}
	return NewChainSegment(bc.chain[len(bc.chain)-n : len(bc.chain)]...)
}

// GetLatest retrieves the "n" earliest blocks from this
// ChainSegment.
// If the segment is shorter than n, the whole segment gets returned.
func (bc *ChainSegment) GetEarliest(n int) *ChainSegment {
	if n > bc.Len() {
		n = bc.Len()
	}
	return NewChainSegment(bc.chain[:n]...)
}

func (bc *ChainSegment) NewSegmentRight(ctx context.Context, c client.Sync, num int) (*ChainSegment, error) {
	rightMost := bc.Latest()
	if rightMost == nil {
		return nil, ErrEmpty
	}
	chain := []*types.Header{}
	previous := rightMost
	for i := 1; i <= num; i++ {
		blockNum := new(big.Int).Sub(rightMost.Number, big.NewInt(int64(i)))
		h, err := c.HeaderByNumber(ctx, blockNum)
		if err != nil {
			return nil, err
		}
		if h.Hash().Cmp(previous.ParentHash) != 0 {
			// the server has a different chain state than this segment,
			// so it is part of a reorged away chain-segment
			return nil, ErrReorg
		}
		chain = append(chain, h)
		previous = h
	}
	return NewChainSegment(chain...), nil
}

func (bc *ChainSegment) ExtendLeft(ctx context.Context, c client.Sync, num int) (*ChainSegment, error) {
	leftMost := bc.Earliest()
	if leftMost == nil {
		return nil, ErrEmpty
	}
	for num > 0 {
		blockNum := new(big.Int).Sub(leftMost.Number, big.NewInt(int64(1)))
		// OPTIM: we do cap the max poll number when calling this method,
		// but then we make one request per block anyways.
		// This doesn't make sense, but there currently is no batching
		// for retrieving ranges of headers.
		h, err := c.HeaderByNumber(ctx, blockNum)
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve header by number (#%d): %w", blockNum.Uint64(), err)
		}
		if h.Hash().Cmp(leftMost.ParentHash) != 0 {
			// The server has a different chain state than this segment,
			// so it is part of a reorged away chain-segment.
			// This can also happen when the server reorged during this loop
			// and we now polled the parent with an unexpected hash.
			return nil, ErrReorg
		}
		bc.chain = append([]*types.Header{h}, bc.chain...)
		leftMost = h
		num--
	}
	return bc, nil
}
