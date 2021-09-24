package decryptor

import (
	"bytes"
	"crypto/rand"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	"gotest.tools/v3/assert"

	"github.com/shutter-network/shutter/shlib/shcrypto/shbls"
)

func TestHashChain(t *testing.T) {
	inputs := [][][]byte{
		{},
		{[]byte("value")},
		{[]byte("v1"), []byte("v2")},
		{[]byte("v1"), []byte("v2"), []byte("v3")},
	}
	expectedOutputs := [][]byte{
		make([]byte, 32),
		crypto.Keccak256([]byte("value"), make([]byte, 32)),
		crypto.Keccak256([]byte("v2"), crypto.Keccak256([]byte("v1"), make([]byte, 32))),
		crypto.Keccak256([]byte("v3"), crypto.Keccak256([]byte("v2"), crypto.Keccak256([]byte("v1"), make([]byte, 32)))),
	}

	for i := 0; i < len(inputs); i++ {
		output := hashChain(inputs[i])
		assert.Check(t, bytes.Equal(output.Bytes(), expectedOutputs[i]))
	}
}

func TestSigning(t *testing.T) {
	d := decryptionSigningData{
		instanceID:     1,
		epochID:        2,
		cipherBatch:    [][]byte{[]byte("ctx1"), []byte("ctx2")},
		decryptedBatch: [][]byte{[]byte("dtx1"), []byte("dtx2")},
	}
	secretKey, publicKey, err := shbls.RandomKeyPair(rand.Reader)
	assert.NilError(t, err)

	sig := d.sign(secretKey)
	assert.Check(t, d.verify(sig, publicKey))

	modD := decryptionSigningData{
		instanceID:     1,
		epochID:        2,
		cipherBatch:    [][]byte{},
		decryptedBatch: [][]byte{},
	}
	assert.Check(t, d.hash() != modD.hash())

	modSig := modD.sign(secretKey)
	assert.Check(t, modD.verify(modSig, publicKey))
	assert.Check(t, !d.verify(modSig, publicKey))
}
