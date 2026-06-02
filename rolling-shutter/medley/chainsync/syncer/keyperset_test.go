package syncer

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestMissingIndices(t *testing.T) {
	tests := []struct {
		name  string
		known []int64
		total uint64
		want  []uint64
	}{
		{
			name:  "nothing known, fetch all",
			known: nil,
			total: 4,
			want:  []uint64{0, 1, 2, 3},
		},
		{
			name:  "nothing known, empty contract",
			known: nil,
			total: 0,
			want:  nil,
		},
		{
			name:  "all known",
			known: []int64{0, 1, 2, 3},
			total: 4,
			want:  nil,
		},
		{
			name:  "gap at the start",
			known: []int64{3, 4, 5},
			total: 6,
			want:  []uint64{0, 1, 2},
		},
		{
			name:  "gap at the end",
			known: []int64{0, 1, 2},
			total: 6,
			want:  []uint64{3, 4, 5},
		},
		{
			name:  "gap in the middle",
			known: []int64{0, 1, 5, 6},
			total: 7,
			want:  []uint64{2, 3, 4},
		},
		{
			name:  "known indices beyond total are ignored",
			known: []int64{0, 1, 10, 20},
			total: 4,
			want:  []uint64{2, 3},
		},
		{
			name:  "duplicate known indices",
			known: []int64{1, 1, 2, 2},
			total: 4,
			want:  []uint64{0, 3},
		},
		{
			name:  "unordered known indices",
			known: []int64{4, 1, 3},
			total: 5,
			want:  []uint64{0, 2},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.DeepEqual(t, missingIndices(tc.known, tc.total), tc.want)
		})
	}
}
