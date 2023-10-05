package testkeygen

import (
	"math/rand"
	"testing"

	"gotest.tools/assert"

	"github.com/shutter-network/shutter/shlib/shcrypto"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/epochid"
)

// TestKeyGenerator is a helper tool to generate secret and public eon and epoch keys and key
// shares. It will generate a new eon key every eonInterval epochs.
type TestKeyGenerator struct {
	t           *testing.T
	eonInterval uint64
	eonKeyGen   map[uint64]*EonKeys
	NumKeypers  uint64
	Threshold   uint64
}

func NewTestKeyGenerator(t *testing.T, numKeypers uint64, threshold uint64) *TestKeyGenerator {
	t.Helper()
	return &TestKeyGenerator{
		t:           t,
		eonInterval: 100, // 0 stands for infinity
		eonKeyGen:   make(map[uint64]*EonKeys),
		NumKeypers:  numKeypers,
		Threshold:   threshold,
	}
}

// getEonIndex computes the index of the EON key to be used for the given epochID. We generate a new
// eon key every eonInterval epochs.
func (tkg *TestKeyGenerator) getEonIndex(epochID epochid.EpochID) uint64 {
	if tkg.eonInterval == 0 {
		return 0
	}

	return epochID.Big().Uint64() / tkg.eonInterval
}

func (tkg *TestKeyGenerator) EonKeysForEpoch(epochID epochid.EpochID) *EonKeys {
	tkg.t.Helper()
	var err error
	eonIndex := tkg.getEonIndex(epochID)
	res, ok := tkg.eonKeyGen[eonIndex]
	if !ok {
		res, err = NewEonKeys(
			rand.New(rand.NewSource(int64(eonIndex))), //nolint:gosec
			tkg.NumKeypers,
			tkg.Threshold,
		)
		assert.NilError(tkg.t, err)
		tkg.eonKeyGen[eonIndex] = res
	}
	return res
}

func (tkg *TestKeyGenerator) EonPublicKeyShare(
	epochID epochid.EpochID,
	keyperIndex uint64,
) *shcrypto.EonPublicKeyShare {
	tkg.t.Helper()
	return tkg.EonKeysForEpoch(epochID).keyperShares[keyperIndex].eonPublicKeyShare
}

func (tkg *TestKeyGenerator) EonPublicKey(epochID epochid.EpochID) *shcrypto.EonPublicKey {
	tkg.t.Helper()
	return tkg.EonKeysForEpoch(epochID).publicKey
}

func (tkg *TestKeyGenerator) EonSecretKeyShare(
	epochID epochid.EpochID,
	keyperIndex uint64,
) *shcrypto.EonSecretKeyShare {
	tkg.t.Helper()
	return tkg.EonKeysForEpoch(epochID).keyperShares[keyperIndex].eonSecretKeyShare
}

func (tkg *TestKeyGenerator) EpochSecretKeyShare(
	epochID epochid.EpochID,
	keyperIndex uint64,
) *shcrypto.EpochSecretKeyShare {
	tkg.t.Helper()
	return tkg.EonKeysForEpoch(epochID).keyperShares[keyperIndex].ComputeEpochSecretKeyShare(epochID)
}

func (tkg *TestKeyGenerator) EpochSecretKey(epochID epochid.EpochID) *shcrypto.EpochSecretKey {
	tkg.t.Helper()
	epochSecretKey, err := tkg.EonKeysForEpoch(epochID).EpochSecretKey(epochID)
	assert.NilError(tkg.t, err)
	return epochSecretKey
}
