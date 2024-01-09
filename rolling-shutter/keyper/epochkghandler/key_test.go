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
	identityPreimages := []identitypreimage.IdentityPreimage{}
	for i := 0; i < 3; i++ {
		identityPreimage := identitypreimage.Uint64ToIdentityPreimage(uint64(i))
		identityPreimages = append(identityPreimages, identityPreimage)
	}
	keyperIndex := uint64(1)

	tkg := testsetup.InitializeEon(ctx, t, dbpool, config, keyperIndex)

	var handler p2p.MessageHandler = &DecryptionKeyHandler{config: config, dbpool: dbpool}
	encodedDecryptionKeys := [][]byte{}
	for _, identityPreimage := range identityPreimages {
		encodedDecryptionKey := tkg.EpochSecretKey(identityPreimage).Marshal()
		encodedDecryptionKeys = append(encodedDecryptionKeys, encodedDecryptionKey)
	}

	// send a decryption key and check that it gets inserted
	keys := []*p2pmsg.Key{}
	for i, identityPreimage := range identityPreimages {
		key := &p2pmsg.Key{
			Identity: identityPreimage.Bytes(),
			Key:      encodedDecryptionKeys[i],
		}
		keys = append(keys, key)
	}
	msgs := p2ptest.MustHandleMessage(t, handler, ctx, &p2pmsg.DecryptionKeys{
		InstanceID: config.GetInstanceID(),
		Eon:        eon,
		Keys:       keys,
	})
	assert.Check(t, len(msgs) == 0)
	for i, identityPreimage := range identityPreimages {
		key, err := queries.GetDecryptionKey(ctx, database.GetDecryptionKeyParams{
			Eon:     int64(eon),
			EpochID: identityPreimage.Bytes(),
		})
		assert.NilError(t, err)
		assert.Check(t, bytes.Equal(key.DecryptionKey, encodedDecryptionKeys[i]))
	}
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
	secondIdentityPreimage := identitypreimage.BigToIdentityPreimage(common.Big1)
	wrongIdentityPreimage := identitypreimage.BigToIdentityPreimage(common.Big2)
	tkg := testsetup.InitializeEon(ctx, t, dbpool, config, keyperIndex)
	secretKey := tkg.EpochSecretKey(identityPreimage).Marshal()
	secondSecretKey := tkg.EpochSecretKey(secondIdentityPreimage).Marshal()

	var handler p2p.MessageHandler = &DecryptionKeyHandler{config: config, dbpool: dbpool}
	tests := []struct {
		name             string
		validationResult pubsub.ValidationResult
		msg              *p2pmsg.DecryptionKeys
	}{
		{
			name:             "valid decryption key",
			validationResult: pubsub.ValidationAccept,
			msg: &p2pmsg.DecryptionKeys{
				InstanceID: config.GetInstanceID(),
				Eon:        eon,
				Keys: []*p2pmsg.Key{
					{
						Identity: identityPreimage.Bytes(),
						Key:      secretKey,
					},
				},
			},
		},
		{
			name:             "invalid decryption key wrong epoch",
			validationResult: pubsub.ValidationReject,
			msg: &p2pmsg.DecryptionKeys{
				InstanceID: config.GetInstanceID(),
				Eon:        eon,
				Keys: []*p2pmsg.Key{
					{
						Identity: wrongIdentityPreimage.Bytes(),
						Key:      secretKey,
					},
				},
			},
		},
		{
			name:             "invalid decryption key wrong instance ID",
			validationResult: pubsub.ValidationReject,
			msg: &p2pmsg.DecryptionKeys{
				InstanceID: config.GetInstanceID() + 1,
				Eon:        eon,
				Keys: []*p2pmsg.Key{
					{
						Identity: identityPreimage.Bytes(),
						Key:      secretKey,
					},
				},
			},
		},
		{
			name:             "invalid decryption key empty",
			validationResult: pubsub.ValidationReject,
			msg: &p2pmsg.DecryptionKeys{
				InstanceID: config.GetInstanceID(),
				Eon:        eon,
				Keys:       []*p2pmsg.Key{},
			},
		},
		{
			name:             "invalid decryption key duplicate",
			validationResult: pubsub.ValidationReject,
			msg: &p2pmsg.DecryptionKeys{
				InstanceID: config.GetInstanceID(),
				Eon:        eon,
				Keys: []*p2pmsg.Key{
					{
						Identity: identityPreimage.Bytes(),
						Key:      secretKey,
					},
					{
						Identity: identityPreimage.Bytes(),
						Key:      secretKey,
					},
				},
			},
		},
		{
			name:             "invalid decryption key unordered",
			validationResult: pubsub.ValidationReject,
			msg: &p2pmsg.DecryptionKeys{
				InstanceID: config.GetInstanceID(),
				Eon:        eon,
				Keys: []*p2pmsg.Key{
					{
						Identity: secondIdentityPreimage.Bytes(),
						Key:      secondSecretKey,
					},
					{
						Identity: identityPreimage.Bytes(),
						Key:      secretKey,
					},
				},
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p2ptest.MustValidateMessageResult(t, tc.validationResult, handler, ctx, tc.msg)
		})
	}
}
