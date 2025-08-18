package shutterservice

import (
	"context"
	"math"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/shutter-network/contracts/v2/bindings/shutterregistry"
	"gotest.tools/assert"

	corekeyperdatabase "github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/database"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/epochkghandler"
	servicedatabase "github.com/shutter-network/rolling-shutter/rolling-shutter/keyperimpl/shutterservice/database"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/broker"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/chainsync/event"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/configuration"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/encodeable/keys"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/encodeable/number"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/identitypreimage"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/testsetup"
)

func TestProcessBlockSuccess(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	ctx := context.Background()

	dbpool, dbclose := testsetup.NewTestDBPool(ctx, t, servicedatabase.Definition)
	t.Cleanup(dbclose)

	serviceDB := servicedatabase.New(dbpool)
	coreKeyperDB := corekeyperdatabase.New(dbpool)

	privateKey, sender, _ := generateRandomAccount()

	decryptionTriggerChannel := make(chan *broker.Event[*epochkghandler.DecryptionTrigger])

	kpr := &Keyper{
		dbpool: dbpool,
		config: &Config{
			Chain: &ChainConfig{
				Node: &configuration.EthnodeConfig{
					PrivateKey: &keys.ECDSAPrivate{
						Key: privateKey,
					},
				},
			},
		},
		decryptionTriggerChannel: decryptionTriggerChannel,
	}

	blockHash, _ := generateRandom32Bytes()
	blockTimestamp := time.Now().Add(5 * time.Second).Unix()
	blockNumber := 102
	activationBlockNumber := 100

	identityPrefix, _ := generateRandom32Bytes()
	identity := computeIdentity(&shutterregistry.ShutterregistryIdentityRegistered{
		IdentityPrefix: [32]byte(identityPrefix),
		Sender:         sender,
	})

	err := coreKeyperDB.InsertEon(ctx, corekeyperdatabase.InsertEonParams{
		Eon:                   int64(config.GetEon()),
		Height:                0,
		ActivationBlockNumber: int64(activationBlockNumber),
		KeyperConfigIndex:     0,
	})
	assert.NilError(t, err)

	_, err = serviceDB.InsertIdentityRegisteredEvent(ctx, servicedatabase.InsertIdentityRegisteredEventParams{
		BlockNumber:    int64(activationBlockNumber + 1),
		BlockHash:      blockHash,
		TxIndex:        1,
		LogIndex:       1,
		Eon:            int64(config.GetEon()),
		IdentityPrefix: identityPrefix,
		Sender:         sender.Hex(),
		Timestamp:      time.Now().Unix(),
		Identity:       identity,
	})
	assert.NilError(t, err)

	assert.NilError(t, err)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case ev := <-decryptionTriggerChannel:
				assert.Equal(t, ev.Value.BlockNumber, uint64(activationBlockNumber+1))
				assert.DeepEqual(t, ev.Value.IdentityPreimages, []identitypreimage.IdentityPreimage{identity})
			}
		}
	}()

	err = kpr.processNewBlock(ctx, &event.LatestBlock{
		Number: &number.BlockNumber{
			Int: big.NewInt(int64(blockNumber)),
		},
		BlockHash: common.Hash(blockHash),
		Header: &types.Header{
			Time:   uint64(blockTimestamp),
			Number: big.NewInt(int64(blockNumber)),
		},
	})
	assert.NilError(t, err)
}

func TestShouldTriggerDecryption(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	ctx := context.Background()

	dbpool, dbclose := testsetup.NewTestDBPool(ctx, t, servicedatabase.Definition)
	t.Cleanup(dbclose)

	coreKeyperDB := corekeyperdatabase.New(dbpool)

	privateKey, _, _ := generateRandomAccount()

	decryptionTriggerChannel := make(chan *broker.Event[*epochkghandler.DecryptionTrigger])

	activationBlockNumber := 100
	eventTimestamp := time.Now().Unix()
	blockNumber := 100
	blockTimestamp := time.Now().Add(5 * time.Second).Unix()

	if blockTimestamp < 0 {
		t.Fatalf("blockTimestamp is negative: %d", blockTimestamp)
	}

	if eventTimestamp < 0 {
		t.Fatalf("eventTimestamp is negative: %d", eventTimestamp)
	}

	kpr := &Keyper{
		dbpool: dbpool,
		config: &Config{
			Chain: &ChainConfig{
				Node: &configuration.EthnodeConfig{
					PrivateKey: &keys.ECDSAPrivate{
						Key: privateKey,
					},
				},
			},
		},
		decryptionTriggerChannel: decryptionTriggerChannel,
	}

	err := coreKeyperDB.InsertBatchConfig(ctx, corekeyperdatabase.InsertBatchConfigParams{
		KeyperConfigIndex: 0,
		Keypers:           []string{kpr.config.GetAddress().Hex()},
		Threshold:         1,
	})
	assert.NilError(t, err)

	eon := config.GetEon()
	if eon > math.MaxInt64 {
		t.Fatalf("Eon is too large: %d", eon)
	}

	err = coreKeyperDB.InsertEon(ctx, corekeyperdatabase.InsertEonParams{
		Eon:                   int64(eon),
		Height:                0,
		ActivationBlockNumber: int64(activationBlockNumber),
		KeyperConfigIndex:     0,
	})
	assert.NilError(t, err)

	trigger := kpr.shouldTriggerDecryption(
		ctx,
		servicedatabase.IdentityRegisteredEvent{
			Eon:         int64(eon),
			BlockNumber: int64(blockNumber),
			Timestamp:   eventTimestamp,
		},
		&event.LatestBlock{
			Number: &number.BlockNumber{
				Int: big.NewInt(int64(blockNumber)),
			},
			Header: &types.Header{
				Time:   uint64(blockTimestamp),
				Number: big.NewInt(int64(blockNumber)),
			},
		},
	)
	assert.Equal(t, trigger, true)
}

func TestShouldNotTriggerDecryption(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	ctx := context.Background()

	dbpool, dbclose := testsetup.NewTestDBPool(ctx, t, servicedatabase.Definition)
	t.Cleanup(dbclose)

	privateKey, _, _ := generateRandomAccount()

	decryptionTriggerChannel := make(chan *broker.Event[*epochkghandler.DecryptionTrigger])

	eventTimestamp := time.Now().Unix()
	blockNumber := 100
	blockTimestamp := time.Now().Unix()

	if blockTimestamp < 0 {
		t.Fatalf("blockTimestamp is negative: %d", blockTimestamp)
	}

	if eventTimestamp < 0 {
		t.Fatalf("eventTimestamp is negative: %d", eventTimestamp)
	}

	kpr := &Keyper{
		dbpool: dbpool,
		config: &Config{
			Chain: &ChainConfig{
				Node: &configuration.EthnodeConfig{
					PrivateKey: &keys.ECDSAPrivate{
						Key: privateKey,
					},
				},
			},
		},
		decryptionTriggerChannel: decryptionTriggerChannel,
	}

	trigger := kpr.shouldTriggerDecryption(
		ctx,
		servicedatabase.IdentityRegisteredEvent{
			Timestamp: eventTimestamp,
		},
		&event.LatestBlock{
			Number: &number.BlockNumber{
				Int: big.NewInt(int64(blockNumber)),
			},
			Header: &types.Header{
				Time:   uint64(blockTimestamp),
				Number: big.NewInt(int64(blockNumber)),
			},
		},
	)

	assert.Equal(t, trigger, false)
}

func TestShouldTriggerDecryptionDifferentEon(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	ctx := context.Background()

	dbpool, dbclose := testsetup.NewTestDBPool(ctx, t, servicedatabase.Definition)
	t.Cleanup(dbclose)

	coreKeyperDB := corekeyperdatabase.New(dbpool)

	privateKey, _, _ := generateRandomAccount()

	decryptionTriggerChannel := make(chan *broker.Event[*epochkghandler.DecryptionTrigger])

	activationBlockNumber := 100
	eventTimestamp := time.Now().Unix()
	blockNumber := 100
	blockTimestamp := time.Now().Add(5 * time.Second).Unix()

	if blockTimestamp < 0 {
		t.Fatalf("blockTimestamp is negative: %d", blockTimestamp)
	}

	kpr := &Keyper{
		dbpool: dbpool,
		config: &Config{
			Chain: &ChainConfig{
				Node: &configuration.EthnodeConfig{
					PrivateKey: &keys.ECDSAPrivate{
						Key: privateKey,
					},
				},
			},
		},
		decryptionTriggerChannel: decryptionTriggerChannel,
	}

	eon := config.GetEon()
	if eon > math.MaxInt64-1 {
		t.Fatalf("Eon is too large: %d", eon)
	}

	err := coreKeyperDB.InsertBatchConfig(ctx, corekeyperdatabase.InsertBatchConfigParams{
		KeyperConfigIndex: 0,
		Keypers:           []string{kpr.config.GetAddress().Hex()},
		Threshold:         1,
	})
	assert.NilError(t, err)

	// Insert eon 0
	err = coreKeyperDB.InsertEon(ctx, corekeyperdatabase.InsertEonParams{
		Eon:                   int64(eon),
		Height:                0,
		ActivationBlockNumber: int64(activationBlockNumber),
		KeyperConfigIndex:     0,
	})
	assert.NilError(t, err)

	// Insert eon 1
	err = coreKeyperDB.InsertEon(ctx, corekeyperdatabase.InsertEonParams{
		Eon:                   int64(eon + 1),
		Height:                1,
		ActivationBlockNumber: int64(activationBlockNumber + 50),
		KeyperConfigIndex:     1,
	})
	assert.NilError(t, err)

	// Test with event from eon 0, but current block is in eon 1
	trigger := kpr.shouldTriggerDecryption(
		ctx,
		servicedatabase.IdentityRegisteredEvent{
			Eon:         int64(eon), // Event from eon 0
			BlockNumber: int64(blockNumber),
			Timestamp:   eventTimestamp,
		},
		&event.LatestBlock{
			Number: &number.BlockNumber{
				Int: big.NewInt(int64(blockNumber + 100)), // Block in eon 1
			},
			Header: &types.Header{
				Time:   uint64(blockTimestamp),
				Number: big.NewInt(int64(blockNumber + 100)),
			},
		},
	)
	assert.Equal(t, trigger, true)
}

func TestShouldNotTriggerDecryptionBeforeActivation(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	ctx := context.Background()

	dbpool, dbclose := testsetup.NewTestDBPool(ctx, t, servicedatabase.Definition)
	t.Cleanup(dbclose)

	coreKeyperDB := corekeyperdatabase.New(dbpool)

	privateKey, _, _ := generateRandomAccount()

	decryptionTriggerChannel := make(chan *broker.Event[*epochkghandler.DecryptionTrigger])

	activationBlockNumber := 200 // Eon activates at block 200
	eventTimestamp := time.Now().Unix()
	blockNumber := 150 // Current block is 150, before activation
	blockTimestamp := time.Now().Add(5 * time.Second).Unix()

	if blockTimestamp < 0 {
		t.Fatalf("blockTimestamp is negative: %d", blockTimestamp)
	}

	if eventTimestamp < 0 {
		t.Fatalf("eventTimestamp is negative: %d", eventTimestamp)
	}

	kpr := &Keyper{
		dbpool: dbpool,
		config: &Config{
			Chain: &ChainConfig{
				Node: &configuration.EthnodeConfig{
					PrivateKey: &keys.ECDSAPrivate{
						Key: privateKey,
					},
				},
			},
		},
		decryptionTriggerChannel: decryptionTriggerChannel,
	}

	err := coreKeyperDB.InsertBatchConfig(ctx, corekeyperdatabase.InsertBatchConfigParams{
		KeyperConfigIndex: 0,
		Keypers:           []string{kpr.config.GetAddress().Hex()},
		Threshold:         1,
	})
	assert.NilError(t, err)

	// Insert eon 0 that activates at block 200
	err = coreKeyperDB.InsertEon(ctx, corekeyperdatabase.InsertEonParams{
		Eon:                   int64(config.GetEon()),
		Height:                0,
		ActivationBlockNumber: int64(activationBlockNumber),
		KeyperConfigIndex:     0,
	})
	assert.NilError(t, err)

	// Test with event from eon 0, but current block (150) is before activation (200)
	trigger := kpr.shouldTriggerDecryption(
		ctx,
		servicedatabase.IdentityRegisteredEvent{
			Eon:         int64(config.GetEon()), // Event from eon 0
			BlockNumber: int64(blockNumber),
			Timestamp:   eventTimestamp,
		},
		&event.LatestBlock{
			Number: &number.BlockNumber{
				Int: big.NewInt(int64(blockNumber)), // Block 150, before activation at 200
			},
			Header: &types.Header{
				Time:   uint64(blockTimestamp),
				Number: big.NewInt(int64(blockNumber)),
			},
		},
	)
	assert.Equal(t, trigger, false)
}
