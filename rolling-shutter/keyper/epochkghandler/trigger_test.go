package epochkghandler

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"gotest.tools/assert"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/db/chainobsdb"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/db/kprdb"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/epochid"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/testdb"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2p"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2p/p2ptest"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2pmsg"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/shdb"
)

func TestHandleDecryptionTriggerIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	db, dbpool, closedb := testdb.NewKeyperTestDB(ctx, t)
	defer closedb()

	epochID := epochid.Uint64ToEpochID(50)
	keyperIndex := uint64(1)

	initializeEon(ctx, t, dbpool, keyperIndex)
	var handler p2p.MessageHandler = &DecryptionTriggerHandler{config: config, dbpool: dbpool}
	// send decryption key share when first trigger is received
	trigger, err := p2pmsg.NewSignedDecryptionTrigger(
		config.GetInstanceID(),
		epochID,
		0,
		make([]byte, 32),
		config.GetCollatorKey(),
	)
	assert.NilError(t, err)
	msgs := p2ptest.MustHandleMessage(t, handler, ctx, trigger)
	share, err := db.GetDecryptionKeyShare(ctx, kprdb.GetDecryptionKeyShareParams{
		Eon:         int64(config.GetEon()),
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
	msgs = p2ptest.MustHandleMessage(t, handler, ctx, trigger)
	assert.NilError(t, err)
	assert.Check(t, len(msgs) == 0)
}

func TestTriggerValidatorIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	_, dbpool, closedb := testdb.NewKeyperTestDB(ctx, t)
	defer closedb()

	var handler p2p.MessageHandler = &DecryptionTriggerHandler{config: config, dbpool: dbpool}
	collatorKey1, err := ethcrypto.GenerateKey()
	assert.NilError(t, err)
	collatorAddress1 := ethcrypto.PubkeyToAddress(collatorKey1.PublicKey)

	collatorKey2, err := ethcrypto.GenerateKey()
	assert.NilError(t, err)
	collatorAddress2 := ethcrypto.PubkeyToAddress(collatorKey2.PublicKey)

	// Make a db with collator 1 from a certain block and collator 2 afterwards
	activationBlk1 := uint64(0)
	epochID1, _ := epochid.BigToEpochID(common.Big0)
	activationBlk2 := uint64(123)
	epochID2, _ := epochid.BigToEpochID(common.Big1)
	assert.NilError(t, err)
	collator1 := shdb.EncodeAddress(collatorAddress1)
	collator2 := shdb.EncodeAddress(collatorAddress2)
	err = chainobsdb.New(dbpool).InsertChainCollator(ctx, chainobsdb.InsertChainCollatorParams{
		ActivationBlockNumber: int64(activationBlk1),
		Collator:              collator1,
	})
	assert.NilError(t, err)
	err = chainobsdb.New(dbpool).InsertChainCollator(ctx, chainobsdb.InsertChainCollatorParams{
		ActivationBlockNumber: int64(activationBlk2),
		Collator:              collator2,
	})
	assert.NilError(t, err)

	tests := []struct {
		name        string
		valid       bool
		instanceID  uint64
		epochID     epochid.EpochID
		blockNumber uint64
		privKey     *ecdsa.PrivateKey
	}{
		{
			name:        "valid trigger collator 1",
			valid:       true,
			instanceID:  config.GetInstanceID(),
			epochID:     epochID1,
			blockNumber: activationBlk1,
			privKey:     collatorKey1,
		},
		{
			name:        "valid trigger collator 2",
			valid:       true,
			instanceID:  config.GetInstanceID(),
			epochID:     epochID2,
			blockNumber: activationBlk2,
			privKey:     collatorKey2,
		},
		{
			name:        "invalid trigger wrong collator 1",
			valid:       false,
			instanceID:  config.GetInstanceID(),
			epochID:     epochID2,
			blockNumber: activationBlk2,
			privKey:     collatorKey1,
		},
		{
			name:        "invalid trigger wrong collator 2",
			valid:       false,
			instanceID:  config.GetInstanceID(),
			epochID:     epochID1,
			blockNumber: activationBlk1,
			privKey:     collatorKey2,
		},
		{
			name:        "invalid trigger wrong instanceID",
			valid:       false,
			instanceID:  config.GetInstanceID() + 1,
			epochID:     epochID1,
			blockNumber: activationBlk1,
			privKey:     collatorKey1,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			msg, err := p2pmsg.NewSignedDecryptionTrigger(
				tc.instanceID,
				tc.epochID,
				tc.blockNumber,
				[]byte{},
				tc.privKey,
			)
			assert.NilError(t, err)
			p2ptest.MustValidateMessageResult(t, tc.valid, handler, ctx, msg)
		})
	}
}
