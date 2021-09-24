package kprdb

import (
	"github.com/ethereum/go-ethereum/common"

	"github.com/shutter-network/shutter/shuttermint/shdb"
)

func (bc *KeyperTendermintBatchConfig) KeyperIndex(addr common.Address) (uint64, bool) {
	hexaddr := shdb.EncodeAddress(addr)
	for i, a := range bc.Keypers {
		if a == hexaddr {
			return uint64(i), true
		}
	}
	return uint64(0), false
}
