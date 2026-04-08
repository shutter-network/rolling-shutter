package shutterservice

import (
	"bytes"
	"context"
	"database/sql"
	"math"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/jackc/pgx/v4/pgxpool"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
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

const (
	testKeyperConfigIndex   int64 = 7
	testKeyperConfigIndex32 int32 = 7
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

	decryptionTriggerChannel := make(chan *broker.Event[*epochkghandler.DecryptionTrigger], 1)

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
	const (
		activationBlockNumber              = 100
		activationBlockNumberUint64 uint64 = 100
	)
	eon := config.GetEon()
	if eon > math.MaxInt64 {
		t.Fatalf("Eon is too large: %d", eon)
	}
	eonInt64 := int64(eon)

	identityPrefix, _ := generateRandom32Bytes()
	identity := identityPrefix
	identity = append(identity, sender.Bytes()...)

	insertBatchConfig(ctx, t, coreKeyperDB, []string{kpr.config.GetAddress().Hex()}, int64(activationBlockNumber))
	insertEon(ctx, t, coreKeyperDB, eonInt64, int64(activationBlockNumber))
	insertDKGResult(ctx, t, coreKeyperDB, eonInt64, true)

	_, err := serviceDB.InsertIdentityRegisteredEvent(ctx, servicedatabase.InsertIdentityRegisteredEventParams{
		BlockNumber:    int64(activationBlockNumber + 1),
		BlockHash:      blockHash,
		TxIndex:        1,
		LogIndex:       1,
		Eon:            testKeyperConfigIndex,
		IdentityPrefix: identityPrefix,
		Sender:         sender.Hex(),
		Timestamp:      time.Now().Unix(),
		Identity:       identity,
	})
	assert.NilError(t, err)

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

	select {
	case ev := <-decryptionTriggerChannel:
		assert.Equal(t, ev.Value.BlockNumber, activationBlockNumberUint64)
		assert.DeepEqual(t, ev.Value.IdentityPreimages, []identitypreimage.IdentityPreimage{identity})
	case <-time.After(2 * time.Second):
		t.Fatal("expected decryption trigger")
	}
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

	const activationBlockNumber = 100
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

	eon := config.GetEon()
	if eon > math.MaxInt64 {
		t.Fatalf("Eon is too large: %d", eon)
	}

	insertBatchConfig(ctx, t, coreKeyperDB, []string{kpr.config.GetAddress().Hex()}, int64(activationBlockNumber))
	insertEon(ctx, t, coreKeyperDB, int64(eon), int64(activationBlockNumber))
	insertDKGResult(ctx, t, coreKeyperDB, int64(eon), true)

	trigger, err := kpr.shouldTriggerDecryption(
		ctx,
		servicedatabase.IdentityRegisteredEvent{
			Eon:         testKeyperConfigIndex,
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

	const (
		activationBlockNumber                   = 100
		laterActivationBlockNumber              = activationBlockNumber + 50
		laterActivationBlockNumberUint64 uint64 = laterActivationBlockNumber
	)
	eventTimestamp := time.Now().Unix()
	blockNumber := 100
	blockTimestamp := time.Now().Add(5 * time.Second).Unix()
	eon := config.GetEon()
	if eon > math.MaxInt64-1 {
		t.Fatalf("Eon is too large: %d", eon)
	}
	earlierEon := int64(eon)
	laterEon := earlierEon + 1

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

	insertBatchConfig(ctx, t, coreKeyperDB, []string{kpr.config.GetAddress().Hex()}, int64(activationBlockNumber))
	insertEon(ctx, t, coreKeyperDB, earlierEon, int64(activationBlockNumber))
	insertDKGResult(ctx, t, coreKeyperDB, earlierEon, true)
	insertEon(ctx, t, coreKeyperDB, laterEon, int64(laterActivationBlockNumber))
	insertDKGResult(ctx, t, coreKeyperDB, laterEon, true)

	registeredEvent := servicedatabase.IdentityRegisteredEvent{
		Eon:         testKeyperConfigIndex,
		BlockNumber: int64(blockNumber),
		Timestamp:   eventTimestamp,
		Identity:    b32(0x01),
	}

	trigger, err := kpr.shouldTriggerDecryption(
		ctx,
		registeredEvent,
		&event.LatestBlock{
			Number: &number.BlockNumber{
				Int: big.NewInt(int64(blockNumber + 100)),
			},
			Header: &types.Header{
				Time:   uint64(blockTimestamp), //nolint:gosec
				Number: big.NewInt(int64(blockNumber + 100)),
			},
		},
	)
	assert.NilError(t, err)
	assert.Equal(t, trigger, true)

	triggers, err := kpr.createTriggersFromIdentityRegisteredEvents(
		ctx,
		[]servicedatabase.IdentityRegisteredEvent{registeredEvent},
		nil,
	)
	assert.NilError(t, err)
	assert.Equal(t, len(triggers), 1)
	assert.Equal(t, triggers[0].BlockNumber, laterActivationBlockNumberUint64)
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

	eon := config.GetEon()
	if eon > math.MaxInt64 {
		t.Fatalf("Eon is too large: %d", eon)
	}

	insertBatchConfig(ctx, t, coreKeyperDB, []string{kpr.config.GetAddress().Hex()}, int64(activationBlockNumber))
	insertEon(ctx, t, coreKeyperDB, int64(eon), int64(activationBlockNumber))
	insertDKGResult(ctx, t, coreKeyperDB, int64(eon), true)

	trigger, err := kpr.shouldTriggerDecryption(
		ctx,
		servicedatabase.IdentityRegisteredEvent{
			Eon:         testKeyperConfigIndex,
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

func TestShouldNotTriggerDecryptionWithoutSuccessfulDKG(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	ctx := context.Background()

	dbpool, dbclose := testsetup.NewTestDBPool(ctx, t, servicedatabase.Definition)
	t.Cleanup(dbclose)

	coreKeyperDB := corekeyperdatabase.New(dbpool)

	privateKey, _, _ := generateRandomAccount()

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
	}

	activationBlockNumber := int64(100)
	eon := config.GetEon()
	if eon > math.MaxInt64 {
		t.Fatalf("Eon is too large: %d", eon)
	}
	eonInt64 := int64(eon)
	blockTimestamp := time.Now().Add(5 * time.Second).Unix()
	if blockTimestamp < 0 {
		t.Fatalf("blockTimestamp is negative: %d", blockTimestamp)
	}

	insertBatchConfig(ctx, t, coreKeyperDB, []string{kpr.config.GetAddress().Hex()}, activationBlockNumber)
	insertEon(ctx, t, coreKeyperDB, eonInt64, activationBlockNumber)
	insertDKGResult(ctx, t, coreKeyperDB, eonInt64, false)

	trigger, err := kpr.shouldTriggerDecryption(
		ctx,
		servicedatabase.IdentityRegisteredEvent{
			Eon:         testKeyperConfigIndex,
			BlockNumber: activationBlockNumber + 1,
			Timestamp:   time.Now().Unix(),
		},
		&event.LatestBlock{
			Number: &number.BlockNumber{
				Int: big.NewInt(activationBlockNumber + 10),
			},
			Header: &types.Header{
				Time:   uint64(blockTimestamp),
				Number: big.NewInt(activationBlockNumber + 10),
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
	_ = sender

	return kpr, servicedatabase.New(dbpool), 1
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
			Identity:       b32(r.identity),
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

func insertBatchConfig(
	ctx context.Context,
	t *testing.T,
	coreKeyperDB *corekeyperdatabase.Queries,
	keypers []string,
	activationBlockNumber int64,
) {
	t.Helper()

	err := coreKeyperDB.InsertBatchConfig(ctx, corekeyperdatabase.InsertBatchConfigParams{
		KeyperConfigIndex:     testKeyperConfigIndex32,
		Keypers:               keypers,
		Threshold:             1,
		ActivationBlockNumber: activationBlockNumber,
	})
	assert.NilError(t, err)
}

func insertEon(
	ctx context.Context,
	t *testing.T,
	coreKeyperDB *corekeyperdatabase.Queries,
	eon int64,
	activationBlockNumber int64,
) {
	t.Helper()

	err := coreKeyperDB.InsertEon(ctx, corekeyperdatabase.InsertEonParams{
		Eon:                   eon,
		Height:                0,
		ActivationBlockNumber: activationBlockNumber,
		KeyperConfigIndex:     testKeyperConfigIndex,
	})
	assert.NilError(t, err)
}

func insertDKGResult(
	ctx context.Context,
	t *testing.T,
	coreKeyperDB *corekeyperdatabase.Queries,
	eon int64,
	success bool,
) {
	t.Helper()

	err := coreKeyperDB.InsertDKGResult(ctx, corekeyperdatabase.InsertDKGResultParams{
		Eon:        eon,
		Success:    success,
		Error:      sql.NullString{},
		PureResult: []byte{},
	})
	assert.NilError(t, err)
}

func b32(last byte) []byte {
	b := make([]byte, 32)
	b[31] = last
	return b
}
