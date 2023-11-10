package epochkghandler

import (
	"bytes"
	"context"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"gotest.tools/assert"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/db/kprdb"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/identitypreimage"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/testdb"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2p"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2p/p2ptest"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2pmsg"
)

func TestHandleDecryptionKeyIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	ctx := context.Background()
	db, dbpool, closedb := testdb.NewKeyperTestDB(ctx, t)
	defer closedb()

	eon := config.GetEon()
	identityPreimage := identitypreimage.Uint64ToIdentityPreimage(50)
	keyperIndex := uint64(1)

	tkg := initializeEon(ctx, t, dbpool, keyperIndex)

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
	key, err := db.GetDecryptionKey(ctx, kprdb.GetDecryptionKeyParams{
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
	_, dbpool, closedb := testdb.NewKeyperTestDB(ctx, t)
	defer closedb()

	keyperIndex := uint64(1)
	eon := config.GetEon()
	identityPreimage, _ := identitypreimage.BigToIdentityPreimage(common.Big0)
	wrongIdentityPreimage, _ := identitypreimage.BigToIdentityPreimage(common.Big1)
	tkg := initializeEon(ctx, t, dbpool, keyperIndex)
	secretKey := tkg.EpochSecretKey(identityPreimage).Marshal()

	var handler p2p.MessageHandler = &DecryptionKeyHandler{config: config, dbpool: dbpool}
	tests := []struct {
		name  string
		valid bool
		msg   *p2pmsg.DecryptionKey
	}{
		{
			name:  "valid decryption key",
			valid: true,
			msg: &p2pmsg.DecryptionKey{
				InstanceID: config.GetInstanceID(),
				Eon:        eon,
				EpochID:    identityPreimage.Bytes(),
				Key:        secretKey,
			},
		},
		{
			name:  "invalid decryption key wrong epoch",
			valid: false,
			msg: &p2pmsg.DecryptionKey{
				InstanceID: config.GetInstanceID(),
				Eon:        eon,
				EpochID:    wrongIdentityPreimage.Bytes(),
				Key:        secretKey,
			},
		},
		{
			name:  "invalid decryption key wrong instance ID",
			valid: false,
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
			p2ptest.MustValidateMessageResult(t, tc.valid, handler, ctx, tc.msg)
		})
	}
}
