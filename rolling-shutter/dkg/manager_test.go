package dkg

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/ecies"
	"github.com/jackc/pgx/v4"
	"gotest.tools/v3/assert"

	obskeyperdb "github.com/shutter-network/rolling-shutter/rolling-shutter/chainobserver/db/keyper"
	corekeyperdb "github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/database"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/testsetup"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/shdb"
)

// TestProcessDKGReturnsErrorWhenDkgContractIsNull asserts the strict-config
// acceptance criterion: an eons row with NULL `dkg_contract` causes
// processDKG to return a non-nil error and write no tx_outbox row.
func TestProcessDKGReturnsErrorWhenDkgContractIsNull(t *testing.T) {
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

	err = mgr.processDKG(ctx, eon, blockNumber)
	assert.Assert(t, err != nil, "expected error when dkg_contract is NULL")

	pending, err := coreQueries.GetPendingTxs(ctx)
	assert.NilError(t, err)
	assert.Equal(t, 0, len(pending), "no tx_outbox row may be written when dkg_contract is NULL")
}

// TestProcessDKGReturnsErrorWhenPhaseParamsAreNull asserts the strict-config
// acceptance criterion: an eons row with NULL `phase_length` and/or NULL
// `lead_length` causes processDKG to return a non-nil error and write no
// tx_outbox row.
func TestProcessDKGReturnsErrorWhenPhaseParamsAreNull(t *testing.T) {
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

	err = mgr.processDKG(ctx, eon, blockNumber)
	assert.Assert(t, err != nil, "expected error when phase_length/lead_length is NULL")

	pending, err := coreQueries.GetPendingTxs(ctx)
	assert.NilError(t, err)
	assert.Equal(t, 0, len(pending), "no tx_outbox row may be written when phase params are NULL")
}

// TestHandleBlockSkipsEonWhenNotAMember asserts the membership-filter
// acceptance criterion: HandleBlock returns nil and writes no tx_outbox row
// when the manager's address is not present in the keyper set, even though
// the eon row is otherwise fully configured (dkg_contract + phase params
// set) and the block falls inside the Dealing phase. The filter now lives in
// `activeDKGs`, so we exercise the full `HandleBlock` path.
func TestHandleBlockSkipsEonWhenNotAMember(t *testing.T) {
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

	err = mgr.HandleBlock(ctx, blockNumber)
	assert.NilError(t, err, "HandleBlock must return nil when the keyper is filtered out")

	pending, err := coreQueries.GetPendingTxs(ctx)
	assert.NilError(t, err)
	assert.Equal(t, 0, len(pending), "no tx_outbox row may be written when not a member")
}

// TestProcessDKGWritesNothingPastMaxRetries is the end-to-end wire-through
// acceptance test for the MAX_RETRIES ceiling: an eons row with
// `max_retries = 2` plus a block advanced into what would be retry 3's
// Dealing window must produce zero `tx_outbox` and zero `dkg_sent_actions`
// rows. The retry-ceiling filter now lives in `activeDKGs`, and `PhaseAt`
// retains its `retryCounter >= maxRetries` guard as belt-and-braces — either
// gate suffices on its own, so exercising `processDKG` directly still yields
// a no-op.
func TestProcessDKGWritesNothingPastMaxRetries(t *testing.T) {
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
		// PhaseAt must return PhaseNone and processDKG must be a no-op.
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

	err = mgr.processDKG(ctx, eon, blockNumber)
	assert.NilError(t, err, "processDKG must return nil (no error) once past max_retries")

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

// TestHandleBlockSkipsSupersededKeyperSet is the end-to-end wire-through for
// the supersession filter (Seam 2 case a): two Keyper Sets with activation
// blocks 100 and 120; advancing to a block past the successor's activation
// must leave the older set silent — no `tx_outbox` row and no
// `dkg_sent_actions` row for any of the four phases.
func TestHandleBlockSkipsSupersededKeyperSet(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	ctx := context.Background()

	dbpool, dbclose := testsetup.NewTestDBPool(ctx, t, corekeyperdb.Definition)
	t.Cleanup(dbclose)

	const (
		olderIndex      int64 = 1
		newerIndex      int64 = 2
		olderActivation int64 = 100
		newerActivation int64 = 120
		phaseLength     int64 = 10
		leadLength      int64 = 2
		maxRetries      int64 = 10
		// blockNumber lands in what would be the older set's retry-1 Dealing
		// window (retry 0 dealing at [98, 108), retry 1 dealing at [138, 148)).
		// 130 is past the successor's activation (120), so supersession must
		// keep the older set silent even though a naive read of the phase
		// arithmetic would place the older set outside PhaseNone at 138+.
		blockNumber uint64 = 138
	)

	ownECDSA, err := crypto.GenerateKey()
	assert.NilError(t, err)
	ownAddr := crypto.PubkeyToAddress(ownECDSA.PublicKey)

	coreQueries := corekeyperdb.New(dbpool)
	for _, e := range []struct {
		idx        int64
		activation int64
	}{
		{olderIndex, olderActivation},
		{newerIndex, newerActivation},
	} {
		err = coreQueries.InsertEon(ctx, corekeyperdb.InsertEonParams{
			Eon:                   e.idx,
			ActivationBlockNumber: e.activation,
			KeyperConfigIndex:     e.idx,
			DkgContract:           sql.NullString{String: "0xd0000000000000000000000000000000000000aa", Valid: true},
			PhaseLength:           sql.NullInt64{Int64: phaseLength, Valid: true},
			LeadLength:            sql.NullInt64{Int64: leadLength, Valid: true},
			MaxRetries:            maxRetries,
		})
		assert.NilError(t, err)
	}

	obsQueries := obskeyperdb.New(dbpool)
	for _, e := range []struct {
		idx        int64
		activation int64
	}{
		{olderIndex, olderActivation},
		{newerIndex, newerActivation},
	} {
		err = obsQueries.InsertKeyperSet(ctx, obskeyperdb.InsertKeyperSetParams{
			KeyperConfigIndex:     e.idx,
			ActivationBlockNumber: e.activation,
			Keypers:               shdb.EncodeAddresses([]common.Address{ownAddr}),
			Threshold:             1,
		})
		assert.NilError(t, err)
	}

	mgr := New(Config{
		DBPool:            dbpool,
		OwnAddress:        ownAddr,
		ECIESPrivateKey:   ecies.ImportECDSA(ownECDSA),
		ECIESRegistryAddr: common.HexToAddress("0xe0000000000000000000000000000000000000bb"),
	})

	err = mgr.HandleBlock(ctx, blockNumber)
	assert.NilError(t, err)

	pending, err := coreQueries.GetPendingTxs(ctx)
	assert.NilError(t, err)
	olderLabelMarker := fmt.Sprintf("ksi=%d ", olderIndex)
	for _, tx := range pending {
		assert.Assert(t, !strings.Contains(tx.Label, olderLabelMarker),
			"no tx_outbox row may be written for the superseded older set (label=%q)", tx.Label)
	}

	for retry := int64(0); retry < maxRetries; retry++ {
		for _, action := range []string{ActionDealing, ActionAccusing, ActionApologizing, ActionFinalizing} {
			sent, err := coreQueries.ExistsDKGSentAction(ctx, corekeyperdb.ExistsDKGSentActionParams{
				KeyperSetIndex: olderIndex,
				RetryCounter:   retry,
				Action:         action,
			})
			assert.NilError(t, err)
			assert.Assert(t, !sent,
				"no dkg_sent_actions row may be written for superseded set (retry=%d, action=%s)", retry, action)
		}
	}
}

// TestHandleDKGSuccessOnSupersededSetStillWritesResult is the end-to-end
// wire-through for the peer-driven success case (Seam 2 case b): after the
// successor has activated, a `DKGSucceeded` event arriving for the older
// (superseded) set must still write the `dkg_result` row exactly as it does
// today, and the next HandleBlock must still produce no submissions for the
// older set.
func TestHandleDKGSuccessOnSupersededSetStillWritesResult(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	ctx := context.Background()

	dbpool, dbclose := testsetup.NewTestDBPool(ctx, t, corekeyperdb.Definition)
	t.Cleanup(dbclose)

	const (
		olderIndex      int64  = 1
		newerIndex      int64  = 2
		olderActivation int64  = 100
		newerActivation int64  = 120
		phaseLength     int64  = 10
		leadLength      int64  = 2
		maxRetries      int64  = 10
		blockNumber     uint64 = 138
		retryCounter    int64  = 0
	)

	ownECDSA, err := crypto.GenerateKey()
	assert.NilError(t, err)
	ownAddr := crypto.PubkeyToAddress(ownECDSA.PublicKey)

	coreQueries := corekeyperdb.New(dbpool)
	for _, e := range []struct {
		idx        int64
		activation int64
	}{
		{olderIndex, olderActivation},
		{newerIndex, newerActivation},
	} {
		err = coreQueries.InsertEon(ctx, corekeyperdb.InsertEonParams{
			Eon:                   e.idx,
			ActivationBlockNumber: e.activation,
			KeyperConfigIndex:     e.idx,
			DkgContract:           sql.NullString{String: "0xd0000000000000000000000000000000000000aa", Valid: true},
			PhaseLength:           sql.NullInt64{Int64: phaseLength, Valid: true},
			LeadLength:            sql.NullInt64{Int64: leadLength, Valid: true},
			MaxRetries:            maxRetries,
		})
		assert.NilError(t, err)
	}

	obsQueries := obskeyperdb.New(dbpool)
	for _, e := range []struct {
		idx        int64
		activation int64
	}{
		{olderIndex, olderActivation},
		{newerIndex, newerActivation},
	} {
		err = obsQueries.InsertKeyperSet(ctx, obskeyperdb.InsertKeyperSetParams{
			KeyperConfigIndex:     e.idx,
			ActivationBlockNumber: e.activation,
			Keypers:               shdb.EncodeAddresses([]common.Address{ownAddr}),
			Threshold:             1,
		})
		assert.NilError(t, err)
	}

	mgr := New(Config{
		DBPool:            dbpool,
		OwnAddress:        ownAddr,
		ECIESPrivateKey:   ecies.ImportECDSA(ownECDSA),
		ECIESRegistryAddr: common.HexToAddress("0xe0000000000000000000000000000000000000bb"),
	})

	// Peer-driven success on the (already superseded) older set.
	err = dbpool.BeginFunc(ctx, func(tx pgx.Tx) error {
		return mgr.HandleDKGSuccess(ctx, tx, olderIndex, retryCounter)
	})
	assert.NilError(t, err)

	succeeded, err := coreQueries.ExistsDKGResultSuccess(ctx, olderIndex)
	assert.NilError(t, err)
	assert.Assert(t, succeeded, "HandleDKGSuccess must write dkg_result row even for a superseded set")

	// Next HandleBlock must still be a no-op for the older set: the
	// "already-succeeded" filter now also excludes it, but the supersession
	// filter alone was already enough.
	err = mgr.HandleBlock(ctx, blockNumber)
	assert.NilError(t, err)

	pending, err := coreQueries.GetPendingTxs(ctx)
	assert.NilError(t, err)
	olderLabelMarker := fmt.Sprintf("ksi=%d ", olderIndex)
	for _, tx := range pending {
		assert.Assert(t, !strings.Contains(tx.Label, olderLabelMarker),
			"no tx_outbox row may be written for the superseded older set after peer-driven success (label=%q)", tx.Label)
	}

	for retry := int64(0); retry < maxRetries; retry++ {
		for _, action := range []string{ActionDealing, ActionAccusing, ActionApologizing, ActionFinalizing} {
			sent, err := coreQueries.ExistsDKGSentAction(ctx, corekeyperdb.ExistsDKGSentActionParams{
				KeyperSetIndex: olderIndex,
				RetryCounter:   retry,
				Action:         action,
			})
			assert.NilError(t, err)
			assert.Assert(t, !sent,
				"no dkg_sent_actions row may be written for superseded set (retry=%d, action=%s)", retry, action)
		}
	}
}
