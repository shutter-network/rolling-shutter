package serviceztypes

import (
	"crypto/ecdsa"

	"github.com/pkg/errors"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/identitypreimage"
)

type IdentityPreimage struct {
	Bytes []byte `ssz-size:"52"`
}

type DecryptionSignatureData struct {
	InstanceID        uint64
	Eon               uint64
	IdentityPreimages []IdentityPreimage `ssz-max:"1024"`
}

func NewDecryptionSignatureData(
	instanceID uint64,
	eon uint64,
	identityPreimages []identitypreimage.IdentityPreimage,
) (*DecryptionSignatureData, error) {
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

	return &DecryptionSignatureData{
		InstanceID:        instanceID,
		Eon:               eon,
		IdentityPreimages: wrappedPreimages,
	}, nil
}

func (d *DecryptionSignatureData) ComputeSignature(key *ecdsa.PrivateKey) ([]byte, error) {
	h, err := d.HashTreeRoot()
	if err != nil {
		return nil, errors.Wrap(err, "failed to compute hash tree root of slot decryption signature data")
	}
	return crypto.Sign(h[:], key)
}

func (d *DecryptionSignatureData) CheckSignature(signature []byte, address common.Address) (bool, error) {
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
