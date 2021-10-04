package medley

import (
	"io"
	"math/big"
	"math/rand"
	"testing"

	"gotest.tools/assert"

	"github.com/shutter-network/shutter/shlib/shcrypto"
)

type TestKeyGenerator struct {
	t *testing.T

	randReader io.Reader

	eonInterval        uint64
	eonPublicKeys      []*shcrypto.EonPublicKey
	eonSecretKeyShares []*shcrypto.EonSecretKeyShare
}

func NewTestKeyGenerator(t *testing.T) *TestKeyGenerator {
	t.Helper()

	src := rand.NewSource(0)
	randReader := rand.New(src)

	return &TestKeyGenerator{
		t: t,

		randReader: randReader,

		eonInterval:        100, // 0 stands for infinity
		eonPublicKeys:      []*shcrypto.EonPublicKey{},
		eonSecretKeyShares: []*shcrypto.EonSecretKeyShare{},
	}
}

func (tkg *TestKeyGenerator) populateNextEonKeys() {
	p, err := shcrypto.RandomPolynomial(tkg.randReader, 0)
	assert.NilError(tkg.t, err)
	eonPublicKey := shcrypto.ComputeEonPublicKey([]*shcrypto.Gammas{p.Gammas()})

	v := p.EvalForKeyper(0)
	eonSecretKeyShare := shcrypto.ComputeEonSecretKeyShare([]*big.Int{v})

	tkg.eonPublicKeys = append(tkg.eonPublicKeys, eonPublicKey)
	tkg.eonSecretKeyShares = append(tkg.eonSecretKeyShares, eonSecretKeyShare)
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

func (tkg *TestKeyGenerator) EonIndex(epochID uint64) uint64 {
	tkg.t.Helper()

	if tkg.eonInterval == 0 {
		return 0
	}
	return epochID / tkg.eonInterval
}

func (tkg *TestKeyGenerator) EonPublicKey(epochID uint64) *shcrypto.EonPublicKey {
	tkg.t.Helper()

	eonIndex := tkg.EonIndex(epochID)
	tkg.populateEonKeysUntilEon(eonIndex)
	return tkg.eonPublicKeys[eonIndex]
}

func (tkg *TestKeyGenerator) EonSecretKeyShare(epochID uint64) *shcrypto.EonSecretKeyShare {
	tkg.t.Helper()

	eonIndex := tkg.EonIndex(epochID)
	tkg.populateEonKeysUntilEon(eonIndex)
	return tkg.eonSecretKeyShares[eonIndex]
}

func (tkg *TestKeyGenerator) EpochSecretKey(epochID uint64) *shcrypto.EpochSecretKey {
	tkg.t.Helper()

	tkg.populateEonKeysUntilEpoch(epochID)

	epochIDG1 := shcrypto.ComputeEpochID(epochID)
	epochSecretKeyShare := shcrypto.ComputeEpochSecretKeyShare(tkg.EonSecretKeyShare(epochID), epochIDG1)
	epochSecretKey, err := shcrypto.ComputeEpochSecretKey(
		[]int{0},
		[]*shcrypto.EpochSecretKeyShare{epochSecretKeyShare},
		1,
	)
	assert.NilError(tkg.t, err)
	return epochSecretKey
}
