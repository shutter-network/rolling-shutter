package bitfield

import (
	"reflect"
	"sort"
	"testing"

	"gotest.tools/v3/assert"
)

func TestBitfield(t *testing.T) {
	tests := [][]int32{{1, 2, 3, 4, 5, 6}, {1}, {8}, {9}, {1, 2, 12, 14, 25}, {6, 5, 4, 3, 2, 1}, {45, 9}}
	for _, test := range tests {
		roundTripSet := GetIndexes(MakeBitfieldFromArray(test))
		sort.Slice(roundTripSet, func(i, j int) bool { return roundTripSet[i] < roundTripSet[j] })
		sort.Slice(test, func(i, j int) bool { return test[i] < test[j] })
		assert.Check(t, reflect.DeepEqual(test, roundTripSet))
	}
}
