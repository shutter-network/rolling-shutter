package gnosis

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	validatorRegistryBindings "github.com/shutter-network/gnosh-contracts/gnoshcontracts/validatorregistry"
	blst "github.com/supranational/blst/bindings/go"
	"gotest.tools/assert"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyperimpl/gnosis/database"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/beaconapiclient"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/testsetup"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/validatorregistry"
)

func TestLegacyValidatorRegisterFilterEvent(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	msg := &validatorregistry.LegacyRegistrationMessage{
		Version:                  0,
		ChainID:                  2,
		ValidatorRegistryAddress: common.HexToAddress("0x1234567890123456789012345678901234567890"),
		ValidatorIndex:           3,
		Nonce:                    0,
		IsRegistration:           true,
	}
	ctx := context.Background()

	var ikm [32]byte
	privkey := blst.KeyGen(ikm[:])
	pubkey := new(blst.P1Affine).From(privkey)

	sig := validatorregistry.CreateSignature(privkey, msg)
	url := mockBeaconClient(t, hex.EncodeToString(pubkey.Compress()))

	cl, err := beaconapiclient.New(url)
	assert.NilError(t, err)

	dbpool, dbclose := testsetup.NewTestDBPool(ctx, t, database.Definition)
	t.Cleanup(dbclose)

	vs := ValidatorSyncer{
		BeaconAPIClient: cl,
		DBPool:          dbpool,
		ChainID:         msg.ChainID,
	}

	events := []*validatorRegistryBindings.ValidatorregistryUpdated{{
		Signature: sig.Compress(),
		Message:   msg.Marshal(),
		Raw: types.Log{
			Address: common.HexToAddress("0x1234567890123456789012345678901234567890"),
		},
	}}

	finalEvents, err := vs.filterEvents(ctx, events)
	assert.NilError(t, err)

	assert.DeepEqual(t, finalEvents, events)
}

func TestAggregateValidatorRegisterFilterEvent(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	msg := &validatorregistry.AggregateRegistrationMessage{
		Version:                  1,
		ChainID:                  2,
		ValidatorRegistryAddress: common.HexToAddress("0x1234567890123456789012345678901234567890"),
		ValidatorIndex:           3,
		Nonce:                    0,
		Count:                    1,
		IsRegistration:           true,
	}
	ctx := context.Background()

	var ikm [32]byte
	var sks []*blst.SecretKey
	var pks []*blst.P1Affine
	for i := 0; i < int(msg.Count); i++ {
		privkey := blst.KeyGen(ikm[:])
		pubkey := new(blst.P1Affine).From(privkey)
		sks = append(sks, privkey)
		pks = append(pks, pubkey)
	}

	sig := validatorregistry.CreateAggregateSignature(sks, msg)
	url := mockBeaconClient(t, hex.EncodeToString(pks[0].Compress()))

	cl, err := beaconapiclient.New(url)
	assert.NilError(t, err)

	dbpool, dbclose := testsetup.NewTestDBPool(ctx, t, database.Definition)
	t.Cleanup(dbclose)

	vs := ValidatorSyncer{
		BeaconAPIClient: cl,
		DBPool:          dbpool,
		ChainID:         msg.ChainID,
	}

	events := []*validatorRegistryBindings.ValidatorregistryUpdated{{
		Signature: sig.Compress(),
		Message:   msg.Marshal(),
		Raw: types.Log{
			Address: common.HexToAddress("0x1234567890123456789012345678901234567890"),
		},
	}}

	finalEvents, err := vs.filterEvents(ctx, events)
	assert.NilError(t, err)

	assert.DeepEqual(t, finalEvents, events)
}

func TestValidatorRegisterWithInvalidNonce(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	msg := &validatorregistry.LegacyRegistrationMessage{
		Version:                  0,
		ChainID:                  2,
		ValidatorRegistryAddress: common.HexToAddress("0x1234567890123456789012345678901234567890"),
		ValidatorIndex:           3,
		Nonce:                    1, // This nonce will be invalid
		IsRegistration:           true,
	}
	ctx := context.Background()

	var ikm [32]byte
	privkey := blst.KeyGen(ikm[:])
	pubkey := new(blst.P1Affine).From(privkey)

	sig := validatorregistry.CreateSignature(privkey, msg)
	url := mockBeaconClient(t, hex.EncodeToString(pubkey.Compress()))

	cl, err := beaconapiclient.New(url)
	assert.NilError(t, err)

	dbpool, dbclose := testsetup.NewTestDBPool(ctx, t, database.Definition)
	t.Cleanup(dbclose)

	// Insert a previous registration with a higher nonce
	db := database.New(dbpool)
	err = db.InsertValidatorRegistration(ctx, database.InsertValidatorRegistrationParams{
		BlockNumber:    1,
		BlockHash:      []byte{1, 2, 3},
		TxIndex:        0,
		LogIndex:       0,
		ValidatorIndex: 3,
		Nonce:          2, // Higher nonce than the event
		IsRegistration: true,
	})
	assert.NilError(t, err)

	vs := ValidatorSyncer{
		BeaconAPIClient: cl,
		DBPool:          dbpool,
		ChainID:         msg.ChainID,
	}

	events := []*validatorRegistryBindings.ValidatorregistryUpdated{{
		Signature: sig.Compress(),
		Message:   msg.Marshal(),
		Raw: types.Log{
			Address:     common.HexToAddress("0x1234567890123456789012345678901234567890"),
			BlockNumber: 2,
			TxIndex:     0,
			Index:       0,
		},
	}}

	finalEvents, err := vs.filterEvents(ctx, events)
	assert.NilError(t, err)

	// The event should be filtered out due to invalid nonce
	assert.Equal(t, len(finalEvents), 0)
}

func TestValidatorRegisterWithUnknownValidator(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	msg := &validatorregistry.LegacyRegistrationMessage{
		Version:                  0,
		ChainID:                  2,
		ValidatorRegistryAddress: common.HexToAddress("0x1234567890123456789012345678901234567890"),
		ValidatorIndex:           3,
		Nonce:                    1,
		IsRegistration:           true,
	}
	ctx := context.Background()

	var ikm [32]byte
	privkey := blst.KeyGen(ikm[:])

	sig := validatorregistry.CreateSignature(privkey, msg)

	// Create a mock beacon client that returns not found for the validator
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, err := w.Write([]byte(`{"finalized": false, "data": null}`))
		assert.NilError(t, err)
	}))
	defer server.Close()

	cl, err := beaconapiclient.New(server.URL)
	assert.NilError(t, err)

	dbpool, dbclose := testsetup.NewTestDBPool(ctx, t, database.Definition)
	t.Cleanup(dbclose)

	vs := ValidatorSyncer{
		BeaconAPIClient: cl,
		DBPool:          dbpool,
		ChainID:         msg.ChainID,
	}

	events := []*validatorRegistryBindings.ValidatorregistryUpdated{{
		Signature: sig.Compress(),
		Message:   msg.Marshal(),
		Raw: types.Log{
			Address:     common.HexToAddress("0x1234567890123456789012345678901234567890"),
			BlockNumber: 2,
			TxIndex:     0,
			Index:       0,
		},
	}}

	finalEvents, err := vs.filterEvents(ctx, events)
	assert.NilError(t, err)

	// The event should be filtered out because the validator is unknown
	assert.Equal(t, len(finalEvents), 0)
}

func mockBeaconClient(t *testing.T, pubKeyHex string) string {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		x := beaconapiclient.GetValidatorByIndexResponse{
			Finalized: true,
			Data: beaconapiclient.ValidatorData{
				Validator: beaconapiclient.Validator{
					PubkeyHex: pubKeyHex,
				},
			},
		}
		res, err := json.Marshal(x)
		assert.NilError(t, err)
		w.WriteHeader(http.StatusOK)
		_, err = w.Write(res)
		assert.NilError(t, err)
	})).URL
}
