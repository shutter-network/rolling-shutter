package epochkghandler

import (
	"context"
	"testing"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"gotest.tools/assert"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/database"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/identitypreimage"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/testsetup"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2p"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2pmsg"
)

func BenchmarkDecryptionKeySharesValidationIntegration(b *testing.B) {
	ctx := context.Background()

	dbpool, dbclose := testsetup.NewTestDBPool(ctx, b, database.Definition)
	b.Cleanup(dbclose)

	keyperIndex := uint64(1)
	tkg := testsetup.InitializeEon(ctx, b, dbpool, config, keyperIndex)

	keyperConfigIndex := uint64(1)
	identityPreimages := []identitypreimage.IdentityPreimage{}
	for i := 0; i < 10000; i++ {
		identityPreimage := identitypreimage.Uint64ToIdentityPreimage(uint64(i))
		identityPreimages = append(identityPreimages, identityPreimage)
	}
	identityPreimages = sort
	var handler p2p.MessageHandler = &DecryptionKeyShareHandler{config: config, dbpool: dbpool}

	shares := []*p2pmsg.KeyShare{}
	for _, identityPreimage := range identityPreimages {
		share := &p2pmsg.KeyShare{
			EpochID: identityPreimage.Bytes(),
			Share:   tkg.EpochSecretKeyShare(identityPreimage, 0).Marshal(),
		}
		shares = append(shares, share)
	}
	msg := &p2pmsg.DecryptionKeyShares{
		InstanceID:  config.GetInstanceID(),
		Eon:         keyperConfigIndex,
		KeyperIndex: 0,
		Shares:      shares,
	}

	validationResult, err := handler.ValidateMessage(ctx, msg)
	assert.NilError(b, err)
	assert.Equal(b, pubsub.ValidationAccept, validationResult)
	_, err = handler.HandleMessage(ctx, msg)
	assert.NilError(b, err)

	shares2 := []*p2pmsg.KeyShare{}
	for _, identityPreimage := range identityPreimages {
		share := &p2pmsg.KeyShare{
			EpochID: identityPreimage.Bytes(),
			Share:   tkg.EpochSecretKeyShare(identityPreimage, 2).Marshal(),
		}
		shares2 = append(shares2, share)
	}
	msg2 := &p2pmsg.DecryptionKeyShares{
		InstanceID:  config.GetInstanceID(),
		Eon:         keyperConfigIndex,
		KeyperIndex: 2,
		Shares:      shares2,
	}

	b.ResetTimer()
	validationResult, err = handler.ValidateMessage(ctx, msg2)
	assert.NilError(b, err)
	assert.Equal(b, pubsub.ValidationAccept, validationResult)
	_, err = handler.HandleMessage(ctx, msg2)
	assert.NilError(b, err)
	b.StopTimer()

	assert.NilError(b, err)
	assert.Equal(b, pubsub.ValidationAccept, validationResult)
}
