package epochkghandler

import (
	"bytes"
	"context"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"gotest.tools/assert"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/epochid"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/testdb"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2p"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2pmsg"
)

func TestHandleDecryptionKeyShareIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	ctx := context.Background()
	db, dbpool, closedb := testdb.NewKeyperTestDB(ctx, t)
	defer closedb()

	epochID := epochid.Uint64ToEpochID(50)
	keyperIndex := uint64(1)

	tkg := initializeEon(ctx, t, db, keyperIndex)
	var handler p2p.MessageHandler = &DecryptionKeyShareHandler{config: config, dbpool: dbpool}
	encodedDecryptionKey := tkg.EpochSecretKey(epochID).Marshal()

	// threshold is two, so no outgoing message after first input
	msgs, err := handler.HandleMessage(ctx, &p2pmsg.DecryptionKeyShare{
		InstanceID:  0,
		EpochID:     epochID.Bytes(),
		KeyperIndex: 0,
		Share:       tkg.EpochSecretKeyShare(epochID, 0).Marshal(),
	})
	assert.NilError(t, err)
	assert.Check(t, len(msgs) == 0)

	// second message pushes us over the threshold (note that we didn't send a trigger, so the
	// share of the handler itself doesn't count)
	msgs, err = handler.HandleMessage(ctx, &p2pmsg.DecryptionKeyShare{
		InstanceID:  0,
		EpochID:     epochID.Bytes(),
		KeyperIndex: 2,
		Share:       tkg.EpochSecretKeyShare(epochID, 2).Marshal(),
	})
	assert.NilError(t, err)
	assert.Check(t, len(msgs) == 1)
	msg, ok := msgs[0].(*p2pmsg.DecryptionKey)
	assert.Check(t, ok)
	assert.Check(t, msg.InstanceID == config.GetInstanceID())
	assert.Check(t, bytes.Equal(msg.EpochID, epochID.Bytes()))
	assert.Check(t, bytes.Equal(msg.Key, encodedDecryptionKey))
}

func TestDecryptionKeyshareValidatorIntegration(t *testing.T) {
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
	keyshare := tkg.EpochSecretKeyShare(epochID, keyperIndex).Marshal()
	var kpr p2p.MessageHandler = &DecryptionKeyShareHandler{config: config, dbpool: dbpool}

	tests := []struct {
		name  string
		valid bool
		msg   *p2pmsg.DecryptionKeyShare
	}{
		{
			name:  "valid decryption key share",
			valid: true,
			msg: &p2pmsg.DecryptionKeyShare{
				InstanceID:  config.GetInstanceID(),
				Eon:         eon,
				EpochID:     epochID.Bytes(),
				KeyperIndex: keyperIndex,
				Share:       keyshare,
			},
		},
		{
			name:  "invalid decryption key share wrong epoch",
			valid: false,
			msg: &p2pmsg.DecryptionKeyShare{
				InstanceID:  config.GetInstanceID(),
				Eon:         eon,
				EpochID:     wrongEpochID.Bytes(),
				KeyperIndex: keyperIndex,
				Share:       keyshare,
			},
		},
		{
			name:  "invalid decryption key share wrong instance ID",
			valid: false,
			msg: &p2pmsg.DecryptionKeyShare{
				InstanceID:  config.GetInstanceID() + 1,
				Eon:         eon,
				EpochID:     epochID.Bytes(),
				KeyperIndex: keyperIndex,
				Share:       keyshare,
			},
		},
		{
			name:  "invalid decryption key share wrong keyper index",
			valid: false,
			msg: &p2pmsg.DecryptionKeyShare{
				InstanceID:  config.GetInstanceID(),
				Eon:         eon,
				EpochID:     epochID.Bytes(),
				KeyperIndex: keyperIndex + 1,
				Share:       keyshare,
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			validationResult, err := kpr.ValidateMessage(ctx, tc.msg)
			if tc.valid {
				assert.NilError(t, err)
			}
			assert.Equal(t, validationResult, tc.valid,
				"validate failed valid=%t msg=%+v type=%T", tc.valid, tc.msg, tc.msg)
		})
	}
}
