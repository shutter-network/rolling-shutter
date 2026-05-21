package dkg

import (
	"context"
	"crypto/rand"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/ecies"
	"gotest.tools/v3/assert"

	obskeyperdb "github.com/shutter-network/rolling-shutter/rolling-shutter/chainobserver/db/keyper"
	corekeyperdb "github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/database"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/testsetup"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/shdb"
)

// TestMaybeDealPersistsInitialStateAndIsIdempotent exercises the two
// acceptance criteria for the maybeDeal refactor:
//
//  1. First invocation writes a dkg_initial_states row plus the own poly
//     commitment, the per-receiver poly evals (including the self-eval), and
//     a tx_outbox row for submitDealing.
//  2. Second invocation is a no-op — no new rows are inserted in any of those
//     tables, because the dkg_initial_states row already exists.
func TestMaybeDealPersistsInitialStateAndIsIdempotent(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	ctx := context.Background()

	dbpool, dbclose := testsetup.NewTestDBPool(ctx, t, corekeyperdb.Definition)
	t.Cleanup(dbclose)

	const (
		keyperConfigIndex int64 = 7
		retryCounter      int64 = 0
	)

	// Build a 3-keyper set where we are the keyper at index 1.
	ownECDSA, err := crypto.GenerateKey()
	assert.NilError(t, err)
	ownAddr := crypto.PubkeyToAddress(ownECDSA.PublicKey)
	other0, err := crypto.GenerateKey()
	assert.NilError(t, err)
	other0Addr := crypto.PubkeyToAddress(other0.PublicKey)
	other2, err := crypto.GenerateKey()
	assert.NilError(t, err)
	other2Addr := crypto.PubkeyToAddress(other2.PublicKey)
	keypers := []common.Address{other0Addr, ownAddr, other2Addr}

	obsQueries := obskeyperdb.New(dbpool)
	err = obsQueries.InsertKeyperSet(ctx, obskeyperdb.InsertKeyperSetParams{
		KeyperConfigIndex:     keyperConfigIndex,
		ActivationBlockNumber: 0,
		Keypers:               shdb.EncodeAddresses(keypers),
		Threshold:             2,
	})
	assert.NilError(t, err)

	coreQueries := corekeyperdb.New(dbpool)
	// All three keypers must have an ECIES public key registered — encryption
	// of per-receiver evals (and the self-eval) reads this table.
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
		DKGContractAddr:   dkgAddr,
		ECIESRegistryAddr: common.HexToAddress("0xe0000000000000000000000000000000000000bb"),
	})

	// Sanity: nothing exists yet.
	_, err = coreQueries.GetDKGInitialState(ctx, corekeyperdb.GetDKGInitialStateParams{
		KeyperConfigIndex: keyperConfigIndex,
		RetryCounter:      retryCounter,
	})
	assert.Assert(t, err != nil, "no initial state row expected before maybeDeal")

	// First invocation: writes rows.
	err = mgr.maybeDeal(ctx, dkgAddr, keyperConfigIndex, retryCounter)
	assert.NilError(t, err)

	initial, err := coreQueries.GetDKGInitialState(ctx, corekeyperdb.GetDKGInitialStateParams{
		KeyperConfigIndex: keyperConfigIndex,
		RetryCounter:      retryCounter,
	})
	assert.NilError(t, err)
	assert.Assert(t, len(initial.PuredkgBytes) > 0, "initial state blob should not be empty")
	// Blob must decode back to a usable PureDKG with the self-eval populated.
	roundtrip, err := shdb.DecodePureDKG(initial.PuredkgBytes)
	assert.NilError(t, err)
	assert.Equal(t, uint64(1), roundtrip.Keyper)
	assert.Assert(t, roundtrip.Evals[1] != nil, "self-eval should be set on persisted state")

	commitments, err := coreQueries.GetDKGPolyCommitments(ctx, corekeyperdb.GetDKGPolyCommitmentsParams{
		KeyperConfigIndex: keyperConfigIndex,
		RetryCounter:      retryCounter,
	})
	assert.NilError(t, err)
	assert.Equal(t, 1, len(commitments))
	assert.Equal(t, int64(1), commitments[0].KeyperIndex)

	evals, err := coreQueries.GetDKGPolyEvals(ctx, corekeyperdb.GetDKGPolyEvalsParams{
		KeyperConfigIndex: keyperConfigIndex,
		RetryCounter:      retryCounter,
	})
	assert.NilError(t, err)
	// One row per receiver, including the self-row: 3 keypers → 3 rows.
	assert.Equal(t, 3, len(evals))

	pending, err := coreQueries.GetPendingTxs(ctx)
	assert.NilError(t, err)
	assert.Equal(t, 1, len(pending))
	assert.Equal(t, dkgAddr.Hex(), pending[0].ToAddress)
	firstOutboxID := pending[0].ID

	// Second invocation: idempotent — no new rows in any tracked table.
	err = mgr.maybeDeal(ctx, dkgAddr, keyperConfigIndex, retryCounter)
	assert.NilError(t, err)

	commitmentsAfter, err := coreQueries.GetDKGPolyCommitments(ctx, corekeyperdb.GetDKGPolyCommitmentsParams{
		KeyperConfigIndex: keyperConfigIndex,
		RetryCounter:      retryCounter,
	})
	assert.NilError(t, err)
	assert.Equal(t, len(commitments), len(commitmentsAfter), "no extra commitment rows on idempotent call")

	evalsAfter, err := coreQueries.GetDKGPolyEvals(ctx, corekeyperdb.GetDKGPolyEvalsParams{
		KeyperConfigIndex: keyperConfigIndex,
		RetryCounter:      retryCounter,
	})
	assert.NilError(t, err)
	assert.Equal(t, len(evals), len(evalsAfter), "no extra poly eval rows on idempotent call")

	pendingAfter, err := coreQueries.GetPendingTxs(ctx)
	assert.NilError(t, err)
	assert.Equal(t, 1, len(pendingAfter), "no extra tx_outbox rows on idempotent call")
	assert.Equal(t, firstOutboxID, pendingAfter[0].ID, "same outbox row as before")
}

// TestMaybeDealNoopWhenInitialStateExists asserts that the dkg_initial_states
// row is the sole idempotency signal: even if commitment/eval/outbox rows are
// absent, presence of the initial-state row alone short-circuits maybeDeal.
func TestMaybeDealNoopWhenInitialStateExists(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	ctx := context.Background()

	dbpool, dbclose := testsetup.NewTestDBPool(ctx, t, corekeyperdb.Definition)
	t.Cleanup(dbclose)

	const (
		keyperConfigIndex int64 = 9
		retryCounter      int64 = 0
	)

	ownECDSA, err := crypto.GenerateKey()
	assert.NilError(t, err)
	ownAddr := crypto.PubkeyToAddress(ownECDSA.PublicKey)
	obsQueries := obskeyperdb.New(dbpool)
	err = obsQueries.InsertKeyperSet(ctx, obskeyperdb.InsertKeyperSetParams{
		KeyperConfigIndex:     keyperConfigIndex,
		ActivationBlockNumber: 0,
		Keypers:               shdb.EncodeAddresses([]common.Address{ownAddr}),
		Threshold:             1,
	})
	assert.NilError(t, err)

	// Pre-seed dkg_initial_states with arbitrary bytes (the contents are not
	// inspected by maybeDeal's idempotency check).
	pseudoBlob := make([]byte, 8)
	_, err = rand.Read(pseudoBlob)
	assert.NilError(t, err)
	coreQueries := corekeyperdb.New(dbpool)
	err = coreQueries.InsertDKGInitialState(ctx, corekeyperdb.InsertDKGInitialStateParams{
		KeyperConfigIndex: keyperConfigIndex,
		RetryCounter:      retryCounter,
		PuredkgBytes:      pseudoBlob,
	})
	assert.NilError(t, err)

	dkgAddr := common.HexToAddress("0xd0000000000000000000000000000000000000aa")
	mgr := New(Config{
		DBPool:            dbpool,
		OwnAddress:        ownAddr,
		ECIESPrivateKey:   ecies.ImportECDSA(ownECDSA),
		DKGContractAddr:   dkgAddr,
		ECIESRegistryAddr: common.HexToAddress("0xe0000000000000000000000000000000000000bb"),
	})

	err = mgr.maybeDeal(ctx, dkgAddr, keyperConfigIndex, retryCounter)
	assert.NilError(t, err)

	commitments, err := coreQueries.GetDKGPolyCommitments(ctx, corekeyperdb.GetDKGPolyCommitmentsParams{
		KeyperConfigIndex: keyperConfigIndex,
		RetryCounter:      retryCounter,
	})
	assert.NilError(t, err)
	assert.Equal(t, 0, len(commitments), "no commitment row should be written when initial state already exists")

	pending, err := coreQueries.GetPendingTxs(ctx)
	assert.NilError(t, err)
	assert.Equal(t, 0, len(pending), "no outbox row should be written when initial state already exists")
}
