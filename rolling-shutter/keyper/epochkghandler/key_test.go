package epochkghandler

import (
	"bytes"
	"context"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"gotest.tools/assert"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/database"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/identitypreimage"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/testsetup"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2p"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2p/p2ptest"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2pmsg"
)

func TestHandleDecryptionKeyIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	ctx := context.Background()

	dbpool, dbclose := testsetup.NewTestDBPool(ctx, t, database.Definition)
	t.Cleanup(dbclose)

	queries := database.New(dbpool)

	eon := config.GetEon()
	identityPreimage := identitypreimage.Uint64ToIdentityPreimage(50)
	keyperIndex := uint64(1)

	tkg := testsetup.InitializeEon(ctx, t, dbpool, config, keyperIndex)

	var handler p2p.MessageHandler = &DecryptionKeyHandler{config: config, dbpool: dbpool}
	encodedDecryptionKey := tkg.EpochSecretKey(identityPreimage).Marshal()

	// send a decryption key and check that it gets inserted
	msgs := p2ptest.MustHandleMessage(t, handler, ctx, &p2pmsg.DecryptionKey{
		InstanceID: config.GetInstanceID(),
		Eon:        eon,
		EpochID:    identityPreimage.Bytes(),
		Key:        encodedDecryptionKey,
	})
	assert.Check(t, len(msgs) == 0)
	key, err := queries.GetDecryptionKey(ctx, database.GetDecryptionKeyParams{
		Eon:     int64(eon),
		EpochID: identityPreimage.Bytes(),
	})
	assert.NilError(t, err)
	assert.Check(t, bytes.Equal(key.DecryptionKey, encodedDecryptionKey))
}

func TestDecryptionKeyValidatorIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	dbpool, dbclose := testsetup.NewTestDBPool(ctx, t, database.Definition)
	t.Cleanup(dbclose)

	keyperIndex := uint64(1)
	eon := config.GetEon()
	identityPreimage := identitypreimage.BigToIdentityPreimage(common.Big0)
	wrongIdentityPreimage := identitypreimage.BigToIdentityPreimage(common.Big1)
	tkg := testsetup.InitializeEon(ctx, t, dbpool, config, keyperIndex)
	secretKey := tkg.EpochSecretKey(identityPreimage).Marshal()

	var handler p2p.MessageHandler = &DecryptionKeyHandler{config: config, dbpool: dbpool}
	tests := []struct {
		name             string
		validationResult pubsub.ValidationResult
		msg              *p2pmsg.DecryptionKey
	}{
		{
			name:             "valid decryption key",
			validationResult: pubsub.ValidationAccept,
			msg: &p2pmsg.DecryptionKey{
				InstanceID: config.GetInstanceID(),
				Eon:        eon,
				EpochID:    identityPreimage.Bytes(),
				Key:        secretKey,
			},
		},
		{
			name:             "invalid decryption key wrong epoch",
			validationResult: pubsub.ValidationReject,
			msg: &p2pmsg.DecryptionKey{
				InstanceID: config.GetInstanceID(),
				Eon:        eon,
				EpochID:    wrongIdentityPreimage.Bytes(),
				Key:        secretKey,
			},
		},
		{
			name:             "invalid decryption key wrong instance ID",
			validationResult: pubsub.ValidationReject,
			msg: &p2pmsg.DecryptionKey{
				InstanceID: config.GetInstanceID() + 1,
				Eon:        eon,
				EpochID:    identityPreimage.Bytes(),
				Key:        secretKey,
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p2ptest.MustValidateMessageResult(t, tc.validationResult, handler, ctx, tc.msg)
		})
	}
}
