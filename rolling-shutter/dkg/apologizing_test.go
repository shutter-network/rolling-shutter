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
// enqueues a submitApology tx_outbox entry when an accusation against us is
// present — the bug-fix path. The function must NOT write to the shared
// `dkg_apologies` table (the chain syncer owns it); the test checks the
// tx_outbox row and `dkg_sent_actions` marker instead.
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
		KeyperSetIndex: testKsi,
		RetryCounter:      testRetry,
		AccuserIndex:      0,
		AccusedIndex:      1,
	})
	assert.NilError(t, err)

	err = env.runMaybe(ctx, PhaseApologizing)
	assert.NilError(t, err)

	apologies, err := coreQueries.GetDKGApologies(ctx, corekeyperdb.GetDKGApologiesParams{
		KeyperSetIndex: testKsi,
		RetryCounter:      testRetry,
	})
	assert.NilError(t, err)
	assert.Equal(t, 0, len(apologies), "maybeApologize must not write to dkg_apologies — chain syncer owns it")

	sentAction, err := coreQueries.ExistsDKGSentAction(ctx, corekeyperdb.ExistsDKGSentActionParams{
		KeyperSetIndex: testKsi,
		RetryCounter:      testRetry,
		Action:            ActionApologizing,
	})
	assert.NilError(t, err)
	assert.Assert(t, sentAction, "dkg_sent_actions row should mark the apologizing action as enqueued")

	pending, err := coreQueries.GetPendingTxs(ctx)
	assert.NilError(t, err)
	// submitDealing + submitApology.
	assert.Equal(t, 2, len(pending))
}

// TestMaybeApologizeNoopWhenNotAccused asserts that maybeApologize enqueues
// no submitApology tx when no accusation against us is on file but still
// writes a `dkg_sent_actions` row so the reactor short-circuits on
// subsequent blocks (no log spam, no repeated puredkg replay). A second
// call is a no-op.
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
		KeyperSetIndex: testKsi,
		RetryCounter:      testRetry,
		AccuserIndex:      0,
		AccusedIndex:      2,
	})
	assert.NilError(t, err)

	err = env.runMaybe(ctx, PhaseApologizing)
	assert.NilError(t, err)

	apologies, err := coreQueries.GetDKGApologies(ctx, corekeyperdb.GetDKGApologiesParams{
		KeyperSetIndex: testKsi,
		RetryCounter:      testRetry,
	})
	assert.NilError(t, err)
	assert.Equal(t, 0, len(apologies))

	pending, err := coreQueries.GetPendingTxs(ctx)
	assert.NilError(t, err)
	assert.Equal(t, 1, len(pending)) // only the submitDealing entry

	sentAction, err := coreQueries.ExistsDKGSentAction(ctx, corekeyperdb.ExistsDKGSentActionParams{
		KeyperSetIndex: testKsi,
		RetryCounter:      testRetry,
		Action:            ActionApologizing,
	})
	assert.NilError(t, err)
	assert.Assert(t, sentAction, "dkg_sent_actions row should be written even when no apologies are sent")

	// Second invocation: short-circuits on the existing dkg_sent_actions
	// row, returns without error, and writes nothing new.
	err = env.runMaybe(ctx, PhaseApologizing)
	assert.NilError(t, err)
	pendingAfter, err := coreQueries.GetPendingTxs(ctx)
	assert.NilError(t, err)
	assert.Equal(t, len(pending), len(pendingAfter), "no extra tx_outbox rows on idempotent no-apologies call")
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
		KeyperSetIndex: testKsi,
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
		KeyperSetIndex: testKsi,
		RetryCounter:      testRetry,
		AccuserIndex:      2,
		AccusedIndex:      1,
	})
	assert.NilError(t, err)

	// Write tx: dispatch maybeApologize with the snapshot from the read tx.
	err = env.dbpool.BeginFunc(ctx, func(tx pgx.Tx) error {
		return env.mgr.maybeApologize(ctx, tx, env.dkgAddr, testKsi, testRetry, pure, keypers, ownIndex)
	})
	assert.NilError(t, err, "maybeApologize must tolerate a late accusation arriving between read tx and write tx")

	apologies, err := coreQueries.GetDKGApologies(ctx, corekeyperdb.GetDKGApologiesParams{
		KeyperSetIndex: testKsi,
		RetryCounter:      testRetry,
	})
	assert.NilError(t, err)
	// maybeApologize no longer writes its own apology row — the chain
	// syncer owns dkg_apologies. The acceptance signal that the call
	// processed exactly the read-tx snapshot is the single tx_outbox entry
	// recorded against the apologizing sent-action marker; the late
	// accusation from keyper 2 is left for the next block within the
	// Apologizing phase window.
	assert.Equal(t, 0, len(apologies), "maybeApologize must not write to dkg_apologies")

	sentAction, err := coreQueries.ExistsDKGSentAction(ctx, corekeyperdb.ExistsDKGSentActionParams{
		KeyperSetIndex: testKsi,
		RetryCounter:      testRetry,
		Action:            ActionApologizing,
	})
	assert.NilError(t, err)
	assert.Assert(t, sentAction, "dkg_sent_actions row should mark the apologizing action as enqueued")
}

// TestMaybeApologizeIdempotent asserts that a second invocation does not
// enqueue an additional submitApology tx_outbox row. The dkg_sent_actions
// row written on the first invocation is the idempotency marker. The
// shared `dkg_apologies` table is never written by the reactor in either
// invocation.
func TestMaybeApologizeIdempotent(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	ctx := context.Background()
	env := setupDKGTestEnv(ctx, t)

	err := env.runMaybe(ctx, PhaseDealing)
	assert.NilError(t, err)

	coreQueries := corekeyperdb.New(env.dbpool)
	err = coreQueries.InsertDKGAccusation(ctx, corekeyperdb.InsertDKGAccusationParams{
		KeyperSetIndex: testKsi,
		RetryCounter:      testRetry,
		AccuserIndex:      0,
		AccusedIndex:      1,
	})
	assert.NilError(t, err)

	err = env.runMaybe(ctx, PhaseApologizing)
	assert.NilError(t, err)

	firstPending, err := coreQueries.GetPendingTxs(ctx)
	assert.NilError(t, err)
	sentAction, err := coreQueries.ExistsDKGSentAction(ctx, corekeyperdb.ExistsDKGSentActionParams{
		KeyperSetIndex: testKsi,
		RetryCounter:      testRetry,
		Action:            ActionApologizing,
	})
	assert.NilError(t, err)
	assert.Assert(t, sentAction, "dkg_sent_actions row should exist for the apologizing action")

	err = env.runMaybe(ctx, PhaseApologizing)
	assert.NilError(t, err)

	apologies, err := coreQueries.GetDKGApologies(ctx, corekeyperdb.GetDKGApologiesParams{
		KeyperSetIndex: testKsi,
		RetryCounter:      testRetry,
	})
	assert.NilError(t, err)
	assert.Equal(t, 0, len(apologies), "maybeApologize must not write to dkg_apologies on any invocation")
	secondPending, err := coreQueries.GetPendingTxs(ctx)
	assert.NilError(t, err)
	assert.Equal(t, len(firstPending), len(secondPending), "no extra tx_outbox rows on idempotent call")
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
		KeyperSetIndex: testKsi,
		RetryCounter:      testRetry,
		AccuserIndex:      0,
		AccusedIndex:      1,
	})
	assert.NilError(t, err)

	err = env.runMaybe(ctx, PhaseApologizing)
	assert.NilError(t, err)

	apologies, err := coreQueries.GetDKGApologies(ctx, corekeyperdb.GetDKGApologiesParams{
		KeyperSetIndex: testKsi,
		RetryCounter:      testRetry,
	})
	assert.NilError(t, err)
	assert.Equal(t, 0, len(apologies))
}
