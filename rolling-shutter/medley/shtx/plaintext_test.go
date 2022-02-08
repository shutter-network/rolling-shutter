package shtx

import (
	"bytes"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"gotest.tools/v3/assert"
)

func TestRoundTripEncodingPlaintext(t *testing.T) {
	tx := makeExamplePlaintextTx()
	encoded, err := tx.Encode()
	assert.NilError(t, err)
	decoded, err := decodePlaintextTx(encoded)
	assert.NilError(t, err)
	assertEqualPlaintextTx(t, tx, decoded)
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

func assertEqualPlaintextTx(t *testing.T, tx1 *PlaintextTransaction, tx2 *PlaintextTransaction) {
	t.Helper()
	if tx1.Nonce.Cmp(tx2.Nonce) != 0 {
		t.Errorf("Nonce differ")
	}
	if tx1.Value.Cmp(tx2.Value) != 0 {
		t.Errorf("Value differ")
	}
	if tx1.GasLimit.Cmp(tx2.GasLimit) != 0 {
		t.Errorf("GasLimit differ")
	}
	if tx1.InclusionFeePerGas.Cmp(tx2.InclusionFeePerGas) != 0 {
		t.Errorf("InclusionFeePerGas differ")
	}
	if tx1.ExecutionFeePerGas.Cmp(tx2.ExecutionFeePerGas) != 0 {
		t.Errorf("ExecutionFeePerGas differ")
	}
	if !bytes.Equal(tx1.Signature, tx2.Signature) {
		t.Errorf("Signature differ")
	}
	if !bytes.Equal(tx1.Calldata, tx2.Calldata) {
		t.Errorf("Signature differ")
	}
	if !bytes.Equal(tx1.Receiver.Bytes(), tx2.Receiver.Bytes()) {
		t.Errorf("Receiver differ")
	}
}
