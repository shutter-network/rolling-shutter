package syncer

import (
	"context"
	"errors"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/chainsync/chainsegment"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/chainsync/database"
)

type ChainCache interface {
	Get(context.Context) (*chainsegment.ChainSegment, error)
	Update(context.Context, QueryContext) error
}

var ErrEmpy = errors.New("chain-cache empty")

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

func (mcc *MemoryChainCache) Get(ctx context.Context) (*chainsegment.ChainSegment, error) {
	if mcc.chain == nil {
		return nil, ErrEmpy
	}
	return mcc.chain, nil

}
func (mcc *MemoryChainCache) Update(ctx context.Context, qCtx QueryContext) error {
	newSegment := []*types.Header{}
	if mcc.chain != nil {
		// OPTIM: can be implemented more efficient, but mainly used for testing
		removeHashes := map[common.Hash]struct{}{}
		if qCtx.Remove != nil {
			for _, header := range qCtx.Remove.Get() {
				removeHashes[header.Hash()] = struct{}{}
			}
		}
		for _, header := range mcc.chain.Get() {
			_, remove := removeHashes[header.Hash()]
			if !remove {
				newSegment = append(newSegment, header)
			}
		}
		if qCtx.Update != nil {
			for _, header := range qCtx.Update.Get() {
				newSegment = append(newSegment, header)
			}
		}
		if len(newSegment) > mcc.maxSize {
			//TODO: check for oneoff
			newSegment = newSegment[len(newSegment)-mcc.maxSize:]
		}
	} else {
		if qCtx.Update == nil {
			return nil
		}
		newSegment = qCtx.Update.Get()
	}
	mcc.chain = chainsegment.NewChainSegment(newSegment...)
	return nil
}

func NewDatabaseChainCache(db *pgxpool.Pool) *DatabaseChainCache {
	return &DatabaseChainCache{
		dbpool: db,
	}
}

type DatabaseChainCache struct {
	dbpool *pgxpool.Pool
}

func EncodeHeader(h *types.Header) ([]byte, error) {
	return rlp.EncodeToBytes(h)
}
func DecodeHeader(b []byte) (*types.Header, error) {
	h := new(types.Header)
	err := rlp.DecodeBytes(b, h)
	if err != nil {
		return nil, err
	}
	return h, nil
}

func (f *DatabaseChainCache) Get(ctx context.Context) (*chainsegment.ChainSegment, error) {
	headers := []*types.Header{}
	err := f.dbpool.BeginFunc(ctx, func(tx pgx.Tx) error {
		q := database.New(tx)
		blocks, err := q.GetSyncedBlocks(ctx)
		if err != nil {
			return err
		}
		// TODO: sanity check hashes / parent hashes
		for _, block := range blocks {
			h, err := DecodeHeader(block.Header)
			if err != nil {
				return fmt.Errorf("error decoding header: %w", err)
			}
			headers = append(headers, h)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return chainsegment.NewChainSegment(headers...), nil
}

func (f *DatabaseChainCache) Update(ctx context.Context, qCtx QueryContext) error {
	return f.dbpool.BeginFunc(ctx, func(tx pgx.Tx) error {
		q := database.New(tx)
		for _, header := range qCtx.Remove.Get() {
			err := q.DeleteBlock(ctx, header.Hash().Bytes())
			if err != nil {
				return err
			}
		}
		updateHeader := qCtx.Update.Get()
		for _, header := range updateHeader {
			b, err := EncodeHeader(header)
			if err != nil {
				return err
			}
			err = q.InsertBlock(ctx,
				database.InsertBlockParams{
					//TODO: check overflow
					BlockNumber: header.Number.Int64(),
					BlockHash:   header.Hash().Bytes(),
					ParentHash:  header.ParentHash.Bytes(),
					Header:      b,
				})
			if err != nil {
				return err
			}
		}

		if len(updateHeader) != 0 {
			// TODO: overflow/underflow
			earliestBlockNum := updateHeader[0].Number.Int64()
			evictBefore := earliestBlockNum - MaxSyncedBlockCacheSize
			err := q.EvictBefore(ctx, evictBefore)
			if err != nil {
				return err
			}
		}
		return nil
	})
}
