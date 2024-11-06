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
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyperimpl/gnosis/database"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/beaconapiclient"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/testsetup"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/validatorregistry"
	blst "github.com/supranational/blst/bindings/go"
	"gotest.tools/assert"
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

	events := []*validatorRegistryBindings.ValidatorregistryUpdated{&validatorRegistryBindings.ValidatorregistryUpdated{
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

	events := []*validatorRegistryBindings.ValidatorregistryUpdated{&validatorRegistryBindings.ValidatorregistryUpdated{
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

func mockBeaconClient(t *testing.T, pubKeyHex string) string {
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
		w.Write(res)
	})).URL
}
