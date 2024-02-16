package gnosis

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/binary"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/identitypreimage"
)

type SlotDecryptionSignatureData struct {
	InstanceID        uint64
	Eon               uint64
	Slot              uint64
	TxPointer         uint64
	IdentityPreimages []identitypreimage.IdentityPreimage
}

func HashSlotDecryptionSignatureData(data *SlotDecryptionSignatureData) common.Hash {
	// all fields are fixed size and identity preimages are ordered, so there is no malleability if we just append all fields
	buf := new(bytes.Buffer)
	_ = binary.Write(buf, binary.BigEndian, data.InstanceID)
	_ = binary.Write(buf, binary.BigEndian, data.Eon)
	_ = binary.Write(buf, binary.BigEndian, data.Slot)
	_ = binary.Write(buf, binary.BigEndian, data.TxPointer)
	for _, preimage := range data.IdentityPreimages {
		_ = binary.Write(buf, binary.BigEndian, preimage)
	}
	return crypto.Keccak256Hash(buf.Bytes())
}

func ComputeSlotDecryptionSignature(
	data *SlotDecryptionSignatureData,
	key *ecdsa.PrivateKey,
) ([]byte, error) {
	h := HashSlotDecryptionSignatureData(data)
	return crypto.Sign(h.Bytes(), key)
}

func CheckSlotDecryptionSignature(data *SlotDecryptionSignatureData, signature []byte, address common.Address) (bool, error) {
	h := HashSlotDecryptionSignatureData(data)
	signerPubkey, err := crypto.SigToPub(h.Bytes(), signature)
	if err != nil {
		return false, err
	}
	signerAddress := crypto.PubkeyToAddress(*signerPubkey)
	return signerAddress == address, nil
}
