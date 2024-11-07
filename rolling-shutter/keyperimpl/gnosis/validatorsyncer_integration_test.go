package gnosis

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	validatorRegistryBindings "github.com/shutter-network/gnosh-contracts/gnoshcontracts/validatorregistry"
	"gotest.tools/assert"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyperimpl/gnosis/database"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/beaconapiclient"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/testsetup"
)

func TestAggregateValidationWithData(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	ctx := context.Background()
	url := mockBeaconClientWithJSONData(t)
	cl, err := beaconapiclient.New(url)
	assert.NilError(t, err)
	dbpool, dbclose := testsetup.NewTestDBPool(ctx, t, database.Definition)
	t.Cleanup(dbclose)
	vs := ValidatorSyncer{
		BeaconAPIClient: cl,
		DBPool:          dbpool,
		ChainID:         10200,
	}

	msg := readMsg(t)

	message, err := hex.DecodeString(msg["message"][2:])
	assert.NilError(t, err)

	signature, err := hex.DecodeString(msg["signature"][2:])
	assert.NilError(t, err)

	events := []*validatorRegistryBindings.ValidatorregistryUpdated{{
		Signature: signature,
		Message:   message,
		Raw: types.Log{
			Address: common.HexToAddress("0xa9289A3Dd14FEBe10611119bE81E5d35eAaC3084"),
		},
	}}

	finalEvents, err := vs.filterEvents(ctx, events)
	assert.NilError(t, err)

	assert.DeepEqual(t, len(finalEvents), 1)
}

func mockBeaconClientWithJSONData(t *testing.T) string {
	t.Helper()
	jsonFile, err := os.Open("../../../testdata/validatorInfo.json")
	assert.NilError(t, err)
	defer jsonFile.Close()

	byteValue, _ := io.ReadAll(jsonFile)
	var result map[string]string
	json.Unmarshal([]byte(byteValue), &result)

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(r.URL.Path, "/")

		x := beaconapiclient.GetValidatorByIndexResponse{
			Finalized: true,
			Data: beaconapiclient.ValidatorData{
				Validator: beaconapiclient.Validator{
					PubkeyHex: result[parts[len(parts)-1]],
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

func readMsg(t *testing.T) map[string]string {
	t.Helper()
	jsonFile, err := os.Open("../../../testdata/signedRegistrations.json")
	assert.NilError(t, err)
	defer jsonFile.Close()

	byteValue, _ := io.ReadAll(jsonFile)
	var result map[string]string
	json.Unmarshal([]byte(byteValue), &result)
	return result
}
