package testkeygen

// Run the benchmarks with pprof enabled with:
//
//     go test -bench=. ./medley/testkeygen -cpuprofile profile.out
//
// and view results with
//
//     go tool pprof -http=: profile.out
//
import (
	"io"
	"math/rand"
	"testing"

	"github.com/pkg/errors"
	"gotest.tools/assert"

	"github.com/shutter-network/shutter/shlib/shcrypto"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/identitypreimage"
)

const (
	transactionsPerBlock   = 100
	transactionSizeInBytes = 250
	numKeypers             = 50
	threshold              = 34
)

var random io.Reader = rand.New(rand.NewSource(11)) //nolint:gosec

type BlockBuilder struct {
	cleartextTransactions [][]byte
	encryptedTransactions [][]byte
	eonkeys               *EonKeys
}

func (bb *BlockBuilder) genTx() error {
	var err error
	cleartextTx := make([]byte, transactionSizeInBytes)
	_, err = random.Read(cleartextTx)
	if err != nil {
		return err
	}

	sigma, err := shcrypto.RandomSigma(random)
	if err != nil {
		return errors.Wrap(err, "failed to generate random sigma")
	}

	identityPreimage := identitypreimage.Uint64ToIdentityPreimage(uint64(len(bb.cleartextTransactions)))
	msg := shcrypto.Encrypt(cleartextTx,
		bb.eonkeys.publicKey,
		shcrypto.ComputeEpochID(identityPreimage.Bytes()),
		sigma)

	bb.cleartextTransactions = append(bb.cleartextTransactions, cleartextTx)
	bb.encryptedTransactions = append(bb.encryptedTransactions, msg.Marshal())
	return nil
}

func NewBlockBuilder() (*BlockBuilder, error) {
	var err error
	bb := &BlockBuilder{}
	bb.eonkeys, err = NewEonKeys(random, numKeypers, threshold)
	if err != nil {
		return nil, err
	}
	for i := 0; i < transactionsPerBlock; i++ {
		err = bb.genTx()
		if err != nil {
			return nil, err
		}
	}

	return bb, nil
}

// BenchmarkKeyperComputeSecretShares benchmarks the work the keyper has to do to generate the
// secret shares for one block, where each transaction is encrypted for a different identityPreimage.
func BenchmarkKeyperComputeSecretShares(b *testing.B) {
	ek, err := NewEonKeys(random, numKeypers, threshold)
	assert.NilError(b, err)
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		for i := 0; i < transactionsPerBlock; i++ {
			ek.keyperShares[0].ComputeEpochSecretKeyShare(identitypreimage.Uint64ToIdentityPreimage(uint64(i)))
		}
	}
}

// BenchmarkSecretKeyGeneration benchmarks the generation of the secret key from the key shares
// sent by the keypers.
func BenchmarkSecretKeyGeneration(b *testing.B) {
	ek, err := NewEonKeys(random, numKeypers, threshold)
	assert.NilError(b, err)
	keyperIndices := []int{}
	for i := uint64(0); i < ek.Threshold; i++ {
		keyperIndices = append(keyperIndices, int(i))
	}

	shares := ek.getEpochSecretKeyShares(identitypreimage.Uint64ToIdentityPreimage(55), keyperIndices)
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_, err := shcrypto.ComputeEpochSecretKey(
			keyperIndices,
			shares,
			ek.Threshold,
		)
		assert.NilError(b, err)
	}
}

func decryptBlock(b *testing.B, bb *BlockBuilder, keyperIndices []int, shares [][]*shcrypto.EpochSecretKeyShare) {
	b.Helper()
	lagrangeCoeffs := shcrypto.NewLagrangeCoeffs(keyperIndices)
	for i := 0; i < len(bb.encryptedTransactions); i++ {
		secretKey, err := lagrangeCoeffs.ComputeEpochSecretKey(shares[i])
		assert.NilError(b, err)
		message := &shcrypto.EncryptedMessage{}
		err = message.Unmarshal(bb.encryptedTransactions[i])
		assert.NilError(b, err)
		decryptedBytes, err := message.Decrypt(secretKey)
		assert.NilError(b, err)
		assert.DeepEqual(b, decryptedBytes, bb.cleartextTransactions[i])
	}
}

func BenchmarkFullBlock(b *testing.B) {
	bb, err := NewBlockBuilder()
	assert.NilError(b, err)

	keyperIndices := []int{}
	for i := uint64(0); i < bb.eonkeys.Threshold; i++ {
		keyperIndices = append(keyperIndices, int(i))
	}

	shares := [][]*shcrypto.EpochSecretKeyShare{}
	for i := 0; i < len(bb.encryptedTransactions); i++ {
		shares = append(shares, bb.eonkeys.getEpochSecretKeyShares(identitypreimage.Uint64ToIdentityPreimage(uint64(i)), keyperIndices))
	}

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		decryptBlock(b, bb, keyperIndices, shares)
	}
}
