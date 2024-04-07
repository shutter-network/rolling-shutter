package database

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/shdb"
)

// GetIndex returns the index of the given address in the KeyperSet.
func (s *KeyperSet) GetIndex(address common.Address) (uint64, error) {
	encodedAddress := shdb.EncodeAddress(address)
	for i, m := range s.Keypers {
		if m == encodedAddress {
			return uint64(i), nil
		}
	}
	return 0, errors.Errorf("keyper %s not found", address.String())
}

// Contains checks if the given address is present in the KeyperSet.
// It returns true if the address is found, otherwise false.
func (s *KeyperSet) Contains(address common.Address) bool {
	encodedAddress := shdb.EncodeAddress(address)
	for _, m := range s.Keypers {
		if m == encodedAddress {
			return true
		}
	}
	return false
}

// GetSubset returns a subset of addresses from the KeyperSet based on the given indices.
// The return value is ordered according to the order of the given indices. If indices contains
// duplicates, the return value will do so as well. If at least one of the given indices is out of
// range, an error is returned.
func (s *KeyperSet) GetSubset(indices []uint64) ([]common.Address, error) {
	subset := []common.Address{}
	for _, i := range indices {
		if i >= uint64(len(s.Keypers)) {
			return nil, errors.Errorf("keyper index %d out of range (size %d)", i, len(s.Keypers))
		}
		addressStr := s.Keypers[i]
		address, err := shdb.DecodeAddress(addressStr)
		if err != nil {
			return nil, err
		}
		subset = append(subset, address)
	}
	return subset, nil
}
