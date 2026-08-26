package dkg

import (
	"testing"

	"gotest.tools/v3/assert"
)

// maxRetriesForTests is a generous ceiling used by tests that do not
// themselves exercise the retry-limit gate. It matches the deploy-time
// default (`DKG_MAX_RETRIES = 10`) and keeps existing per-window assertions
// unaffected by the `retryCounter >= maxRetries` guard added in ticket 02.
const maxRetriesForTests uint64 = 10

func TestPhaseAtBoundaries(t *testing.T) {
	// Choose parameters where the lead length leaves room for a positive
	// dealing-start block so the boundaries are easy to reason about.
	const (
		activationBlock uint64 = 1000
		dkgLeadLength   uint64 = 40
		phaseLength     uint64 = 10
	)
	// Retry 0: dealing starts at block 960 (1000 - 40).
	start := DKGStart(activationBlock, dkgLeadLength, phaseLength, 0)
	assert.Equal(t, int64(960), start)

	cases := []struct {
		name  string
		block uint64
		want  Phase
	}{
		{"before dealing start", 959, PhaseNone},
		{"dealing first block", 960, PhaseDealing},
		{"dealing last block", 969, PhaseDealing},
		{"accusing first block", 970, PhaseAccusing},
		{"accusing last block", 979, PhaseAccusing},
		{"apologizing first block", 980, PhaseApologizing},
		{"apologizing last block", 989, PhaseApologizing},
		{"finalizing first block", 990, PhaseFinalizing},
		{"finalizing last block", 999, PhaseFinalizing},
		{"activation block (already past finalizing of retry 0)", 1000, PhaseDealing},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			retry := CurrentRetryCounter(activationBlock, dkgLeadLength, phaseLength, tc.block)
			got := PhaseAt(activationBlock, dkgLeadLength, phaseLength, maxRetriesForTests, retry, tc.block)
			assert.Equal(t, tc.want, got, "block=%d, retry=%d", tc.block, retry)
		})
	}
}

func TestPhaseAtAcrossRetries(t *testing.T) {
	const (
		activationBlock uint64 = 1000
		dkgLeadLength   uint64 = 40
		phaseLength     uint64 = 10
	)
	cycle := CycleLength(phaseLength) // 40

	// Retry 1 starts at block 1000 (960 + 40).
	assert.Equal(t, int64(1000), DKGStart(activationBlock, dkgLeadLength, phaseLength, 1))

	r := CurrentRetryCounter(activationBlock, dkgLeadLength, phaseLength, 1000)
	assert.Equal(t, uint64(1), r)
	assert.Equal(t, PhaseDealing, PhaseAt(activationBlock, dkgLeadLength, phaseLength, maxRetriesForTests, r, 1000))

	r = CurrentRetryCounter(activationBlock, dkgLeadLength, phaseLength, 1039)
	assert.Equal(t, uint64(1), r)
	assert.Equal(t, PhaseFinalizing, PhaseAt(activationBlock, dkgLeadLength, phaseLength, maxRetriesForTests, r, 1039))

	r = CurrentRetryCounter(activationBlock, dkgLeadLength, phaseLength, 1040)
	assert.Equal(t, uint64(2), r)
	assert.Equal(t, PhaseDealing, PhaseAt(activationBlock, dkgLeadLength, phaseLength, maxRetriesForTests, r, 1040))

	assert.Equal(t,
		DKGStart(activationBlock, dkgLeadLength, phaseLength, 0)+2*int64(cycle), //nolint:gosec // G115: cycle length fits well within int64
		DKGStart(activationBlock, dkgLeadLength, phaseLength, 2),
	)
}

func TestPhaseAtPastInstance(t *testing.T) {
	const (
		activationBlock uint64 = 1000
		dkgLeadLength   uint64 = 40
		phaseLength     uint64 = 10
	)
	got := PhaseAt(activationBlock, dkgLeadLength, phaseLength, maxRetriesForTests, 0, 1000)
	assert.Equal(t, PhaseNone, got)
}

func TestCurrentRetryCounterBeforeStart(t *testing.T) {
	const (
		activationBlock uint64 = 1000
		dkgLeadLength   uint64 = 40
		phaseLength     uint64 = 10
	)
	for _, block := range []uint64{0, 500, 959} {
		r := CurrentRetryCounter(activationBlock, dkgLeadLength, phaseLength, block)
		assert.Equal(t, uint64(0), r, "block=%d", block)
	}
}

func TestPhaseAtZeroPhaseLength(t *testing.T) {
	assert.Equal(t, PhaseNone, PhaseAt(1000, 40, 0, maxRetriesForTests, 0, 1000))
	assert.Equal(t, uint64(0), CurrentRetryCounter(1000, 40, 0, 1000))
}

// TestDispatchPhaseAtDefersFirstBlockOfEachWindow verifies that dispatch is
// deferred by exactly one block at every phase boundary: the first block of
// each window returns PhaseNone, the second block returns the window's phase,
// and all later blocks match PhaseAt unchanged.
func TestDispatchPhaseAtDefersFirstBlockOfEachWindow(t *testing.T) {
	const (
		activationBlock uint64 = 1000
		dkgLeadLength   uint64 = 40
		phaseLength     uint64 = 10
	)
	// Retry 0: dealing starts at block 960 (1000 - 40).
	cases := []struct {
		name  string
		block uint64
		want  Phase
	}{
		{"before dealing start", 959, PhaseNone},
		{"dealing first block deferred", 960, PhaseNone},
		{"dealing second block", 961, PhaseDealing},
		{"dealing last block", 969, PhaseDealing},
		{"accusing first block deferred", 970, PhaseNone},
		{"accusing second block", 971, PhaseAccusing},
		{"apologizing first block deferred", 980, PhaseNone},
		{"apologizing second block", 981, PhaseApologizing},
		{"finalizing first block deferred", 990, PhaseNone},
		{"finalizing second block", 991, PhaseFinalizing},
		{"finalizing last block", 999, PhaseFinalizing},
		{"retry 1 dealing first block deferred", 1000, PhaseNone},
		{"retry 1 dealing second block", 1001, PhaseDealing},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			retry := CurrentRetryCounter(activationBlock, dkgLeadLength, phaseLength, tc.block)
			got := DispatchPhaseAt(activationBlock, dkgLeadLength, phaseLength, maxRetriesForTests, retry, tc.block)
			assert.Equal(t, tc.want, got, "block=%d, retry=%d", tc.block, retry)
		})
	}
}

// TestDispatchPhaseAtPhaseLengthTwo verifies that a phase length of two is
// exempt from the deferral: the window's second block is its last, and a
// transaction triggered there is included one block later at the earliest,
// i.e. never in-phase. Dispatch therefore stays on the first block and
// DispatchPhaseAt degrades to PhaseAt.
func TestDispatchPhaseAtPhaseLengthTwo(t *testing.T) {
	const (
		activationBlock uint64 = 1000
		dkgLeadLength   uint64 = 4
		phaseLength     uint64 = 2
	)
	// Retry 0 starts at block 996: dealing 996-997, accusing 998-999,
	// apologizing 1000-1001, finalizing 1002-1003. Retry 1 deals from 1004.
	for block, want := range map[uint64]Phase{
		995:  PhaseNone,
		996:  PhaseDealing,
		997:  PhaseDealing,
		998:  PhaseAccusing,
		999:  PhaseAccusing,
		1000: PhaseApologizing,
		1001: PhaseApologizing,
		1002: PhaseFinalizing,
		1003: PhaseFinalizing,
		1004: PhaseDealing,
	} {
		retry := CurrentRetryCounter(activationBlock, dkgLeadLength, phaseLength, block)
		got := DispatchPhaseAt(activationBlock, dkgLeadLength, phaseLength, maxRetriesForTests, retry, block)
		assert.Equal(t, want, got, "block=%d, retry=%d", block, retry)
	}
}

// TestDispatchPhaseAtPhaseLengthOne verifies that a phase length of one leaves
// no room to defer, so DispatchPhaseAt degrades to PhaseAt.
func TestDispatchPhaseAtPhaseLengthOne(t *testing.T) {
	const (
		activationBlock uint64 = 1000
		dkgLeadLength   uint64 = 4
		phaseLength     uint64 = 1
	)
	// Retry 0: dealing at 996, accusing at 997, apologizing at 998,
	// finalizing at 999.
	for block, want := range map[uint64]Phase{
		995: PhaseNone,
		996: PhaseDealing,
		997: PhaseAccusing,
		998: PhaseApologizing,
		999: PhaseFinalizing,
	} {
		retry := CurrentRetryCounter(activationBlock, dkgLeadLength, phaseLength, block)
		got := DispatchPhaseAt(activationBlock, dkgLeadLength, phaseLength, maxRetriesForTests, retry, block)
		assert.Equal(t, want, got, "block=%d", block)
	}
}

// TestDispatchPhaseAtWindowStartingBeforeGenesis verifies that the boundary
// test is offset arithmetic, not a lookback on the previous block: a window
// that already covers block 0 (activation smaller than the lead length) is
// mid-window at block 0 and dispatches there, while the next window's first
// block (5) still defers.
func TestDispatchPhaseAtWindowStartingBeforeGenesis(t *testing.T) {
	const (
		activationBlock uint64 = 5
		dkgLeadLength   uint64 = 10
		phaseLength     uint64 = 10
	)
	// Retry 0 dealing starts at block -5, so dealing covers [-5, 5) and
	// accusing [5, 15).
	assert.Equal(t, PhaseDealing, PhaseAt(activationBlock, dkgLeadLength, phaseLength, maxRetriesForTests, 0, 0))
	assert.Equal(t, PhaseDealing, DispatchPhaseAt(activationBlock, dkgLeadLength, phaseLength, maxRetriesForTests, 0, 0))
	assert.Equal(t, PhaseNone, DispatchPhaseAt(activationBlock, dkgLeadLength, phaseLength, maxRetriesForTests, 0, 5))
	assert.Equal(t, PhaseAccusing, DispatchPhaseAt(activationBlock, dkgLeadLength, phaseLength, maxRetriesForTests, 0, 6))
}

// TestPhaseAtRetryCounterAtOrAboveMaxRetries covers the new retry-ceiling
// gate: PhaseAt returns PhaseNone for any retryCounter >= maxRetries, across
// representative block numbers inside each of the four sub-windows. This
// mirrors the on-chain rule enforced by ticket 01.
func TestPhaseAtRetryCounterAtOrAboveMaxRetries(t *testing.T) {
	const (
		activationBlock uint64 = 1000
		dkgLeadLength   uint64 = 40
		phaseLength     uint64 = 10
		maxRetries      uint64 = 2
	)
	// Retry 2 starts at block 1040 (960 + 2*40). Sample one block per phase
	// inside retry 2's cycle and one well past it, plus a retry > maxRetries.
	sampleBlocks := []uint64{1040, 1050, 1060, 1070, 5000}
	for _, retryCounter := range []uint64{maxRetries, maxRetries + 1, maxRetries + 5, 100} {
		for _, block := range sampleBlocks {
			got := PhaseAt(activationBlock, dkgLeadLength, phaseLength, maxRetries, retryCounter, block)
			assert.Equal(t, PhaseNone, got,
				"retryCounter=%d, maxRetries=%d, block=%d", retryCounter, maxRetries, block)
		}
	}
}

// TestPhaseAtRetryCounterBelowMaxRetries verifies that the existing per-window
// behavior is preserved for any retryCounter < maxRetries. This is the
// mirror-case of TestPhaseAtRetryCounterAtOrAboveMaxRetries.
func TestPhaseAtRetryCounterBelowMaxRetries(t *testing.T) {
	const (
		activationBlock uint64 = 1000
		dkgLeadLength   uint64 = 40
		phaseLength     uint64 = 10
		maxRetries      uint64 = 3
	)
	cases := []struct {
		name         string
		retryCounter uint64
		block        uint64
		want         Phase
	}{
		{"retry 0 dealing", 0, 960, PhaseDealing},
		{"retry 0 finalizing", 0, 990, PhaseFinalizing},
		{"retry 1 dealing", 1, 1000, PhaseDealing},
		{"retry 1 apologizing", 1, 1020, PhaseApologizing},
		{"retry 2 dealing", 2, 1040, PhaseDealing},
		{"retry 2 accusing", 2, 1050, PhaseAccusing},
		{"retry 2 finalizing", 2, 1070, PhaseFinalizing},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := PhaseAt(activationBlock, dkgLeadLength, phaseLength, maxRetries, tc.retryCounter, tc.block)
			assert.Equal(t, tc.want, got,
				"retryCounter=%d, maxRetries=%d, block=%d", tc.retryCounter, maxRetries, tc.block)
		})
	}
}
