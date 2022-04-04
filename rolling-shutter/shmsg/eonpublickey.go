package shmsg

import (
	"crypto/ecdsa"
	"encoding/binary"

	"github.com/ethereum/go-ethereum/common"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"golang.org/x/crypto/sha3"
)

var eonPubKeyHashPrefix = []byte{0x19, 'e', 'o', 'n', 'p', 'u', 'b'}

func NewSignedEonPublicKey(
	instanceID uint64, eonPublicKey []byte, activationBlock uint64, keyperIndex uint64, privKey *ecdsa.PrivateKey,
) (*EonPublicKey, error) {
	pubKey := &EonPublicKey{
		InstanceID:      instanceID,
		PublicKey:       eonPublicKey,
		ActivationBlock: activationBlock,
		KeyperIndex:     keyperIndex,
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
	hash.Write(e.PublicKey)
	_ = binary.Write(hash, binary.BigEndian, e.ActivationBlock)
	_ = binary.Write(hash, binary.BigEndian, e.KeyperIndex)
	return hash.Sum(nil)
}
