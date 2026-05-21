package dkg

import (
	"context"
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

// TestMaybeRegisterECIESKeyIsIdempotent exercises the two acceptance
// criteria of MaybeRegisterECIESKey:
//
//  1. First invocation for a keyper set we are a member of writes a single
//     tx_outbox row (a `registerKey` enqueue against the configured registry
//     address).
//  2. Second invocation is a no-op once the `ecies_keys` row exists — the
//     DB cache populated by the host keyper's ECIES syncer is the source of
//     truth for prior registrations.
func TestMaybeRegisterECIESKeyIsIdempotent(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	ctx := context.Background()

	dbpool, dbclose := testsetup.NewTestDBPool(ctx, t, corekeyperdb.Definition)
	t.Cleanup(dbclose)

	const keyperConfigIndex int64 = 11

	ownECDSA, err := crypto.GenerateKey()
	assert.NilError(t, err)
	ownAddr := crypto.PubkeyToAddress(ownECDSA.PublicKey)
	otherECDSA, err := crypto.GenerateKey()
	assert.NilError(t, err)
	otherAddr := crypto.PubkeyToAddress(otherECDSA.PublicKey)

	obsQueries := obskeyperdb.New(dbpool)
	err = obsQueries.InsertKeyperSet(ctx, obskeyperdb.InsertKeyperSetParams{
		KeyperConfigIndex:     keyperConfigIndex,
		ActivationBlockNumber: 0,
		// We are index 1 in this keyper set; the test asserts on `pending[0].ID`
		// rather than the encoded calldata, so the specific index does not
		// matter for the acceptance criteria.
		Keypers:   shdb.EncodeAddresses([]common.Address{otherAddr, ownAddr}),
		Threshold: 2,
	})
	assert.NilError(t, err)

	registryAddr := common.HexToAddress("0xe0000000000000000000000000000000000000bb")
	mgr := New(Config{
		DBPool:            dbpool,
		OwnAddress:        ownAddr,
		ECIESPrivateKey:   ecies.ImportECDSA(ownECDSA),
		ECIESRegistryAddr: registryAddr,
	})

	coreQueries := corekeyperdb.New(dbpool)

	// First invocation: writes exactly one outbox row targeting the registry.
	err = mgr.MaybeRegisterECIESKey(ctx, keyperConfigIndex)
	assert.NilError(t, err)

	pending, err := coreQueries.GetPendingTxs(ctx)
	assert.NilError(t, err)
	assert.Equal(t, 1, len(pending), "first call must enqueue exactly one registerKey tx")
	assert.Equal(t, registryAddr.Hex(), pending[0].ToAddress)
	firstOutboxID := pending[0].ID

	// Second invocation (before the ECIES syncer has indexed the registry
	// event back into ecies_keys): still a no-op, because membership and
	// the ECIES-key check are the only gates. The first run's enqueue is
	// still in the outbox, but that does not affect ExistsECIESKey, so a
	// raw second call would actually re-enqueue. The contract is that the
	// caller's ECIES syncer populates ecies_keys before the next call —
	// emulate that here, then verify the no-op.
	err = coreQueries.UpsertECIESKey(ctx, corekeyperdb.UpsertECIESKeyParams{
		KeyperAddress:  shdb.EncodeAddress(ownAddr),
		EciesPublicKey: shdb.EncodeEciesPublicKey(&ecies.ImportECDSA(ownECDSA).PublicKey),
	})
	assert.NilError(t, err)

	err = mgr.MaybeRegisterECIESKey(ctx, keyperConfigIndex)
	assert.NilError(t, err)

	pendingAfter, err := coreQueries.GetPendingTxs(ctx)
	assert.NilError(t, err)
	assert.Equal(t, 1, len(pendingAfter), "second call must not enqueue another registerKey tx")
	assert.Equal(t, firstOutboxID, pendingAfter[0].ID, "same outbox row as before")
}

// TestMaybeRegisterECIESKeyNotMemberIsNoop asserts that a keyper who is
// not a member of the given keyper set never enqueues a registerKey tx.
func TestMaybeRegisterECIESKeyNotMemberIsNoop(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	ctx := context.Background()

	dbpool, dbclose := testsetup.NewTestDBPool(ctx, t, corekeyperdb.Definition)
	t.Cleanup(dbclose)

	const keyperConfigIndex int64 = 12

	ownECDSA, err := crypto.GenerateKey()
	assert.NilError(t, err)
	ownAddr := crypto.PubkeyToAddress(ownECDSA.PublicKey)
	otherECDSA, err := crypto.GenerateKey()
	assert.NilError(t, err)
	otherAddr := crypto.PubkeyToAddress(otherECDSA.PublicKey)

	obsQueries := obskeyperdb.New(dbpool)
	err = obsQueries.InsertKeyperSet(ctx, obskeyperdb.InsertKeyperSetParams{
		KeyperConfigIndex:     keyperConfigIndex,
		ActivationBlockNumber: 0,
		Keypers:               shdb.EncodeAddresses([]common.Address{otherAddr}),
		Threshold:             1,
	})
	assert.NilError(t, err)

	mgr := New(Config{
		DBPool:            dbpool,
		OwnAddress:        ownAddr,
		ECIESPrivateKey:   ecies.ImportECDSA(ownECDSA),
		ECIESRegistryAddr: common.HexToAddress("0xe0000000000000000000000000000000000000bb"),
	})

	err = mgr.MaybeRegisterECIESKey(ctx, keyperConfigIndex)
	assert.NilError(t, err)

	coreQueries := corekeyperdb.New(dbpool)
	pending, err := coreQueries.GetPendingTxs(ctx)
	assert.NilError(t, err)
	assert.Equal(t, 0, len(pending), "non-member must not enqueue any registerKey tx")
}
