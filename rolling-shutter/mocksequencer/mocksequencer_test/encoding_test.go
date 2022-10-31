package sequencer_test

import (
	"context"
	"testing"

	txtypes "github.com/shutter-network/txtypes/types"
	"gotest.tools/assert"
)

// TestPayloadNotRLPEncoded makes sure that adding the
// decrypted payload field to a ShutterTx does not
// influence the rlp based hash value of the object.
// This is needed since the decrypted payload
// is used to hold the data after decryption
// for internal usage or transaction retrieval
// via RPC,
// but should not be used when rlp-encoding and
// hashing the transaction.
func TestPayloadNotRLPEncoded(t *testing.T) {
	fixtures, err := NewFixtures(context.Background(), 1, false)
	assert.NilError(t, err)

	shtxInner, err := fixtures.MakeShutterTx(0, 1, 42, nil)
	assert.NilError(t, err)

	shtx1 := txtypes.NewTx(shtxInner)
	hsh1 := shtx1.Hash()

	// remove the payload
	// this should not have an impact on anything
	// that relies on the RLP encoding (e.g. Hashing)
	shtxInner.Payload = nil

	shtx2 := txtypes.NewTx(shtxInner)
	hsh2 := shtx2.Hash()

	assert.Equal(t, hsh1.Hex(), hsh2.Hex())
}
