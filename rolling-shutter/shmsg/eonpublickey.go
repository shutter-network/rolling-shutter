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
	candidate := &EonPublicKeyCandidate{
		InstanceID:        instanceID,
		PublicKey:         eonPublicKey,
		ActivationBlock:   activationBlock,
		KeyperConfigIndex: keyperConfigIndex,
		Eon:               eon,
	}
	return candidate.Sign(privKey)
}

func (e *EonPublicKey) RecoverAddress() (common.Address, error) {
	pubkey, err := ethcrypto.SigToPub(e.Candidate.Hash(), e.Signature)
	if err != nil {
		return common.Address{}, err
	}
	return ethcrypto.PubkeyToAddress(*pubkey), nil
}

func (e *EonPublicKey) VerifySignature(address common.Address) (bool, error) {
	recoveredAddress, err := e.RecoverAddress()
	if err != nil {
		return false, err
	}
	return recoveredAddress == address, nil
}

func (e *EonPublicKeyCandidate) Hash() []byte {
	hash := sha3.New256()
	hash.Write(eonPubKeyHashPrefix)
	_ = binary.Write(hash, binary.BigEndian, e.InstanceID)
	_ = binary.Write(hash, binary.BigEndian, e.ActivationBlock)
	_ = binary.Write(hash, binary.BigEndian, e.KeyperConfigIndex)
	_ = binary.Write(hash, binary.BigEndian, e.Eon)
	hash.Write(e.PublicKey)
	return hash.Sum(nil)
}

// Sign signs the eon public key candidate and returns an eon public key.
func (e *EonPublicKeyCandidate) Sign(privKey *ecdsa.PrivateKey) (*EonPublicKey, error) {
	signature, err := ethcrypto.Sign(e.Hash(), privKey)
	if err != nil {
		return nil, err
	}
	return &EonPublicKey{
		Candidate: e,
		Signature: signature,
	}, nil
}

func (e *EonPublicKey) GetInstanceID() uint64 {
	return e.Candidate.GetInstanceID()
}
