package gnosis

import (
	"bytes"
	"testing"

	"gotest.tools/v3/assert"
)

func TestPolyEvalBlobRoundtrip(t *testing.T) {
	cases := [][][]byte{
		{},
		{{0xaa}},
		{{0x01, 0x02, 0x03}, {0x04, 0x05}, {0x06}},
		{{}, {0x42}, {}},
	}
	for _, evals := range cases {
		blob, err := EncodePolyEvalBlob(evals)
		assert.NilError(t, err)
		decoded, err := DecodePolyEvalBlob(blob)
		assert.NilError(t, err)
		assert.Equal(t, len(decoded), len(evals))
		for i := range evals {
			assert.Assert(t, bytes.Equal(decoded[i], evals[i]),
				"entry %d roundtrip mismatch", i)
		}
	}
}

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
