/*Package epochkg implements the epoch key generation given the result of a successful DKG generation with puredkg.
 */
package epochkg

import (
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/shutter-network/shutter/shlib/puredkg"
	"github.com/shutter-network/shutter/shlib/shcrypto"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/identitypreimage"
)

type (
	KeyperIndex = uint64
)

type EpochKG struct {
	Eon             uint64
	NumKeypers      uint64
	Threshold       uint64
	Keyper          KeyperIndex
	SecretKeyShare  *shcrypto.EonSecretKeyShare
	PublicKey       *shcrypto.EonPublicKey
	PublicKeyShares []*shcrypto.EonPublicKeyShare

	SecretShares map[string][]*EpochSecretKeyShare
	SecretKeys   map[string]*shcrypto.EpochSecretKey
}

type EpochSecretKeyShare struct {
	Eon              uint64
	IdentityPreimage identitypreimage.IdentityPreimage
	Sender           KeyperIndex
	Share            *shcrypto.EpochSecretKeyShare
}

func NewEpochKG(puredkgResult *puredkg.Result) *EpochKG {
	return &EpochKG{
		Eon:             puredkgResult.Eon,
		NumKeypers:      puredkgResult.NumKeypers,
		Threshold:       puredkgResult.Threshold,
		Keyper:          puredkgResult.Keyper,
		SecretKeyShare:  puredkgResult.SecretKeyShare,
		PublicKey:       puredkgResult.PublicKey,
		PublicKeyShares: puredkgResult.PublicKeyShares,

		SecretShares: make(map[string][]*EpochSecretKeyShare),
		SecretKeys:   make(map[string]*shcrypto.EpochSecretKey),
	}
}

func (epochkg *EpochKG) ComputeEpochSecretKeyShare(identityPreimage identitypreimage.IdentityPreimage) *shcrypto.EpochSecretKeyShare {
	epochID := shcrypto.ComputeEpochID(identityPreimage.Bytes())
	return shcrypto.ComputeEpochSecretKeyShare(epochkg.SecretKeyShare, epochID)
}

func (epochkg *EpochKG) computeEpochSecretKey(shares []*EpochSecretKeyShare) (*shcrypto.EpochSecretKey, error) {
	var keyperIndices []int
	var epochSecretKeyShares []*shcrypto.EpochSecretKeyShare
	for _, s := range shares {
		keyperIndices = append(keyperIndices, int(s.Sender))
		epochSecretKeyShares = append(epochSecretKeyShares, s.Share)
	}
	return shcrypto.ComputeEpochSecretKey(keyperIndices, epochSecretKeyShares, epochkg.Threshold)
}

func (epochkg *EpochKG) addEpochSecretKeyShare(share *EpochSecretKeyShare) error {
	shares := epochkg.SecretShares[share.IdentityPreimage.String()]
	log.Info().Interface("a", shares).Msg("test")
	for _, s := range shares {
		if s.Sender == share.Sender {
			return errors.Errorf(
				"already have EpochSecretKeyShare from sender %d for epoch %d",
				share.Sender,
				share.IdentityPreimage)
		}
	}
	shares = append(shares, share)
	if len(shares) != int(epochkg.Threshold) {
		epochkg.SecretShares[share.IdentityPreimage.String()] = shares
		return nil
	}

	secretKey, err := epochkg.computeEpochSecretKey(shares)
	delete(epochkg.SecretShares, share.IdentityPreimage.String())
	epochkg.SecretKeys[share.IdentityPreimage.String()] = secretKey // may be nil in the error case
	return err
}

func (epochkg *EpochKG) HandleEpochSecretKeyShare(share *EpochSecretKeyShare) error {
	if _, ok := epochkg.SecretKeys[share.IdentityPreimage.String()]; ok {
		// We already have the key for this epoch
		return nil
	}
	epochID := shcrypto.ComputeEpochID(share.IdentityPreimage.Bytes())
	if !shcrypto.VerifyEpochSecretKeyShare(
		share.Share,
		epochkg.PublicKeyShares[share.Sender],
		epochID,
	) {
		return errors.Errorf(
			"cannot verify epoch secret key share from sender %d for epoch %d",
			share.Sender,
			share.IdentityPreimage)
	}
	err := epochkg.addEpochSecretKeyShare(share)
	if err != nil {
		return err
	}

	return nil
}
