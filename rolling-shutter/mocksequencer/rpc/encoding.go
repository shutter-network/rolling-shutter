package rpc

import (
	"errors"

	"github.com/ethereum/go-ethereum/common"
)

func stringToAddress(addr string) (common.Address, error) {
	if !common.IsHexAddress(addr) {
		var a common.Address
		return a, errors.New("address is no hex address string, can't decode")
	}
	return common.HexToAddress(addr), nil
}
