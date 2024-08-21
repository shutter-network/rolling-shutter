package p2pmsg

import (
	"crypto/ecdsa"
	"encoding/binary"

	"golang.org/x/crypto/sha3"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/identitypreimage"
)

var triggerHashPrefix = []byte{0x19, 't', 'r', 'i', 'g', 'g', 'e', 'r'}

func NewSignedDecryptionTrigger(
	instanceID uint64, identityPreimage identitypreimage.IdentityPreimage, blockNumber uint64, txHash []byte, privKey *ecdsa.PrivateKey,
) (*DecryptionTrigger, error) {
	trigger := &DecryptionTrigger{
		InstanceId:       instanceID,
		EpochId:          identityPreimage.Bytes(),
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
	_ = binary.Write(hash, binary.BigEndian, trigger.InstanceId)
	_ = binary.Write(hash, binary.BigEndian, trigger.EpochId)
	hash.Write(trigger.TransactionsHash)
	return hash.Sum(nil)
}

func HashByteList(l [][]byte) []byte {
	hash := sha3.New256()
	for _, bytes := range l {
		h := sha3.Sum256(bytes)
		hash.Write(h[:])
	}
	return hash.Sum(nil)
}
