package dkg

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v4"
	"gotest.tools/v3/assert"

	"github.com/shutter-network/shutter/shlib/puredkg"

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

	err := env.runMaybe(ctx, PhaseDealing)
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

	err = env.runMaybe(ctx, PhaseApologizing)
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
	err := env.runMaybe(ctx, PhaseDealing)
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

	err = env.runMaybe(ctx, PhaseApologizing)
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

// TestMaybeApologizeToleratesAccusationBetweenReadAndWriteTx asserts the
// split-transaction acceptance criterion from the read/write tx slice:
// when a new Accusation row is inserted between the buildPureDKG read tx
// closing and the write tx opening, maybeApologize must not error. The
// apology produced reflects the snapshot the read tx saw — the late
// accusation is left for the next block to cover.
func TestMaybeApologizeToleratesAccusationBetweenReadAndWriteTx(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	ctx := context.Background()
	env := setupDKGTestEnv(ctx, t)

	err := env.runMaybe(ctx, PhaseDealing)
	assert.NilError(t, err)

	// Keyper 0's accusation is the row the read tx will see.
	coreQueries := corekeyperdb.New(env.dbpool)
	err = coreQueries.InsertDKGAccusation(ctx, corekeyperdb.InsertDKGAccusationParams{
		KeyperConfigIndex: testKsi,
		RetryCounter:      testRetry,
		AccuserIndex:      0,
		AccusedIndex:      1,
	})
	assert.NilError(t, err)

	keypers, ownIndex, threshold := loadKeyperSetParams(ctx, t, env)

	// Read tx: rebuild puredkg with the single accusation visible.
	var pure *puredkg.PureDKG
	err = env.dbpool.BeginFunc(ctx, func(tx pgx.Tx) error {
		p, err := env.mgr.buildPureDKG(ctx, tx, testKsi, testRetry, PhaseApologizing, keypers, ownIndex, threshold)
		pure = p
		return err
	})
	assert.NilError(t, err)
	assert.Assert(t, pure != nil)

	// Race: keyper 2 accuses us after the read tx closed. The write tx
	// will see this row in the DB but maybeApologize works off the stale
	// puredkg snapshot returned above — it must not error.
	err = coreQueries.InsertDKGAccusation(ctx, corekeyperdb.InsertDKGAccusationParams{
		KeyperConfigIndex: testKsi,
		RetryCounter:      testRetry,
		AccuserIndex:      2,
		AccusedIndex:      1,
	})
	assert.NilError(t, err)

	// Write tx: dispatch maybeApologize with the snapshot from the read tx.
	err = env.dbpool.BeginFunc(ctx, func(tx pgx.Tx) error {
		return env.mgr.maybeApologize(ctx, tx, env.dkgAddr, testKsi, testRetry, pure, ownIndex)
	})
	assert.NilError(t, err, "maybeApologize must tolerate a late accusation arriving between read tx and write tx")

	apologies, err := coreQueries.GetDKGApologies(ctx, corekeyperdb.GetDKGApologiesParams{
		KeyperConfigIndex: testKsi,
		RetryCounter:      testRetry,
	})
	assert.NilError(t, err)
	// Exactly one apology — for accuser 0, the only accusation the read tx
	// saw. Keyper 2's accusation arrived too late and is left for the next
	// block within the Apologizing phase window.
	assert.Equal(t, 1, len(apologies))
	assert.Equal(t, int64(1), apologies[0].ApologizerIndex)
	assert.Equal(t, int64(0), apologies[0].AccuserIndex)
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

	err = env.runMaybe(ctx, PhaseApologizing)
	assert.NilError(t, err)

	apologies, err := coreQueries.GetDKGApologies(ctx, corekeyperdb.GetDKGApologiesParams{
		KeyperConfigIndex: testKsi,
		RetryCounter:      testRetry,
	})
	assert.NilError(t, err)
	assert.Equal(t, 0, len(apologies))
}
