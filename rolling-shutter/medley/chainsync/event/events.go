package event

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/encodeable/number"
)

type (
	KeyperSet struct {
		ActivationBlock uint64
		Members         []common.Address
		Threshold       uint64
		Eon             uint64
		// Contract is the deployed `KeyperSet` contract address for this eon.
		// Needed for callers that read per-keyper-set state (e.g.
		// `getDKGContract`) that is not part of `KeyperSetManager`.
		Contract common.Address

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
		Header    *types.Header
	}
	ECIESKey struct {
		Keyper         common.Address
		EciesPublicKey []byte

		AtBlockNumber *number.BlockNumber `json:",omitempty"`
	}
)

// DKGEvent is any event observed from the DKG Contract (live or synthesised
// at startup). Consumers type-switch on the concrete event types below.
type DKGEvent interface {
	isDKGEvent()
}

type DealingEvent struct {
	KeyperSetIndex uint64
	RetryCounter   uint64
	KeyperIndex    uint64
	Commitment     []byte
	PolyEvals      [][]byte

	AtBlockNumber *number.BlockNumber `json:",omitempty"`
}

type AccusationEvent struct {
	KeyperSetIndex uint64
	RetryCounter   uint64
	KeyperIndex    uint64
	AccusedIndices []uint64

	AtBlockNumber *number.BlockNumber `json:",omitempty"`
}

type ApologyEvent struct {
	KeyperSetIndex uint64
	RetryCounter   uint64
	KeyperIndex    uint64
	AccuserIndices []uint64
	PolyEvalData   [][]byte

	AtBlockNumber *number.BlockNumber `json:",omitempty"`
}

type SuccessVoteEvent struct {
	KeyperSetIndex uint64
	RetryCounter   uint64
	KeyperIndex    uint64
	EonPublicKey   []byte

	AtBlockNumber *number.BlockNumber `json:",omitempty"`
}

type SuccessEvent struct {
	KeyperSetIndex uint64
	RetryCounter   uint64
	EonPublicKey   []byte

	AtBlockNumber *number.BlockNumber `json:",omitempty"`
}

func (*DealingEvent) isDKGEvent()     {}
func (*AccusationEvent) isDKGEvent()  {}
func (*ApologyEvent) isDKGEvent()     {}
func (*SuccessVoteEvent) isDKGEvent() {}
func (*SuccessEvent) isDKGEvent()     {}
