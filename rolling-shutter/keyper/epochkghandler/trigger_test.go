package epochkghandler

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp/cmpopts"
	"gotest.tools/assert"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/database"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/broker"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/identitypreimage"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/service"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/testsetup"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2p/p2ptest"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2pmsg"
)

func TestHandleDecryptionTriggerIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()

	dbpool, dbclose := testsetup.NewTestDBPool(ctx, t, database.Definition)
	t.Cleanup(dbclose)
	db := database.New(dbpool)

	identityPreimage := identitypreimage.Uint64ToIdentityPreimage(50)
	keyperIndex := uint64(1)
	keyperConfigIndex := int64(1)

	testsetup.InitializeEon(ctx, t, dbpool, config, keyperIndex)

	decrTrigChan := make(chan *broker.Event[*DecryptionTrigger])

	messaging, err := p2ptest.NewTestMessaging()
	assert.NilError(t, err)

	ksh := &KeyShareHandler{
		InstanceID:    config.GetInstanceID(),
		KeyperAddress: config.GetAddress(),
		DBPool:        dbpool,
		Messaging:     messaging,
		Trigger:       decrTrigChan,
	}
	group, cleanup := service.RunBackground(
		ctx,
		ksh,
	)
	assert.NilError(t, err)

	trig := &DecryptionTrigger{
		BlockNumber:       42,
		IdentityPreimages: []identitypreimage.IdentityPreimage{identityPreimage},
	}
	decrTrigChan <- broker.NewEvent(trig)
	close(decrTrigChan)
	err = group.Wait()
	cleanup()
	assert.NilError(t, err)

	// send decryption key share when first trigger is received
	share, err := db.GetDecryptionKeyShare(ctx, database.GetDecryptionKeyShareParams{
		Eon:         keyperConfigIndex,
		EpochID:     identityPreimage.Bytes(),
		KeyperIndex: int64(keyperIndex),
	})
	assert.NilError(t, err)

	msg, ok := messaging.SentMessages[0].Message.(*p2pmsg.DecryptionKeyShares)
	assert.Check(t, ok)
	assert.DeepEqual(t, msg.GetShares(),
		[]*p2pmsg.KeyShare{
			{
				EpochID: identityPreimage.Bytes(),
				Share:   share.DecryptionKeyShare,
			},
		},
		cmpopts.IgnoreUnexported(p2pmsg.KeyShare{}),
	)
}
