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

func TestHandleDecryptionKeyShareIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	ctx := context.Background()

	dbpool, dbclose := testsetup.NewTestDBPool(ctx, t, database.Definition)
	t.Cleanup(dbclose)

	identityPreimages := []identitypreimage.IdentityPreimage{}
	for i := 0; i < 3; i++ {
		identityPreimage := identitypreimage.Uint64ToIdentityPreimage(uint64(i))
		identityPreimages = append(identityPreimages, identityPreimage)
	}
	keyperIndex := uint64(1)
	keyperConfigIndex := uint64(1)

	tkg := testsetup.InitializeEon(ctx, t, dbpool, config, keyperIndex)
	var handler p2p.MessageHandler = &DecryptionKeyShareHandler{config: config, dbpool: dbpool}
	encodedDecryptionKeys := [][]byte{}
	for _, identityPreimage := range identityPreimages {
		encodedDecryptionKey := tkg.EpochSecretKey(identityPreimage).Marshal()
		encodedDecryptionKeys = append(encodedDecryptionKeys, encodedDecryptionKey)
	}

	// threshold is two, so no outgoing message after first input
	shares := []*p2pmsg.KeyShare{}
	for _, identityPreimage := range identityPreimages {
		share := &p2pmsg.KeyShare{
			EpochID: identityPreimage.Bytes(),
			Share:   tkg.EpochSecretKeyShare(identityPreimage, 0).Marshal(),
		}
		shares = append(shares, share)
	}
	msgs := p2ptest.MustHandleMessage(t, handler, ctx, &p2pmsg.DecryptionKeyShares{
		InstanceID:  config.GetInstanceID(),
		Eon:         keyperConfigIndex,
		KeyperIndex: 0,
		Shares:      shares,
	})
	assert.Check(t, len(msgs) == 0)

	// second message pushes us over the threshold (note that we didn't send a trigger, so the
	// share of the handler itself doesn't count)
	shares = []*p2pmsg.KeyShare{}
	for _, identityPreimage := range identityPreimages {
		share := &p2pmsg.KeyShare{
			EpochID: identityPreimage.Bytes(),
			Share:   tkg.EpochSecretKeyShare(identityPreimage, 2).Marshal(),
		}
		shares = append(shares, share)
	}
	msgs = p2ptest.MustHandleMessage(t, handler, ctx, &p2pmsg.DecryptionKeyShares{
		InstanceID:  config.GetInstanceID(),
		Eon:         keyperConfigIndex,
		KeyperIndex: 2,
		Shares:      shares,
	})
	assert.Assert(t, len(msgs) == 1)
	msg, ok := msgs[0].(*p2pmsg.DecryptionKeys)
	assert.Check(t, ok)
	assert.Check(t, msg.InstanceID == config.GetInstanceID())
	assert.Check(t, len(msg.Keys) == len(identityPreimages))
	for i, key := range msg.Keys {
		assert.Check(t, bytes.Equal(key.Identity, identityPreimages[i].Bytes()))
		assert.Check(t, bytes.Equal(key.Key, encodedDecryptionKeys[i]))
	}
}

func TestDecryptionKeyshareValidatorIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()

	dbpool, dbclose := testsetup.NewTestDBPool(ctx, t, database.Definition)
	t.Cleanup(dbclose)

	keyperIndex := uint64(1)
	keyperConfigIndex := uint64(1)
	identityPreimage := identitypreimage.BigToIdentityPreimage(common.Big0)
	secondIdentityPreimage := identitypreimage.BigToIdentityPreimage(common.Big1)
	wrongIdentityPreimage := identitypreimage.BigToIdentityPreimage(common.Big2)
	tkg := testsetup.InitializeEon(ctx, t, dbpool, config, keyperIndex)
	keyshare := tkg.EpochSecretKeyShare(identityPreimage, keyperIndex).Marshal()
	secondKeyshare := tkg.EpochSecretKeyShare(secondIdentityPreimage, keyperIndex).Marshal()
	var handler p2p.MessageHandler = &DecryptionKeyShareHandler{config: config, dbpool: dbpool}

	tests := []struct {
		name             string
		validationResult pubsub.ValidationResult
		msg              *p2pmsg.DecryptionKeyShares
	}{
		{
			name:             "valid decryption key share",
			validationResult: pubsub.ValidationAccept,
			msg: &p2pmsg.DecryptionKeyShares{
				InstanceID:  config.GetInstanceID(),
				Eon:         keyperConfigIndex,
				KeyperIndex: keyperIndex,
				Shares: []*p2pmsg.KeyShare{
					{
						EpochID: identityPreimage.Bytes(),
						Share:   keyshare,
					},
					{
						EpochID: secondIdentityPreimage.Bytes(),
						Share:   secondKeyshare,
					},
				},
			},
		},
		{
			name:             "invalid decryption key share wrong epoch",
			validationResult: pubsub.ValidationReject,
			msg: &p2pmsg.DecryptionKeyShares{
				InstanceID:  config.GetInstanceID(),
				Eon:         keyperConfigIndex,
				KeyperIndex: keyperIndex,
				Shares: []*p2pmsg.KeyShare{
					{
						EpochID: wrongIdentityPreimage.Bytes(),
						Share:   keyshare,
					},
				},
			},
		},
		{
			name:             "invalid decryption key share wrong instance ID",
			validationResult: pubsub.ValidationReject,
			msg: &p2pmsg.DecryptionKeyShares{
				InstanceID:  config.GetInstanceID() + 1,
				Eon:         keyperConfigIndex,
				KeyperIndex: keyperIndex,
				Shares: []*p2pmsg.KeyShare{
					{
						EpochID: identityPreimage.Bytes(),
						Share:   keyshare,
					},
				},
			},
		},
		{
			name:             "invalid decryption key share wrong keyper index",
			validationResult: pubsub.ValidationReject,
			msg: &p2pmsg.DecryptionKeyShares{
				InstanceID:  config.GetInstanceID(),
				Eon:         keyperConfigIndex,
				KeyperIndex: keyperIndex + 1,
				Shares: []*p2pmsg.KeyShare{
					{
						EpochID: identityPreimage.Bytes(),
						Share:   keyshare,
					},
				},
			},
		},
		{
			name:             "invalid decryption key share empty",
			validationResult: pubsub.ValidationReject,
			msg: &p2pmsg.DecryptionKeyShares{
				InstanceID:  config.GetInstanceID(),
				Eon:         keyperConfigIndex,
				KeyperIndex: keyperIndex,
				Shares:      []*p2pmsg.KeyShare{},
			},
		},
		{
			name:             "invalid decryption key share unordered",
			validationResult: pubsub.ValidationReject,
			msg: &p2pmsg.DecryptionKeyShares{
				InstanceID:  config.GetInstanceID(),
				Eon:         keyperConfigIndex,
				KeyperIndex: keyperIndex,
				Shares: []*p2pmsg.KeyShare{
					{
						EpochID: secondIdentityPreimage.Bytes(),
						Share:   secondKeyshare,
					},
					{
						EpochID: identityPreimage.Bytes(),
						Share:   keyshare,
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
