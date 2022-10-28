package encoding

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
)

type ChainConfig struct {
	ChainID      *big.Int `json:"chainId"`                // chainId identifies the current chain and is used for replay protection
	ShutterBlock *big.Int `json:"shutterBlock,omitempty"` // Shutter switch block (nil = no fork, 0 = already on shutter)
}

func StringToAddress(addr string) (common.Address, error) {
	if !common.IsHexAddress(addr) {
		var a common.Address
		return a, errors.New("address is no hex address string, can't decode")
	}
	return common.HexToAddress(addr), nil
}
