package shutterservice

import (
	"crypto/ecdsa"
	"math/rand"
	"testing"

	cryptoRand "crypto/rand"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	registryBindings "github.com/shutter-network/contracts/v2/bindings/shutterregistry"
	"gotest.tools/assert"
)

func TestFilterIdentityRegisteredEvent(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	events := make([]*registryBindings.ShutterregistryIdentityRegistered, 2)
	for i := 0; i < 2; i++ {
		identityPrefix, err := generateRandomBytes(32)
		assert.NilError(t, err)
		sender, err := generateRandomAddress()
		assert.NilError(t, err)
		events[i] = &registryBindings.ShutterregistryIdentityRegistered{
			Eon:            uint64(i),
			IdentityPrefix: [32]byte(identityPrefix),
			Sender:         sender,
			Timestamp:      rand.Uint64(),
		}
	}

	rs := RegistrySyncer{}

	finalEvents := rs.filterEvents(events)
	assert.DeepEqual(t, finalEvents, events)
}

func generateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := cryptoRand.Read(b)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func generateRandomAddress() (common.Address, error) {
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		return common.Address{}, err
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return common.Address{}, err
	}
	address := crypto.PubkeyToAddress(*publicKeyECDSA)

	return address, nil
}
