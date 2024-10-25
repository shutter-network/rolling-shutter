package chaincache

import (
	"context"
	"errors"
	"fmt"
	"math"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"

	database "github.com/shutter-network/rolling-shutter/rolling-shutter/chainobserver/db/keyper"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/chainsync/chainsegment"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/chainsync/syncer"
)

const MaxSyncedBlockCacheSize = 100

var _ syncer.ChainCache = &DatabaseChainCache{}

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
	if len(headers) == 0 {
		return nil, syncer.ErrEmpy
	}
	return chainsegment.NewChainSegment(headers...), nil
}

func (f *DatabaseChainCache) GetHeaderByHash(ctx context.Context, blockHash common.Hash) (*types.Header, error) {
	var header *types.Header
	err := f.dbpool.BeginFunc(ctx, func(tx pgx.Tx) error {
		q := database.New(tx)
		block, err := q.GetSyncedBlockByHash(ctx, blockHash.Bytes())
		if err != nil {
			return err
		}
		header, err = DecodeHeader(block.Header)
		return err
	})
	if err != nil && err != pgx.ErrNoRows {
		return nil, fmt.Errorf("failed to query block header from database cache: %w", err)
	}
	if err == pgx.ErrNoRows {
		// we don't have the header cached
		err = nil
	}
	return header, err
}

func (f *DatabaseChainCache) Update(ctx context.Context, update syncer.ChainUpdateContext) error {
	return f.dbpool.BeginFunc(ctx, func(tx pgx.Tx) error {
		q := database.New(tx)
		for _, header := range update.Remove.Get() {
			// Not strictly necessary if the QueryContext.Delete
			// have the corresponding reorged blocks in QueryContext.Update
			err := q.DeleteSyncedBlockByHash(ctx, header.Hash().Bytes())
			if err != nil {
				return err
			}
		}
		updateHeader := update.Append.Get()
		for _, header := range updateHeader {
			h, err := EncodeHeader(header)
			if err != nil {
				return err
			}
			if !header.Number.IsInt64() {
				return fmt.Errorf("block number int64 overflow")
			}
			if header.Time > math.MaxInt64 {
				return errors.New("block time int64 overflow")
			}
			if err := q.InsertSyncedBlock(ctx,
				database.InsertSyncedBlockParams{
					BlockHash:   header.Hash().Bytes(),
					ParentHash:  header.ParentHash.Bytes(),
					BlockNumber: header.Number.Int64(),
					Timestamp:   int64(header.Time),
					Header:      h,
				}); err != nil {
				return err
			}
		}

		if len(updateHeader) != 0 {
			earliestBlockNum := updateHeader[0].Number
			if !earliestBlockNum.IsInt64() {
				return fmt.Errorf("earliest block number int64 overflow")
			}
			evictBefore := earliestBlockNum.Int64() - MaxSyncedBlockCacheSize
			err := q.EvictSyncedBlocksBefore(ctx, evictBefore)
			if err != nil {
				return err
			}
		}
		return nil
	})
}
