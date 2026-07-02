package shutterservice

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/jackc/pgx/v4/pgxpool"
	"gotest.tools/assert"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyperimpl/shutterservice/database"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyperimpl/shutterservice/help"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/testsetup"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/testutils/rpccount"
)

// TestMultiEventSyncerRPCCount instruments the RPC calls
// MultiEventSyncer.Sync issues when it has to fetch logs for N active
// EventTriggerRegisteredEvent rows over one sync range.
//
// The pre-S2 code path issues one FilterLogs per active trigger (N calls per
// range); the post-S2 code path unions all triggers' (address, topic[0])
// criteria into a single combined FilterLogs. This test does not assert
// specific numbers; it t.Logs the counts so the same test file compiled on
// both branches produces a clean before/after comparison.
//
// To keep the test source portable across the S2 interface change on
// EventProcessor, only TriggerProcessor is registered. That is where the
// bulk of the reduction lives anyway (registry processor issued one call in
// both variants).
func TestMultiEventSyncerRPCCount(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	// Use context.Background() for the DB pool: t.Cleanup runs after
	// deferred cancels in the test body, and dbclose needs a live context to
	// tear the schema down.
	dbpool, dbclose := testsetup.NewTestDBPool(context.Background(), t, database.Definition)
	t.Cleanup(dbclose)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
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

	// Emit one instance of each event type so there is at least one matching
	// log per registered trigger to fetch. Distinct blocks help the syncer
	// see logs in multiple sub-ranges of the run.
	_, err = emitter.EmitTwo(auth, big.NewInt(1))
	assert.NilError(t, err)
	rpc.Sim.Commit()

	_, err = emitter.EmitFour(auth, big.NewInt(1), big.NewInt(2), big.NewInt(3), []byte{0xAA})
	assert.NilError(t, err)
	rpc.Sim.Commit()

	_, err = emitter.EmitFive(auth, big.NewInt(1), big.NewInt(2), big.NewInt(3), []byte{0xAA}, []byte{0xBB})
	assert.NilError(t, err)
	rpc.Sim.Commit()

	_, err = emitter.EmitSix(auth, big.NewInt(1), "hello", auth.From, []byte{0xAA}, big.NewInt(9), []byte{0xBB})
	assert.NilError(t, err)
	rpc.Sim.Commit()

	_, err = emitter.EmitSingleIdx(auth, big.NewInt(1), []byte{0xCC}, big.NewInt(2))
	assert.NilError(t, err)
	rpc.Sim.Commit()

	// A few empty blocks after the last emission so the range extends past
	// the events.
	for i := 0; i < 3; i++ {
		rpc.Sim.Commit()
	}

	tipHeader, err := rpc.Client.HeaderByNumber(ctx, nil)
	assert.NilError(t, err)
	tipBlock := tipHeader.Number.Uint64()

	parsedABI, err := help.EmitterMetaData.GetAbi()
	assert.NilError(t, err)

	// Insert N=5 active EventTriggerRegisteredEvent rows, one per Emitter
	// event type. Each row's Definition is a valid EventTriggerDefinition
	// with a BytesEq predicate on topic[0] matching that event's topic hash,
	// so pre-S2 TriggerProcessor issues one FilterLogs per row.
	const numTriggers = 5
	eventNames := []string{"Two", "Four", "Five", "Six", "SingleIdx"}
	queries := database.New(dbpool)
	for i, name := range eventNames {
		topic := parsedABI.Events[name].ID
		def := EventTriggerDefinition{
			Contract: contractAddress,
			LogPredicates: []LogPredicate{
				{
					LogValueRef: LogValueRef{Dynamic: false, Offset: 0},
					ValuePredicate: ValuePredicate{
						Op:       BytesEq,
						ByteArgs: [][]byte{topic.Bytes()},
					},
				},
			},
		}
		assert.NilError(t, def.Validate(), "definition for %s must validate", name)
		defBytes := def.MarshalBytes()

		identityPrefix := make([]byte, 32)
		identityPrefix[31] = byte(i + 1)
		identity := crypto.Keccak256(
			append(append(identityPrefix, auth.From.Bytes()...), defBytes...),
		)

		_, err = queries.InsertEventTriggerRegisteredEvent(ctx, database.InsertEventTriggerRegisteredEventParams{
			BlockNumber:           1,
			BlockHash:             []byte{0x01},
			TxIndex:               0,
			LogIndex:              int64(i),
			Eon:                   7,
			IdentityPrefix:        identityPrefix,
			Sender:                auth.From.Hex(),
			Definition:            defBytes,
			ExpirationBlockNumber: 1_000_000,
			Identity:              identity,
		})
		assert.NilError(t, err, "insert trigger row for %s", name)
	}

	// Sanity: verify the rows are considered active.
	active, err := queries.GetActiveEventTriggerRegisteredEvents(ctx, 1)
	assert.NilError(t, err)
	assert.Equal(t, len(active), numTriggers, "expected %d active triggers", numTriggers)

	tp := NewTriggerProcessor(rpc.Client, dbpool)

	// SyncStartBlockNumber=0 means the first Sync will request the full
	// [1..tipBlock] range; DefaultMaxRequestBlockRange (10000) is large
	// enough that this is one range.
	syncer, err := NewMultiEventSyncer(
		dbpool,
		rpc.Client,
		0,
		[]EventProcessor{tp},
	)
	assert.NilError(t, err)

	// Reset the counter so only the Sync-driven traffic is measured.
	rpc.Reset()

	header, err := rpc.Client.HeaderByNumber(ctx, new(big.Int).SetUint64(tipBlock))
	assert.NilError(t, err)

	err = syncer.Sync(ctx, header)
	assert.NilError(t, err)

	counts := rpc.Counts()
	t.Logf(
		"MultiEventSyncer S2: tipBlock=%d numTriggers=%d",
		tipBlock, numTriggers,
	)
	rpc.LogCounts(t, "MultiEventSyncer S2 RPC counts")
	t.Logf("MultiEventSyncer S2 headline: eth_getLogs=%d eth_getBlockByNumber=%d",
		counts["eth_getLogs"], counts["eth_getBlockByNumber"])

	// Confirm that at least one FiredTrigger was recorded so we know the
	// pipeline actually processed logs and the RPC counts are not artefacts
	// of an aborted run.
	fired, err := countFiredTriggers(ctx, dbpool)
	assert.NilError(t, err)
	assert.Assert(t, fired > 0, "expected at least one fired trigger, got %d", fired)
}

func countFiredTriggers(ctx context.Context, dbpool *pgxpool.Pool) (int, error) {
	var n int
	err := dbpool.QueryRow(ctx, "SELECT COUNT(*) FROM fired_triggers").Scan(&n)
	return n, err
}
