package database

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"gotest.tools/v3/assert"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/shdb"
)

func makeTestKeyperSet() KeyperSet {
	return KeyperSet{
		KeyperConfigIndex:     0,
		ActivationBlockNumber: 0,
		Keypers: []string{
			shdb.EncodeAddress(common.HexToAddress("0x0000000000000000000000000000000000000000")),
			shdb.EncodeAddress(common.HexToAddress("0x5555555555555555555555555555555555555555")),
			shdb.EncodeAddress(common.HexToAddress("0xaAaAaAaaAaAaAaaAaAAAAAAAAaaaAaAaAaaAaaAa")),
		},
		Threshold: 2,
	}
}

func TestKeyperSetGetIndex(t *testing.T) {
	keyperSet := makeTestKeyperSet()
	addresses, err := shdb.DecodeAddresses(keyperSet.Keypers)
	assert.NilError(t, err)

	for i, address := range addresses {
		index, err := keyperSet.GetIndex(address)
		assert.NilError(t, err)
		assert.Equal(t, uint64(i), index)
	}
	_, err = keyperSet.GetIndex(common.HexToAddress("0xffffffffffffffffffffffffffffffffffffffff"))
	assert.ErrorContains(t, err, "keyper 0xFFfFfFffFFfffFFfFFfFFFFFffFFFffffFfFFFfF not found")
}

func TestKeyperSetContains(t *testing.T) {
	keyperSet := makeTestKeyperSet()
	addresses, err := shdb.DecodeAddresses(keyperSet.Keypers)
	assert.NilError(t, err)

	for _, address := range addresses {
		assert.Assert(t, keyperSet.Contains(address))
	}
	assert.Assert(t, !keyperSet.Contains(common.HexToAddress("0xffffffffffffffffffffffffffffffffffffffff")))
}

func TestKeyperSetSubset(t *testing.T) {
	keyperSet := makeTestKeyperSet()
	testCases := []struct {
		indices []uint64
		valid   bool
	}{
		{indices: []uint64{0, 1, 2}, valid: true},
		{indices: []uint64{}, valid: true},
		{indices: []uint64{1, 0}, valid: true},
		{indices: []uint64{0, 0}, valid: true},
		{indices: []uint64{0, 0, 0, 0}, valid: true},
		{indices: []uint64{3}, valid: false},
	}

	for _, tc := range testCases {
		subset, err := keyperSet.GetSubset(tc.indices)
		if tc.valid {
			assert.Assert(t, len(subset) == len(tc.indices))
			for _, i := range tc.indices {
				assert.Assert(t, shdb.EncodeAddress(subset[i]) == keyperSet.Keypers[tc.indices[i]])
			}
		} else {
			assert.Assert(t, err != nil)
		}
	}
}
