package shtx

import (
	"bytes"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	gocmp "github.com/google/go-cmp/cmp"
	"gotest.tools/v3/assert"

	"github.com/shutter-network/shutter/shlib/shtest"
)

func TestRoundTripEncodingPlaintext(t *testing.T) {
	tx := makeExamplePlaintextTx()
	encoded, err := tx.Encode()
	assert.NilError(t, err)
	decoded, err := decodePlaintextTx(encoded)
	assert.NilError(t, err)
	assert.DeepEqual(t, tx, decoded, shtest.BigIntComparer, gocmp.Comparer(bytes.Equal))
}

func makeExamplePlaintextTx() *PlaintextTransaction {
	return &PlaintextTransaction{
		Receiver:           common.Address{},
		Calldata:           []byte{1, 2, 3, 4},
		Value:              big.NewInt(1111),
		GasLimit:           big.NewInt(222222),
		InclusionFeePerGas: big.NewInt(3333),
		ExecutionFeePerGas: big.NewInt(3456),
		Nonce:              big.NewInt(999),
		Signature:          []byte{8, 9, 10, 11},
	}
}
