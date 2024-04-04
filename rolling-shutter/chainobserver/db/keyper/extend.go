package database

import (
	"github.com/ethereum/go-ethereum/common"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/shdb"
)

func (s *KeyperSet) Contains(address common.Address) bool {
	encodedAddress := shdb.EncodeAddress(address)
	for _, m := range s.Keypers {
		if m == encodedAddress {
			return true
		}
	}
	return false
}
