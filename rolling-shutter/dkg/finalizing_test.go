package dkg

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v4"
	"gotest.tools/v3/assert"

	corekeyperdb "github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/database"
)

// TestMaybeFinalizeWritesSentActionNotDKGResult exercises the success path of
// maybeFinalize when all three dealers are honest:
//
//  1. First invocation enqueues a submitSuccessVote tx_outbox row and writes a
//     dkg_sent_actions row for ActionFinalizing. It must NOT write dkg_result —
//     that row is the exclusive responsibility of HandleDKGSuccess.
//  2. Second invocation is a no-op — ExistsDKGSentAction short-circuits so no
//     new rows are inserted.
func TestMaybeFinalizeWritesSentActionNotDKGResult(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	ctx := context.Background()
	env := setupDKGTestEnv(ctx, t)

	err := env.runMaybe(ctx, PhaseDealing)
	assert.NilError(t, err)
	env.insertForeignDealing(ctx, t, 0)
	env.insertForeignDealing(ctx, t, 2)

	err = env.runMaybe(ctx, PhaseFinalizing)
	assert.NilError(t, err)

	coreQueries := corekeyperdb.New(env.dbpool)

	// dkg_result must NOT be present — HandleDKGSuccess owns that write.
	_, err = coreQueries.GetDKGResult(ctx, testKsi)
	assert.ErrorContains(t, err, "no rows", "maybeFinalize must not write dkg_result")

	firstPending, err := coreQueries.GetPendingTxs(ctx)
	assert.NilError(t, err)
	// submitDealing + submitSuccessVote.
	assert.Equal(t, 2, len(firstPending))

	sentAction, err := coreQueries.ExistsDKGSentAction(ctx, corekeyperdb.ExistsDKGSentActionParams{
		KeyperConfigIndex: testKsi,
		RetryCounter:      testRetry,
		Action:            ActionFinalizing,
	})
	assert.NilError(t, err)
	assert.Assert(t, sentAction, "dkg_sent_actions row should exist for the finalizing action")

	// Second invocation: ExistsDKGSentAction short-circuits; no new rows.
	err = env.runMaybe(ctx, PhaseFinalizing)
	assert.NilError(t, err)

	secondPending, err := coreQueries.GetPendingTxs(ctx)
	assert.NilError(t, err)
	assert.Equal(t, len(firstPending), len(secondPending), "no extra tx_outbox rows on idempotent call")
}

// TestHandleDKGSuccessRetry1AfterRetry0Failed asserts the key correctness
// property: a keyper that locally finalized retry 0 (sent a success vote) is
// not blocked from participating in retry 1 when retry 0 fails on chain, and
// HandleDKGSuccess for retry 1 populates dkg_result.pure_result.
func TestHandleDKGSuccessRetry1AfterRetry0Failed(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	ctx := context.Background()
	env := setupDKGTestEnv(ctx, t)

	const retry0 int64 = 0
	const retry1 int64 = 1

	// --- Retry 0 ---
	// Local keyper deals and all three keypers' data is available.
	err := env.runMaybeRetry(ctx, PhaseDealing, retry0)
	assert.NilError(t, err)
	env.insertForeignDealingRetry(ctx, t, 0, retry0)
	env.insertForeignDealingRetry(ctx, t, 2, retry0)

	// maybeFinalize enqueues the success vote for retry 0.
	err = env.runMaybeRetry(ctx, PhaseFinalizing, retry0)
	assert.NilError(t, err)

	coreQueries := corekeyperdb.New(env.dbpool)

	// Retry 0 failed on chain: no dkg_result row exists yet.
	_, err = coreQueries.GetDKGResult(ctx, testKsi)
	assert.ErrorContains(t, err, "no rows", "no dkg_result row before HandleDKGSuccess")

	// --- Retry 1 ---
	// Local keyper deals again (new dkg_initial_states row for retry 1).
	err = env.runMaybeRetry(ctx, PhaseDealing, retry1)
	assert.NilError(t, err)
	env.insertForeignDealingRetry(ctx, t, 0, retry1)
	env.insertForeignDealingRetry(ctx, t, 2, retry1)

	// maybeFinalize for retry 1 must succeed — ExistsDKGSentAction checks
	// (ksi, retry=1, "finalizing") which has no row yet.
	err = env.runMaybeRetry(ctx, PhaseFinalizing, retry1)
	assert.NilError(t, err)

	sentAction1, err := coreQueries.ExistsDKGSentAction(ctx, corekeyperdb.ExistsDKGSentActionParams{
		KeyperConfigIndex: testKsi,
		RetryCounter:      retry1,
		Action:            ActionFinalizing,
	})
	assert.NilError(t, err)
	assert.Assert(t, sentAction1, "dkg_sent_actions row should exist for retry 1 finalizing")

	// Chain emits DKGSucceeded for retry 1: HandleDKGSuccess must write
	// dkg_result with a non-nil pure_result.
	err = env.dbpool.BeginFunc(ctx, func(tx pgx.Tx) error {
		return env.mgr.HandleDKGSuccess(ctx, tx, testKsi, retry1)
	})
	assert.NilError(t, err)

	result, err := coreQueries.GetDKGResult(ctx, testKsi)
	assert.NilError(t, err)
	assert.Assert(t, result.Success, "dkg_result row must mark success")
	assert.Assert(t, len(result.PureResult) > 0, "pure_result must be populated for a participating keyper")

	// Idempotency: second HandleDKGSuccess call is a no-op.
	err = env.dbpool.BeginFunc(ctx, func(tx pgx.Tx) error {
		return env.mgr.HandleDKGSuccess(ctx, tx, testKsi, retry1)
	})
	assert.NilError(t, err)
}
