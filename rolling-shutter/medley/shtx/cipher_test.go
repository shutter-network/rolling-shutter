package shtx

import (
	"bytes"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	gocmp "github.com/google/go-cmp/cmp"
	"gotest.tools/v3/assert"

	"github.com/shutter-network/shutter/shlib/shtest"
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
	assert.DeepEqual(t, tx, decoded, shtest.BigIntComparer, gocmp.Comparer(bytes.Equal))
}

func makeExampleCipherTx() *CipherTransaction {
	return &CipherTransaction{
		EncryptedPayload:   []byte{1, 2, 3, 4},
		GasLimit:           big.NewInt(33333),
		InclusionFeePerGas: big.NewInt(4444),
		ExecutionFeePerGas: big.NewInt(3333),
		Nonce:              big.NewInt(55555),
	}
}
