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
		MaxRetries:  10,
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
		MaxRetries: 10,
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
		MaxRetries:            10,
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

// TestHandleEonWritesNothingPastMaxRetries is the end-to-end wire-through
// acceptance test for ticket 02: an eons row with `max_retries = 2` plus a
// block advanced into what would be retry 3's Dealing window must produce
// zero `tx_outbox` and zero `dkg_sent_actions` rows. This proves that the
// `max_retries` column threads through `phaseParamsForEon` into `PhaseAt`
// and that the existing `PhaseNone` early-return is the single gate for the
// ceiling — no new branches added to `handleEon`.
func TestHandleEonWritesNothingPastMaxRetries(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	ctx := context.Background()

	dbpool, dbclose := testsetup.NewTestDBPool(ctx, t, corekeyperdb.Definition)
	t.Cleanup(dbclose)

	const (
		keyperConfigIndex int64 = 13
		activationBlock   int64 = 100
		phaseLength       int64 = 10
		leadLength        int64 = 40
		maxRetries        int64 = 2
		// Cycle length = 4 * phaseLength = 40. DKG start for retry 0 is
		// activationBlock - leadLength = 60. Block 185 lands in retry 3's
		// Dealing window ([180, 190)) — retry counter 3 >= maxRetries 2, so
		// PhaseAt must return PhaseNone and handleEon must be a no-op.
		blockNumber uint64 = 185
	)

	ownECDSA, err := crypto.GenerateKey()
	assert.NilError(t, err)
	ownAddr := crypto.PubkeyToAddress(ownECDSA.PublicKey)

	coreQueries := corekeyperdb.New(dbpool)
	err = coreQueries.InsertEon(ctx, corekeyperdb.InsertEonParams{
		Eon:                   keyperConfigIndex,
		ActivationBlockNumber: activationBlock,
		KeyperConfigIndex:     keyperConfigIndex,
		DkgContract:           sql.NullString{String: "0xd0000000000000000000000000000000000000aa", Valid: true},
		PhaseLength:           sql.NullInt64{Int64: phaseLength, Valid: true},
		LeadLength:            sql.NullInt64{Int64: leadLength, Valid: true},
		MaxRetries:            maxRetries,
	})
	assert.NilError(t, err)

	obsQueries := obskeyperdb.New(dbpool)
	err = obsQueries.InsertKeyperSet(ctx, obskeyperdb.InsertKeyperSetParams{
		KeyperConfigIndex:     keyperConfigIndex,
		ActivationBlockNumber: activationBlock,
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

	// Sanity: without the max-retries gate, retry 3's Dealing window would
	// otherwise trigger a dispatch. Verify the retry counter arithmetic is
	// what the test intends.
	retry := CurrentRetryCounter(uint64(activationBlock), uint64(leadLength), uint64(phaseLength), blockNumber)
	assert.Equal(t, uint64(3), retry, "test setup expects retry counter 3 at this block")

	err = mgr.handleEon(ctx, eon, blockNumber)
	assert.NilError(t, err, "handleEon must return nil (no error) once past max_retries")

	pending, err := coreQueries.GetPendingTxs(ctx)
	assert.NilError(t, err)
	assert.Equal(t, 0, len(pending), "no tx_outbox row may be written once retry counter >= max_retries")

	for _, action := range []string{ActionDealing, ActionAccusing, ActionApologizing, ActionFinalizing} {
		//nolint:gosec // G115: retry counter fits well within int64
		sentAction, err := coreQueries.ExistsDKGSentAction(ctx, corekeyperdb.ExistsDKGSentActionParams{
			KeyperSetIndex: keyperConfigIndex,
			RetryCounter:   int64(retry),
			Action:         action,
		})
		assert.NilError(t, err)
		assert.Assert(t, !sentAction,
			"no dkg_sent_actions row may be written for action=%s once past max_retries", action)
	}
}
