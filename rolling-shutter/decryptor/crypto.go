package decryptor

import (
	"encoding/binary"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/shutter-network/shutter/shlib/shcrypto"
	"github.com/shutter-network/shutter/shlib/shcrypto/shbls"
)

type DecryptionSigningData struct {
	InstanceID     uint64
	EpochID        uint64
	CipherBatch    [][]byte
	DecryptedBatch [][]byte
}

// hashChain computes a hash over the given slice. The empty slice is mapped to the zero hash. A
// non-empty slice is mapped to `keccak256(data[-1], hashChain(data[:-1]))``.
func hashChain(data [][]byte) common.Hash {
	h := make([]byte, 32)
	for _, d := range data {
		h = crypto.Keccak256(d, h)
	}
	return common.BytesToHash(h)
}

// Hash computes the Hash over the whole struct, which is the data that should be signed.
func (d DecryptionSigningData) Hash() common.Hash {
	s := crypto.NewKeccakState()
	b := make([]byte, 8)

	binary.BigEndian.PutUint64(b, d.InstanceID)
	s.Write(b)

	binary.BigEndian.PutUint64(b, d.EpochID)
	s.Write(b)

	s.Write(hashChain(d.CipherBatch).Bytes())
	s.Write(hashChain(d.DecryptedBatch).Bytes())

	h := s.Sum([]byte{})
	return common.BytesToHash(h)
}

// Sign signs the data in the struct with the given secret key.
func (d DecryptionSigningData) Sign(secretKey *shbls.SecretKey) *shbls.Signature {
	return shbls.Sign(d.Hash().Bytes(), secretKey)
}

// Verify checks that the given public key created the given signature over the data in the struct.
func (d DecryptionSigningData) Verify(signature *shbls.Signature, publicKey *shbls.PublicKey) bool {
	return shbls.Verify(signature, publicKey, d.Hash().Bytes())
}

func decryptCipherBatch(cipherBatch [][]byte, key *shcrypto.EpochSecretKey) [][]byte {
	decryptedBatch := make([][]byte, len(cipherBatch))

	for i, tx := range cipherBatch {
		encryptedMessage := shcrypto.EncryptedMessage{}
		if err := encryptedMessage.Unmarshal(tx); err != nil {
			decryptedBatch[i] = []byte{}
			continue
		}

		decryptedTx, err := encryptedMessage.Decrypt(key)
		if err != nil {
			decryptedBatch[i] = []byte{}
			continue
		}

		decryptedBatch[i] = decryptedTx
	}

	return decryptedBatch
}
