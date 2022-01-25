package shtx

import (
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
