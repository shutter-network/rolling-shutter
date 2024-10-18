package tester

import (
	"context"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	gethLog "github.com/ethereum/go-ethereum/log"
	"github.com/shutter-network/shop-contracts/bindings"
	"golang.org/x/exp/slog"
	"gotest.tools/v3/assert"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/chainsync/syncer"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/service"
)

// This serves as a smoketest for the home-baked PackKeyBroadcast function
// in the test and the handler parsing.
func TestPackAndParseEvent(t *testing.T) {
	headers := MakeChain(1, common.BigToHash(big.NewInt(0)), 1, 42)
	header := *headers[0]
	eon := uint64(42)
	key := []byte("thisisasecretbytekey")

	eventLog, err := PackKeyBroadcast(eon, key, header)
	assert.NilError(t, err)

	log := gethLog.NewLogger(slog.Default().Handler())
	testHandler, err := NewTestKeyBroadcastHandler(log)
	assert.NilError(t, err)
	h, err := syncer.WrapHandler(testHandler)
	assert.NilError(t, err)

	evAny, err := h.Parse(*eventLog)
	assert.NilError(t, err)

	ev, ok := evAny.(bindings.KeyBroadcastContractEonKeyBroadcast)
	assert.Assert(t, ok)
	assert.Equal(t, ev.Eon, eon)
	assert.DeepEqual(t, ev.Key, key)
	assert.DeepEqual(t, ev.Raw.BlockHash, header.Hash())
}

func TestReorg(t *testing.T) { //nolint: funlen,gocyclo
	log := gethLog.NewLogger(slog.Default().Handler())

	var originalEvents LogFactory = func(relativeIndex int, header *types.Header) ([]types.Log, error) {
		if relativeIndex == 1 {
			// shouldn't be removed, is in non-reorged chainsegment
			log, err := PackKeyBroadcast(1, []byte("key1"), *header)
			if err != nil || log == nil {
				return nil, err
			}
			return []types.Log{*log}, err
		}
		if relativeIndex == 7 {
			// should be removed, is in reorged chainsegment
			log, err := PackKeyBroadcast(2, []byte("key2"), *header)
			if err != nil || log == nil {
				return nil, err
			}
			return []types.Log{*log}, err
		}
		return nil, nil
	}
	var updateEvents LogFactory = func(relativeIndex int, header *types.Header) ([]types.Log, error) {
		if relativeIndex == 1 {
			// shouldn't be removed, is in non-reorged chainsegment
			log, err := PackKeyBroadcast(3, []byte("key3"), *header)
			if err != nil || log == nil {
				return nil, err
			}
			return []types.Log{*log}, err
		}
		if relativeIndex == 7 {
			// shouldn't be removed, is in non-reorged chainsegment
			log, err := PackKeyBroadcast(4, []byte("key4"), *header)
			if err != nil || log == nil {
				return nil, err
			}
			return []types.Log{*log}, err
		}
		return nil, nil
	}
	chain := MakeChainSegments(t,
		MakeChainSegmentsArgs{
			Original: MakeChainSegmentsChain{
				Length:     10,
				LogFactory: originalEvents,
			},
			Update: MakeChainSegmentsChain{
				Length:     10,
				LogFactory: updateEvents,
			},
			BranchOffBlock:        5,
			UpdateSegmentLength:   1,
			Reorg:                 true,
			ClientNoProgressHeads: true,
		},
	)

	f := syncer.NewFetcher(chain.Client, syncer.NewMemoryChainCache(50, nil))

	keyBroadcastHandler, err := NewTestKeyBroadcastHandler(log)
	assert.NilError(t, err)
	h, err := syncer.WrapHandler(keyBroadcastHandler)
	assert.NilError(t, err)

	chainUpdateHandler, chainUpdateHandlerChannel, err := NewTestChainUpdateHandler(log)
	assert.NilError(t, err)

	f.RegisterContractEventHandler(h)
	f.RegisterChainUpdateHandler(chainUpdateHandler)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// we have to be able to query a latest head
	ok := chain.Controller.ProgressHead()
	assert.Assert(t, ok)
	group, deferFn := service.RunBackground(ctx, f)
	defer deferFn()
	defer func() {
		if ctx.Err() != nil {
			// only wait for the error when the deadline
			// raised
			err := group.Wait()
			if err != nil {
				err = fmt.Errorf("Fetcher failed during test: %w", err)
			}
			assert.NilError(t, err)
		}
	}()
	chain.Controller.WaitSubscribed(ctx)

	for {
		ok := chain.Controller.ProgressHead()
		if !ok {
			break
		}
		err := chain.Controller.EmitLatestHead(ctx)
		assert.NilError(t, err)
		// Wait for the handler to be finished with processing
		select {
		case <-chainUpdateHandlerChannel:
		case <-ctx.Done():
			t.FailNow()
		}
	}
	uptodateEons := keyBroadcastHandler.GetEons()
	t.Logf("eons: %v", uptodateEons)
	for _, eon := range []uint64{1, 3, 4} {
		_, ok := uptodateEons[eon]
		assert.Assert(t, ok)
	}
	_ = group
	// group.Wait()
}
