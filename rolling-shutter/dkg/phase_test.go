package dkg

import (
	"testing"

	"gotest.tools/v3/assert"
)

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
			got := PhaseAt(activationBlock, dkgLeadLength, phaseLength, retry, tc.block)
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
	assert.Equal(t, PhaseDealing, PhaseAt(activationBlock, dkgLeadLength, phaseLength, r, 1000))

	r = CurrentRetryCounter(activationBlock, dkgLeadLength, phaseLength, 1039)
	assert.Equal(t, uint64(1), r)
	assert.Equal(t, PhaseFinalizing, PhaseAt(activationBlock, dkgLeadLength, phaseLength, r, 1039))

	r = CurrentRetryCounter(activationBlock, dkgLeadLength, phaseLength, 1040)
	assert.Equal(t, uint64(2), r)
	assert.Equal(t, PhaseDealing, PhaseAt(activationBlock, dkgLeadLength, phaseLength, r, 1040))

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
	got := PhaseAt(activationBlock, dkgLeadLength, phaseLength, 0, 1000)
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
	assert.Equal(t, PhaseNone, PhaseAt(1000, 40, 0, 0, 1000))
	assert.Equal(t, uint64(0), CurrentRetryCounter(1000, 40, 0, 1000))
}
