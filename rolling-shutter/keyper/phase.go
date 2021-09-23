package keyper

import "github.com/shutter-network/shutter/shlib/puredkg"

// PhaseLength is used to store the accumulated lengths of the DKG phases.
type PhaseLength struct {
	Off         int64
	Dealing     int64
	Accusing    int64
	Apologizing int64
}

// NewConstantPhaseLength creates a new phase length definition where each phase has the same
// length.
func NewConstantPhaseLength(l int64) PhaseLength {
	return PhaseLength{
		Off:         0 * l,
		Dealing:     1 * l,
		Accusing:    2 * l,
		Apologizing: 3 * l,
	}
}

func (plen *PhaseLength) getPhaseAtHeight(height int64, eonStartHeight int64) puredkg.Phase {
	if height < eonStartHeight+plen.Off {
		return puredkg.Off
	}
	if height < eonStartHeight+plen.Dealing {
		return puredkg.Dealing
	}
	if height < eonStartHeight+plen.Accusing {
		return puredkg.Accusing
	}
	if height < eonStartHeight+plen.Apologizing {
		return puredkg.Apologizing
	}
	return puredkg.Finalized
}
