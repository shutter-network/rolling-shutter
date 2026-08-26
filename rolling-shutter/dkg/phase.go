// Package dkg implements the DKG participation logic as a database-driven
// reactor. The host keyper calls HandleBlock on every new block; the manager
// iterates all active eons and writes any required actions (DKG message rows
// and tx_outbox entries) back to the database. The package owns no chainsync
// subscriptions and makes no live chain calls — see ADR 0004.
package dkg

// Phase mirrors the on-chain enum defined in DKGContract.sol. The numeric
// values match the contract so that values returned by `DKGContract.currentPhase`
// can be compared directly. PhaseNone is used for blocks before the dealing
// window starts (negative offset) and for blocks past the finalizing window.
type Phase uint8

const (
	PhaseNone Phase = iota
	PhaseDealing
	PhaseAccusing
	PhaseApologizing
	PhaseFinalizing
)

func (p Phase) String() string {
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

// DKGStart returns the first block number at which the DKG instance
// `(keyperSetIndex, retryCounter)` enters its Dealing phase. The formula
// matches `DKGContract.dkgStart` in DKGContract.sol; using int64 internally
// allows the result to be negative when the activation block is smaller than
// the lead length (which the on-chain contract handles via int256).
func DKGStart(activationBlock, dkgLeadLength, phaseLength uint64, retryCounter uint64) int64 {
	cycle := CycleLength(phaseLength)
	//nolint:gosec // G115: block numbers and DKG params fit well within int64
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
//
// `maxRetries` is the on-chain retry ceiling for this keyper set (see the
// `MAX_RETRIES` glossary entry): the valid retry-counter range is
// `{0, ..., maxRetries - 1}`. Any `retryCounter >= maxRetries` returns
// PhaseNone before any window arithmetic. `maxRetries` is positioned before
// `retryCounter` to mirror the DKG contract's argument ordering.
func PhaseAt(activationBlock, dkgLeadLength, phaseLength, maxRetries, retryCounter, currentBlock uint64) Phase {
	if retryCounter >= maxRetries {
		return PhaseNone
	}
	if phaseLength == 0 {
		return PhaseNone
	}
	start := DKGStart(activationBlock, dkgLeadLength, phaseLength, retryCounter)
	offset := int64(currentBlock) - start //nolint:gosec // G115: block number fits well within int64
	if offset < 0 {
		return PhaseNone
	}
	pl := int64(phaseLength) //nolint:gosec // G115: phase length fits well within int64
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

// DispatchPhaseAt returns the phase whose action may be dispatched at
// `currentBlock`. It matches PhaseAt except on the first block of each phase
// window, where it returns PhaseNone so that dispatch is deferred to the
// window's second block. Dispatch decisions go through this function rather
// than PhaseAt so that the deferral cannot be forgotten at a call site.
//
// The deferral exists because gas estimation races the phase boundary: an RPC
// node can announce block N via newHead while `eth_estimateGas` still executes
// against the state of N-1. A message enqueued on the boundary block then
// reverts with WrongPhase during estimation and is permanently marked failed.
// Waiting one block gives the node's state a full block interval to catch up.
// The contract window itself is unchanged, so with phaseLength L the remaining
// L-1 blocks are ample for inclusion.
//
// Windows with phaseLength <= 2 are exempt: they have no block that is both
// past the boundary and still has a successor inside the window, since a
// transaction triggered by block B is included at B+1 at the earliest. For
// those, dispatch happens on the window's first block, which is the only
// choice that can land in-phase at all (L == 2) or the only block there is
// (L == 1).
func DispatchPhaseAt(activationBlock, dkgLeadLength, phaseLength, maxRetries, retryCounter, currentBlock uint64) Phase {
	phase := PhaseAt(activationBlock, dkgLeadLength, phaseLength, maxRetries, retryCounter, currentBlock)
	if phase == PhaseNone || phaseLength <= 2 {
		return phase
	}
	// `phase != PhaseNone` guarantees the offset is inside the four windows,
	// so it is non-negative and a zero remainder means "first block of a
	// window". This mirrors `blocksInto` in the contract's DKGState script.
	start := DKGStart(activationBlock, dkgLeadLength, phaseLength, retryCounter)
	offset := int64(currentBlock) - start //nolint:gosec // G115: block number fits well within int64
	pl := int64(phaseLength)              //nolint:gosec // G115: phase length fits well within int64
	if offset%pl == 0 {
		return PhaseNone
	}
	return phase
}

// CurrentRetryCounter derives the active retry counter from block arithmetic.
// Each failed cycle advances the counter by one; the counter is never stored
// in the database. A block before `DKGStart(..., 0)` returns 0 since the
// loop has not begun yet (matching the contract's behavior of treating early
// blocks as "Phase.None" within retry 0).
func CurrentRetryCounter(activationBlock, dkgLeadLength, phaseLength, currentBlock uint64) uint64 {
	cycle := CycleLength(phaseLength)
	if cycle == 0 {
		return 0
	}
	start := DKGStart(activationBlock, dkgLeadLength, phaseLength, 0)
	offset := int64(currentBlock) - start //nolint:gosec // G115: block number fits well within int64
	if offset < 0 {
		return 0
	}
	return uint64(offset) / cycle
}
