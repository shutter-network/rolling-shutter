package dkg

import (
	"context"
	"database/sql"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/ecies"
	"github.com/jackc/pgx/v4/pgxpool"
	"gotest.tools/v3/assert"

	obskeyperdb "github.com/shutter-network/rolling-shutter/rolling-shutter/chainobserver/db/keyper"
	corekeyperdb "github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/database"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/testsetup"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/shdb"
)

// activeDKGsSetFixture describes a Keyper Set + eons row pair inserted by
// the table-driven activeDKGs tests. `maxRetries` is zero-defaulted to
// `fixtureMaxRetries` so cases that don't care about the retry ceiling can
// omit it.
type activeDKGsSetFixture struct {
	index           int64
	activationBlock int64
	includeOwn      bool
	insertResult    bool // pre-seed a dkg_result success row for this ksi.
	maxRetries      int64
}

// The same phase params are used for every fixture row unless a case
// overrides them per-set. Cycle length = 4 * phaseLength = 40 blocks;
// retry 0 dealing window opens at activationBlock - leadLength.
const (
	fixturePhaseLength int64 = 10
	fixtureLeadLength  int64 = 2
	fixtureMaxRetries  int64 = 10
)

func insertActiveDKGsFixture(
	ctx context.Context,
	t *testing.T,
	dbpool *pgxpool.Pool,
	ownAddr common.Address,
	sets []activeDKGsSetFixture,
) {
	t.Helper()

	// One extra non-member address to pad keyper sets when `includeOwn` is
	// false. Regenerated per call so no cross-case leakage is possible.
	otherECDSA, err := crypto.GenerateKey()
	assert.NilError(t, err)
	otherAddr := crypto.PubkeyToAddress(otherECDSA.PublicKey)

	coreQueries := corekeyperdb.New(dbpool)
	obsQueries := obskeyperdb.New(dbpool)

	for _, s := range sets {
		var keypers []common.Address
		if s.includeOwn {
			keypers = []common.Address{ownAddr, otherAddr}
		} else {
			keypers = []common.Address{otherAddr}
		}
		err := obsQueries.InsertKeyperSet(ctx, obskeyperdb.InsertKeyperSetParams{
			KeyperConfigIndex:     s.index,
			ActivationBlockNumber: s.activationBlock,
			Keypers:               shdb.EncodeAddresses(keypers),
			Threshold:             1,
		})
		assert.NilError(t, err)

		maxRetries := s.maxRetries
		if maxRetries == 0 {
			maxRetries = fixtureMaxRetries
		}
		err = coreQueries.InsertEon(ctx, corekeyperdb.InsertEonParams{
			Eon:                   s.index,
			ActivationBlockNumber: s.activationBlock,
			KeyperConfigIndex:     s.index,
			DkgContract:           sql.NullString{String: "0xd0000000000000000000000000000000000000aa", Valid: true},
			PhaseLength:           sql.NullInt64{Int64: fixturePhaseLength, Valid: true},
			LeadLength:            sql.NullInt64{Int64: fixtureLeadLength, Valid: true},
			MaxRetries:            maxRetries,
		})
		assert.NilError(t, err)

		if s.insertResult {
			err := coreQueries.InsertDKGResult(ctx, corekeyperdb.InsertDKGResultParams{
				Eon:     s.index,
				Success: true,
			})
			assert.NilError(t, err)
		}
	}
}

// TestActiveDKGsFilters is a table-driven test that walks the four filters —
// membership, prior success, supersession, and retry ceiling — through every
// case listed in docs/dkg-supersede/02-skip-superseded-keyper-sets.md's
// "Seam 1" checklist. Each row builds a fresh test database, inserts the
// declared keyper sets and eons, and asserts activeDKGs returns exactly the
// expected keyper-set indices.
//
//nolint:funlen // table body inlined for readability; every case is one clause of the ticket-02 checklist.
func TestActiveDKGsFilters(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	// Test-cases derive from the ticket-02 test checklist. `blockNumber` is
	// picked so retry counter arithmetic never trips the retry ceiling unless
	// a case sets a low `maxRetries` on its fixture set.
	cases := []struct {
		name        string
		sets        []activeDKGsSetFixture
		blockNumber uint64
		wantIndices []int64
	}{
		{
			name: "single set member: included",
			sets: []activeDKGsSetFixture{
				{index: 1, activationBlock: 100, includeOwn: true},
			},
			blockNumber: 105,
			wantIndices: []int64{1},
		},
		{
			name: "single set non-member: excluded",
			sets: []activeDKGsSetFixture{
				{index: 1, activationBlock: 100, includeOwn: false},
			},
			blockNumber: 105,
			wantIndices: nil,
		},
		{
			name: "two sets both live: only higher-index included",
			sets: []activeDKGsSetFixture{
				{index: 1, activationBlock: 100, includeOwn: true},
				{index: 2, activationBlock: 120, includeOwn: true},
			},
			blockNumber: 130,
			wantIndices: []int64{2},
		},
		{
			name: "successor scheduled but not yet live: both included",
			sets: []activeDKGsSetFixture{
				{index: 1, activationBlock: 100, includeOwn: true},
				{index: 2, activationBlock: 200, includeOwn: true},
			},
			blockNumber: 110,
			wantIndices: []int64{1, 2},
		},
		{
			name: "boundary: activation_M == currentBlock excludes older",
			sets: []activeDKGsSetFixture{
				{index: 1, activationBlock: 100, includeOwn: true},
				{index: 2, activationBlock: 120, includeOwn: true},
			},
			blockNumber: 120,
			wantIndices: []int64{2},
		},
		{
			name: "superseded set with prior success row: still excluded",
			sets: []activeDKGsSetFixture{
				{index: 1, activationBlock: 100, includeOwn: true, insertResult: true},
				{index: 2, activationBlock: 120, includeOwn: true},
			},
			blockNumber: 130,
			wantIndices: []int64{2},
		},
		{
			name: "member of older set but not successor: older still excluded",
			sets: []activeDKGsSetFixture{
				{index: 1, activationBlock: 100, includeOwn: true},
				{index: 2, activationBlock: 120, includeOwn: false},
			},
			blockNumber: 130,
			wantIndices: nil,
		},
		{
			name: "retry counter past MAX_RETRIES with no successor: excluded",
			sets: []activeDKGsSetFixture{
				// activation=100, lead=2, phase=10, cycle=40, maxRetries=2.
				// Block 185 lands in retry 3's dealing window (>= max), so
				// the retry-ceiling filter excludes this set.
				{index: 1, activationBlock: 100, includeOwn: true, maxRetries: 2},
			},
			blockNumber: 185,
			wantIndices: nil,
		},
		{
			name: "no live keyper set: future scheduled set included",
			sets: []activeDKGsSetFixture{
				{index: 1, activationBlock: 500, includeOwn: true},
			},
			blockNumber: 100,
			wantIndices: []int64{1},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			dbpool, dbclose := testsetup.NewTestDBPool(ctx, t, corekeyperdb.Definition)
			t.Cleanup(dbclose)

			ownECDSA, err := crypto.GenerateKey()
			assert.NilError(t, err)
			ownAddr := crypto.PubkeyToAddress(ownECDSA.PublicKey)

			insertActiveDKGsFixture(ctx, t, dbpool, ownAddr, tc.sets)

			mgr := New(Config{
				DBPool:            dbpool,
				OwnAddress:        ownAddr,
				ECIESPrivateKey:   ecies.ImportECDSA(ownECDSA),
				ECIESRegistryAddr: common.HexToAddress("0xe0000000000000000000000000000000000000bb"),
			})

			active, _, err := mgr.activeDKGs(ctx, tc.blockNumber)
			assert.NilError(t, err)

			var gotIndices []int64
			for _, e := range active {
				gotIndices = append(gotIndices, e.KeyperConfigIndex)
			}
			assert.DeepEqual(t, tc.wantIndices, gotIndices)
		})
	}
}
