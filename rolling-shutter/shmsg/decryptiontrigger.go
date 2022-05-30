package shmsg

import (
	"crypto/ecdsa"
	"encoding/binary"

	"golang.org/x/crypto/sha3"
)

var triggerHashPrefix = []byte{0x19, 't', 'r', 'i', 'g', 'g', 'e', 'r'}

func NewSignedDecryptionTrigger(
	instanceID uint64, epochID uint64, blockNumber uint64, txHash []byte, privKey *ecdsa.PrivateKey,
) (*DecryptionTrigger, error) {
	trigger := &DecryptionTrigger{
		InstanceID:       instanceID,
		EpochID:          epochID,
		BlockNumber:      blockNumber,
		TransactionsHash: txHash,
	}
	err := Sign(trigger, privKey)
	if err != nil {
		return nil, err
	}
	return trigger, nil
}

func (trigger *DecryptionTrigger) SetSignature(s []byte) {
	trigger.Signature = s
}

func (trigger *DecryptionTrigger) Hash() []byte {
	hash := sha3.New256()
	hash.Write(triggerHashPrefix)
	_ = binary.Write(hash, binary.BigEndian, trigger.InstanceID)
	_ = binary.Write(hash, binary.BigEndian, trigger.EpochID)
	hash.Write(trigger.TransactionsHash)
	return hash.Sum(nil)
}

func HashTransactions(transactions [][]byte) []byte {
	hash := sha3.New256()
	for _, transaction := range transactions {
		h := sha3.Sum256(transaction)
		hash.Write(h[:])
	}
	return hash.Sum(nil)
}
