package testkeygen

import (
	"io"
	"math/big"
	"math/rand"
	"testing"

	"gotest.tools/assert"

	"github.com/shutter-network/shutter/shlib/shcrypto"
	"github.com/shutter-network/shutter/shuttermint/medley/epochid"
)

// TestKeyGenerator is a helper tool to generate secret and public eon and epoch keys and key
// shares. It will generate a new eon key every eonInterval epochs.
type TestKeyGenerator struct {
	t *testing.T

	randReader io.Reader

	eonInterval        uint64
	eonPublicKeyShares [][]*shcrypto.EonPublicKeyShare
	eonPublicKeys      []*shcrypto.EonPublicKey
	eonSecretKeyShares [][]*shcrypto.EonSecretKeyShare

	NumKeypers uint64
	Threshold  uint64
}

func NewTestKeyGenerator(t *testing.T, numKeypers uint64, threshold uint64) *TestKeyGenerator {
	t.Helper()

	src := rand.NewSource(0)
	randReader := rand.New(src)

	return &TestKeyGenerator{
		t: t,

		randReader: randReader,

		eonInterval:        100, // 0 stands for infinity
		eonPublicKeyShares: [][]*shcrypto.EonPublicKeyShare{},
		eonPublicKeys:      []*shcrypto.EonPublicKey{},
		eonSecretKeyShares: [][]*shcrypto.EonSecretKeyShare{},

		NumKeypers: numKeypers,
		Threshold:  threshold,
	}
}

func (tkg *TestKeyGenerator) populateNextEonKeys() {
	ps := []*shcrypto.Polynomial{}
	gammas := []*shcrypto.Gammas{}
	for i := 0; i < int(tkg.NumKeypers); i++ {
		p, err := shcrypto.RandomPolynomial(tkg.randReader, tkg.Threshold-1)
		assert.NilError(tkg.t, err)

		ps = append(ps, p)
		gammas = append(gammas, p.Gammas())
	}
	eonPublicKey := shcrypto.ComputeEonPublicKey(gammas)

	eonPublicKeyShares := []*shcrypto.EonPublicKeyShare{}
	eonSecretKeyShares := []*shcrypto.EonSecretKeyShare{}
	for i := 0; i < int(tkg.NumKeypers); i++ {
		x := shcrypto.KeyperX(i)
		vs := []*big.Int{}
		for j := 0; j < int(tkg.NumKeypers); j++ {
			v := ps[j].Eval(x)
			vs = append(vs, v)
		}
		eonSecretKeyShare := shcrypto.ComputeEonSecretKeyShare(vs)
		eonPublicKeyShare := shcrypto.ComputeEonPublicKeyShare(i, gammas)

		eonSecretKeyShares = append(eonSecretKeyShares, eonSecretKeyShare)
		eonPublicKeyShares = append(eonPublicKeyShares, eonPublicKeyShare)
	}

	tkg.eonPublicKeys = append(tkg.eonPublicKeys, eonPublicKey)
	tkg.eonSecretKeyShares = append(tkg.eonSecretKeyShares, eonSecretKeyShares)
	tkg.eonPublicKeyShares = append(tkg.eonPublicKeyShares, eonPublicKeyShares)
}

func (tkg *TestKeyGenerator) populateEonKeysUntilEon(eonIndex uint64) {
	for uint64(len(tkg.eonPublicKeys)) <= eonIndex {
		tkg.populateNextEonKeys()
	}
}

func (tkg *TestKeyGenerator) populateEonKeysUntilEpoch(epochID uint64) {
	eonIndex := tkg.EonIndex(epochID)
	tkg.populateEonKeysUntilEon(eonIndex)
}

// EonIndex computes the index of the EON key to be used for the given epochID. We generate a new
// eon key every eonInterval epochs.
func (tkg *TestKeyGenerator) EonIndex(epochID uint64) uint64 {
	tkg.t.Helper()

	if tkg.eonInterval == 0 {
		return 0
	}

	return uint64(epochid.SequenceNumber(epochID)) / tkg.eonInterval
}

func (tkg *TestKeyGenerator) EonPublicKeyShare(epochID uint64, keyperIndex uint64) *shcrypto.EonPublicKeyShare {
	tkg.t.Helper()

	eonIndex := tkg.EonIndex(epochID)
	tkg.populateEonKeysUntilEon(eonIndex)
	return tkg.eonPublicKeyShares[eonIndex][keyperIndex]
}

func (tkg *TestKeyGenerator) EonPublicKey(epochID uint64) *shcrypto.EonPublicKey {
	tkg.t.Helper()

	eonIndex := tkg.EonIndex(epochID)
	tkg.populateEonKeysUntilEon(eonIndex)
	return tkg.eonPublicKeys[eonIndex]
}

func (tkg *TestKeyGenerator) EonSecretKeyShare(epochID uint64, keyperIndex uint64) *shcrypto.EonSecretKeyShare {
	tkg.t.Helper()

	eonIndex := tkg.EonIndex(epochID)
	tkg.populateEonKeysUntilEon(eonIndex)
	return tkg.eonSecretKeyShares[eonIndex][keyperIndex]
}

func (tkg *TestKeyGenerator) EpochSecretKeyShare(epochID uint64, keyperIndex uint64) *shcrypto.EpochSecretKeyShare {
	tkg.t.Helper()
	eonKeyShare := tkg.EonSecretKeyShare(epochID, keyperIndex)
	epochIDG1 := shcrypto.ComputeEpochID(epochID)
	return shcrypto.ComputeEpochSecretKeyShare(eonKeyShare, epochIDG1)
}

func (tkg *TestKeyGenerator) EpochSecretKey(epochID uint64) *shcrypto.EpochSecretKey {
	tkg.t.Helper()

	keyperIndices := []int{}
	epochSecretKeyShares := []*shcrypto.EpochSecretKeyShare{}
	for i := uint64(0); i < tkg.Threshold; i++ {
		keyperIndices = append(keyperIndices, int(i))
		epochSecretKeyShares = append(epochSecretKeyShares, tkg.EpochSecretKeyShare(epochID, i))
	}
	epochSecretKey, err := shcrypto.ComputeEpochSecretKey(
		keyperIndices,
		epochSecretKeyShares,
		tkg.Threshold,
	)
	assert.NilError(tkg.t, err)
	return epochSecretKey
}
