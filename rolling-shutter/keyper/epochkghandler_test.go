package keyper

import (
	"bytes"
	"context"
	"database/sql"
	"testing"

	"gotest.tools/assert"

	"github.com/shutter-network/shutter/shlib/puredkg"
	"github.com/shutter-network/shutter/shuttermint/keyper/kprdb"
	"github.com/shutter-network/shutter/shuttermint/medley"
	"github.com/shutter-network/shutter/shuttermint/shdb"
	"github.com/shutter-network/shutter/shuttermint/shmsg"
)

func newTestConfig(t *testing.T) Config {
	t.Helper()

	c := Config{
		InstanceID: 0,
	}
	err := c.GenerateNewKeys()
	assert.NilError(t, err)
	return c
}

func initializeEon(ctx context.Context, t *testing.T, db *kprdb.Queries, config Config) {
	eon := uint64(0)
	keypers := []string{
		"0x0000000000000000000000000000000000000000",
		config.Address().Hex(),
		"0x1111111111111111111111111111111111111111",
	}
	threshold := uint64(1)
	keyperIndex := uint64(1)

	tkg := medley.NewTestKeyGenerator(t, 3, 2)
	dkgResult := puredkg.Result{
		Eon:            eon,
		NumKeypers:     uint64(len(keypers)),
		Threshold:      threshold,
		Keyper:         keyperIndex,
		SecretKeyShare: tkg.EonSecretKeyShare(0, 0),
	}
	dkgResultEncoded, err := shdb.EncodePureDKGResult(&dkgResult)
	assert.NilError(t, err)

	err = db.InsertBatchConfig(ctx, kprdb.InsertBatchConfigParams{
		ConfigIndex: 1,
		Height:      0,
		Keypers:     keypers,
		Threshold:   2,
	})
	assert.NilError(t, err)
	err = db.InsertEon(ctx, kprdb.InsertEonParams{
		Eon:         0,
		Height:      0,
		BatchIndex:  shdb.EncodeUint64(0),
		ConfigIndex: 1,
	})
	assert.NilError(t, err)
	err = db.InsertDKGResult(ctx, kprdb.InsertDKGResultParams{
		Eon:        0,
		Success:    true,
		Error:      sql.NullString{},
		PureResult: dkgResultEncoded,
	})
	assert.NilError(t, err)
}

func TestHandleDecryptionTriggerIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	db, closedb := medley.NewKeyperTestDB(ctx, t)
	defer closedb()

	config := newTestConfig(t)
	initializeEon(ctx, t, db, config)
	epochKGHandler := epochKGHandler{
		config: config,
		db:     db,
	}

	epochID := uint64(100)
	keyperIndex := uint64(1)

	// send decryption key share when first trigger is received
	trigger := &decryptionTrigger{
		EpochID:    epochID,
		InstanceID: 0,
	}
	msgs, err := epochKGHandler.handleDecryptionTrigger(ctx, trigger)
	assert.NilError(t, err)
	share, err := db.GetDecryptionKeyShare(ctx, kprdb.GetDecryptionKeyShareParams{
		EpochID:     shdb.EncodeUint64(epochID),
		KeyperIndex: int64(keyperIndex),
	})
	assert.Check(t, len(msgs) == 1)
	msg, ok := msgs[0].(*shmsg.DecryptionKeyShare)
	assert.Check(t, ok)
	assert.Check(t, msg.InstanceID == 0)
	assert.Check(t, msg.EpochID == epochID)
	assert.Check(t, msg.KeyperIndex == keyperIndex)
	assert.Check(t, bytes.Equal(msg.Share, share.DecryptionKeyShare))

	// don't send share when trigger is received again
	msgs, err = epochKGHandler.handleDecryptionTrigger(ctx, trigger)
	assert.NilError(t, err)
	assert.Check(t, len(msgs) == 0)
}
