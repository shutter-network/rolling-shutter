package decryptor

import (
	"crypto/rand"
	"testing"

	"gotest.tools/v3/assert"

	"github.com/shutter-network/shutter/shlib/shcrypto/shbls"
)

func TestSigning(t *testing.T) {
	d := decryptionSigningData{
		instanceID:     1,
		epochID:        2,
		cipherBatch:    []byte("cipher"),
		decryptedBatch: []byte("decrypted"),
	}
	secretKey, publicKey, err := shbls.RandomKeyPair(rand.Reader)
	assert.NilError(t, err)

	sig := d.sign(secretKey)
	assert.Check(t, d.verify(sig, publicKey))

	modD := decryptionSigningData{
		instanceID:     1,
		epochID:        2,
		cipherBatch:    []byte("cipher"),
		decryptedBatch: []byte("different"),
	}
	assert.Check(t, d.hash() != modD.hash())

	modSig := modD.sign(secretKey)
	assert.Check(t, modD.verify(modSig, publicKey))
	assert.Check(t, !d.verify(modSig, publicKey))
}
