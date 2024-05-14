// This package contains SSZ-encodable types used by the Gnosis keyper.
// The encodings are automatically generated using FastSSZ (https://github.com/ferranbt/fastssz).
// Command: `$ go run sszgen/*.go --path <path to package>`
package gnosisssztypes

import (
	"crypto/ecdsa"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/pkg/errors"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/identitypreimage"
)

type IdentityPreimage struct {
	Bytes []byte `ssz-size:"52"`
}

type SlotDecryptionSignatureData struct {
	InstanceID        uint64
	Eon               uint64
	Slot              uint64
	TxPointer         uint64
	IdentityPreimages []IdentityPreimage `ssz-max:"1024"`
}

func NewSlotDecryptionSignatureData(
	instanceID uint64,
	eon uint64,
	slot uint64,
	txPointer uint64,
	identityPreimages []identitypreimage.IdentityPreimage,
) (*SlotDecryptionSignatureData, error) {
	if len(identityPreimages) > 1024 {
		return nil, errors.New("too many identity preimages")
	}

	wrappedPreimages := []IdentityPreimage{}
	for _, preimage := range identityPreimages {
		wrappedPreimage := IdentityPreimage{
			Bytes: preimage.Bytes(),
		}
		wrappedPreimages = append(wrappedPreimages, wrappedPreimage)
	}

	return &SlotDecryptionSignatureData{
		InstanceID:        instanceID,
		Eon:               eon,
		Slot:              slot,
		TxPointer:         txPointer,
		IdentityPreimages: wrappedPreimages,
	}, nil
}

func (d *SlotDecryptionSignatureData) ComputeSignature(key *ecdsa.PrivateKey) ([]byte, error) {
	h, err := d.HashTreeRoot()
	if err != nil {
		return nil, errors.Wrap(err, "failed to compute hash tree root of slot decryption signature data")
	}
	return crypto.Sign(h[:], key)
}

func (d *SlotDecryptionSignatureData) CheckSignature(signature []byte, address common.Address) (bool, error) {
	h, err := d.HashTreeRoot()
	if err != nil {
		return false, errors.Wrap(err, "failed to compute hash tree root of slot decryption signature data")
	}
	signerPubkey, err := crypto.SigToPub(h[:], signature)
	if err != nil {
		return false, errors.Wrap(err, "failed to recover public key from slot decryption signature")
	}
	signerAddress := crypto.PubkeyToAddress(*signerPubkey)
	return signerAddress == address, nil
}
