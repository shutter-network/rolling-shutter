// Package dkgphase contains the PhaseLength struct, which is used to determine the DKG phase given
// a block number.
package dkgphase

import (
	"fmt"
	"strconv"

	"github.com/shutter-network/shutter/shlib/puredkg"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/encodeable"
)

// PhaseLength is used to store the accumulated lengths of the DKG phases.
type PhaseLength struct {
	off         int64
	dealing     int64
	accusing    int64
	apologizing int64
}

// NewConstantPhaseLength creates a new phase length definition where each phase has the same
// length.
func NewConstantPhaseLength(l int64) *PhaseLength {
	return &PhaseLength{
		off:         0 * l,
		dealing:     1 * l,
		accusing:    2 * l,
		apologizing: 3 * l,
	}
}

func (plen *PhaseLength) UnmarshalText(b []byte) error {
	phaseLenInt, err := strconv.ParseInt(string(b), 10, 64)
	if err != nil {
		return err
	}
	*plen = *NewConstantPhaseLength(phaseLenInt)
	return nil
}

func (plen PhaseLength) String() string {
	return encodeable.String(plen)
}

func (plen PhaseLength) MarshalText() (text []byte, err error) {
	return []byte(fmt.Sprint(plen.dealing)), nil
}

func (plen *PhaseLength) GetPhaseAtHeight(height int64, eonStartHeight int64) puredkg.Phase {
	if height < eonStartHeight+plen.off {
		return puredkg.Off
	}
	if height < eonStartHeight+plen.dealing {
		return puredkg.Dealing
	}
	if height < eonStartHeight+plen.accusing {
		return puredkg.Accusing
	}
	if height < eonStartHeight+plen.apologizing {
		return puredkg.Apologizing
	}
	return puredkg.Finalized
}
