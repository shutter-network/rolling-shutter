package shutterservice

import (
	"bytes"
	"context"
	"math"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/jackc/pgx/v4/pgxpool"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"gotest.tools/assert"

	obskeyper "github.com/shutter-network/rolling-shutter/rolling-shutter/chainobserver/db/keyper"
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
	identity := identityPrefix
	identity = append(identity, sender.Bytes()...)

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

	trigger, err := kpr.shouldTriggerDecryption(
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
	assert.NilError(t, err)
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

	trigger, err := kpr.shouldTriggerDecryption(
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
	assert.NilError(t, err)
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
		Eon:                   int64(eon + 1), //nolint:gosec
		Height:                1,
		ActivationBlockNumber: int64(activationBlockNumber + 50),
		KeyperConfigIndex:     1,
	})
	assert.NilError(t, err)

	// Test with event from eon 0, but current block is in eon 1
	trigger, err := kpr.shouldTriggerDecryption(
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
				Time:   uint64(blockTimestamp), //nolint:gosec
				Number: big.NewInt(int64(blockNumber + 100)),
			},
		},
	)
	assert.NilError(t, err)
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

	// Insert eon 0 that activates at block 200
	err = coreKeyperDB.InsertEon(ctx, corekeyperdatabase.InsertEonParams{
		Eon:                   int64(eon),
		Height:                0,
		ActivationBlockNumber: int64(activationBlockNumber),
		KeyperConfigIndex:     0,
	})
	assert.NilError(t, err)

	// Test with event from eon 0, but current block (150) is before activation (200)
	trigger, err := kpr.shouldTriggerDecryption(
		ctx,
		servicedatabase.IdentityRegisteredEvent{
			Eon:         int64(eon), // Event from eon 0
			BlockNumber: int64(blockNumber),
			Timestamp:   eventTimestamp,
		},
		&event.LatestBlock{
			Number: &number.BlockNumber{
				Int: big.NewInt(int64(blockNumber)), // Block 150, before activation at 200
			},
			Header: &types.Header{
				Time:   uint64(blockTimestamp), //nolint:gosec
				Number: big.NewInt(int64(blockNumber)),
			},
		},
	)
	assert.NilError(t, err)
	assert.Equal(t, trigger, false)
}

func setupEventBasedOrderingTest(
	ctx context.Context,
	t *testing.T,
	dbpool *pgxpool.Pool,
) (*Keyper, *servicedatabase.Queries, int64) {
	t.Helper()

	const keyperIndex = uint64(1)
	testsetup.InitializeEon(ctx, t, dbpool, config, keyperIndex)

	eon := config.GetEon()
	if eon > math.MaxInt64 {
		t.Fatalf("Eon is too large: %d", eon)
	}
	eonInt64 := int64(eon)

	privateKey, sender, err := generateRandomAccount()
	assert.NilError(t, err)

	kpr := &Keyper{
		dbpool: dbpool,
		config: &Config{
			Chain: &ChainConfig{
				Node: &configuration.EthnodeConfig{
					PrivateKey: &keys.ECDSAPrivate{Key: privateKey},
				},
			},
		},
	}

	err = obskeyper.New(dbpool).InsertKeyperSet(ctx, obskeyper.InsertKeyperSetParams{
		KeyperConfigIndex:     1,
		ActivationBlockNumber: 0,
		Keypers:               []string{sender.Hex()},
		Threshold:             1,
	})
	assert.NilError(t, err)

	return kpr, servicedatabase.New(dbpool), eonInt64
}

func TestFiredTriggersProducesOrderedShares(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	dbpool, dbclose := testsetup.NewTestDBPool(ctx, t, servicedatabase.Definition)
	t.Cleanup(dbclose)

	kpr, serviceDB, eon := setupEventBasedOrderingTest(ctx, t, dbpool)

	type row struct {
		identity byte
		prefix   byte
		sender   string
	}

	inserted := []row{
		{identity: 0x04, prefix: 0x14, sender: "0x0000000000000000000000000000000000000011"},
		{identity: 0x03, prefix: 0x13, sender: "0x0000000000000000000000000000000000000011"},
		{identity: 0x02, prefix: 0x12, sender: "0x0000000000000000000000000000000000000011"},
		{identity: 0x01, prefix: 0x11, sender: "0x0000000000000000000000000000000000000011"},
		{identity: 0x05, prefix: 0x15, sender: "0x0000000000000000000000000000000000000011"},
	}

	for i, r := range inserted {
		_, err := serviceDB.InsertEventTriggerRegisteredEvent(ctx, servicedatabase.InsertEventTriggerRegisteredEventParams{
			BlockNumber:           int64(100 + i),
			BlockHash:             []byte{byte(100 + i)},
			TxIndex:               0,
			LogIndex:              0,
			Eon:                   eon,
			IdentityPrefix:        b32(r.prefix),
			Sender:                r.sender,
			Definition:            []byte{0x01},
			ExpirationBlockNumber: 10_000,
			Identity:              b32(r.identity),
		})
		assert.NilError(t, err)

		err = serviceDB.InsertFiredTrigger(ctx, servicedatabase.InsertFiredTriggerParams{
			Eon:            eon,
			IdentityPrefix: b32(r.prefix),
			Sender:         r.sender,
			BlockNumber:    int64(200 + i),
			BlockHash:      []byte{byte(200 + i)},
			TxIndex:        0,
			LogIndex:       0,
		})
		assert.NilError(t, err)
	}

	triggers, err := kpr.prepareEventBasedTriggers(ctx)
	assert.NilError(t, err)
	assert.Equal(t, len(triggers), 1)
	assert.Equal(t, len(triggers[0].IdentityPreimages), len(inserted))
	for i := 1; i < len(triggers[0].IdentityPreimages); i++ {
		assert.Assert(t, bytes.Compare(
			triggers[0].IdentityPreimages[i-1],
			triggers[0].IdentityPreimages[i],
		) < 0)
	}

	coreDB := corekeyperdatabase.New(dbpool)
	triggerBlockNumber := triggers[0].BlockNumber
	if triggerBlockNumber > math.MaxInt64 {
		t.Fatalf("BlockNumber is too large: %d", triggerBlockNumber)
	}

	triggerEon, err := coreDB.GetEonForBlockNumber(ctx, int64(triggerBlockNumber))
	assert.NilError(t, err)

	keyShareHandler := &epochkghandler.KeyShareHandler{
		InstanceID:           config.GetInstanceID(),
		KeyperAddress:        config.GetAddress(),
		MaxNumKeysPerMessage: config.GetMaxNumKeysPerMessage(),
		DBPool:               dbpool,
	}
	msg, err := keyShareHandler.ConstructDecryptionKeyShares(ctx, triggerEon, triggers[0].IdentityPreimages)
	assert.NilError(t, err)

	validator := epochkghandler.NewDecryptionKeyShareHandler(config, dbpool)
	res, err := validator.ValidateMessage(ctx, msg)
	assert.Equal(t, res, pubsub.ValidationAccept)
	assert.NilError(t, err)
}

func b32(last byte) []byte {
	b := make([]byte, 32)
	b[31] = last
	return b
}
