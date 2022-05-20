package shmsg

import (
	"crypto/ecdsa"

	"github.com/ethereum/go-ethereum/common"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
)

type Signable interface {
	Hash() []byte
	SetSignature([]byte)
	GetSignature() []byte
}

func Sign(s Signable, privKey *ecdsa.PrivateKey) error {
	signature, err := ethcrypto.Sign(s.Hash(), privKey)
	if err != nil {
		return err
	}
	s.SetSignature(signature)
	return nil
}

func RecoverAddress(s Signable) (common.Address, error) {
	pubkey, err := ethcrypto.SigToPub(s.Hash(), s.GetSignature())
	if err != nil {
		return common.Address{}, err
	}
	return ethcrypto.PubkeyToAddress(*pubkey), nil
}

func VerifySignature(s Signable, address common.Address) (bool, error) {
	recoveredAddress, err := RecoverAddress(s)
	if err != nil {
		return false, err
	}
	return recoveredAddress == address, nil
}
