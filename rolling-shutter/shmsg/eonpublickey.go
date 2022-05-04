package shmsg

import (
	"crypto/ecdsa"
	"encoding/binary"

	"github.com/ethereum/go-ethereum/common"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
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
	pubKey := &EonPublicKey{
		InstanceID:        instanceID,
		PublicKey:         eonPublicKey,
		ActivationBlock:   activationBlock,
		KeyperIndex:       keyperIndex,
		KeyperConfigIndex: keyperConfigIndex,
		Eon:               eon,
	}
	var err error

	pubKey.Signature, err = ethcrypto.Sign(pubKey.Hash(), privKey)
	if err != nil {
		return nil, err
	}
	return pubKey, nil
}

func (e *EonPublicKey) VerifySignature(address common.Address) (bool, error) {
	pubkey, err := ethcrypto.SigToPub(e.Hash(), e.Signature)
	if err != nil {
		return false, err
	}
	recoveredAddress := ethcrypto.PubkeyToAddress(*pubkey)
	return recoveredAddress == address, nil
}

func (e *EonPublicKey) Hash() []byte {
	hash := sha3.New256()
	hash.Write(eonPubKeyHashPrefix)
	_ = binary.Write(hash, binary.BigEndian, e.InstanceID)
	_ = binary.Write(hash, binary.BigEndian, e.ActivationBlock)
	_ = binary.Write(hash, binary.BigEndian, e.KeyperIndex)
	_ = binary.Write(hash, binary.BigEndian, e.KeyperConfigIndex)
	_ = binary.Write(hash, binary.BigEndian, e.Eon)
	hash.Write(e.PublicKey)
	return hash.Sum(nil)
}
