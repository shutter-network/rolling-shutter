package shutterservice

import (
	"context"
	"crypto/ecdsa"
	cryptoRand "crypto/rand"
	"math"
	"math/rand/v2"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/jackc/pgx/v4"
	"gotest.tools/assert"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyperimpl/shutterservice/database"
	registryBindings "github.com/shutter-network/rolling-shutter/rolling-shutter/keyperimpl/shutterservice/help"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/testsetup"
)

func TestFilterEventTriggerRegisteredEvents(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	events := make([]*registryBindings.ShutterRegistryEventTriggerRegistered, 2)
	for i := 0; i < 2; i++ {
		identityPrefix, err := generateRandom32Bytes()
		assert.NilError(t, err)
		_, sender, err := generateRandomAccount()
		assert.NilError(t, err)
		def, err := generateRandomEventTriggerDefinition()
		events[i] = &registryBindings.ShutterRegistryEventTriggerRegistered{
			Eon:               uint64(i),
			IdentityPrefix:    [32]byte(identityPrefix),
			Sender:            sender,
			TriggerDefinition: def.MarshalBytes(),
			Ttl:               rand.Uint64(),
		}
	}

	rs := RegistrySyncer{}

	finalEvents := rs.filterEvents(events)
	assert.DeepEqual(t, finalEvents, events)
}

func TestInsertEventTriggerRegisteredEvents(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	events := make([]*registryBindings.ShutterRegistryEventTriggerRegistered, 2)
	for i := 0; i < 2; i++ {
		identityPrefix, err := generateRandom32Bytes()
		assert.NilError(t, err)
		_, sender, err := generateRandomAccount()
		assert.NilError(t, err)
		def, err := generateRandomEventTriggerDefinition()
		assert.NilError(t, err)
		events[i] = &registryBindings.ShutterRegistryEventTriggerRegistered{
			Eon:               uint64(i),
			IdentityPrefix:    [32]byte(identityPrefix),
			Sender:            sender,
			TriggerDefinition: def.MarshalBytes(),
			Ttl:               rand.Uint64() % math.MaxInt64,
		}
	}

	dbpool, dbclose := testsetup.NewTestDBPool(ctx, t, database.Definition)
	t.Cleanup(dbclose)

	rs := RegistrySyncer{}

	err := dbpool.BeginFunc(ctx, func(tx pgx.Tx) error {
		err := rs.insertEventTriggerRegisteredEvents(ctx, tx, events)
		return err
	})
	assert.NilError(t, err)
}

func generateRandomEventTriggerDefinition() (*EventTriggerDefinition, error) {
	_, randomContract, err := generateRandomAccount()
	if err != nil {
		return nil, err
	}
	randomTopic, err := generateRandom32Bytes()
	if err != nil {
		return nil, err
	}
	randomSig, err := generateRandom32Bytes()
	if err != nil {
		return nil, err
	}
	def := EventTriggerDefinition{
		Contract: randomContract,
		Signature: EvtSignature{
			hashed: (*common.Hash)(randomSig),
		},
		Conditions: []Condition{
			{
				Constraint: MatchConstraint{
					target: randomTopic,
				},
				Location: TopicData{
					number: 1,
				},
			},
		},
	}
	return &def, nil
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
