package shutterservice

import (
	"context"
	"crypto/ecdsa"
	"testing"
	"time"

	cryptoRand "crypto/rand"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/jackc/pgx/v4"
	registryBindings "github.com/shutter-network/contracts/v2/bindings/shutterregistry"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyperimpl/shutterservice/database"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/testsetup"
	"gotest.tools/assert"
)

func TestFilterIdentityRegisteredEvents(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	events := make([]*registryBindings.ShutterregistryIdentityRegistered, 2)
	for i := 0; i < 2; i++ {
		identityPrefix, err := generateRandom32Bytes()
		assert.NilError(t, err)
		_, sender, err := generateRandomAccount()
		assert.NilError(t, err)
		events[i] = &registryBindings.ShutterregistryIdentityRegistered{
			Eon:            uint64(i),
			IdentityPrefix: [32]byte(identityPrefix),
			Sender:         sender,
			Timestamp:      uint64(time.Now().Unix()),
		}
	}

	rs := RegistrySyncer{}

	finalEvents := rs.filterEvents(events)
	assert.DeepEqual(t, finalEvents, events)
}

func TestInsertIdentityRegisteredEvents(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	events := make([]*registryBindings.ShutterregistryIdentityRegistered, 2)
	for i := 0; i < 2; i++ {
		identityPrefix, err := generateRandom32Bytes()
		assert.NilError(t, err)
		_, sender, err := generateRandomAccount()
		assert.NilError(t, err)
		events[i] = &registryBindings.ShutterregistryIdentityRegistered{
			Eon:            uint64(i),
			IdentityPrefix: [32]byte(identityPrefix),
			Sender:         sender,
			Timestamp:      uint64(time.Now().Unix()),
		}
	}

	dbpool, dbclose := testsetup.NewTestDBPool(ctx, t, database.Definition)
	t.Cleanup(dbclose)

	rs := RegistrySyncer{}

	err := dbpool.BeginFunc(ctx, func(tx pgx.Tx) error {
		err := rs.insertIdentityRegisteredEvents(ctx, tx, events)
		return err
	})
	assert.NilError(t, err)
}

func generateRandom32Bytes() ([]byte, error) {
	b := make([]byte, 32)
	_, err := cryptoRand.Read(b)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func generateRandomAccount() (*ecdsa.PrivateKey, common.Address, error) {
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		return nil, common.Address{}, err
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, common.Address{}, err
	}
	address := crypto.PubkeyToAddress(*publicKeyECDSA)

	return privateKey, address, nil
}
