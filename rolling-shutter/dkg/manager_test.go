package dkg

import (
	"context"
	"database/sql"
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

// TestHandleEonReturnsErrorWhenDkgContractIsNull asserts the strict-config
// acceptance criterion: an eons row with NULL `dkg_contract` causes
// handleEon to return a non-nil error and write no tx_outbox row.
func TestHandleEonReturnsErrorWhenDkgContractIsNull(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	ctx := context.Background()

	dbpool, dbclose := testsetup.NewTestDBPool(ctx, t, corekeyperdb.Definition)
	t.Cleanup(dbclose)

	const (
		keyperConfigIndex int64  = 7
		phaseLength       int64  = 10
		leadLength        int64  = 2
		blockNumber       uint64 = 5 // within Dealing phase
	)

	ownECDSA, err := crypto.GenerateKey()
	assert.NilError(t, err)
	ownAddr := crypto.PubkeyToAddress(ownECDSA.PublicKey)

	coreQueries := corekeyperdb.New(dbpool)
	err = coreQueries.InsertEon(ctx, corekeyperdb.InsertEonParams{
		Eon:                   keyperConfigIndex,
		ActivationBlockNumber: 0,
		KeyperConfigIndex:     keyperConfigIndex,
		// DkgContract intentionally NULL.
		PhaseLength: sql.NullInt64{Int64: phaseLength, Valid: true},
		LeadLength:  sql.NullInt64{Int64: leadLength, Valid: true},
	})
	assert.NilError(t, err)

	// Membership check fires before dkg_contract validation, so the keyper
	// set must include our address for the dkg_contract error to surface.
	obsQueries := obskeyperdb.New(dbpool)
	err = obsQueries.InsertKeyperSet(ctx, obskeyperdb.InsertKeyperSetParams{
		KeyperConfigIndex:     keyperConfigIndex,
		ActivationBlockNumber: 0,
		Keypers:               shdb.EncodeAddresses([]common.Address{ownAddr}),
		Threshold:             1,
	})
	assert.NilError(t, err)

	mgr := New(Config{
		DBPool:            dbpool,
		OwnAddress:        ownAddr,
		ECIESPrivateKey:   ecies.ImportECDSA(ownECDSA),
		ECIESRegistryAddr: common.HexToAddress("0xe0000000000000000000000000000000000000bb"),
	})

	eon, err := coreQueries.GetEon(ctx, keyperConfigIndex)
	assert.NilError(t, err)

	err = mgr.handleEon(ctx, eon, blockNumber)
	assert.Assert(t, err != nil, "expected error when dkg_contract is NULL")

	pending, err := coreQueries.GetPendingTxs(ctx)
	assert.NilError(t, err)
	assert.Equal(t, 0, len(pending), "no tx_outbox row may be written when dkg_contract is NULL")
}

// TestHandleEonReturnsErrorWhenPhaseParamsAreNull asserts the strict-config
// acceptance criterion: an eons row with NULL `phase_length` and/or NULL
// `lead_length` causes handleEon to return a non-nil error and write no
// tx_outbox row.
func TestHandleEonReturnsErrorWhenPhaseParamsAreNull(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	ctx := context.Background()

	dbpool, dbclose := testsetup.NewTestDBPool(ctx, t, corekeyperdb.Definition)
	t.Cleanup(dbclose)

	const (
		keyperConfigIndex int64  = 8
		blockNumber       uint64 = 5
	)

	ownECDSA, err := crypto.GenerateKey()
	assert.NilError(t, err)
	ownAddr := crypto.PubkeyToAddress(ownECDSA.PublicKey)

	coreQueries := corekeyperdb.New(dbpool)
	err = coreQueries.InsertEon(ctx, corekeyperdb.InsertEonParams{
		Eon:                   keyperConfigIndex,
		ActivationBlockNumber: 0,
		KeyperConfigIndex:     keyperConfigIndex,
		DkgContract:           sql.NullString{String: "0xd0000000000000000000000000000000000000aa", Valid: true},
		// PhaseLength and LeadLength intentionally NULL.
	})
	assert.NilError(t, err)

	mgr := New(Config{
		DBPool:            dbpool,
		OwnAddress:        ownAddr,
		ECIESPrivateKey:   ecies.ImportECDSA(ownECDSA),
		ECIESRegistryAddr: common.HexToAddress("0xe0000000000000000000000000000000000000bb"),
	})

	eon, err := coreQueries.GetEon(ctx, keyperConfigIndex)
	assert.NilError(t, err)

	err = mgr.handleEon(ctx, eon, blockNumber)
	assert.Assert(t, err != nil, "expected error when phase_length/lead_length is NULL")

	pending, err := coreQueries.GetPendingTxs(ctx)
	assert.NilError(t, err)
	assert.Equal(t, 0, len(pending), "no tx_outbox row may be written when phase params are NULL")
}

// TestHandleEonReturnsNilWhenNotAMember asserts the membership-ordering
// acceptance criterion: handleEon returns nil and writes no tx_outbox row
// when the manager's address is not present in the keyper set, even though
// the eon row is otherwise fully configured (dkg_contract + phase params
// set) and the block falls inside the Dealing phase.
func TestHandleEonReturnsNilWhenNotAMember(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	ctx := context.Background()

	dbpool, dbclose := testsetup.NewTestDBPool(ctx, t, corekeyperdb.Definition)
	t.Cleanup(dbclose)

	const (
		keyperConfigIndex int64  = 11
		phaseLength       int64  = 10
		leadLength        int64  = 2
		blockNumber       uint64 = 5 // within Dealing phase
	)

	ownECDSA, err := crypto.GenerateKey()
	assert.NilError(t, err)
	ownAddr := crypto.PubkeyToAddress(ownECDSA.PublicKey)

	// Two other keypers — deliberately NOT including ownAddr.
	other0, err := crypto.GenerateKey()
	assert.NilError(t, err)
	other0Addr := crypto.PubkeyToAddress(other0.PublicKey)
	other1, err := crypto.GenerateKey()
	assert.NilError(t, err)
	other1Addr := crypto.PubkeyToAddress(other1.PublicKey)

	coreQueries := corekeyperdb.New(dbpool)
	err = coreQueries.InsertEon(ctx, corekeyperdb.InsertEonParams{
		Eon:                   keyperConfigIndex,
		ActivationBlockNumber: 0,
		KeyperConfigIndex:     keyperConfigIndex,
		DkgContract:           sql.NullString{String: "0xd0000000000000000000000000000000000000aa", Valid: true},
		PhaseLength:           sql.NullInt64{Int64: phaseLength, Valid: true},
		LeadLength:            sql.NullInt64{Int64: leadLength, Valid: true},
	})
	assert.NilError(t, err)

	obsQueries := obskeyperdb.New(dbpool)
	err = obsQueries.InsertKeyperSet(ctx, obskeyperdb.InsertKeyperSetParams{
		KeyperConfigIndex:     keyperConfigIndex,
		ActivationBlockNumber: 0,
		Keypers:               shdb.EncodeAddresses([]common.Address{other0Addr, other1Addr}),
		Threshold:             1,
	})
	assert.NilError(t, err)

	mgr := New(Config{
		DBPool:            dbpool,
		OwnAddress:        ownAddr,
		ECIESPrivateKey:   ecies.ImportECDSA(ownECDSA),
		ECIESRegistryAddr: common.HexToAddress("0xe0000000000000000000000000000000000000bb"),
	})

	eon, err := coreQueries.GetEon(ctx, keyperConfigIndex)
	assert.NilError(t, err)

	err = mgr.handleEon(ctx, eon, blockNumber)
	assert.NilError(t, err, "handleEon must return nil when not a member")

	pending, err := coreQueries.GetPendingTxs(ctx)
	assert.NilError(t, err)
	assert.Equal(t, 0, len(pending), "no tx_outbox row may be written when not a member")
}
