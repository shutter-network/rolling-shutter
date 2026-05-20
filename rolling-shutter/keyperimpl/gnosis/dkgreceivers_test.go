package gnosis

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestReceiverIndicesForSender(t *testing.T) {
	cases := []struct {
		n      uint64
		sender uint64
		want   []uint64
	}{
		{n: 1, sender: 0, want: []uint64{}},
		{n: 3, sender: 0, want: []uint64{1, 2}},
		{n: 3, sender: 1, want: []uint64{0, 2}},
		{n: 4, sender: 3, want: []uint64{0, 1, 2}},
	}
	for _, tc := range cases {
		got := ReceiverIndicesForSender(tc.n, tc.sender)
		assert.Equal(t, len(got), len(tc.want))
		for i := range tc.want {
			assert.Equal(t, got[i], tc.want[i])
		}
	}
}
