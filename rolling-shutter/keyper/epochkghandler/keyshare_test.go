package epochkghandler

import (
	"bytes"
	"context"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"gotest.tools/assert"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/identitypreimage"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/testdb"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2p"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2p/p2ptest"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2pmsg"
)

func TestHandleDecryptionKeyShareIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	ctx := context.Background()
	_, dbpool, closedb := testdb.NewKeyperTestDB(ctx, t)
	defer closedb()

	identityPreimage := identitypreimage.Uint64ToIdentityPreimage(50)
	keyperIndex := uint64(1)

	tkg := initializeEon(ctx, t, dbpool, keyperIndex)
	var handler p2p.MessageHandler = &DecryptionKeyShareHandler{config: config, dbpool: dbpool}
	encodedDecryptionKey := tkg.EpochSecretKey(identityPreimage).Marshal()

	// threshold is two, so no outgoing message after first input
	msgs := p2ptest.MustHandleMessage(t, handler, ctx, &p2pmsg.DecryptionKeyShares{
		InstanceID:  config.GetInstanceID(),
		Eon:         config.GetEon(),
		KeyperIndex: 0,
		Shares: []*p2pmsg.KeyShare{{
			EpochID: identityPreimage.Bytes(),
			Share:   tkg.EpochSecretKeyShare(identityPreimage, 0).Marshal(),
		}},
	})
	assert.Check(t, len(msgs) == 0)

	// second message pushes us over the threshold (note that we didn't send a trigger, so the
	// share of the handler itself doesn't count)
	msgs = p2ptest.MustHandleMessage(t, handler, ctx, &p2pmsg.DecryptionKeyShares{
		InstanceID:  config.GetInstanceID(),
		Eon:         config.GetEon(),
		KeyperIndex: 2,
		Shares: []*p2pmsg.KeyShare{{
			EpochID: identityPreimage.Bytes(),
			Share:   tkg.EpochSecretKeyShare(identityPreimage, 2).Marshal(),
		}},
	})
	assert.Assert(t, len(msgs) == 1)
	msg, ok := msgs[0].(*p2pmsg.DecryptionKey)
	assert.Check(t, ok)
	assert.Check(t, msg.InstanceID == config.GetInstanceID())
	assert.Check(t, bytes.Equal(msg.EpochID, identityPreimage.Bytes()))
	assert.Check(t, bytes.Equal(msg.Key, encodedDecryptionKey))
}

func TestDecryptionKeyshareValidatorIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	_, dbpool, closedb := testdb.NewKeyperTestDB(ctx, t)
	defer closedb()

	keyperIndex := uint64(1)
	eon := config.GetEon()
	identityPreimage, _ := identitypreimage.BigToIdentityPreimage(common.Big0)
	wrongIdentityPreimage, _ := identitypreimage.BigToIdentityPreimage(common.Big1)
	tkg := initializeEon(ctx, t, dbpool, keyperIndex)
	keyshare := tkg.EpochSecretKeyShare(identityPreimage, keyperIndex).Marshal()
	var handler p2p.MessageHandler = &DecryptionKeyShareHandler{config: config, dbpool: dbpool}

	tests := []struct {
		name  string
		valid bool
		msg   *p2pmsg.DecryptionKeyShares
	}{
		{
			name:  "valid decryption key share",
			valid: true,
			msg: &p2pmsg.DecryptionKeyShares{
				InstanceID:  config.GetInstanceID(),
				Eon:         eon,
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
			name:  "invalid decryption key share wrong epoch",
			valid: false,
			msg: &p2pmsg.DecryptionKeyShares{
				InstanceID:  config.GetInstanceID(),
				Eon:         eon,
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
			name:  "invalid decryption key share wrong instance ID",
			valid: false,
			msg: &p2pmsg.DecryptionKeyShares{
				InstanceID:  config.GetInstanceID() + 1,
				Eon:         eon,
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
			name:  "invalid decryption key share wrong keyper index",
			valid: false,
			msg: &p2pmsg.DecryptionKeyShares{
				InstanceID:  config.GetInstanceID(),
				Eon:         eon,
				KeyperIndex: keyperIndex + 1,
				Shares: []*p2pmsg.KeyShare{
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
			p2ptest.MustValidateMessageResult(t, tc.valid, handler, ctx, tc.msg)
		})
	}
}
