package epochkghandler

import (
	"context"
	"math/big"
	"testing"

	"github.com/jackc/pgx/v4/pgxpool"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/rs/zerolog"
	"gotest.tools/assert"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/database"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/identitypreimage"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/testkeygen"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/testsetup"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2p"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2pmsg"
)

// The number of identity preimages to generate for each of the benchmark runs. Note that this must
// be smaller than config.GetMaxNumKeysPerMessage(), otherwise the benchmarks will fail.
const numIdentityPreimages = 1000

func prepareBenchmark(ctx context.Context, b *testing.B, dbpool *pgxpool.Pool) (*testkeygen.EonKeys, []identitypreimage.IdentityPreimage) {
	b.Helper()
	zerolog.SetGlobalLevel(zerolog.Disabled)
	keyperIndex := uint64(1)
	identityPreimages := []identitypreimage.IdentityPreimage{}
	for i := 0; i < numIdentityPreimages; i++ {
		b := make([]byte, 52)
		big.NewInt(int64(i)).FillBytes(b)
		identityPreimage := identitypreimage.IdentityPreimage(b)
		identityPreimages = append(identityPreimages, identityPreimage)
	}

	keys := testsetup.InitializeEon(ctx, b, dbpool, config, keyperIndex)
	return keys, identityPreimages
}

func prepareKeysBenchmark(ctx context.Context, b *testing.B, dbpool *pgxpool.Pool) (p2p.MessageHandler, *p2pmsg.DecryptionKeys) {
	b.Helper()
	zerolog.SetGlobalLevel(zerolog.Disabled)
	keys, identityPreimages := prepareBenchmark(ctx, b, dbpool)

	encodedDecryptionKeys := [][]byte{}
	for _, identityPreimage := range identityPreimages {
		decryptionKey, err := keys.EpochSecretKey(identityPreimage)
		assert.NilError(b, err)
		encodedDecryptionKey := decryptionKey.Marshal()
		encodedDecryptionKeys = append(encodedDecryptionKeys, encodedDecryptionKey)
	}
	decryptionKeys := []*p2pmsg.Key{}
	for i, identityPreimage := range identityPreimages {
		key := &p2pmsg.Key{
			IdentityPreimage: identityPreimage.Bytes(),
			Key:              encodedDecryptionKeys[i],
		}
		decryptionKeys = append(decryptionKeys, key)
	}
	msg := &p2pmsg.DecryptionKeys{
		InstanceId: config.GetInstanceID(),
		Eon:        1,
		Keys:       decryptionKeys,
	}

	var handler p2p.MessageHandler = &DecryptionKeyHandler{config: config, dbpool: dbpool}

	return handler, msg
}

func prepareKeySharesBenchmark(
	ctx context.Context,
	b *testing.B,
	dbpool *pgxpool.Pool,
	isSecond bool,
) (p2p.MessageHandler, *p2pmsg.DecryptionKeyShares) {
	b.Helper()
	zerolog.SetGlobalLevel(zerolog.Disabled)
	keys, identityPreimages := prepareBenchmark(ctx, b, dbpool)
	var handler p2p.MessageHandler = &DecryptionKeyShareHandler{config: config, dbpool: dbpool}

	if isSecond {
		shares := []*p2pmsg.KeyShare{}
		keyperIndex := 0
		for _, identityPreimage := range identityPreimages {
			share := &p2pmsg.KeyShare{
				IdentityPreimage: identityPreimage.Bytes(),
				Share:            keys.EpochSecretKeyShare(identityPreimage, keyperIndex).Marshal(),
			}
			shares = append(shares, share)
		}
		msg := &p2pmsg.DecryptionKeyShares{
			InstanceId:  config.GetInstanceID(),
			Eon:         1,
			KeyperIndex: uint64(keyperIndex),
			Shares:      shares,
		}
		validationResult, err := handler.ValidateMessage(ctx, msg)
		assert.NilError(b, err)
		assert.Check(b, validationResult == pubsub.ValidationAccept)
		_, err = handler.HandleMessage(ctx, msg)
		assert.NilError(b, err)
	}

	keyperIndex := 2
	shares := []*p2pmsg.KeyShare{}
	for _, identityPreimage := range identityPreimages {
		share := &p2pmsg.KeyShare{
			IdentityPreimage: identityPreimage.Bytes(),
			Share:            keys.EpochSecretKeyShare(identityPreimage, keyperIndex).Marshal(),
		}
		shares = append(shares, share)
	}
	msg := &p2pmsg.DecryptionKeyShares{
		InstanceId:  config.GetInstanceID(),
		Eon:         1,
		KeyperIndex: uint64(keyperIndex),
		Shares:      shares,
	}

	return handler, msg
}

func BenchmarkValidateKeysIntegration(b *testing.B) {
	ctx := context.Background()
	dbpool, dbclose := testsetup.NewTestDBPool(ctx, b, database.Definition)
	b.Cleanup(dbclose)
	handler, msg := prepareKeysBenchmark(ctx, b, dbpool)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StartTimer()
		validationResult, err := handler.ValidateMessage(ctx, msg)
		b.StopTimer()
		assert.NilError(b, err)
		assert.Check(b, validationResult == pubsub.ValidationAccept)
	}
}

func BenchmarkHandleKeysIntegration(b *testing.B) {
	ctx := context.Background()
	dbpool, dbclose := testsetup.NewTestDBPool(ctx, b, database.Definition)
	b.Cleanup(dbclose)
	handler, msg := prepareKeysBenchmark(ctx, b, dbpool)

	validationResult, err := handler.ValidateMessage(ctx, msg)
	assert.NilError(b, err)
	assert.Check(b, validationResult == pubsub.ValidationAccept)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StartTimer()
		_, err = handler.HandleMessage(ctx, msg)
		b.StopTimer()
		assert.NilError(b, err)
	}
}

func BenchmarkValidateFirstKeySharesIntegration(b *testing.B) {
	ctx := context.Background()
	dbpool, dbclose := testsetup.NewTestDBPool(ctx, b, database.Definition)
	b.Cleanup(dbclose)
	handler, msg := prepareKeySharesBenchmark(ctx, b, dbpool, false)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StartTimer()
		validationResult, err := handler.ValidateMessage(ctx, msg)
		b.StopTimer()
		assert.NilError(b, err)
		assert.Check(b, validationResult == pubsub.ValidationAccept)
	}
}

func BenchmarkHandleFirstKeySharesIntegration(b *testing.B) {
	ctx := context.Background()
	dbpool, dbclose := testsetup.NewTestDBPool(ctx, b, database.Definition)
	b.Cleanup(dbclose)
	handler, msg := prepareKeySharesBenchmark(ctx, b, dbpool, false)

	validationResult, err := handler.ValidateMessage(ctx, msg)
	assert.NilError(b, err)
	assert.Check(b, validationResult == pubsub.ValidationAccept)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StartTimer()
		_, err = handler.HandleMessage(ctx, msg)
		b.StopTimer()
		assert.NilError(b, err)
	}
}

func BenchmarkValidateSecondKeySharesIntegration(b *testing.B) {
	ctx := context.Background()
	dbpool, dbclose := testsetup.NewTestDBPool(ctx, b, database.Definition)
	b.Cleanup(dbclose)
	handler, msg := prepareKeySharesBenchmark(ctx, b, dbpool, true)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StartTimer()
		validationResult, err := handler.ValidateMessage(ctx, msg)
		b.StopTimer()
		assert.NilError(b, err)
		assert.Check(b, validationResult == pubsub.ValidationAccept)
	}
}

func BenchmarkHandleSecondKeySharesIntegration(b *testing.B) {
	ctx := context.Background()
	dbpool, dbclose := testsetup.NewTestDBPool(ctx, b, database.Definition)
	b.Cleanup(dbclose)
	handler, msg := prepareKeySharesBenchmark(ctx, b, dbpool, true)

	validationResult, err := handler.ValidateMessage(ctx, msg)
	assert.NilError(b, err)
	assert.Check(b, validationResult == pubsub.ValidationAccept)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StartTimer()
		_, err = handler.HandleMessage(ctx, msg)
		b.StopTimer()
		assert.NilError(b, err)
	}
}
