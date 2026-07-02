package eventsyncer_test

import (
	"context"
	"math/big"
	"reflect"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"gotest.tools/assert"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyperimpl/shutterservice/help"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/eventsyncer"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/service"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/testutils/rpccount"
)

// TestEventSyncerRPCCount instruments the RPC calls the EventSyncer issues
// when it is asked to sync a range of blocks that contain events matching
// several registered EventTypes. The test does not assert exact numbers;
// instead it prints them via t.Log for before/after comparison across the B4
// optimization commit.
//
// Setup:
//   - Deploys the Emitter contract on a simulated backend fronted by a
//     JSON-RPC counting proxy.
//   - Registers five EventTypes on the same contract (Two, Four, Five, Six,
//     SingleIdx).
//   - Commits enough blocks that at least one of every event type has been
//     emitted.
//   - Resets the counter, then runs EventSyncer.Start and drains Next() until
//     the syncer has caught up to the tip.
//   - Logs the counts of the JSON-RPC methods the syncer issued during that
//     drain.
func TestEventSyncerRPCCount(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	privateKey, err := crypto.GenerateKey()
	assert.NilError(t, err)
	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, big.NewInt(1337))
	assert.NilError(t, err)

	balance := new(big.Int)
	balance.SetString("1000000000000000000000", 10)
	alloc := types.GenesisAlloc{auth.From: {Balance: balance}}

	rpc := rpccount.NewBackend(t, alloc)

	contractAddress, _, emitter, err := help.DeployEmitter(auth, rpc.Sim.Client())
	assert.NilError(t, err)
	rpc.Sim.Commit()

	// Emit one instance of each event type, sealing a fresh block each time.
	// Five distinct blocks, each carrying a distinct topic0, so every
	// EventType registered below has at least one matching log to fetch.
	tx, err := emitter.EmitTwo(auth, big.NewInt(1))
	assert.NilError(t, err)
	rpc.Sim.Commit()
	_ = tx

	tx, err = emitter.EmitFour(auth, big.NewInt(1), big.NewInt(2), big.NewInt(3), []byte{0xAA})
	assert.NilError(t, err)
	rpc.Sim.Commit()

	tx, err = emitter.EmitFive(auth, big.NewInt(1), big.NewInt(2), big.NewInt(3), []byte{0xAA}, []byte{0xBB})
	assert.NilError(t, err)
	rpc.Sim.Commit()

	tx, err = emitter.EmitSix(auth, big.NewInt(1), "hello", auth.From, []byte{0xAA}, big.NewInt(9), []byte{0xBB})
	assert.NilError(t, err)
	rpc.Sim.Commit()

	tx, err = emitter.EmitSingleIdx(auth, big.NewInt(1), []byte{0xCC}, big.NewInt(2))
	assert.NilError(t, err)
	rpc.Sim.Commit()

	// Add a handful of empty blocks after the last emission so the sync range
	// stretches beyond the emitted events and exercises the polling loop.
	for i := 0; i < 5; i++ {
		rpc.Sim.Commit()
	}

	// Query the current tip so we know when the syncer is caught up.
	tipHeader, err := rpc.Client.HeaderByNumber(ctx, nil)
	assert.NilError(t, err)
	tipBlock := tipHeader.Number.Uint64()

	parsedABI, err := help.EmitterMetaData.GetAbi()
	assert.NilError(t, err)
	boundContract := bind.NewBoundContract(contractAddress, *parsedABI, rpc.Client, rpc.Client, rpc.Client)

	eventNames := []string{"Two", "Four", "Five", "Six", "SingleIdx"}
	eventTypeReflect := map[string]reflect.Type{
		"Two":       reflect.TypeOf(help.EmitterTwo{}),
		"Four":      reflect.TypeOf(help.EmitterFour{}),
		"Five":      reflect.TypeOf(help.EmitterFive{}),
		"Six":       reflect.TypeOf(help.EmitterSix{}),
		"SingleIdx": reflect.TypeOf(help.EmitterSingleIdx{}),
	}
	events := make([]*eventsyncer.EventType, 0, len(eventNames))
	for _, name := range eventNames {
		events = append(events, &eventsyncer.EventType{
			Contract:        boundContract,
			Address:         contractAddress,
			FromBlockNumber: 0,
			ABI:             *parsedABI,
			Name:            name,
			Type:            eventTypeReflect[name],
			Handler:         nil,
		})
	}

	// Reset counts so only the sync loop's RPCs are recorded.
	rpc.Reset()

	syncer := eventsyncer.New(rpc.Client, 0, events, 1, 0)
	group, deferFn := service.RunBackground(ctx, syncer)
	defer func() {
		cancel()
		deferFn()
		_ = group.Wait()
	}()

	// Drain events until the syncer has caught up to the tip.
	drainCtx, drainCancel := context.WithTimeout(ctx, 20*time.Second)
	defer drainCancel()
	seenEvents := 0
	for {
		u, err := syncer.Next(drainCtx)
		if err != nil {
			t.Fatalf("syncer.Next: %v", err)
		}
		if u.Event != nil {
			seenEvents++
			continue
		}
		if u.BlockNumber >= tipBlock {
			break
		}
	}

	assert.Assert(t, seenEvents >= len(eventNames),
		"expected at least %d events, got %d", len(eventNames), seenEvents)

	counts := rpc.Counts()
	t.Logf(
		"EventSyncer B4: tipBlock=%d numEventTypes=%d eventsSeen=%d",
		tipBlock, len(events), seenEvents,
	)
	rpc.LogCounts(t, "EventSyncer B4 RPC counts")

	// Also log the two metrics that matter most in the analysis doc: eth_getLogs
	// and eth_blockNumber. These are the invariants the B4 commit changed.
	t.Logf("EventSyncer B4 headline: eth_getLogs=%d eth_blockNumber=%d",
		counts["eth_getLogs"], counts["eth_blockNumber"])

	_ = common.Address{} // keep import parity with future use
}
