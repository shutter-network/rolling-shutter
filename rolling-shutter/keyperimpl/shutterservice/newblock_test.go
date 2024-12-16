package shutterservice

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/shutter-network/contracts/v2/bindings/shutterregistry"
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
	"github.com/shutter-network/rolling-shutter/rolling-shutter/shdb"
	"gotest.tools/assert"
)

func TestProcessBlockSuccess(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	ctx := context.Background()

	dbpool, dbclose := testsetup.NewTestDBPool(ctx, t, servicedatabase.Definition)
	t.Cleanup(dbclose)

	serviceDB := servicedatabase.New(dbpool)
	obsDB := obskeyper.New(dbpool)
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

	blockHash, _ := generateRandomBytes(32)
	blockTimestamp := time.Now().Add(5 * time.Second).Unix()
	blockNumber := 102
	activationBlockNumber := 100

	identityPrefix, _ := generateRandomBytes(32)
	identity := computeIdentity(&shutterregistry.ShutterregistryIdentityRegistered{
		IdentityPrefix: [32]byte(identityPrefix),
		Sender:         sender,
	})
	keyperConfigIndex := uint64(1)

	err := coreKeyperDB.InsertEon(ctx, corekeyperdatabase.InsertEonParams{
		Eon:                   int64(config.GetEon()),
		Height:                0,
		ActivationBlockNumber: int64(activationBlockNumber),
		KeyperConfigIndex:     1,
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

	err = obsDB.InsertKeyperSet(ctx, obskeyper.InsertKeyperSetParams{
		KeyperConfigIndex:     int64(keyperConfigIndex),
		ActivationBlockNumber: int64(activationBlockNumber),
		Keypers:               shdb.EncodeAddresses([]common.Address{sender}),
		Threshold:             2,
	})
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

	obsDB := obskeyper.New(dbpool)

	privateKey, sender, _ := generateRandomAccount()

	decryptionTriggerChannel := make(chan *broker.Event[*epochkghandler.DecryptionTrigger])

	activationBlockNumber := 100
	keyperConfigIndex := uint64(1)
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

	err := obsDB.InsertKeyperSet(ctx, obskeyper.InsertKeyperSetParams{
		KeyperConfigIndex:     int64(keyperConfigIndex),
		ActivationBlockNumber: int64(activationBlockNumber),
		Keypers:               shdb.EncodeAddresses([]common.Address{sender}),
		Threshold:             2,
	})
	assert.NilError(t, err)

	trigger := kpr.shouldTriggerDecryption(
		ctx,
		obsDB,
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
	assert.Equal(t, trigger, true)
}

func TestShouldNotTriggerDecryption(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	ctx := context.Background()

	dbpool, dbclose := testsetup.NewTestDBPool(ctx, t, servicedatabase.Definition)
	t.Cleanup(dbclose)

	obsDB := obskeyper.New(dbpool)

	privateKey, sender, _ := generateRandomAccount()

	decryptionTriggerChannel := make(chan *broker.Event[*epochkghandler.DecryptionTrigger])

	activationBlockNumber := 100
	keyperConfigIndex := uint64(1)
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

	err := obsDB.InsertKeyperSet(ctx, obskeyper.InsertKeyperSetParams{
		KeyperConfigIndex:     int64(keyperConfigIndex),
		ActivationBlockNumber: int64(activationBlockNumber),
		Keypers:               shdb.EncodeAddresses([]common.Address{sender}),
		Threshold:             2,
	})
	assert.NilError(t, err)

	trigger := kpr.shouldTriggerDecryption(
		ctx,
		obsDB,
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
