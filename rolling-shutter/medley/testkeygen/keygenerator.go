package testkeygen

import (
	"math/rand"
	"testing"

	"gotest.tools/assert"

	"github.com/shutter-network/shutter/shlib/shcrypto"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/identitypreimage"
)

// TestKeyGenerator is a helper tool to generate secret and public eon and epoch keys and key
// shares. It will generate a new eon key every eonInterval epochs.
type TestKeyGenerator struct {
	tb          testing.TB
	eonInterval uint64
	eonKeyGen   map[uint64]*EonKeys
	NumKeypers  uint64
	Threshold   uint64
}

func NewTestKeyGenerator(tb testing.TB, numKeypers uint64, threshold uint64, infiniteInterval bool) *TestKeyGenerator {
	tb.Helper()
	eonInterval := 100
	if infiniteInterval {
		eonInterval = 0 // 0 stands for infinity
	}
	return &TestKeyGenerator{
		tb:          tb,
		eonInterval: uint64(eonInterval),
		eonKeyGen:   make(map[uint64]*EonKeys),
		NumKeypers:  numKeypers,
		Threshold:   threshold,
	}
}

// getEonIndex computes the index of the EON key to be used for the given identityPreimage. We generate a new
// eon key every eonInterval epochs.
func (tkg *TestKeyGenerator) getEonIndex(identityPreimage identitypreimage.IdentityPreimage) uint64 {
	if tkg.eonInterval == 0 {
		return 0
	}

	return identityPreimage.Big().Uint64() / tkg.eonInterval
}

func (tkg *TestKeyGenerator) EonKeysForEpoch(identityPreimage identitypreimage.IdentityPreimage) *EonKeys {
	tkg.tb.Helper()
	var err error
	eonIndex := tkg.getEonIndex(identityPreimage)
	res, ok := tkg.eonKeyGen[eonIndex]
	if !ok {
		res, err = NewEonKeys(
			rand.New(rand.NewSource(int64(eonIndex))), //nolint:gosec
			tkg.NumKeypers,
			tkg.Threshold,
		)
		assert.NilError(tkg.tb, err)
		tkg.eonKeyGen[eonIndex] = res
	}
	return res
}

func (tkg *TestKeyGenerator) EonPublicKeyShare(identityPreimage identitypreimage.IdentityPreimage,
	keyperIndex uint64,
) *shcrypto.EonPublicKeyShare {
	tkg.tb.Helper()
	return tkg.EonKeysForEpoch(identityPreimage).keyperShares[keyperIndex].eonPublicKeyShare
}

func (tkg *TestKeyGenerator) EonPublicKey(identityPreimage identitypreimage.IdentityPreimage) *shcrypto.EonPublicKey {
	tkg.tb.Helper()
	return tkg.EonKeysForEpoch(identityPreimage).publicKey
}

func (tkg *TestKeyGenerator) EonSecretKeyShare(identityPreimage identitypreimage.IdentityPreimage,
	keyperIndex uint64,
) *shcrypto.EonSecretKeyShare {
	tkg.tb.Helper()
	return tkg.EonKeysForEpoch(identityPreimage).keyperShares[keyperIndex].eonSecretKeyShare
}

func (tkg *TestKeyGenerator) EpochSecretKeyShare(identityPreimage identitypreimage.IdentityPreimage,
	keyperIndex uint64,
) *shcrypto.EpochSecretKeyShare {
	tkg.tb.Helper()
	return tkg.EonKeysForEpoch(identityPreimage).keyperShares[keyperIndex].ComputeEpochSecretKeyShare(identityPreimage)
}

func (tkg *TestKeyGenerator) EpochSecretKey(identityPreimage identitypreimage.IdentityPreimage) *shcrypto.EpochSecretKey {
	tkg.tb.Helper()
	epochSecretKey, err := tkg.EonKeysForEpoch(identityPreimage).EpochSecretKey(identityPreimage)
	assert.NilError(tkg.tb, err)
	return epochSecretKey
}
