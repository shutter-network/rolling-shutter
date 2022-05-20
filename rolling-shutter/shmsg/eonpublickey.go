package shmsg

import (
	"crypto/ecdsa"
	"encoding/binary"

	"golang.org/x/crypto/sha3"
)

var eonPubKeyHashPrefix = []byte{0x19, 'e', 'o', 'n', 'p', 'u', 'b'}

// NewSignedEonPublicKey creates a new eon public key and signs it with the given private key.
func NewSignedEonPublicKey(
	instanceID uint64,
	eonPublicKey []byte,
	activationBlock uint64,
	keyperIndex uint64,
	keyperConfigIndex uint64,
	eon uint64,
	privKey *ecdsa.PrivateKey,
) (*EonPublicKey, error) {
	candidate := &EonPublicKey{
		InstanceID:        instanceID,
		PublicKey:         eonPublicKey,
		ActivationBlock:   activationBlock,
		KeyperConfigIndex: keyperConfigIndex,
		Eon:               eon,
	}
	err := Sign(candidate, privKey)
	if err != nil {
		return nil, err
	}
	return candidate, nil
}

func (e *EonPublicKey) SetSignature(s []byte) {
	e.Signature = s
}

func (e *EonPublicKey) Hash() []byte {
	hash := sha3.New256()
	hash.Write(eonPubKeyHashPrefix)
	_ = binary.Write(hash, binary.BigEndian, e.InstanceID)
	_ = binary.Write(hash, binary.BigEndian, e.ActivationBlock)
	_ = binary.Write(hash, binary.BigEndian, e.KeyperConfigIndex)
	_ = binary.Write(hash, binary.BigEndian, e.Eon)
	hash.Write(e.PublicKey)
	return hash.Sum(nil)
}
