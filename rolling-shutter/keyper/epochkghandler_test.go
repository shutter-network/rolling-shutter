package keyper

import (
	"bytes"
	"context"
	"database/sql"
	"testing"

	"gotest.tools/assert"

	"github.com/shutter-network/shutter/shlib/puredkg"
	"github.com/shutter-network/shutter/shlib/shcrypto"
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

func initializeEon(ctx context.Context, t *testing.T, db *kprdb.Queries, config Config, keyperIndex uint64) *medley.TestKeyGenerator {
	t.Helper()

	eon := uint64(0)
	keypers := []string{
		"0x0000000000000000000000000000000000000000",
		config.Address().Hex(),
		"0x1111111111111111111111111111111111111111",
	}

	tkg := medley.NewTestKeyGenerator(t, 3, 2)
	publicKeyShares := []*shcrypto.EonPublicKeyShare{}
	for i := uint64(0); i < tkg.NumKeypers; i++ {
		share := tkg.EonPublicKeyShare(0, i)
		publicKeyShares = append(publicKeyShares, share)
	}
	dkgResult := puredkg.Result{
		Eon:             eon,
		NumKeypers:      tkg.NumKeypers,
		Threshold:       tkg.Threshold,
		Keyper:          keyperIndex,
		SecretKeyShare:  tkg.EonSecretKeyShare(0, keyperIndex),
		PublicKey:       tkg.EonPublicKey(0),
		PublicKeyShares: publicKeyShares,
	}
	dkgResultEncoded, err := shdb.EncodePureDKGResult(&dkgResult)
	assert.NilError(t, err)

	err = db.InsertBatchConfig(ctx, kprdb.InsertBatchConfigParams{
		ConfigIndex: 1,
		Height:      0,
		Keypers:     keypers,
		Threshold:   int32(tkg.Threshold),
	})
	assert.NilError(t, err)
	err = db.InsertEon(ctx, kprdb.InsertEonParams{
		Eon:                   0,
		Height:                0,
		ActivationBlockNumber: 0,
		ConfigIndex:           1,
	})
	assert.NilError(t, err)
	err = db.InsertDKGResult(ctx, kprdb.InsertDKGResultParams{
		Eon:        0,
		Success:    true,
		Error:      sql.NullString{},
		PureResult: dkgResultEncoded,
	})
	assert.NilError(t, err)

	return tkg
}

func TestHandleDecryptionTriggerIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	db, closedb := medley.NewKeyperTestDB(ctx, t)
	defer closedb()

	epochID := uint64(50)
	keyperIndex := uint64(1)

	config := newTestConfig(t)
	initializeEon(ctx, t, db, config, keyperIndex)
	handler := epochKGHandler{
		config: config,
		db:     db,
	}

	// send decryption key share when first trigger is received
	trigger := &decryptionTrigger{
		EpochID:    epochID,
		InstanceID: 0,
	}
	msgs, err := handler.handleDecryptionTrigger(ctx, trigger)
	assert.NilError(t, err)
	share, err := db.GetDecryptionKeyShare(ctx, kprdb.GetDecryptionKeyShareParams{
		EpochID:     shdb.EncodeUint64(epochID),
		KeyperIndex: int64(keyperIndex),
	})
	assert.NilError(t, err)
	assert.Check(t, len(msgs) == 1)
	msg, ok := msgs[0].(*shmsg.DecryptionKeyShare)
	assert.Check(t, ok)
	assert.Check(t, msg.InstanceID == 0)
	assert.Check(t, msg.EpochID == epochID)
	assert.Check(t, msg.KeyperIndex == keyperIndex)
	assert.Check(t, bytes.Equal(msg.Share, share.DecryptionKeyShare))

	// don't send share when trigger is received again
	msgs, err = handler.handleDecryptionTrigger(ctx, trigger)
	assert.NilError(t, err)
	assert.Check(t, len(msgs) == 0)
}

func TestHandleDecryptionKeyShareIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	ctx := context.Background()
	db, closedb := medley.NewKeyperTestDB(ctx, t)
	defer closedb()

	epochID := uint64(50)
	keyperIndex := uint64(1)

	config := newTestConfig(t)
	tkg := initializeEon(ctx, t, db, config, keyperIndex)
	handler := epochKGHandler{
		config: config,
		db:     db,
	}
	encodedDecryptionKey := tkg.EpochSecretKey(epochID).Marshal()

	// threshold is two, so no outgoing message after first input
	msgs, err := handler.handleDecryptionKeyShare(ctx, &decryptionKeyShare{
		instanceID:  0,
		epochID:     epochID,
		keyperIndex: 0,
		share:       tkg.EpochSecretKeyShare(epochID, 0),
	})
	assert.NilError(t, err)
	assert.Check(t, len(msgs) == 0)

	// second message pushes us over the threshold (note that we didn't send a trigger, so the
	// share of the handler itself doesn't count)
	msgs, err = handler.handleDecryptionKeyShare(ctx, &decryptionKeyShare{
		instanceID:  0,
		epochID:     epochID,
		keyperIndex: 2,
		share:       tkg.EpochSecretKeyShare(epochID, 2),
	})
	assert.NilError(t, err)
	assert.Check(t, len(msgs) == 1)
	msg, ok := msgs[0].(*shmsg.DecryptionKey)
	assert.Check(t, ok)
	assert.Check(t, msg.InstanceID == 0)
	assert.Check(t, msg.EpochID == epochID)
	assert.Check(t, bytes.Equal(msg.Key, encodedDecryptionKey))
}

func TestHandleDecryptionKeyIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	ctx := context.Background()
	db, closedb := medley.NewKeyperTestDB(ctx, t)
	defer closedb()

	epochID := uint64(50)
	keyperIndex := uint64(1)

	config := newTestConfig(t)
	tkg := initializeEon(ctx, t, db, config, keyperIndex)
	handler := epochKGHandler{
		config: config,
		db:     db,
	}
	encodedDecryptionKey := tkg.EpochSecretKey(epochID).Marshal()

	// send a decryption key and check that it gets inserted
	msgs, err := handler.handleDecryptionKey(ctx, &decryptionKey{
		instanceID: 0,
		epochID:    epochID,
		key:        tkg.EpochSecretKey(epochID),
	})
	assert.NilError(t, err)
	assert.Check(t, len(msgs) == 0)
	key, err := db.GetDecryptionKey(ctx, shdb.EncodeUint64(epochID))
	assert.NilError(t, err)
	assert.Check(t, bytes.Equal(key.DecryptionKey, encodedDecryptionKey))
}
