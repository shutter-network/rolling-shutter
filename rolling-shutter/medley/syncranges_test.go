package medley

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestGetSyncRanges(t *testing.T) {
	var maxRange uint64 = 3
	testCases := []struct {
		start  uint64
		end    uint64
		ranges [][2]uint64
	}{
		{start: 0, end: 0, ranges: [][2]uint64{{0, 0}}},
		{start: 3, end: 3, ranges: [][2]uint64{{3, 3}}},
		{start: 0, end: 2, ranges: [][2]uint64{{0, 2}}},
		{start: 3, end: 5, ranges: [][2]uint64{{3, 5}}},
		{start: 0, end: 5, ranges: [][2]uint64{{0, 2}, {3, 5}}},
		{start: 3, end: 8, ranges: [][2]uint64{{3, 5}, {6, 8}}},
		{start: 0, end: 1, ranges: [][2]uint64{{0, 1}}},
		{start: 3, end: 4, ranges: [][2]uint64{{3, 4}}},
		{start: 0, end: 4, ranges: [][2]uint64{{0, 2}, {3, 4}}},
		{start: 1, end: 5, ranges: [][2]uint64{{1, 3}, {4, 5}}},
	}
	for _, testCase := range testCases {
		ranges := GetSyncRanges(testCase.start, testCase.end, maxRange)
		assert.DeepEqual(t, ranges, testCase.ranges)
	}
}
