package gnosis

// DKGPhase mirrors the on-chain enum defined in DKGContract.sol. The numeric
// values match the contract so that values returned by `DKGContract.currentPhase`
// can be compared directly. PhaseNone is used for blocks before the dealing
// window starts (negative offset) and for blocks past the finalizing window.
type DKGPhase uint8

const (
	PhaseNone DKGPhase = iota
	PhaseDealing
	PhaseAccusing
	PhaseApologizing
	PhaseFinalizing
)

func (p DKGPhase) String() string {
	switch p {
	case PhaseNone:
		return "None"
	case PhaseDealing:
		return "Dealing"
	case PhaseAccusing:
		return "Accusing"
	case PhaseApologizing:
		return "Apologizing"
	case PhaseFinalizing:
		return "Finalizing"
	default:
		return "Unknown"
	}
}

// InstanceStart returns the first block number at which the DKG Instance
// `(keyperSetIndex, retryCounter)` enters its Dealing phase. The formula
// matches `DKGContract.dkgStart` in DKGContract.sol; using int64 internally
// allows the result to be negative when the activation block is smaller than
// the lead length (which the on-chain contract handles via int256).
func InstanceStart(activationBlock, dkgLeadLength, phaseLength uint64, retryCounter uint64) int64 {
	cycle := CycleLength(phaseLength)
	return int64(activationBlock) - int64(dkgLeadLength) + int64(retryCounter)*int64(cycle)
}

// CycleLength returns the length of a full DKG cycle (one attempt across all
// four phases) for the given per-phase length.
func CycleLength(phaseLength uint64) uint64 {
	return 4 * phaseLength
}

// PhaseAt returns the DKG phase of the instance `(activationBlock, retryCounter)`
// at the given block number. Blocks before the dealing window or past the
// finalizing window return PhaseNone. The boundaries are half-open: a phase
// covers `[start + n*phaseLength, start + (n+1)*phaseLength)` for n = 0..3.
func PhaseAt(activationBlock, dkgLeadLength, phaseLength, retryCounter, currentBlock uint64) DKGPhase {
	if phaseLength == 0 {
		return PhaseNone
	}
	start := InstanceStart(activationBlock, dkgLeadLength, phaseLength, retryCounter)
	offset := int64(currentBlock) - start
	if offset < 0 {
		return PhaseNone
	}
	pl := int64(phaseLength)
	switch {
	case offset < pl:
		return PhaseDealing
	case offset < 2*pl:
		return PhaseAccusing
	case offset < 3*pl:
		return PhaseApologizing
	case offset < 4*pl:
		return PhaseFinalizing
	default:
		return PhaseNone
	}
}

// CurrentRetryCounter derives the active retry counter from block arithmetic.
// Each failed cycle advances the counter by one; the counter is never stored
// in the database. A block before `InstanceStart(..., 0)` returns 0 since the
// loop has not begun yet (matching the contract's behaviour of treating early
// blocks as "Phase.None" within retry 0).
func CurrentRetryCounter(activationBlock, dkgLeadLength, phaseLength, currentBlock uint64) uint64 {
	cycle := CycleLength(phaseLength)
	if cycle == 0 {
		return 0
	}
	start := InstanceStart(activationBlock, dkgLeadLength, phaseLength, 0)
	offset := int64(currentBlock) - start
	if offset < 0 {
		return 0
	}
	return uint64(offset) / cycle
}
