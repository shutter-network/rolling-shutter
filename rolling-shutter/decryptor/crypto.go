package decryptor

import (
	"encoding/binary"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/shutter-network/shutter/shlib/shcrypto"
	"github.com/shutter-network/shutter/shlib/shcrypto/shbls"
)

type decryptionSigningData struct {
	instanceID     uint64
	epochID        uint64
	cipherBatch    []byte
	decryptedBatch []byte
}

// hash computes the hash over the whole struct, which is the data that should be signed.
func (d decryptionSigningData) hash() common.Hash {
	s := crypto.NewKeccakState()
	b := make([]byte, 8)

	binary.BigEndian.PutUint64(b, d.instanceID)
	s.Write(b)

	binary.BigEndian.PutUint64(b, d.epochID)
	s.Write(b)

	binary.BigEndian.PutUint64(b, uint64(len(d.cipherBatch)))
	s.Write(b)
	s.Write(d.cipherBatch)

	binary.BigEndian.PutUint64(b, uint64(len(d.decryptedBatch)))
	s.Write(b)
	s.Write(d.decryptedBatch)

	h := s.Sum([]byte{})
	return common.BytesToHash(h)
}

// sign signs the data in the struct with the given secret key.
func (d decryptionSigningData) sign(secretKey *shbls.SecretKey) *shbls.Signature {
	return shbls.Sign(d.hash().Bytes(), secretKey)
}

// verify checks that the given public key created the given signature over the data in the struct.
func (d decryptionSigningData) verify(signature *shbls.Signature, publicKey *shbls.PublicKey) bool {
	return shbls.Verify(signature, publicKey, d.hash().Bytes())
}

func decryptCipherBatch(cipherBatch []byte, key *shcrypto.EpochSecretKey) []byte {
	// TODO: cipher batches contain many txs that should be decrypted individually. For now, we
	// just pretend it's a single one.
	encryptedMessage := shcrypto.EncryptedMessage{}
	if err := encryptedMessage.Unmarshal(cipherBatch); err != nil {
		return []byte{}
	}

	decryptedBatch, err := encryptedMessage.Decrypt(key)
	if err != nil {
		return []byte{}
	}
	return decryptedBatch
}
