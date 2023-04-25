package testkeygen

import (
	"io"
	"math/big"

	"github.com/shutter-network/shutter/shlib/shcrypto"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/epochid"
)

// KeyperKeyShares holds the public and private key shares of a single keyper.
type KeyperKeyShares struct {
	eonPublicKeyShare *shcrypto.EonPublicKeyShare
	eonSecretKeyShare *shcrypto.EonSecretKeyShare
}

// ComputeEpochSecretKeyShare computes the secret key share for the given epoch.
func (kks *KeyperKeyShares) ComputeEpochSecretKeyShare(epochID epochid.EpochID) *shcrypto.EpochSecretKeyShare {
	epochIDG1 := shcrypto.ComputeEpochID(epochID.Bytes())
	return shcrypto.ComputeEpochSecretKeyShare(kks.eonSecretKeyShare, epochIDG1)
}

// EonKeys holds all keys for one eon.
type EonKeys struct {
	publicKey    *shcrypto.EonPublicKey
	keyperShares []KeyperKeyShares
	NumKeypers   uint64
	Threshold    uint64
}

func NewEonKeys(random io.Reader, numKeypers uint64, threshold uint64) (*EonKeys, error) {
	ps := []*shcrypto.Polynomial{}
	gammas := []*shcrypto.Gammas{}
	for i := 0; i < int(numKeypers); i++ {
		p, err := shcrypto.RandomPolynomial(random, threshold-1)
		if err != nil {
			return nil, err
		}

		ps = append(ps, p)
		gammas = append(gammas, p.Gammas())
	}
	publicKey := shcrypto.ComputeEonPublicKey(gammas)

	shares := []KeyperKeyShares{}
	for i := 0; i < int(numKeypers); i++ {
		x := shcrypto.KeyperX(i)
		vs := []*big.Int{}
		for j := 0; j < int(numKeypers); j++ {
			v := ps[j].Eval(x)
			vs = append(vs, v)
		}
		shares = append(shares, KeyperKeyShares{
			eonSecretKeyShare: shcrypto.ComputeEonSecretKeyShare(vs),
			eonPublicKeyShare: shcrypto.ComputeEonPublicKeyShare(i, gammas),
		})
	}

	return &EonKeys{
		publicKey:    publicKey,
		keyperShares: shares,
		NumKeypers:   numKeypers,
		Threshold:    threshold,
	}, nil
}

func (eonkeys *EonKeys) getEpochSecretKeyShares(
	epochID epochid.EpochID,
	keyperIndices []int,
) []*shcrypto.EpochSecretKeyShare {
	res := make([]*shcrypto.EpochSecretKeyShare, 0, len(keyperIndices))
	for _, i := range keyperIndices {
		res = append(res, eonkeys.keyperShares[i].ComputeEpochSecretKeyShare(epochID))
	}
	return res
}

func (eonkeys *EonKeys) EpochSecretKey(epochID epochid.EpochID) (*shcrypto.EpochSecretKey, error) {
	keyperIndices := []int{}
	for i := uint64(0); i < eonkeys.Threshold; i++ {
		keyperIndices = append(keyperIndices, int(i))
	}
	epochSecretKeyShares := eonkeys.getEpochSecretKeyShares(epochID, keyperIndices)
	return shcrypto.ComputeEpochSecretKey(
		keyperIndices,
		epochSecretKeyShares,
		eonkeys.Threshold,
	)
}
