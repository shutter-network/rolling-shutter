package epochkghandler

import (
	"context"
	"crypto/ecdsa"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"gotest.tools/assert"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/db/chainobsdb"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/epochid"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/testdb"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2pmsg"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/shdb"
)

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
	kpr := New(config.Address, config.InstanceID, dbpool)

	tests := []struct {
		name  string
		valid bool
		msg   *p2pmsg.DecryptionKeyShare
	}{
		{
			name:  "valid decryption key share",
			valid: true,
			msg: &p2pmsg.DecryptionKeyShare{
				InstanceID:  config.InstanceID,
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
				InstanceID:  config.InstanceID,
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
				InstanceID:  config.InstanceID + 1,
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
				InstanceID:  config.InstanceID,
				Eon:         eon,
				EpochID:     epochID.Bytes(),
				KeyperIndex: keyperIndex + 1,
				Share:       keyshare,
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			validationResult, err := kpr.validateDecryptionKeyShare(ctx, tc.msg)
			if tc.valid {
				assert.NilError(t, err)
			}
			assert.Equal(t, validationResult, tc.valid,
				"validate failed valid=%t msg=%+v type=%T", tc.valid, tc.msg, tc.msg)
		})
	}
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

	kpr := New(config.Address, config.InstanceID, dbpool)

	tests := []struct {
		name  string
		valid bool
		msg   *p2pmsg.DecryptionKey
	}{
		{
			name:  "valid decryption key",
			valid: true,
			msg: &p2pmsg.DecryptionKey{
				InstanceID: config.InstanceID,
				Eon:        eon,
				EpochID:    epochID.Bytes(),
				Key:        secretKey,
			},
		},
		{
			name:  "invalid decryption key wrong epoch",
			valid: false,
			msg: &p2pmsg.DecryptionKey{
				InstanceID: config.InstanceID,
				Eon:        eon,
				EpochID:    wrongEpochID.Bytes(),
				Key:        secretKey,
			},
		},
		{
			name:  "invalid decryption key wrong instance ID",
			valid: false,
			msg: &p2pmsg.DecryptionKey{
				InstanceID: config.InstanceID + 1,
				Eon:        eon,
				EpochID:    epochID.Bytes(),
				Key:        secretKey,
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			validationResult, err := kpr.validateDecryptionKey(ctx, tc.msg)
			if tc.valid {
				assert.NilError(t, err)
			}
			assert.Equal(t, validationResult, tc.valid,
				"validate failed valid=%t msg=%+v type=%T", tc.valid, tc.msg, tc.msg)
		})
	}
}

func TestTriggerValidatorIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	_, dbpool, closedb := testdb.NewKeyperTestDB(ctx, t)
	defer closedb()

	kpr := New(config.Address, config.InstanceID, dbpool)

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
			instanceID:  config.InstanceID,
			epochID:     epochID1,
			blockNumber: activationBlk1,
			privKey:     collatorKey1,
		},
		{
			name:        "valid trigger collator 2",
			valid:       true,
			instanceID:  config.InstanceID,
			epochID:     epochID2,
			blockNumber: activationBlk2,
			privKey:     collatorKey2,
		},
		{
			name:        "invalid trigger wrong collator 1",
			valid:       false,
			instanceID:  config.InstanceID,
			epochID:     epochID2,
			blockNumber: activationBlk2,
			privKey:     collatorKey1,
		},
		{
			name:        "invalid trigger wrong collator 2",
			valid:       false,
			instanceID:  config.InstanceID,
			epochID:     epochID1,
			blockNumber: activationBlk1,
			privKey:     collatorKey2,
		},
		{
			name:        "invalid trigger wrong instanceID",
			valid:       false,
			instanceID:  config.InstanceID + 1,
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
			validationResult, err := kpr.validateDecryptionTrigger(ctx, msg)
			if tc.valid {
				assert.NilError(t, err)
			}
			assert.Equal(t, validationResult, tc.valid,
				"validate failed valid=%t", tc.valid)
		})
	}
}
