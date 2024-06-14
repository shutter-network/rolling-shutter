package epochkghandler

// import (
// 	"context"
// 	"testing"

// 	pubsub "github.com/libp2p/go-libp2p-pubsub"
// 	"gotest.tools/assert"

// 	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/database"
// 	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/identitypreimage"
// 	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/testsetup"
// 	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2pmsg"
// )

// func BenchmarkDecryptionKeySharesValidationIntegration(b *testing.B) {
// 	ctx := context.Background()

// 	dbpool, dbclose := testsetup.NewTestDBPool(ctx, b, database.Definition)
// 	b.Cleanup(dbclose)

// 	identityPreimages := []identitypreimage.IdentityPreimage{}
// 	for i := 0; i < 3; i++ {
// 		identityPreimage := identitypreimage.Uint64ToIdentityPreimage(uint64(i))
// 		identityPreimages = append(identityPreimages, identityPreimage)
// 	}
// 	keyperIndex := uint64(1)
// 	keyperConfigIndex := uint64(1)

// 	tkg := testsetup.InitializeEon(ctx, b, dbpool, config, keyperIndex)
// 	handler := &DecryptionKeyShareHandler{config: config, dbpool: dbpool}

// 	shares := []*p2pmsg.KeyShare{}
// 	for _, identityPreimage := range identityPreimages {
// 		share := &p2pmsg.KeyShare{
// 			EpochID: identityPreimage.Bytes(),
// 			Share:   tkg.EpochSecretKeyShare(identityPreimage, 0).Marshal(),
// 		}
// 		shares = append(shares, share)
// 	}

// 	msg := &p2pmsg.DecryptionKeyShares{
// 		InstanceID:  config.GetInstanceID(),
// 		Eon:         keyperConfigIndex,
// 		KeyperIndex: keyperIndex,
// 		Shares:      shares,
// 	}

// 	b.ResetTimer()
// 	validationResult, err := handler.ValidateMessage(ctx, msg)
// 	b.StopTimer()

// 	assert.NilError(b, err)
// 	assert.Equal(b, pubsub.ValidationAccept, validationResult)
// }
