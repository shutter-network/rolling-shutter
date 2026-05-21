package dkg

import (
	"context"
	"testing"

	"gotest.tools/v3/assert"

	corekeyperdb "github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/database"
)

// TestMaybeApologizeEnqueuesApologyWhenAccused asserts that maybeApologize
// writes one apology row plus a tx_outbox entry when an accusation against
// us is present — the bug-fix path. Without our local refactor the
// puredkg's Phase would remain Off after replay and the function would
// silently skip.
func TestMaybeApologizeEnqueuesApologyWhenAccused(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	ctx := context.Background()
	env := setupDKGTestEnv(ctx, t)

	err := env.mgr.maybeDeal(ctx, env.dkgAddr, testKsi, testRetry)
	assert.NilError(t, err)

	// Keyper 0 accuses us (keyper 1). The accusation row drives StartPhase3
	// to emit one apology addressed back to keyper 0.
	coreQueries := corekeyperdb.New(env.dbpool)
	err = coreQueries.InsertDKGAccusation(ctx, corekeyperdb.InsertDKGAccusationParams{
		KeyperConfigIndex: testKsi,
		RetryCounter:      testRetry,
		AccuserIndex:      0,
		AccusedIndex:      1,
	})
	assert.NilError(t, err)

	err = env.mgr.maybeApologize(ctx, env.dkgAddr, testKsi, testRetry)
	assert.NilError(t, err)

	apologies, err := coreQueries.GetDKGApologies(ctx, corekeyperdb.GetDKGApologiesParams{
		KeyperConfigIndex: testKsi,
		RetryCounter:      testRetry,
	})
	assert.NilError(t, err)
	assert.Equal(t, 1, len(apologies))
	assert.Equal(t, int64(1), apologies[0].ApologizerIndex)
	assert.Equal(t, int64(0), apologies[0].AccuserIndex)
	assert.Assert(t, len(apologies[0].PolyEval) > 0, "apology should carry a non-empty poly eval")

	pending, err := coreQueries.GetPendingTxs(ctx)
	assert.NilError(t, err)
	// submitDealing + submitApology.
	assert.Equal(t, 2, len(pending))
}

// TestMaybeApologizeNoopWhenNotAccused asserts that maybeApologize writes
// nothing when no accusation against us is on file — the silent-success
// path.
func TestMaybeApologizeNoopWhenNotAccused(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	ctx := context.Background()
	env := setupDKGTestEnv(ctx, t)
	err := env.mgr.maybeDeal(ctx, env.dkgAddr, testKsi, testRetry)
	assert.NilError(t, err)

	// Accusation against another keyper (0 → 2), not us.
	coreQueries := corekeyperdb.New(env.dbpool)
	err = coreQueries.InsertDKGAccusation(ctx, corekeyperdb.InsertDKGAccusationParams{
		KeyperConfigIndex: testKsi,
		RetryCounter:      testRetry,
		AccuserIndex:      0,
		AccusedIndex:      2,
	})
	assert.NilError(t, err)

	err = env.mgr.maybeApologize(ctx, env.dkgAddr, testKsi, testRetry)
	assert.NilError(t, err)

	apologies, err := coreQueries.GetDKGApologies(ctx, corekeyperdb.GetDKGApologiesParams{
		KeyperConfigIndex: testKsi,
		RetryCounter:      testRetry,
	})
	assert.NilError(t, err)
	assert.Equal(t, 0, len(apologies))

	pending, err := coreQueries.GetPendingTxs(ctx)
	assert.NilError(t, err)
	assert.Equal(t, 1, len(pending)) // only the submitDealing entry
}

// TestMaybeApologizeNoopWithoutInitialState asserts that maybeApologize
// returns silently when `dkg_initial_states` has no row — buildPureDKG
// returns nil and the action is skipped.
func TestMaybeApologizeNoopWithoutInitialState(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	ctx := context.Background()
	env := setupDKGTestEnv(ctx, t)

	coreQueries := corekeyperdb.New(env.dbpool)
	err := coreQueries.InsertDKGAccusation(ctx, corekeyperdb.InsertDKGAccusationParams{
		KeyperConfigIndex: testKsi,
		RetryCounter:      testRetry,
		AccuserIndex:      0,
		AccusedIndex:      1,
	})
	assert.NilError(t, err)

	err = env.mgr.maybeApologize(ctx, env.dkgAddr, testKsi, testRetry)
	assert.NilError(t, err)

	apologies, err := coreQueries.GetDKGApologies(ctx, corekeyperdb.GetDKGApologiesParams{
		KeyperConfigIndex: testKsi,
		RetryCounter:      testRetry,
	})
	assert.NilError(t, err)
	assert.Equal(t, 0, len(apologies))
}
