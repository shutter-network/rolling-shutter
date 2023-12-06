package event

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

type (
	KeyperSet struct {
		ActivationBlock uint64
		Members         []common.Address
		Threshold       uint64
		Eon             uint64
	}
	EonPublicKey struct {
		Eon uint64
		Key []byte
	}
	LatestBlock struct {
		Number    *big.Int
		BlockHash common.Hash
	}
	ShutterState struct {
		Active bool
	}
)
