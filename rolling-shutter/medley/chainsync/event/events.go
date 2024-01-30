package event

import (
	"github.com/ethereum/go-ethereum/common"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/encodeable/number"
)

type (
	KeyperSet struct {
		ActivationBlock uint64
		Members         []common.Address
		Threshold       uint64
		Eon             uint64

		AtBlockNumber *number.BlockNumber `json:",omitempty"`
	}
	EonPublicKey struct {
		Eon uint64
		Key []byte

		AtBlockNumber *number.BlockNumber
	}
	ShutterState struct {
		Active bool

		AtBlockNumber *number.BlockNumber `json:",omitempty"`
	}
	LatestBlock struct {
		Number    *number.BlockNumber
		BlockHash common.Hash
	}
)
