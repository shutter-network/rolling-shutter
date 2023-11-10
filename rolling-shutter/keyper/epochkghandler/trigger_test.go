package epochkghandler

import (
	"context"
	"crypto/ecdsa"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/google/go-cmp/cmp/cmpopts"
	"gotest.tools/assert"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/db/chainobsdb"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/db/kprdb"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/identitypreimage"
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

	identityPreimage := identitypreimage.Uint64ToIdentityPreimage(50)
	keyperIndex := uint64(1)

	initializeEon(ctx, t, dbpool, keyperIndex)
	var handler p2p.MessageHandler = &DecryptionTriggerHandler{config: config, dbpool: dbpool}
	// send decryption key share when first trigger is received
	trigger, err := p2pmsg.NewSignedDecryptionTrigger(
		config.GetInstanceID(),
		identityPreimage,
		0,
		make([]byte, 32),
		config.GetCollatorKey(),
	)
	assert.NilError(t, err)
	msgs := p2ptest.MustHandleMessage(t, handler, ctx, trigger)
	share, err := db.GetDecryptionKeyShare(ctx, kprdb.GetDecryptionKeyShareParams{
		Eon:         int64(config.GetEon()),
		EpochID:     identityPreimage.Bytes(),
		KeyperIndex: int64(keyperIndex),
	})
	assert.NilError(t, err)
	assert.Check(t, len(msgs) == 1)
	msg, ok := msgs[0].(*p2pmsg.DecryptionKeyShares)
	assert.Check(t, ok)
	assert.Check(t, msg.InstanceID == config.GetInstanceID())
	assert.Check(t, msg.KeyperIndex == keyperIndex)
	assert.DeepEqual(t, msg.GetShares(),
		[]*p2pmsg.KeyShare{
			{
				EpochID: identityPreimage.Bytes(),
				Share:   share.DecryptionKeyShare,
			},
		},
		cmpopts.IgnoreUnexported(p2pmsg.KeyShare{}),
	)

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
	identityPreimage1, _ := identitypreimage.BigToIdentityPreimage(common.Big0)
	activationBlk2 := uint64(123)
	identityPreimage2, _ := identitypreimage.BigToIdentityPreimage(common.Big1)
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
		name             string
		valid            bool
		instanceID       uint64
		identityPreimage identitypreimage.IdentityPreimage
		blockNumber      uint64
		privKey          *ecdsa.PrivateKey
	}{
		{
			name:             "valid trigger collator 1",
			valid:            true,
			instanceID:       config.GetInstanceID(),
			identityPreimage: identityPreimage1,
			blockNumber:      activationBlk1,
			privKey:          collatorKey1,
		},
		{
			name:             "valid trigger collator 2",
			valid:            true,
			instanceID:       config.GetInstanceID(),
			identityPreimage: identityPreimage2,
			blockNumber:      activationBlk2,
			privKey:          collatorKey2,
		},
		{
			name:             "invalid trigger wrong collator 1",
			valid:            false,
			instanceID:       config.GetInstanceID(),
			identityPreimage: identityPreimage2,
			blockNumber:      activationBlk2,
			privKey:          collatorKey1,
		},
		{
			name:             "invalid trigger wrong collator 2",
			valid:            false,
			instanceID:       config.GetInstanceID(),
			identityPreimage: identityPreimage1,
			blockNumber:      activationBlk1,
			privKey:          collatorKey2,
		},
		{
			name:             "invalid trigger wrong instanceID",
			valid:            false,
			instanceID:       config.GetInstanceID() + 1,
			identityPreimage: identityPreimage1,
			blockNumber:      activationBlk1,
			privKey:          collatorKey1,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			msg, err := p2pmsg.NewSignedDecryptionTrigger(
				tc.instanceID,
				tc.identityPreimage,
				tc.blockNumber,
				[]byte{},
				tc.privKey,
			)
			assert.NilError(t, err)
			p2ptest.MustValidateMessageResult(t, tc.valid, handler, ctx, msg)
		})
	}
}
