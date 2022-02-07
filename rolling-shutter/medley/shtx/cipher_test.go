package shtx

import (
	"bytes"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	"gotest.tools/v3/assert"
)

func TestSigningHash(t *testing.T) {
	cipherTx := makeExampleCipherTx()
	_, err := cipherTx.SigningHash()
	assert.NilError(t, err)
}

func TestEnvelopeSigner(t *testing.T) {
	tx := makeExampleCipherTx()

	key, err := crypto.GenerateKey()
	assert.NilError(t, err)
	expectedAddrs := crypto.PubkeyToAddress(key.PublicKey).String()

	hash, err := tx.SigningHash()
	assert.NilError(t, err)
	signature, err := crypto.Sign(hash.Bytes(), key)
	assert.NilError(t, err)
	tx.Signature = signature

	recoveredAddrs, err := tx.EnvelopeSigner()

	assert.NilError(t, err)
	assert.Equal(t, recoveredAddrs.String(), expectedAddrs)
}

func TestRoundTripEncodingCipher(t *testing.T) {
	tx := makeExampleCipherTx()

	encoded, err := tx.Encode()
	assert.NilError(t, err)
	decoded, err := decodeCipherTransaction(encoded)
	assert.NilError(t, err)
	assertEqualCipherTx(t, tx, decoded)
}

func makeExampleCipherTx() *CipherTransaction {
	exampleInt := big.NewInt(1234)
	exampleBytes := []byte{1, 2, 3, 4}
	return &CipherTransaction{
		EncryptedPayload:   exampleBytes,
		GasLimit:           exampleInt,
		InclusionFeePerGas: exampleInt,
		ExecutionFeePerGas: exampleInt,
		Nonce:              exampleInt,
	}
}

func assertEqualCipherTx(t *testing.T, tx1 *CipherTransaction, tx2 *CipherTransaction) {
	t.Helper()
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
