package shmsg

import (
	"crypto/ecdsa"
	"encoding/binary"

	"github.com/ethereum/go-ethereum/common"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"golang.org/x/crypto/sha3"
)

var triggerHashPrefix = []byte{0x19, 't', 'r', 'i', 'g', 'g', 'e', 'r'}

func NewSignedDecryptionTrigger(instanceID uint64, epochID uint64, transactions [][]byte,
	privKey *ecdsa.PrivateKey) (*DecryptionTrigger, error) {
	trigger := &DecryptionTrigger{
		InstanceID:       instanceID,
		EpochID:          epochID,
		TransactionsHash: hashTransactions(transactions),
	}
	var err error
	trigger.Signature, err = ethcrypto.Sign(trigger.Hash(), privKey)
	if err != nil {
		return nil, err
	}
	return trigger, nil
}

func (t *DecryptionTrigger) VerifySignature(address common.Address) (bool, error) {
	pubkey, err := ethcrypto.SigToPub(t.Hash(), t.Signature)
	if err != nil {
		return false, err
	}
	recoveredAddress := ethcrypto.PubkeyToAddress(*pubkey)
	return recoveredAddress == address, nil
}

func (t *DecryptionTrigger) Hash() []byte {
	hash := sha3.New256()
	hash.Write(triggerHashPrefix)
	_ = binary.Write(hash, binary.BigEndian, t.InstanceID)
	_ = binary.Write(hash, binary.BigEndian, t.EpochID)
	hash.Write(t.TransactionsHash)
	return hash.Sum(nil)
}

func hashTransactions(transactions [][]byte) []byte {
	hash := sha3.New256()
	for _, transaction := range transactions {
		h := sha3.Sum256(transaction)
		hash.Write(h[:])
	}
	return hash.Sum(nil)
}
