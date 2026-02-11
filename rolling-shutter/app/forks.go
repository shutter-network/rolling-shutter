package app

import "github.com/rs/zerolog/log"

var forkHeightOverrides = map[string]ForkHeightOverrides{
	"shutter-gnosis-1000": {CheckInUpdate: &ForkHeightOverride{
		Eon: uint64Ptr(9),
	}},
	"shutter-chiado-102000": {CheckInUpdate: &ForkHeightOverride{
		Eon: uint64Ptr(13),
	}},
	"shutter-api-gnosis-1001": {CheckInUpdate: &ForkHeightOverride{
		Eon: uint64Ptr(13),
	}},
	"shutter-service-chiado-1000": {CheckInUpdate: &ForkHeightOverride{
		Eon: uint64Ptr(9),
	}},
	"shutter-api-gnosis-1002": {CheckInUpdate: &ForkHeightOverride{
		Eon: uint64Ptr(6),
	}},
}

func int64Ptr(value int64) *int64 {
	return &value
}

func uint64Ptr(value uint64) *uint64 {
	return &value
}

// NewForkHeightsAllEnabled creates a ForkHeights struct, activating all forks
// at genesis.
func NewForkHeightsAllEnabled() *ForkHeights {
	zero := int64(0)
	return &ForkHeights{
		CheckInUpdate: &zero,
	}
}

// NewForkHeightsAllDisabled creates a ForkHeights struct in which all forks
// are set to disabled.
func NewForkHeightsAllDisabled() *ForkHeights {
	return &ForkHeights{
		CheckInUpdate: nil,
	}
}

func (app *ShutterApp) IsCheckInUpdateForkActive() bool {
	var override *ForkHeightOverride
	if overrides, ok := forkHeightOverrides[app.ChainID]; ok {
		override = overrides.CheckInUpdate
	}
	if app.ForkHeights == nil {
		log.Warn().Msg("ForkHeights is nil, assuming all forks disabled")
		return false
	}
	forkHeight := app.ForkHeights.CheckInUpdate
	return isForkActive(forkHeight, override, app.CurrentBlockHeight(), app.EONCounter)
}

// isForkActive checks whether a fork is active.
//
// A fork is active in either of the following cases:
//   - an override is set and the override condition is met
//   - no override is set, a fork height is set, and current block height
//     is greater than or equal to the fork height
//
// Otherwise, the fork is not active, i.e., in any of the following cases:
//   - an override is set but the override condition is not met
//   - no override is set, a fork height is set, but the current block height is
//     less than the fork height
//   - no override is set and no fork height is set
//
// An override condition is met if
//   - an override height is set and the current block height is greater than or
//     equal to the override height
//   - an override eon is set and the current eon is greater than or equal to
//     the override eon
//
// If both override height and override eon are set, the height takes
// precedence. If neither is set, the fork is not active, regardless of the
// fork height.
func isForkActive(forkHeightGenesis *int64, override *ForkHeightOverride, currentBlockHeight int64, currentEon uint64) bool {
	if override != nil {
		if override.Height != nil {
			return currentBlockHeight >= *override.Height
		}
		if override.Eon != nil {
			return currentEon >= *override.Eon
		}
		return false
	}
	if forkHeightGenesis == nil {
		return false
	}
	return currentBlockHeight >= *forkHeightGenesis
}
