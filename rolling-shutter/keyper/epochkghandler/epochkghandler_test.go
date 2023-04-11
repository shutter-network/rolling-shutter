package epochkghandler

import (
	"bytes"
	"context"
	"database/sql"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"gotest.tools/assert"

	"github.com/shutter-network/shutter/shlib/puredkg"
	"github.com/shutter-network/shutter/shlib/shcrypto"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/db/kprdb"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/epochid"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/testdb"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/testkeygen"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2pmsg"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/shdb"
)

type TestConfig struct{}

var config = &TestConfig{}

func (c *TestConfig) GetAddress() common.Address {
	return common.HexToAddress("0x2222222222222222222222222222222222222222")
}

func (c *TestConfig) GetInstanceID() uint64 {
	return 55
}

func initializeEon(
	ctx context.Context,
	t *testing.T,
	db *kprdb.Queries,
	keyperIndex uint64, //nolint:unparam
) *testkeygen.TestKeyGenerator {
	t.Helper()
	eon := uint64(0)
	keypers := []string{
		"0x0000000000000000000000000000000000000000",
		config.GetAddress().Hex(),
		"0x1111111111111111111111111111111111111111",
	}

	tkg := testkeygen.NewTestKeyGenerator(t, 3, 2)
	publicKeyShares := []*shcrypto.EonPublicKeyShare{}
	epochID, _ := epochid.BigToEpochID(common.Big0)
	for i := uint64(0); i < tkg.NumKeypers; i++ {
		share := tkg.EonPublicKeyShare(epochID, i)
		publicKeyShares = append(publicKeyShares, share)
	}
	dkgResult := puredkg.Result{
		Eon:             eon,
		NumKeypers:      tkg.NumKeypers,
		Threshold:       tkg.Threshold,
		Keyper:          keyperIndex,
		SecretKeyShare:  tkg.EonSecretKeyShare(epochID, keyperIndex),
		PublicKey:       tkg.EonPublicKey(epochID),
		PublicKeyShares: publicKeyShares,
	}
	dkgResultEncoded, err := shdb.EncodePureDKGResult(&dkgResult)
	assert.NilError(t, err)

	err = db.InsertBatchConfig(ctx, kprdb.InsertBatchConfigParams{
		KeyperConfigIndex: 1,
		Height:            0,
		Keypers:           keypers,
		Threshold:         int32(tkg.Threshold),
	})
	assert.NilError(t, err)
	err = db.InsertEon(ctx, kprdb.InsertEonParams{
		Eon:                   0,
		Height:                0,
		ActivationBlockNumber: 0,
		KeyperConfigIndex:     1,
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
	db, dbpool, closedb := testdb.NewKeyperTestDB(ctx, t)
	defer closedb()

	epochID := epochid.Uint64ToEpochID(50)
	keyperIndex := uint64(1)

	initializeEon(ctx, t, db, keyperIndex)
	handler := New(config, dbpool)

	// send decryption key share when first trigger is received
	trigger := &p2pmsg.DecryptionTrigger{
		EpochID:    epochID.Bytes(),
		InstanceID: 0,
	}
	msgs, err := handler.handleDecryptionTrigger(ctx, trigger)
	assert.NilError(t, err)
	share, err := db.GetDecryptionKeyShare(ctx, kprdb.GetDecryptionKeyShareParams{
		EpochID:     epochID.Bytes(),
		KeyperIndex: int64(keyperIndex),
	})
	assert.NilError(t, err)
	assert.Check(t, len(msgs) == 1)
	msg, ok := msgs[0].(*p2pmsg.DecryptionKeyShare)
	assert.Check(t, ok)
	assert.Check(t, msg.InstanceID == config.GetInstanceID())
	assert.Check(t, bytes.Equal(msg.EpochID, epochID.Bytes()))
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
	db, dbpool, closedb := testdb.NewKeyperTestDB(ctx, t)
	defer closedb()

	epochID := epochid.Uint64ToEpochID(50)
	keyperIndex := uint64(1)

	tkg := initializeEon(ctx, t, db, keyperIndex)
	handler := New(config, dbpool)
	encodedDecryptionKey := tkg.EpochSecretKey(epochID).Marshal()

	// threshold is two, so no outgoing message after first input
	msgs, err := handler.handleDecryptionKeyShare(ctx, &p2pmsg.DecryptionKeyShare{
		InstanceID:  0,
		EpochID:     epochID.Bytes(),
		KeyperIndex: 0,
		Share:       tkg.EpochSecretKeyShare(epochID, 0).Marshal(),
	})
	assert.NilError(t, err)
	assert.Check(t, len(msgs) == 0)

	// second message pushes us over the threshold (note that we didn't send a trigger, so the
	// share of the handler itself doesn't count)
	msgs, err = handler.handleDecryptionKeyShare(ctx, &p2pmsg.DecryptionKeyShare{
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
	handler := New(config, dbpool)
	encodedDecryptionKey := tkg.EpochSecretKey(epochID).Marshal()

	// send a decryption key and check that it gets inserted
	msgs, err := handler.handleDecryptionKey(ctx, &p2pmsg.DecryptionKey{
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
