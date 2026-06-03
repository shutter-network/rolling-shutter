package dkg

import (
	"context"
	"database/sql"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/ecies"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"gotest.tools/v3/assert"

	"github.com/shutter-network/shutter/shlib/puredkg"

	obskeyperdb "github.com/shutter-network/rolling-shutter/rolling-shutter/chainobserver/db/keyper"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/contract"
	corekeyperdb "github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/database"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/testsetup"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/shdb"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/txsender"
)

// runMaybeDealLocal mirrors the dispatch handleEon performs: look up the
// keyper set, build the puredkg in a read tx, then call maybeDeal in a
// separate write tx. Used by dealing_test.go tests that wire a Manager by
// hand instead of going through dkgTestEnv.
func runMaybeDealLocal(
	ctx context.Context,
	dbpool *pgxpool.Pool,
	mgr *Manager,
	dkgAddr common.Address,
	keyperConfigIndex, retryCounter int64,
) error {
	obsQueries := obskeyperdb.New(dbpool)
	keyperSet, err := obsQueries.GetKeyperSetByKeyperConfigIndex(ctx, keyperConfigIndex)
	if err != nil {
		return err
	}
	ownIndex, err := keyperSet.GetIndex(mgr.cfg.OwnAddress)
	if err != nil {
		return nil
	}
	keypers, err := shdb.DecodeAddresses(keyperSet.Keypers)
	if err != nil {
		return err
	}
	threshold := uint64(keyperSet.Threshold)

	var pure *puredkg.PureDKG
	err = dbpool.BeginFunc(ctx, func(tx pgx.Tx) error {
		p, err := mgr.buildPureDKG(ctx, tx, keyperConfigIndex, retryCounter, PhaseDealing, keypers, ownIndex, threshold)
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
	return dbpool.BeginFunc(ctx, func(tx pgx.Tx) error {
		return mgr.maybeDeal(ctx, tx, dkgAddr, keyperConfigIndex, retryCounter, pure, keypers, ownIndex)
	})
}

// TestMaybeDealPersistsInitialStateAndIsIdempotent exercises the two
// acceptance criteria for the maybeDeal refactor:
//
//  1. First invocation writes a dkg_initial_states row, a tx_outbox row for
//     submitDealing, and a dkg_sent_actions row marking the dealing action
//     as enqueued. It must NOT write to the shared `dkg_poly_commitments`
//     or `dkg_poly_evals` tables — those are populated exclusively by the
//     chain syncer from indexed events.
//  2. Second invocation is a no-op — no new rows are inserted in any of those
//     tables, because the dkg_sent_actions row already exists.
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
		ECIESRegistryAddr: common.HexToAddress("0xe0000000000000000000000000000000000000bb"),
	})

	// Sanity: nothing exists yet.
	_, err = coreQueries.GetDKGInitialState(ctx, corekeyperdb.GetDKGInitialStateParams{
		KeyperSetIndex: keyperConfigIndex,
		RetryCounter:      retryCounter,
	})
	assert.Assert(t, err != nil, "no initial state row expected before maybeDeal")

	// First invocation: writes rows.
	err = runMaybeDealLocal(ctx, dbpool, mgr, dkgAddr, keyperConfigIndex, retryCounter)
	assert.NilError(t, err)

	initial, err := coreQueries.GetDKGInitialState(ctx, corekeyperdb.GetDKGInitialStateParams{
		KeyperSetIndex: keyperConfigIndex,
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
		KeyperSetIndex: keyperConfigIndex,
		RetryCounter:      retryCounter,
	})
	assert.NilError(t, err)
	assert.Equal(t, 0, len(commitments), "maybeDeal must not write to dkg_poly_commitments — chain syncer owns it")

	evals, err := coreQueries.GetDKGPolyEvals(ctx, corekeyperdb.GetDKGPolyEvalsParams{
		KeyperSetIndex: keyperConfigIndex,
		RetryCounter:      retryCounter,
	})
	assert.NilError(t, err)
	assert.Equal(t, 0, len(evals), "maybeDeal must not write to dkg_poly_evals — chain syncer owns it")

	pending, err := coreQueries.GetPendingTxs(ctx)
	assert.NilError(t, err)
	assert.Equal(t, 1, len(pending))
	assert.Equal(t, dkgAddr.Hex(), pending[0].ToAddress)
	firstOutboxID := pending[0].ID

	sentAction, err := coreQueries.ExistsDKGSentAction(ctx, corekeyperdb.ExistsDKGSentActionParams{
		KeyperSetIndex: keyperConfigIndex,
		RetryCounter:      retryCounter,
		Action:            ActionDealing,
	})
	assert.NilError(t, err)
	assert.Assert(t, sentAction, "dkg_sent_actions row should exist for the dealing action")

	// Second invocation: idempotent — no new rows in any tracked table.
	err = runMaybeDealLocal(ctx, dbpool, mgr, dkgAddr, keyperConfigIndex, retryCounter)
	assert.NilError(t, err)

	commitmentsAfter, err := coreQueries.GetDKGPolyCommitments(ctx, corekeyperdb.GetDKGPolyCommitmentsParams{
		KeyperSetIndex: keyperConfigIndex,
		RetryCounter:      retryCounter,
	})
	assert.NilError(t, err)
	assert.Equal(t, 0, len(commitmentsAfter), "dkg_poly_commitments still empty on idempotent call")

	evalsAfter, err := coreQueries.GetDKGPolyEvals(ctx, corekeyperdb.GetDKGPolyEvalsParams{
		KeyperSetIndex: keyperConfigIndex,
		RetryCounter:      retryCounter,
	})
	assert.NilError(t, err)
	assert.Equal(t, 0, len(evalsAfter), "dkg_poly_evals still empty on idempotent call")

	pendingAfter, err := coreQueries.GetPendingTxs(ctx)
	assert.NilError(t, err)
	assert.Equal(t, 1, len(pendingAfter), "no extra tx_outbox rows on idempotent call")
	assert.Equal(t, firstOutboxID, pendingAfter[0].ID, "same outbox row as before")
}

// TestMaybeDealNoopWhenSentActionExists asserts that a pre-seeded
// dkg_sent_actions row for the dealing action short-circuits maybeDeal — no
// commitment, eval, or outbox row is written. This is the new idempotency
// signal that replaces the previous dkg_initial_states check.
func TestMaybeDealNoopWhenSentActionExists(t *testing.T) {
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

	// Pre-seed a tx_outbox row (via EnqueueTx) so the dkg_sent_actions FK is
	// satisfied, then the dkg_sent_actions marker for the dealing action.
	coreQueries := corekeyperdb.New(dbpool)
	var outboxID int64
	err = dbpool.BeginFunc(ctx, func(tx pgx.Tx) error {
		id, inner := txsender.EnqueueTx(ctx, tx, common.HexToAddress("0xd0000000000000000000000000000000000000aa"), []byte{0x01}, nil, "preseed")
		if inner != nil {
			return inner
		}
		outboxID = id
		return corekeyperdb.New(tx).InsertDKGSentAction(ctx, corekeyperdb.InsertDKGSentActionParams{
			KeyperSetIndex: keyperConfigIndex,
			RetryCounter:      retryCounter,
			Action:            ActionDealing,
			TxOutboxID:        sql.NullInt64{Int64: id, Valid: true},
		})
	})
	assert.NilError(t, err)

	dkgAddr := common.HexToAddress("0xd0000000000000000000000000000000000000aa")
	mgr := New(Config{
		DBPool:            dbpool,
		OwnAddress:        ownAddr,
		ECIESPrivateKey:   ecies.ImportECDSA(ownECDSA),
		ECIESRegistryAddr: common.HexToAddress("0xe0000000000000000000000000000000000000bb"),
	})

	err = runMaybeDealLocal(ctx, dbpool, mgr, dkgAddr, keyperConfigIndex, retryCounter)
	assert.NilError(t, err)

	commitments, err := coreQueries.GetDKGPolyCommitments(ctx, corekeyperdb.GetDKGPolyCommitmentsParams{
		KeyperSetIndex: keyperConfigIndex,
		RetryCounter:      retryCounter,
	})
	assert.NilError(t, err)
	assert.Equal(t, 0, len(commitments), "no commitment row should be written when sent action exists")

	pending, err := coreQueries.GetPendingTxs(ctx)
	assert.NilError(t, err)
	assert.Equal(t, 1, len(pending), "only the pre-seeded outbox row should remain")
	assert.Equal(t, outboxID, pending[0].ID)
	_, err = coreQueries.GetDKGInitialState(ctx, corekeyperdb.GetDKGInitialStateParams{
		KeyperSetIndex: keyperConfigIndex,
		RetryCounter:      retryCounter,
	})
	assert.Assert(t, err != nil, "no initial state row should be written when sent action exists")
}

// TestMaybeDealSubstitutesEmptyEvalForMissingECIESKey asserts the graceful-
// fallback behavior added for the "single misconfigured keyper must not block
// the entire DKG" problem: when one receiver has no `ecies_keys` row, the
// dealing is still enqueued, the `polyEvals` array still has length N−1, and
// the slot for the missing receiver is empty bytes (positional semantics
// preserved). Other receivers' slots remain populated with valid ciphertexts.
func TestMaybeDealSubstitutesEmptyEvalForMissingECIESKey(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	ctx := context.Background()

	dbpool, dbclose := testsetup.NewTestDBPool(ctx, t, corekeyperdb.Definition)
	t.Cleanup(dbclose)

	const (
		keyperConfigIndex int64 = 11
		retryCounter      int64 = 0
	)

	// Build a 3-keyper set where we are the keyper at index 1, the receiver
	// at index 0 has no ECIES key, and the receiver at index 2 does.
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
	// Register ECIES keys for self and `other2` only; `other0` is intentionally
	// absent — simulating a keyper that has not (yet) registered.
	for _, kp := range []struct {
		addr common.Address
		key  *ecies.PrivateKey
	}{
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

	err = runMaybeDealLocal(ctx, dbpool, mgr, dkgAddr, keyperConfigIndex, retryCounter)
	assert.NilError(t, err, "missing ECIES key for one receiver must not abort dealing")

	// Dealing was enqueued — initial state, sent-action marker, and tx_outbox
	// row must all be present.
	_, err = coreQueries.GetDKGInitialState(ctx, corekeyperdb.GetDKGInitialStateParams{
		KeyperSetIndex: keyperConfigIndex,
		RetryCounter:      retryCounter,
	})
	assert.NilError(t, err, "initial state row should be written despite missing ECIES key")

	sentAction, err := coreQueries.ExistsDKGSentAction(ctx, corekeyperdb.ExistsDKGSentActionParams{
		KeyperSetIndex: keyperConfigIndex,
		RetryCounter:      retryCounter,
		Action:            ActionDealing,
	})
	assert.NilError(t, err)
	assert.Assert(t, sentAction, "dkg_sent_actions row should be written for the dealing action")

	pending, err := coreQueries.GetPendingTxs(ctx)
	assert.NilError(t, err)
	assert.Equal(t, 1, len(pending), "one submitDealing outbox row expected")
	assert.Equal(t, dkgAddr.Hex(), pending[0].ToAddress)

	// Decode the submitDealing calldata to inspect the polyEvals payload.
	// Calldata layout is [4-byte selector][ABI-encoded args].
	abi, err := contract.DKGContractMetaData.GetAbi()
	assert.NilError(t, err)
	method, ok := abi.Methods["submitDealing"]
	assert.Assert(t, ok, "submitDealing method must be present on the ABI")
	assert.Assert(t, len(pending[0].Data) >= 4, "calldata must contain the 4-byte selector")
	args, err := method.Inputs.Unpack(pending[0].Data[4:])
	assert.NilError(t, err)
	polyEvals, ok := args[4].([][]byte)
	assert.Assert(t, ok, "polyEvals arg must decode to [][]byte")

	// N=3, sender index 1 → receivers [0, 2] → polyEvals length 2.
	// Slot 0 corresponds to receiver 0 (missing key, empty); slot 1 to
	// receiver 2 (present, non-empty ciphertext).
	assert.Equal(t, 2, len(polyEvals), "polyEvals length must equal N-1")
	assert.Equal(t, 0, len(polyEvals[0]), "missing-key slot must be empty bytes")
	assert.Assert(t, len(polyEvals[1]) > 0, "present-key slot must contain ciphertext")
}
