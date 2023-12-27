package rollup

import (
	"context"
	"crypto/ecdsa"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"gotest.tools/assert"

	chainobsdb "github.com/shutter-network/rolling-shutter/rolling-shutter/chainobserver/db/collator"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/epochkghandler"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyperimpl/rollup/database"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/broker"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/identitypreimage"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/service"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/testsetup"
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
	dbpool, dbclose := testsetup.NewTestDBPool(ctx, t, database.Definition)
	t.Cleanup(dbclose)

	identityPreimage := identitypreimage.Uint64ToIdentityPreimage(50)
	keyperIndex := uint64(1)

	cfg := NewConfig()
	err := cfg.SetExampleValues()
	assert.NilError(t, err)
	cfg.InstanceID = config.GetInstanceID()

	triggerHandler := NewDecryptionTriggerHandler(ctx, *cfg, dbpool)

	sender, err := p2ptest.NewTestMessaging()
	assert.NilError(t, err)
	sender.AddMessageHandler(triggerHandler)

	group := service.RunBackground(
		ctx,
		sender,
	)

	testsetup.InitializeEon(ctx, t, dbpool, config, keyperIndex)

	// send decryption key share when first trigger is received
	trigger, err := p2pmsg.NewSignedDecryptionTrigger(
		config.GetInstanceID(),
		identityPreimage,
		42,
		make([]byte, 32),
		config.GetCollatorKey(),
	)
	assert.NilError(t, err)

	sendTimeout := 10 * time.Millisecond

	err = sender.PushMessage(ctx, trigger, sendTimeout)
	assert.NilError(t, err)

	ev, ok := <-triggerHandler.C
	assert.Check(t, ok)
	assert.DeepEqual(t, ev.Value.IdentityPreimages, []identitypreimage.IdentityPreimage{identityPreimage})
	assert.DeepEqual(t, ev.Value.BlockNumber, uint64(42))

	// send the same message again
	err = sender.PushMessage(ctx, trigger, sendTimeout)
	assert.NilError(t, err)

	// close the channel
	sender.StopReceive()

	// assert the second (same) trigger does cause another
	// call in the DecryptionTriggerHandler.
	// The keyper core reading the trigger events can then decide
	// whether duplicate triggers should be acted upon
	ev, ok = <-triggerHandler.C
	assert.Check(t, ok)
	assert.DeepEqual(t, ev.Value.IdentityPreimages, []identitypreimage.IdentityPreimage{identityPreimage})
	assert.DeepEqual(t, ev.Value.BlockNumber, uint64(42))

	err = group.Wait()
	assert.NilError(t, err)
}

func TestTriggerValidatorIntegration(t *testing.T) { //nolint:funlen
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	dbpool, dbclose := testsetup.NewTestDBPool(ctx, t, database.Definition)
	t.Cleanup(dbclose)

	decrTrigChan := make(chan *broker.Event[*epochkghandler.DecryptionTrigger])

	cfg := NewConfig()
	err := cfg.SetExampleValues()
	assert.NilError(t, err)
	cfg.InstanceID = config.GetInstanceID()

	var handler p2p.MessageHandler = &DecryptionTriggerHandler{trigger: decrTrigChan, config: *cfg, dbpool: dbpool}
	collatorKey1, err := ethcrypto.GenerateKey()
	assert.NilError(t, err)
	collatorAddress1 := ethcrypto.PubkeyToAddress(collatorKey1.PublicKey)

	collatorKey2, err := ethcrypto.GenerateKey()
	assert.NilError(t, err)
	collatorAddress2 := ethcrypto.PubkeyToAddress(collatorKey2.PublicKey)

	// Make a db with collator 1 from a certain block and collator 2 afterwards
	activationBlk1 := uint64(0)
	idPreimage1 := identitypreimage.BigToIdentityPreimage(common.Big0)
	activationBlk2 := uint64(123)
	idPreimage2 := identitypreimage.BigToIdentityPreimage(common.Big1)
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
			identityPreimage: idPreimage1,
			blockNumber:      activationBlk1,
			privKey:          collatorKey1,
		},
		{
			name:             "valid trigger collator 2",
			valid:            true,
			instanceID:       config.GetInstanceID(),
			identityPreimage: idPreimage2,
			blockNumber:      activationBlk2,
			privKey:          collatorKey2,
		},
		{
			name:             "invalid trigger wrong collator 1",
			valid:            false,
			instanceID:       config.GetInstanceID(),
			identityPreimage: idPreimage2,
			blockNumber:      activationBlk2,
			privKey:          collatorKey1,
		},
		{
			name:             "invalid trigger wrong collator 2",
			valid:            false,
			instanceID:       config.GetInstanceID(),
			identityPreimage: idPreimage1,
			blockNumber:      activationBlk1,
			privKey:          collatorKey2,
		},
		{
			name:             "invalid trigger wrong instanceID",
			valid:            false,
			instanceID:       config.GetInstanceID() + 1,
			identityPreimage: idPreimage1,
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
