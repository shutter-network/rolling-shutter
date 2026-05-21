package dkg

import (
	"context"
	"testing"

	"gotest.tools/v3/assert"

	corekeyperdb "github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/database"
)

// TestMaybeFinalizeWritesSentActionAndDKGResult exercises the success path of
// maybeFinalize when all three dealers are honest:
//
//  1. First invocation writes a dkg_result success row, a tx_outbox row for
//     submitSuccessVote, and a dkg_sent_actions row for the finalizing
//     action.
//  2. Second invocation is a no-op — ExistsDKGResultSuccess short-circuits
//     the call so no new rows are inserted.
func TestMaybeFinalizeWritesSentActionAndDKGResult(t *testing.T) {
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
	result, err := coreQueries.GetDKGResult(ctx, testKsi)
	assert.NilError(t, err)
	assert.Assert(t, result.Success, "dkg_result row should mark success after maybeFinalize")

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

	// Second invocation: ExistsDKGResultSuccess short-circuits; no new rows.
	err = env.runMaybe(ctx, PhaseFinalizing)
	assert.NilError(t, err)

	secondPending, err := coreQueries.GetPendingTxs(ctx)
	assert.NilError(t, err)
	assert.Equal(t, len(firstPending), len(secondPending), "no extra tx_outbox rows on idempotent call")
}
