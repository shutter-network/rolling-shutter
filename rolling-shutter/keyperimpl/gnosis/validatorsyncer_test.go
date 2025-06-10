package gnosis

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
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
	url := mockBeaconClient(t, hex.EncodeToString(pubkey.Compress()), msg.ValidatorIndex)

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
	url := mockBeaconClient(t, hex.EncodeToString(pks[0].Compress()), msg.ValidatorIndex)

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
	url := mockBeaconClient(t, hex.EncodeToString(pubkey.Compress()), msg.ValidatorIndex)

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

func TestValidatorRegisterWithUnorderedIndices(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	msg := &validatorregistry.AggregateRegistrationMessage{
		Version:                  1,
		ChainID:                  2,
		ValidatorRegistryAddress: common.HexToAddress("0x1234567890123456789012345678901234567890"),
		ValidatorIndex:           3,
		Nonce:                    0,
		Count:                    2,
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

	// Create a mock beacon client that returns validators in a different order
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// The message requests indices [3, 4] but we return them in reverse order [4, 3]
		x := beaconapiclient.GetValidatorByIndexResponse{
			Finalized: true,
			Data: []beaconapiclient.ValidatorData{
				{
					Index: 4,
					Validator: beaconapiclient.Validator{
						PubkeyHex: hex.EncodeToString(pks[1].Compress()),
					},
				},
				{
					Index: 3,
					Validator: beaconapiclient.Validator{
						PubkeyHex: hex.EncodeToString(pks[0].Compress()),
					},
				},
			},
		}
		res, err := json.Marshal(x)
		assert.NilError(t, err)
		w.WriteHeader(http.StatusOK)
		_, err = w.Write(res)
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
			Address: common.HexToAddress("0x1234567890123456789012345678901234567890"),
		},
	}}

	finalEvents, err := vs.filterEvents(ctx, events)
	assert.NilError(t, err)

	// The event should still be accepted despite the different order
	assert.DeepEqual(t, finalEvents, events)
}

func TestValidatorRegisterWithManyIndices(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	msg := &validatorregistry.AggregateRegistrationMessage{
		Version:                  1,
		ChainID:                  2,
		ValidatorRegistryAddress: common.HexToAddress("0x1234567890123456789012345678901234567890"),
		ValidatorIndex:           3,
		Nonce:                    0,
		Count:                    100, // More than 64 indices
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

	// Create a mock beacon client that handles multiple chunks of indices
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Parse the indices from the query parameters
		query := r.URL.Query()
		indices := query["id"]

		// Create response data for this chunk
		var data []beaconapiclient.ValidatorData
		for _, indexStr := range indices {
			index, err := strconv.ParseUint(indexStr, 10, 64)
			assert.NilError(t, err)

			// Use the index to determine which pubkey to use
			pubkeyIndex := int(index) - 3 // Since we start from index 3
			data = append(data, beaconapiclient.ValidatorData{
				Index: index,
				Validator: beaconapiclient.Validator{
					PubkeyHex: hex.EncodeToString(pks[pubkeyIndex].Compress()),
				},
			})
		}

		x := beaconapiclient.GetValidatorByIndexResponse{
			Finalized: true,
			Data:      data,
		}
		res, err := json.Marshal(x)
		assert.NilError(t, err)
		w.WriteHeader(http.StatusOK)
		_, err = w.Write(res)
		assert.NilError(t, err)

		requestCount++
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
			Address: common.HexToAddress("0x1234567890123456789012345678901234567890"),
		},
	}}

	finalEvents, err := vs.filterEvents(ctx, events)
	assert.NilError(t, err)

	// The event should be accepted
	assert.DeepEqual(t, finalEvents, events)

	// Verify that we made the expected number of requests
	// For 100 indices with max 64 per request, we should make 2 requests
	expectedRequests := 2
	assert.Equal(t, requestCount, expectedRequests, "Expected %d requests for %d indices", expectedRequests, msg.Count)
}

func mockBeaconClient(t *testing.T, pubKeyHex string, index uint64) string {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		x := beaconapiclient.GetValidatorByIndexResponse{
			Finalized: true,
			Data: []beaconapiclient.ValidatorData{
				{
					Index: index,
					Validator: beaconapiclient.Validator{
						PubkeyHex: pubKeyHex,
					},
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
