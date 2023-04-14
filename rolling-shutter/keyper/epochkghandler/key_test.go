package epochkghandler

import (
	"bytes"
	"context"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"gotest.tools/assert"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/db/kprdb"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/epochid"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/testdb"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2p"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2pmsg"
)

func TestHandleDecryptionKeyIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	ctx := context.Background()
	db, dbpool, closedb := testdb.NewKeyperTestDB(ctx, t)
	defer closedb()

	eon := uint64(2)
	epochID := epochid.Uint64ToEpochID(50)
	keyperIndex := uint64(1)

	tkg := initializeEon(ctx, t, db, keyperIndex)

	var handler p2p.MessageHandler = &DecryptionKeyHandler{config: config, dbpool: dbpool}
	encodedDecryptionKey := tkg.EpochSecretKey(epochID).Marshal()

	// send a decryption key and check that it gets inserted
	msgs, err := handler.HandleMessage(ctx, &p2pmsg.DecryptionKey{
		InstanceID: 0,
		Eon:        eon,
		EpochID:    epochID.Bytes(),
		Key:        encodedDecryptionKey,
	})
	assert.NilError(t, err)
	assert.Check(t, len(msgs) == 0)
	key, err := db.GetDecryptionKey(ctx, kprdb.GetDecryptionKeyParams{
		Eon:     int64(eon),
		EpochID: epochID.Bytes(),
	})
	assert.NilError(t, err)
	assert.Check(t, bytes.Equal(key.DecryptionKey, encodedDecryptionKey))
}

func TestDecryptionKeyValidatorIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	db, dbpool, closedb := testdb.NewKeyperTestDB(ctx, t)
	defer closedb()

	keyperIndex := uint64(1)
	eon := uint64(0)
	epochID, _ := epochid.BigToEpochID(common.Big0)
	wrongEpochID, _ := epochid.BigToEpochID(common.Big1)
	tkg := initializeEon(ctx, t, db, keyperIndex)
	secretKey := tkg.EpochSecretKey(epochID).Marshal()

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
				EpochID:    epochID.Bytes(),
				Key:        secretKey,
			},
		},
		{
			name:  "invalid decryption key wrong epoch",
			valid: false,
			msg: &p2pmsg.DecryptionKey{
				InstanceID: config.GetInstanceID(),
				Eon:        eon,
				EpochID:    wrongEpochID.Bytes(),
				Key:        secretKey,
			},
		},
		{
			name:  "invalid decryption key wrong instance ID",
			valid: false,
			msg: &p2pmsg.DecryptionKey{
				InstanceID: config.GetInstanceID() + 1,
				Eon:        eon,
				EpochID:    epochID.Bytes(),
				Key:        secretKey,
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			validationResult, err := handler.ValidateMessage(ctx, tc.msg)
			if tc.valid {
				assert.NilError(t, err)
			}
			assert.Equal(t, validationResult, tc.valid,
				"validate failed valid=%t msg=%+v type=%T", tc.valid, tc.msg, tc.msg)
		})
	}
}
