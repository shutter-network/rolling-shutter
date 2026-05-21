package dkg

import (
	"context"
	"crypto/rand"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/ecies"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"gotest.tools/v3/assert"

	"github.com/shutter-network/shutter/shlib/puredkg"

	obskeyperdb "github.com/shutter-network/rolling-shutter/rolling-shutter/chainobserver/db/keyper"
	corekeyperdb "github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/database"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/testsetup"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/shdb"
)

// dkgTestEnv bundles the artefacts a maybeAccuse / maybeApologize test
// needs: a three-keyper set with the local Manager at index 1, all three
// ECIES keys registered, and a ready-to-call `mgr.maybeDeal` /
// `mgr.maybeAccuse` etc.
type dkgTestEnv struct {
	dbpool   *pgxpool.Pool
	mgr      *Manager
	dkgAddr  common.Address
	ownECIES *ecies.PrivateKey
}

const (
	testKsi   int64 = 7
	testRetry int64 = 0
)

func setupDKGTestEnv(ctx context.Context, t *testing.T) *dkgTestEnv {
	t.Helper()
	dbpool, dbclose := testsetup.NewTestDBPool(ctx, t, corekeyperdb.Definition)
	t.Cleanup(dbclose)

	ownECDSA, err := crypto.GenerateKey()
	assert.NilError(t, err)
	other0, err := crypto.GenerateKey()
	assert.NilError(t, err)
	other2, err := crypto.GenerateKey()
	assert.NilError(t, err)
	ownAddr := crypto.PubkeyToAddress(ownECDSA.PublicKey)
	other0Addr := crypto.PubkeyToAddress(other0.PublicKey)
	other2Addr := crypto.PubkeyToAddress(other2.PublicKey)
	keypers := []common.Address{other0Addr, ownAddr, other2Addr}

	obsQueries := obskeyperdb.New(dbpool)
	err = obsQueries.InsertKeyperSet(ctx, obskeyperdb.InsertKeyperSetParams{
		KeyperConfigIndex:     testKsi,
		ActivationBlockNumber: 0,
		Keypers:               shdb.EncodeAddresses(keypers),
		Threshold:             2,
	})
	assert.NilError(t, err)

	coreQueries := corekeyperdb.New(dbpool)
	for _, kp := range []struct {
		addr common.Address
		key  *ecies.PrivateKey
	}{
		{addr: other0Addr, key: ecies.ImportECDSA(other0)},
		{addr: ownAddr, key: ecies.ImportECDSA(ownECDSA)},
		{addr: other2Addr, key: ecies.ImportECDSA(other2)},
	} {
		err = coreQueries.UpsertECIESKey(ctx, corekeyperdb.UpsertECIESKeyParams{
			KeyperAddress:  shdb.EncodeAddress(kp.addr),
			EciesPublicKey: shdb.EncodeEciesPublicKey(&kp.key.PublicKey),
		})
		assert.NilError(t, err)
	}

	dkgAddr := common.HexToAddress("0xd0000000000000000000000000000000000000aa")
	mgr := New(Config{
		DBPool:            dbpool,
		OwnAddress:        ownAddr,
		ECIESPrivateKey:   ecies.ImportECDSA(ownECDSA),
		ECIESRegistryAddr: common.HexToAddress("0xe0000000000000000000000000000000000000bb"),
	})

	return &dkgTestEnv{
		dbpool:   dbpool,
		mgr:      mgr,
		dkgAddr:  dkgAddr,
		ownECIES: ecies.ImportECDSA(ownECDSA),
	}
}

// runMaybe wraps a single dispatch the way `handleEon` does: look up the
// keyper set, open a read transaction for `buildPureDKG`, then open a
// separate write transaction for the matching maybe-function. Returns nil
// silently if the manager would not participate (non-member or no initial
// state for non-Dealing phases).
func (env *dkgTestEnv) runMaybe(ctx context.Context, phase Phase) error {
	obsQueries := obskeyperdb.New(env.dbpool)
	keyperSet, err := obsQueries.GetKeyperSetByKeyperConfigIndex(ctx, testKsi)
	if err != nil {
		return err
	}
	ownIndex, err := keyperSet.GetIndex(env.mgr.cfg.OwnAddress)
	if err != nil {
		return nil
	}
	keypers, err := shdb.DecodeAddresses(keyperSet.Keypers)
	if err != nil {
		return err
	}
	threshold := uint64(keyperSet.Threshold)

	var pure *puredkg.PureDKG
	err = env.dbpool.BeginFunc(ctx, func(tx pgx.Tx) error {
		p, err := env.mgr.buildPureDKG(ctx, tx, testKsi, testRetry, phase, keypers, ownIndex, threshold)
		if err != nil {
			return err
		}
		pure = p
		return nil
	})
	if err != nil {
		return err
	}
	if pure == nil {
		return nil
	}
	return env.dbpool.BeginFunc(ctx, func(tx pgx.Tx) error {
		switch phase {
		case PhaseDealing:
			return env.mgr.maybeDeal(ctx, tx, env.dkgAddr, testKsi, testRetry, pure, keypers, ownIndex)
		case PhaseAccusing:
			return env.mgr.maybeAccuse(ctx, tx, env.dkgAddr, testKsi, testRetry, pure, ownIndex)
		case PhaseApologizing:
			return env.mgr.maybeApologize(ctx, tx, env.dkgAddr, testKsi, testRetry, pure, ownIndex)
		case PhaseFinalizing:
			return env.mgr.maybeFinalize(ctx, tx, env.dkgAddr, testKsi, testRetry, pure, ownIndex)
		}
		return nil
	})
}

// insertForeignDealing simulates `dealerIdx` having dealt: insert their poly
// commitment and the encrypted-for-us PolyEval row. Both are derived from a
// fresh puredkg run by the test in-process.
func (env *dkgTestEnv) insertForeignDealing(ctx context.Context, t *testing.T, dealerIdx uint64) {
	t.Helper()
	const eon = uint64(testKsi)
	p := puredkg.NewPureDKG(eon, 3, 2, dealerIdx)
	commit, evals, err := p.StartPhase1Dealing()
	assert.NilError(t, err)

	coreQueries := corekeyperdb.New(env.dbpool)
	err = coreQueries.InsertDKGPolyCommitment(ctx, corekeyperdb.InsertDKGPolyCommitmentParams{
		KeyperConfigIndex: testKsi,
		RetryCounter:      testRetry,
		KeyperIndex:       int64(dealerIdx),
		Commitment:        commit.Gammas.Marshal(),
	})
	assert.NilError(t, err)

	for _, ev := range evals {
		if ev.Receiver != 1 {
			continue
		}
		ciphertext, err := ecies.Encrypt(rand.Reader, &env.ownECIES.PublicKey, ev.Eval.Bytes(), nil, nil)
		assert.NilError(t, err)
		err = coreQueries.InsertDKGPolyEval(ctx, corekeyperdb.InsertDKGPolyEvalParams{
			KeyperConfigIndex: testKsi,
			RetryCounter:      testRetry,
			SenderIndex:       int64(dealerIdx),
			ReceiverIndex:     1,
			EncryptedEval:     ciphertext,
		})
		assert.NilError(t, err)
	}
}

// TestMaybeAccuseEnqueuesAccusationForMissingDealing asserts that maybeAccuse
// emits one accusation per dealer whose PolyEval is missing — the bug-fix
// path. Without our local refactor `pure.Phase` would remain Off after
// replay and the function would silently skip.
func TestMaybeAccuseEnqueuesAccusationForMissingDealing(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	ctx := context.Background()
	env := setupDKGTestEnv(ctx, t)

	// Local keyper deals — populates dkg_initial_states.
	err := env.runMaybe(ctx, PhaseDealing)
	assert.NilError(t, err)

	// Neither keyper 0 nor keyper 2 deals: their commitment+eval rows are
	// absent. maybeAccuse must accuse both.
	err = env.runMaybe(ctx, PhaseAccusing)
	assert.NilError(t, err)

	coreQueries := corekeyperdb.New(env.dbpool)
	accusations, err := coreQueries.GetDKGAccusations(ctx, corekeyperdb.GetDKGAccusationsParams{
		KeyperConfigIndex: testKsi,
		RetryCounter:      testRetry,
	})
	assert.NilError(t, err)
	assert.Equal(t, 2, len(accusations), "expected one accusation per missing dealer")
	got := map[int64]bool{}
	for _, a := range accusations {
		assert.Equal(t, int64(1), a.AccuserIndex)
		got[a.AccusedIndex] = true
	}
	assert.Assert(t, got[0] && got[2], "expected accusations against keypers 0 and 2")

	pending, err := coreQueries.GetPendingTxs(ctx)
	assert.NilError(t, err)
	// One submitDealing entry from maybeDeal, plus one submitAccusation.
	assert.Equal(t, 2, len(pending))
}

// TestMaybeAccuseNoopWhenAllDealersHonest asserts the silent-success path:
// when every dealer's commitment + eval is present and valid, maybeAccuse
// writes nothing.
func TestMaybeAccuseNoopWhenAllDealersHonest(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	ctx := context.Background()
	env := setupDKGTestEnv(ctx, t)

	err := env.runMaybe(ctx, PhaseDealing)
	assert.NilError(t, err)
	env.insertForeignDealing(ctx, t, 0)
	env.insertForeignDealing(ctx, t, 2)

	err = env.runMaybe(ctx, PhaseAccusing)
	assert.NilError(t, err)

	coreQueries := corekeyperdb.New(env.dbpool)
	accusations, err := coreQueries.GetDKGAccusations(ctx, corekeyperdb.GetDKGAccusationsParams{
		KeyperConfigIndex: testKsi,
		RetryCounter:      testRetry,
	})
	assert.NilError(t, err)
	assert.Equal(t, 0, len(accusations))

	pending, err := coreQueries.GetPendingTxs(ctx)
	assert.NilError(t, err)
	// Only the submitDealing entry from maybeDeal — no submitAccusation.
	assert.Equal(t, 1, len(pending))
}

// TestMaybeAccuseIdempotent asserts that a second invocation does not write
// additional accusation or outbox rows.
func TestMaybeAccuseIdempotent(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	ctx := context.Background()
	env := setupDKGTestEnv(ctx, t)
	err := env.runMaybe(ctx, PhaseDealing)
	assert.NilError(t, err)
	err = env.runMaybe(ctx, PhaseAccusing)
	assert.NilError(t, err)

	coreQueries := corekeyperdb.New(env.dbpool)
	first, err := coreQueries.GetDKGAccusations(ctx, corekeyperdb.GetDKGAccusationsParams{
		KeyperConfigIndex: testKsi,
		RetryCounter:      testRetry,
	})
	assert.NilError(t, err)
	firstPending, err := coreQueries.GetPendingTxs(ctx)
	assert.NilError(t, err)
	sentAction, err := coreQueries.ExistsDKGSentAction(ctx, corekeyperdb.ExistsDKGSentActionParams{
		KeyperConfigIndex: testKsi,
		RetryCounter:      testRetry,
		Action:            ActionAccusing,
	})
	assert.NilError(t, err)
	assert.Assert(t, sentAction, "dkg_sent_actions row should exist for the accusing action")

	err = env.runMaybe(ctx, PhaseAccusing)
	assert.NilError(t, err)

	second, err := coreQueries.GetDKGAccusations(ctx, corekeyperdb.GetDKGAccusationsParams{
		KeyperConfigIndex: testKsi,
		RetryCounter:      testRetry,
	})
	assert.NilError(t, err)
	secondPending, err := coreQueries.GetPendingTxs(ctx)
	assert.NilError(t, err)

	assert.Equal(t, len(first), len(second))
	assert.Equal(t, len(firstPending), len(secondPending))
}
