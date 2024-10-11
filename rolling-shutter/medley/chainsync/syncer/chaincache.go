package syncer

import (
	"context"
	"errors"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/chainsync/chainsegment"
)

type ChainCache interface {
	Get(context.Context) (*chainsegment.ChainSegment, error)
	Update(context.Context, ChainUpdateContext) error
	GetHeaderByHash(context.Context, common.Hash) (*types.Header, error)
}

var ErrEmpy = errors.New("chain-cache empty")

var _ ChainCache = &MemoryChainCache{}

func NewMemoryChainCache(maxSize int, chain *chainsegment.ChainSegment) *MemoryChainCache {
	// chain can be nil
	return &MemoryChainCache{
		chain:   chain,
		maxSize: maxSize,
	}
}

type MemoryChainCache struct {
	chain   *chainsegment.ChainSegment
	maxSize int
}

func (mcc *MemoryChainCache) Get(_ context.Context) (*chainsegment.ChainSegment, error) {
	if mcc.chain == nil {
		return nil, ErrEmpy
	}
	return mcc.chain, nil
}

func (mcc *MemoryChainCache) GetHeaderByHash(_ context.Context, h common.Hash) (*types.Header, error) {
	return mcc.chain.GetHeaderByHash(h), nil
}

func (mcc *MemoryChainCache) Update(_ context.Context, update ChainUpdateContext) error {
	newSegment := []*types.Header{}
	if mcc.chain != nil {
		// OPTIM: can be implemented more efficient, but mainly used for testing
		removeHashes := map[common.Hash]struct{}{}
		if update.Remove != nil {
			for _, header := range update.Remove.Get() {
				removeHashes[header.Hash()] = struct{}{}
			}
		}
		for _, header := range mcc.chain.Get() {
			_, remove := removeHashes[header.Hash()]
			if !remove {
				newSegment = append(newSegment, header)
			}
		}
		if update.Append != nil {
			newSegment = append(newSegment, update.Append.Get()...)
		}
		if len(newSegment) > mcc.maxSize {
			// TODO: check for oneoff.
			newSegment = newSegment[len(newSegment)-mcc.maxSize:]
		}
	} else {
		if update.Append == nil {
			return nil
		}
		newSegment = update.Append.Get()
	}
	mcc.chain = chainsegment.NewChainSegment(newSegment...)
	return nil
}
