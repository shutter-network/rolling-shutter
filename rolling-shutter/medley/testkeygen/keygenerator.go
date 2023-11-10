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
	t           *testing.T
	eonInterval uint64
	eonKeyGen   map[uint64]*EonKeys
	NumKeypers  uint64
	Threshold   uint64
}

func NewTestKeyGenerator(t *testing.T, numKeypers uint64, threshold uint64, infiniteInterval bool) *TestKeyGenerator {
	t.Helper()
	generator := &TestKeyGenerator{
		t:           t,
		eonInterval: 100, // 0 stands for infinity
		eonKeyGen:   make(map[uint64]*EonKeys),
		NumKeypers:  numKeypers,
		Threshold:   threshold,
	}
	if infiniteInterval {
		generator.eonInterval = 0
	}
	return generator
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
	tkg.t.Helper()
	var err error
	eonIndex := tkg.getEonIndex(identityPreimage)
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

func (tkg *TestKeyGenerator) EonPublicKeyShare(identityPreimage identitypreimage.IdentityPreimage,
	keyperIndex uint64,
) *shcrypto.EonPublicKeyShare {
	tkg.t.Helper()
	return tkg.EonKeysForEpoch(identityPreimage).keyperShares[keyperIndex].eonPublicKeyShare
}

func (tkg *TestKeyGenerator) EonPublicKey(identityPreimage identitypreimage.IdentityPreimage) *shcrypto.EonPublicKey {
	tkg.t.Helper()
	return tkg.EonKeysForEpoch(identityPreimage).publicKey
}

func (tkg *TestKeyGenerator) EonSecretKeyShare(identityPreimage identitypreimage.IdentityPreimage,
	keyperIndex uint64,
) *shcrypto.EonSecretKeyShare {
	tkg.t.Helper()
	return tkg.EonKeysForEpoch(identityPreimage).keyperShares[keyperIndex].eonSecretKeyShare
}

func (tkg *TestKeyGenerator) EpochSecretKeyShare(identityPreimage identitypreimage.IdentityPreimage,
	keyperIndex uint64,
) *shcrypto.EpochSecretKeyShare {
	tkg.t.Helper()
	return tkg.EonKeysForEpoch(identityPreimage).keyperShares[keyperIndex].ComputeEpochSecretKeyShare(identityPreimage)
}

func (tkg *TestKeyGenerator) EpochSecretKey(identityPreimage identitypreimage.IdentityPreimage) *shcrypto.EpochSecretKey {
	tkg.t.Helper()
	epochSecretKey, err := tkg.EonKeysForEpoch(identityPreimage).EpochSecretKey(identityPreimage)
	assert.NilError(tkg.t, err)
	return epochSecretKey
}
