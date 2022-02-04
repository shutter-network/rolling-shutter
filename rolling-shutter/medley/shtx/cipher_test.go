package shtx

import (
	"bytes"
	"math/big"
	"testing"

	"gotest.tools/v3/assert"
)

func TestSigningHash(t *testing.T) {
	exampleInt := big.NewInt(1234)
	exampleBytes := []byte{1, 2, 3, 4}
	cipherTx := CipherTransaction{
		EncryptedPayload:   exampleBytes,
		GasLimit:           exampleInt,
		InclusionFeePerGas: exampleInt,
		ExecutionFeePerGas: exampleInt,
		Nonce:              exampleInt,
		Signature:          exampleBytes,
	}
	_, err := cipherTx.SigningHash()
	assert.NilError(t, err)
}

func TestRoundTripEncodingCipher(t *testing.T) {
	exampleInt := big.NewInt(1234)
	exampleBytes := []byte{1, 2, 3, 4}
	tx := &CipherTransaction{
		EncryptedPayload:   exampleBytes,
		GasLimit:           exampleInt,
		InclusionFeePerGas: exampleInt,
		ExecutionFeePerGas: exampleInt,
		Nonce:              exampleInt,
		Signature:          exampleBytes,
	}
	encoded, err := tx.Encode()
	assert.NilError(t, err)
	decoded, err := DecodeCipherTransaction(encoded)
	assert.NilError(t, err)
	assertEqualCipherTx(t, tx, decoded)
}

func assertEqualCipherTx(t *testing.T, tx1 *CipherTransaction, tx2 *CipherTransaction) {
	if tx1.Nonce.Cmp(tx2.Nonce) != 0 {
		t.Errorf("Nonce differ")
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
	if !bytes.Equal(tx1.EncryptedPayload, tx2.EncryptedPayload) {
		t.Errorf("EncryptedPayload differ")
	}
}
