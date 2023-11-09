package testkeygen

import (
	"crypto/rand"

	"github.com/shutter-network/shutter/shlib/shcrypto"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/epochid"
)

// KeyGenerator is a helper tool to generate secret and public eon and epoch keys and key
// shares. It will generate a new eon key every eonInterval epochs.
type KeyGenerator struct {
	eonInterval uint64
	eonKeyGen   map[uint64]*EonKeys
	NumKeypers  uint64
	Threshold   uint64
}

func NewKeyGenerator(numKeypers uint64, threshold uint64) *KeyGenerator {
	return &KeyGenerator{
		eonInterval: 100, // 0 stands for infinity
		eonKeyGen:   make(map[uint64]*EonKeys),
		NumKeypers:  numKeypers,
		Threshold:   threshold,
	}
}

// getEonIndex computes the index of the EON key to be used for the given epochID. We generate a new
// eon key every eonInterval epochs.
func (kg *KeyGenerator) getEonIndex(epochID epochid.EpochID) uint64 {
	if kg.eonInterval == 0 {
		return 0
	}

	return epochID.Big().Uint64() / kg.eonInterval
}

func (kg *KeyGenerator) EonKeysForEpoch(epochID epochid.EpochID) *EonKeys {
	eonIndex := kg.getEonIndex(epochID)
	res, ok := kg.eonKeyGen[eonIndex]
	var err error
	if !ok {
		res, err = NewEonKeys(
			rand.Reader,
			kg.NumKeypers,
			kg.Threshold,
		)
		if err != nil {
			return nil
		}
		kg.eonKeyGen[eonIndex] = res
	}
	return res
}

func (kg *KeyGenerator) EonPublicKeyShare(epochID epochid.EpochID, keyperIndex uint64) *shcrypto.EonPublicKeyShare {
	return kg.EonKeysForEpoch(epochID).keyperShares[keyperIndex].eonPublicKeyShare
}

func (kg *KeyGenerator) EonPublicKey(epochID epochid.EpochID) *shcrypto.EonPublicKey {
	return kg.EonKeysForEpoch(epochID).publicKey
}

func (kg *KeyGenerator) EonSecretKeyShare(epochID epochid.EpochID, keyperIndex uint64) *shcrypto.EonSecretKeyShare {
	return kg.EonKeysForEpoch(epochID).keyperShares[keyperIndex].eonSecretKeyShare
}

func (kg *KeyGenerator) EpochSecretKeyShare(epochID epochid.EpochID, keyperIndex uint64) *shcrypto.EpochSecretKeyShare {
	return kg.EonKeysForEpoch(epochID).keyperShares[keyperIndex].ComputeEpochSecretKeyShare(epochID)
}

func (kg *KeyGenerator) EpochSecretKey(epochID epochid.EpochID) *shcrypto.EpochSecretKey {
	epochSecretKey, err := kg.EonKeysForEpoch(epochID).EpochSecretKey(epochID)
	if err != nil {
		panic(err)
	}
	return epochSecretKey
}

func (kg *KeyGenerator) RandomEpochID(epochbytes []byte) epochid.EpochID {
	_, err := rand.Read(epochbytes)
	if err != nil {
		panic(err)
	}

	epochID, err := epochid.BytesToEpochID(epochbytes)
	if err != nil {
		panic(err)
	}
	return epochID
}

func (kg *KeyGenerator) RandomSigma() (shcrypto.Block, error) {
	return shcrypto.RandomSigma(rand.Reader)
}
