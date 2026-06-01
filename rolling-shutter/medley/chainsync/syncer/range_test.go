package syncer

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestIndicesToRanges(t *testing.T) {
	tests := []struct {
		name    string
		indices []int64
		want    []IndexRange
	}{
		{
			name:    "empty",
			indices: nil,
			want:    nil,
		},
		{
			name:    "single",
			indices: []int64{3},
			want:    []IndexRange{{3, 3}},
		},
		{
			name:    "contiguous",
			indices: []int64{0, 1, 2, 3},
			want:    []IndexRange{{0, 3}},
		},
		{
			name:    "gap in middle",
			indices: []int64{0, 1, 3, 4},
			want:    []IndexRange{{0, 1}, {3, 4}},
		},
		{
			name:    "all separate",
			indices: []int64{1, 3, 5},
			want:    []IndexRange{{1, 1}, {3, 3}, {5, 5}},
		},
		{
			name:    "starts above zero",
			indices: []int64{5, 6, 7},
			want:    []IndexRange{{5, 7}},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.DeepEqual(t, IndicesToRanges(tc.indices), tc.want)
		})
	}
}

func TestComplementRanges(t *testing.T) {
	tests := []struct {
		name  string
		known []IndexRange
		total uint64
		want  []IndexRange
	}{
		{
			name:  "nothing known, fetch all",
			known: nil,
			total: 4,
			want:  []IndexRange{{0, 3}},
		},
		{
			name:  "nothing known, empty contract",
			known: nil,
			total: 0,
			want:  nil,
		},
		{
			name:  "all known",
			known: []IndexRange{{0, 7}},
			total: 8,
			want:  nil,
		},
		{
			name:  "historical gap only",
			known: []IndexRange{{5, 7}},
			total: 8,
			want:  []IndexRange{{0, 4}},
		},
		{
			name:  "historical gap plus new entries",
			known: []IndexRange{{5, 7}},
			total: 10,
			want:  []IndexRange{{0, 4}, {8, 9}},
		},
		{
			name:  "gap in middle",
			known: []IndexRange{{0, 2}, {5, 6}},
			total: 9,
			want:  []IndexRange{{3, 4}, {7, 8}},
		},
		{
			name:  "new entries only",
			known: []IndexRange{{0, 5}},
			total: 8,
			want:  []IndexRange{{6, 7}},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.DeepEqual(t, complementRanges(tc.known, tc.total), tc.want)
		})
	}
}
