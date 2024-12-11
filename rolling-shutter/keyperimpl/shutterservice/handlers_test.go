package shutterservice

import (
	"context"
	"encoding/binary"
	"math/rand"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	obskeyperdatabase "github.com/shutter-network/rolling-shutter/rolling-shutter/chainobserver/db/keyper"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyperimpl/shutterservice/database"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyperimpl/shutterservice/serviceztypes"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/testsetup"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2p"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2pmsg"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/shdb"
	"gotest.tools/assert"
)

func TestHandleDecryptionKeySharesThresholdNotReached(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	ctx := context.Background()

	dbpool, dbclose := testsetup.NewTestDBPool(ctx, t, database.Definition)
	t.Cleanup(dbclose)

	obsKeyperDB := obskeyperdatabase.New(dbpool)
	privateKey, sender, err := generateRandomAccount()
	assert.NilError(t, err)
	keyperIndex := uint64(1)
	keyperConfigIndex := uint64(1)

	err = obsKeyperDB.InsertKeyperSet(ctx, obskeyperdatabase.InsertKeyperSetParams{
		KeyperConfigIndex:     int64(keyperConfigIndex),
		ActivationBlockNumber: rand.Int63(),
		Keypers:               shdb.EncodeAddresses([]common.Address{sender}),
		Threshold:             2,
	})
	assert.NilError(t, err)

	identityPreimages := []serviceztypes.IdentityPreimage{}
	for i := 0; i < 3; i++ {
		identityPreimage := serviceztypes.IdentityPreimage{
			Bytes: intTo32ByteArray(i),
		}
		identityPreimages = append(identityPreimages, identityPreimage)
	}

	decryptionData := &serviceztypes.DecryptionSignatureData{
		InstanceID:        config.GetInstanceID(),
		Eon:               keyperConfigIndex,
		IdentityPreimages: identityPreimages,
	}

	signature, err := decryptionData.ComputeSignature(privateKey)
	assert.NilError(t, err)

	keys := testsetup.InitializeEon(ctx, t, dbpool, config, keyperIndex)
	var handler p2p.MessageHandler = &DecryptionKeySharesHandler{dbpool: dbpool}
	// threshold is two, so no outgoing message after first input
	shares := []*p2pmsg.KeyShare{}
	for _, identityPreimage := range identityPreimages {
		share := &p2pmsg.KeyShare{
			IdentityPreimage: identityPreimage.Bytes,
			Share:            keys.EpochSecretKeyShare(identityPreimage.Bytes, 0).Marshal(),
		}
		shares = append(shares, share)
	}
	msg := &p2pmsg.DecryptionKeyShares{
		InstanceId:  config.GetInstanceID(),
		Eon:         keyperConfigIndex,
		KeyperIndex: 0,
		Shares:      shares,
		Extra: &p2pmsg.DecryptionKeyShares_Service{
			Service: &p2pmsg.ShutterServiceDecryptionKeySharesExtra{
				Signature: signature,
			},
		},
	}
	validation, err := handler.ValidateMessage(ctx, msg)
	assert.NilError(t, err, "validation returned error")
	assert.Equal(t, validation, pubsub.ValidationAccept)
	msgs, err := handler.HandleMessage(ctx, msg)
	assert.NilError(t, err)
	assert.Equal(t, len(msgs), 0)
}

func intTo32ByteArray(num int) []byte {
	b := make([]byte, 32)
	tempBuffer := make([]byte, 8)
	binary.BigEndian.PutUint64(tempBuffer, uint64(num))
	copy(b[24:], tempBuffer)
	return b
}
