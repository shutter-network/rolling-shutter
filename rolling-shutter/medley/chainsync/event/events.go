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

	// DKGEventKind identifies which DKG Contract event a DKGEvent represents.
	DKGEventKind int

	// DKGEvent is a unified envelope for any DKG Contract event observed
	// (live or synthesised at startup). Fields are populated according to Kind.
	DKGEvent struct {
		Kind           DKGEventKind
		KeyperSetIndex uint64
		RetryCounter   uint64
		KeyperIndex    uint64

		Commitment     []byte
		PolyEvals      [][]byte
		AccusedIndices []uint64
		AccuserIndices []uint64
		PolyEvalData   [][]byte
		EonPublicKey   []byte

		AtBlockNumber *number.BlockNumber `json:",omitempty"`
	}
)

const (
	DKGEventKindDealing DKGEventKind = iota
	DKGEventKindAccusation
	DKGEventKindApology
	DKGEventKindSuccessVote
	DKGEventKindSuccess
)
