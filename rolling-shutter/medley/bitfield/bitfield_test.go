package bitfield

import (
	"sort"
	"testing"

	"gotest.tools/v3/assert"
)

func TestBitfield(t *testing.T) {
	tests := [][]int32{
		{1, 2, 3, 4, 5, 6},
		{1},
		{8},
		{9},
		{1, 2, 12, 14, 25},
		{6, 5, 4, 3, 2, 1},
		{45, 9},
		{0},
		{},
		{0, 1, 2, 3, 4, 5, 6, 7, 8},
		{0, 1, 2, 3, 4, 5, 6, 7},
	}
	for _, test := range tests {
		bf := MakeBitfieldFromIndex(test...)
		roundTripSet := bf.GetIndexes()
		if roundTripSet == nil {
			roundTripSet = []int32{}
		}
		sort.Slice(test, func(i, j int) bool { return test[i] < test[j] })
		assert.DeepEqual(t, test, roundTripSet)
	}
}
