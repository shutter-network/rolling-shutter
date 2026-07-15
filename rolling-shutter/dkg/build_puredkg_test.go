package dkg

import (
	"context"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/jackc/pgx/v4"
	"gotest.tools/v3/assert"

	"github.com/shutter-network/shutter/shlib/puredkg"

	obskeyperdb "github.com/shutter-network/rolling-shutter/rolling-shutter/chainobserver/db/keyper"
	corekeyperdb "github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/database"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/shdb"
)

// loadKeyperSetParams looks up keypers/ownIndex/threshold for testKsi via the
// observer-db, matching the pool query processDKG runs before opening the
// buildPureDKG read tx. Tests call this once and pass the values into
// buildPureDKG so the function itself performs no observer-db reads.
func loadKeyperSetParams(ctx context.Context, t *testing.T, env *dkgTestEnv) ([]common.Address, uint64, uint64) {
	t.Helper()
	obsQueries := obskeyperdb.New(env.dbpool)
	keyperSet, err := obsQueries.GetKeyperSetByKeyperConfigIndex(ctx, testKsi)
	assert.NilError(t, err)
	ownIndex, err := keyperSet.GetIndex(env.mgr.cfg.OwnAddress)
	assert.NilError(t, err)
	keypers, err := shdb.DecodeAddresses(keyperSet.Keypers)
	assert.NilError(t, err)
	return keypers, ownIndex, uint64(keyperSet.Threshold) //nolint:gosec // G115: threshold is a small positive integer
}

// TestBuildPureDKGReturnsExpectedPhasePerBlockPhase asserts the per-phase
// reconstruction contract: each `blockPhase` input must leave the returned
// puredkg one step short of the phase the corresponding maybe-function will
// transition into (so `StartPhaseN` advances cleanly without tripping
// puredkg's `setPhase` invariant).
func TestBuildPureDKGReturnsExpectedPhasePerBlockPhase(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	ctx := context.Background()
	env := setupDKGTestEnv(ctx, t)

	err := env.runMaybe(ctx, PhaseDealing)
	assert.NilError(t, err)
	keypers, ownIndex, threshold := loadKeyperSetParams(ctx, t, env)

	cases := []struct {
		name      string
		phase     Phase
		wantPhase puredkg.Phase
	}{
		{"PhaseDealing", PhaseDealing, puredkg.Off},
		{"PhaseAccusing", PhaseAccusing, puredkg.Dealing},
		{"PhaseApologizing", PhaseApologizing, puredkg.Accusing},
		{"PhaseFinalizing", PhaseFinalizing, puredkg.Apologizing},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var pure *puredkg.PureDKG
			err := env.dbpool.BeginFunc(ctx, func(tx pgx.Tx) error {
				p, err := env.mgr.buildPureDKG(ctx, tx, testKsi, testRetry, tc.phase, keypers, ownIndex, threshold)
				pure = p
				return err
			})
			assert.NilError(t, err)
			assert.Assert(t, pure != nil, "puredkg should be returned for member at phase %s", tc.name)
			assert.Equal(t, tc.wantPhase, pure.Phase)
		})
	}
}

// TestBuildPureDKGNilWithoutInitialStateForNonDealingPhases asserts the
// "Keyper never dealt" branch: without a `dkg_initial_states` row,
// non-Dealing phases return nil so the caller skips silently.
func TestBuildPureDKGNilWithoutInitialStateForNonDealingPhases(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	ctx := context.Background()
	env := setupDKGTestEnv(ctx, t)
	keypers, ownIndex, threshold := loadKeyperSetParams(ctx, t, env)

	for _, p := range []Phase{PhaseAccusing, PhaseApologizing, PhaseFinalizing} {
		t.Run(p.String(), func(t *testing.T) {
			var pure *puredkg.PureDKG
			err := env.dbpool.BeginFunc(ctx, func(tx pgx.Tx) error {
				got, err := env.mgr.buildPureDKG(ctx, tx, testKsi, testRetry, p, keypers, ownIndex, threshold)
				pure = got
				return err
			})
			assert.NilError(t, err)
			assert.Assert(t, pure == nil, "expected nil puredkg for %s without initial state", p)
		})
	}
}

// TestBuildPureDKGDealingDoesNotRequireInitialState asserts the Dealing
// branch builds a fresh PureDKG regardless of whether `dkg_initial_states`
// has a row — `maybeDeal` is what populates it.
func TestBuildPureDKGDealingDoesNotRequireInitialState(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	ctx := context.Background()
	env := setupDKGTestEnv(ctx, t)

	// Sanity: no initial state row exists yet.
	coreQueries := corekeyperdb.New(env.dbpool)
	_, err := coreQueries.GetDKGInitialState(ctx, corekeyperdb.GetDKGInitialStateParams{
		KeyperSetIndex: testKsi,
		RetryCounter:   testRetry,
	})
	assert.Assert(t, err != nil)

	keypers, ownIndex, threshold := loadKeyperSetParams(ctx, t, env)
	var pure *puredkg.PureDKG
	err = env.dbpool.BeginFunc(ctx, func(tx pgx.Tx) error {
		p, err := env.mgr.buildPureDKG(ctx, tx, testKsi, testRetry, PhaseDealing, keypers, ownIndex, threshold)
		pure = p
		return err
	})
	assert.NilError(t, err)
	assert.Assert(t, pure != nil)
	assert.Equal(t, puredkg.Off, pure.Phase)
}
