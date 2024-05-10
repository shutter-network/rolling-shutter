package syncer

import (
	"context"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
	"gotest.tools/v3/assert"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/chainsync/client"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/chainsync/event"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/encodeable/number"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/service"
)

func MakeChain(start int64, startParent common.Hash, numHeader uint, seed int64) []*types.Header {
	n := numHeader
	parent := startParent
	num := big.NewInt(start)
	h := []*types.Header{}

	// change the hashes for different seeds
	mixinh := common.BigToHash(big.NewInt(seed))
	for n > 0 {
		head := &types.Header{
			ParentHash: parent,
			Number:     num,
			MixDigest:  mixinh,
		}
		h = append(h, head)
		num = new(big.Int).Add(num, big.NewInt(1))
		parent = head.Hash()
		n--
	}
	return h
}

func TestReorg(t *testing.T) {
	headersBeforeReorg := MakeChain(1, common.BigToHash(big.NewInt(0)), 10, 42)
	branchOff := headersBeforeReorg[5]
	// block number 5 will be reorged
	headersReorgBranch := MakeChain(branchOff.Number.Int64()+1, branchOff.Hash(), 10, 43)
	clnt, ctl := client.NewTestClient()
	ctl.AppendNextHeaders(headersBeforeReorg...)
	ctl.AppendNextHeaders(headersReorgBranch...)

	handlerBlock := make(chan *event.LatestBlock, 1)

	h := &UnsafeHeadSyncer{
		Client: clnt,
		Log:    log.New(),
		Handler: func(_ context.Context, ev *event.LatestBlock) error {
			handlerBlock <- ev
			return nil
		},
		SyncedHandler:      []ManualFilterHandler{},
		SyncStartBlock:     number.NewBlockNumber(nil),
		FetchActiveAtStart: false,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	service.RunBackground(ctx, h)

	// intitial sync is independent of the subscription,
	// this will get polled from the eth client
	b := <-handlerBlock
	assert.Assert(t, b.Number.Cmp(headersBeforeReorg[0].Number) == 0)
	idx := 1
	for {
		ok := ctl.ProgressHead()
		assert.Assert(t, ok)
		err := ctl.EmitEvents(ctx)
		assert.NilError(t, err)

		b = <-handlerBlock
		assert.Equal(t, b.Number.Uint64(), headersBeforeReorg[idx].Number.Uint64(), fmt.Sprintf("block number equal for idx %d", idx))
		assert.Equal(t, b.BlockHash, headersBeforeReorg[idx].Hash())
		idx++
		if idx == len(headersBeforeReorg) {
			break
		}
	}
	ok := ctl.ProgressHead()
	assert.Assert(t, ok)
	err := ctl.EmitEvents(ctx)
	assert.NilError(t, err)
	b = <-handlerBlock
	// now the reorg should have happened.
	// the handler should have emitted an "artificial" latest head
	// event for the block BEFORE the re-orged block
	assert.Equal(t, b.Number.Uint64(), headersReorgBranch[0].Number.Uint64()-1, "block number equal for reorg")
	assert.Equal(t, b.BlockHash, headersReorgBranch[0].ParentHash)
	idx = 0
	for ctl.ProgressHead() {
		assert.Assert(t, ok)
		err := ctl.EmitEvents(ctx)
		assert.NilError(t, err)

		b := <-handlerBlock
		assert.Equal(t, b.Number.Uint64(), headersReorgBranch[idx].Number.Uint64(), fmt.Sprintf("block number equal for idx %d", idx))
		assert.Equal(t, b.BlockHash, headersReorgBranch[idx].Hash())
		idx++
		if idx == len(headersReorgBranch) {
			break
		}
	}
}
