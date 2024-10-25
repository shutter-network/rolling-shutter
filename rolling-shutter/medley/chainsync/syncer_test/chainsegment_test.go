package tester

import (
	"context"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	gethLog "github.com/ethereum/go-ethereum/log"
	"golang.org/x/exp/slog"
	"gotest.tools/v3/assert"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/chainsync/chainsegment"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/chainsync/client"
)

func TestExtendLeft(t *testing.T) {
	headers := MakeChain(1, common.BigToHash(big.NewInt(0)), 10, 42)
	log := gethLog.NewLogger(slog.Default().Handler())
	clnt, ctl := client.NewTestClient(log)

	for _, h := range headers {
		ctl.AppendNextHeader(h)
		ctl.ProgressHead()
	}
	latest := chainsegment.NewChainSegment(headers[9])
	_, err := latest.ExtendLeft(context.Background(), clnt, 9)
	assert.NilError(t, err)
	assert.Equal(t, len(latest.Get()), 10)
}

func TestUpdateLatest(t *testing.T) { //nolint: funlen
	tests := map[string]struct {
		chainArgs MakeChainSegmentsArgs

		expectedFullLength int

		expectedUpdatedLength      int
		expectedUpdatedEarliestNum int

		expectedRemovedLength      int
		expectedRemovedEarliestNum int

		expectedErrorString string
	}{
		"close gap and reorg": {
			chainArgs: MakeChainSegmentsArgs{
				Original: MakeChainSegmentsChain{
					Length: 10,
				},
				Update: MakeChainSegmentsChain{
					Length: 15,
				},
				BranchOffBlock:      5,
				UpdateSegmentLength: 1,
				Reorg:               true,
			},

			expectedFullLength:         20,
			expectedUpdatedLength:      15,
			expectedUpdatedEarliestNum: 5,

			expectedRemovedLength:      5,
			expectedRemovedEarliestNum: 5,
			expectedErrorString:        "",
		},
		"no gap and reorg": {
			chainArgs: MakeChainSegmentsArgs{
				Original: MakeChainSegmentsChain{
					Length: 10,
				},
				Update: MakeChainSegmentsChain{
					Length: 15,
				},
				BranchOffBlock:      5,
				UpdateSegmentLength: 15,
				Reorg:               true,
			},

			expectedFullLength: 20,

			expectedUpdatedLength:      15,
			expectedUpdatedEarliestNum: 5,

			expectedRemovedLength:      5,
			expectedRemovedEarliestNum: 5,
			expectedErrorString:        "",
		},
		"overlap and reorg": {
			chainArgs: MakeChainSegmentsArgs{
				Original: MakeChainSegmentsChain{
					Length: 10,
				},
				Update: MakeChainSegmentsChain{
					Length: 15,
				},
				BranchOffBlock:      5,
				UpdateSegmentLength: 18,
				Reorg:               true,
			},
			expectedFullLength: 20,

			expectedUpdatedLength:      15,
			expectedUpdatedEarliestNum: 5,

			expectedRemovedLength:      5,
			expectedRemovedEarliestNum: 5,
			expectedErrorString:        "",
		},
		"append no reorg": {
			chainArgs: MakeChainSegmentsArgs{
				Original: MakeChainSegmentsChain{
					Length: 5,
				},
				Update: MakeChainSegmentsChain{
					Length: 10,
				},
				BranchOffBlock: 0,
				// no gap, perfect alignment
				UpdateSegmentLength: 5,
				Reorg:               false,
			},

			expectedFullLength:         10,
			expectedUpdatedLength:      5,
			expectedUpdatedEarliestNum: 5,
			expectedRemovedLength:      0,
			expectedRemovedEarliestNum: -1,
			expectedErrorString:        "",
		},
		"close gap no reorg": {
			chainArgs: MakeChainSegmentsArgs{
				Original: MakeChainSegmentsChain{
					Length: 5,
				},
				Update: MakeChainSegmentsChain{
					Length: 10,
				},
				BranchOffBlock: 0,
				// gap of 3
				UpdateSegmentLength: 2,
				Reorg:               false,
			},
			expectedFullLength:         10,
			expectedUpdatedLength:      5,
			expectedUpdatedEarliestNum: 5,
			expectedRemovedLength:      0,
			expectedRemovedEarliestNum: -1,
			expectedErrorString:        "",
		},
		"overlap no reorg": {
			chainArgs: MakeChainSegmentsArgs{
				Original: MakeChainSegmentsChain{
					Length: 5,
				},
				Update: MakeChainSegmentsChain{
					Length: 10,
				},
				BranchOffBlock: 0,
				// overlap of 3
				UpdateSegmentLength: 8,
				Reorg:               false,
			},

			expectedFullLength: 10,
			// overlap shouldn't be updated
			expectedUpdatedLength:      5,
			expectedUpdatedEarliestNum: 5,
			expectedRemovedLength:      0,
			expectedRemovedEarliestNum: -1,
			expectedErrorString:        "",
		},
		"full overlap no reorg": {
			chainArgs: MakeChainSegmentsArgs{
				Original: MakeChainSegmentsChain{
					Length: 10,
				},
				Update: MakeChainSegmentsChain{
					Length: 10,
				},
				BranchOffBlock: 0,
				// full overlap
				UpdateSegmentLength: 10,
				Reorg:               false,
			},

			expectedFullLength: 10,
			// overlap shouldn't be updated
			expectedUpdatedLength:      0,
			expectedUpdatedEarliestNum: -1,
			expectedRemovedLength:      0,
			expectedRemovedEarliestNum: -1,
			expectedErrorString:        "",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			chain := MakeChainSegments(t, test.chainArgs)
			// this should poll all the header and detect the reorg
			assert.Assert(t, chain.UpdateSegment.Len() > 0)
			assert.Assert(t, chain.OriginalSegment.Len() > 0)
			result, err := chain.OriginalSegment.UpdateLatest(context.Background(), chain.Client, chain.UpdateSegment)
			assert.NilError(t, err)
			full := result.FullSegment
			removed := result.RemovedSegment
			updated := result.UpdatedSegment
			assert.Assert(t, full != nil)
			assert.Assert(t, full.Len() > 0)
			if test.expectedErrorString != "" {
				assert.ErrorContains(t, err, test.expectedErrorString)
				return
			}
			assert.NilError(t, err)
			assert.Assert(t, full != nil)
			assert.Assert(t, chain.OriginalSegment != nil)

			assert.Equal(t, full.Earliest().Number.Cmp(big.NewInt(0)), 0)
			for i, h := range full.Get() {
				t.Logf("full: index=%d, num=%d", i, h.Number.Uint64())
			}
			assert.Equal(t, full.Latest().Number.Cmp(big.NewInt(int64(test.expectedFullLength)-1)), 0)

			if updated != nil {
				assert.Assert(t, updated.Len() > 0)
				assert.Equal(t, updated.Len(), test.expectedUpdatedLength)
				assert.Equal(t, updated.Earliest().Number.Cmp(big.NewInt(int64(test.expectedUpdatedEarliestNum))), 0)
				assert.Equal(t, updated.Latest().Number.Cmp(big.NewInt(int64(test.expectedUpdatedEarliestNum+test.expectedUpdatedLength)-1)), 0)
			} else {
				assert.Equal(t, 0, test.expectedUpdatedLength)
			}
			if removed != nil {
				assert.Assert(t, removed.Len() > 0)
				assert.Equal(t, removed.Len(), test.expectedRemovedLength)
				assert.Equal(t, removed.Earliest().Number.Cmp(big.NewInt(int64(test.expectedRemovedEarliestNum))), 0)
				assert.Equal(t, removed.Latest().Number.Cmp(big.NewInt(int64(test.expectedRemovedEarliestNum+test.expectedRemovedLength)-1)), 0)
			} else {
				assert.Equal(t, 0, test.expectedRemovedLength)
			}
		})
	}
}

type LogFactory func(relativeIndex int, header *types.Header) ([]types.Log, error)

type MakeChainSegmentsChain struct {
	Length     int
	LogFactory LogFactory
}

type MakeChainSegmentsArgs struct {
	Original MakeChainSegmentsChain
	Update   MakeChainSegmentsChain

	BranchOffBlock int
	// This will cut the updated chain segment
	// so that it only has headers from
	// `latest update header - UpdateSegmentLength`
	// until
	// `latest update header`, while the client
	// still knows the state of all update headers
	UpdateSegmentLength int
	// uses a different mixin-seed value in the update chain,
	// and induces a reorg
	Reorg bool
	// Will not set the internal state of the client
	// to the latest head of the updated chain.
	// If this is true, the Controller.ProgressHead()
	// or Controller.ProgressAllHeads() have to
	// be called manually so that blocks can
	// be queried from the client.
	ClientNoProgressHeads bool
}

type MakeChainSegmentsResult struct {
	Client          client.Sync
	Controller      *client.TestClientController
	OriginalSegment *chainsegment.ChainSegment
	UpdateSegment   *chainsegment.ChainSegment
}

func MakeChainSegments(t *testing.T, args MakeChainSegmentsArgs) *MakeChainSegmentsResult {
	t.Helper()

	var oldHeaders []*types.Header
	newHeaders := []*types.Header{}
	assert.Assert(t, args.BranchOffBlock < args.Original.Length)
	newChainlength := args.Update.Length + args.BranchOffBlock
	assert.Assert(t, args.UpdateSegmentLength <= newChainlength)
	oldHeaders = MakeChain(0, common.BigToHash(big.NewInt(0)), uint(args.Original.Length), 42)
	// TODO: header events

	// use different seed for the reorg chain to change the hashes
	parentHash := common.BigToHash(big.NewInt(0))
	if args.BranchOffBlock != 0 {
		parentHash = oldHeaders[args.BranchOffBlock-1].Hash()
	}
	var seed int64 = 42
	if args.Reorg {
		seed = 442
	}
	reorgHeaders := MakeChain(int64(args.BranchOffBlock), parentHash, uint(args.Update.Length), seed)
	newHeaders = append(newHeaders, oldHeaders[:args.BranchOffBlock]...)
	newHeaders = append(newHeaders, reorgHeaders...)

	// Make some assertions about the constructed chains
	assert.Equal(t, len(oldHeaders), args.Original.Length)
	assert.Equal(t, len(reorgHeaders), args.Update.Length)
	assert.Equal(t, len(newHeaders), newChainlength)
	assert.Assert(t, oldHeaders[len(oldHeaders)-1].Number.Cmp(big.NewInt(int64(len(oldHeaders)-1))) == 0)
	assert.Assert(t, reorgHeaders[0].Number.Cmp(big.NewInt(int64(args.BranchOffBlock))) == 0)

	assert.Assert(t, reorgHeaders[len(reorgHeaders)-1].Number.Cmp(big.NewInt(int64(args.BranchOffBlock+args.Update.Length-1))) == 0)

	log := gethLog.NewLogger(slog.Default().Handler())
	testClient, testClientController := client.NewTestClient(log)

	for i, h := range oldHeaders {
		var logs []types.Log
		if args.Original.LogFactory != nil {
			var err error
			logs, err = args.Original.LogFactory(i, h)
			assert.NilError(t, err)
		}
		testClientController.AppendNextHeader(h, logs...)
	}
	for i, h := range reorgHeaders {
		var logs []types.Log
		if args.Update.LogFactory != nil {
			var err error
			logs, err = args.Update.LogFactory(i, h)
			assert.NilError(t, err)
		}
		testClientController.AppendNextHeader(h, logs...)
	}
	if !args.ClientNoProgressHeads {
		testClientController.ProgressAllHeads()
	}
	original := chainsegment.NewChainSegment(oldHeaders...)
	updateHeaders := newHeaders[len(newHeaders)-args.UpdateSegmentLength:]
	update := chainsegment.NewChainSegment(updateHeaders...)
	assert.Assert(t, update.Len() == args.UpdateSegmentLength)
	assert.Assert(t, update.Len() > 0)
	assert.Assert(t, original.Len() > 0)
	assert.Assert(t, update.Len() > 0)

	return &MakeChainSegmentsResult{
		Client:          testClient,
		Controller:      testClientController,
		OriginalSegment: original,
		UpdateSegment:   update,
	}
}

func TestReplaceWholeSegment(t *testing.T) {
	headers := MakeChain(1, common.BigToHash(big.NewInt(0)), 5, 42)
	reorg := MakeChain(1, common.BigToHash(big.NewInt(0)), 5, 422)

	cs := chainsegment.NewChainSegment(headers...)
	rcs := chainsegment.NewChainSegment(reorg...)
	remove, update := cs.DiffLeftAligned(rcs)

	assert.Equal(t, len(remove.Get()), 5)
	assert.Equal(t, len(update.Get()), 5)
}
