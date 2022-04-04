package kprdb

import (
	"github.com/ethereum/go-ethereum/common"

	"github.com/shutter-network/shutter/shuttermint/shdb"
)

func GetKeyperIndex(addr common.Address, keypers []string) (uint64, bool) {
	hexaddr := shdb.EncodeAddress(addr)
	for i, a := range keypers {
		if a == hexaddr {
			return uint64(i), true
		}
	}
	return uint64(0), false
}

func (bc *TendermintBatchConfig) KeyperIndex(addr common.Address) (uint64, bool) {
	return GetKeyperIndex(addr, bc.Keypers)
}
