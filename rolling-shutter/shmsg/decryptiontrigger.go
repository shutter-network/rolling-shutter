package shmsg

import (
	"crypto/ecdsa"
	"encoding/binary"

	"github.com/ethereum/go-ethereum/common"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"golang.org/x/crypto/sha3"
)

const HashPrefix = "decryptionTriggerPrefix"

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
	triggerBytes := make([]byte, 16)
	triggerBytes = append(triggerBytes, []byte(HashPrefix)...)
	binary.LittleEndian.PutUint64(triggerBytes, t.InstanceID)
	binary.LittleEndian.PutUint64(triggerBytes[8:], t.EpochID)
	triggerBytes = append(triggerBytes, t.TransactionsHash...)
	hash := sha3.Sum256(triggerBytes)
	return hash[:]
}

func hashTransactions(transactions [][]byte) []byte {
	concatenatedHashedTransactions := []byte{}
	for _, transaction := range transactions {
		transactionHash := sha3.Sum256(transaction)
		concatenatedHashedTransactions = append(concatenatedHashedTransactions, transactionHash[:]...)
	}
	hash := sha3.Sum256(concatenatedHashedTransactions)
	return hash[:]
}
