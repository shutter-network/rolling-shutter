package epochkghandler

import (
	"context"
	"database/sql"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"gotest.tools/assert"

	"github.com/shutter-network/shutter/shlib/puredkg"
	"github.com/shutter-network/shutter/shlib/shcrypto"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/db/kprdb"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/epochid"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/testkeygen"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/shdb"
)

type TestConfig struct{}

var config = &TestConfig{}

func (c *TestConfig) GetAddress() common.Address {
	return common.HexToAddress("0x2222222222222222222222222222222222222222")
}

func (c *TestConfig) GetInstanceID() uint64 {
	return 55
}

func initializeEon(
	ctx context.Context,
	t *testing.T,
	db *kprdb.Queries,
	keyperIndex uint64, //nolint:unparam
) *testkeygen.TestKeyGenerator {
	t.Helper()
	eon := uint64(0)
	keypers := []string{
		"0x0000000000000000000000000000000000000000",
		config.GetAddress().Hex(),
		"0x1111111111111111111111111111111111111111",
	}

	tkg := testkeygen.NewTestKeyGenerator(t, 3, 2)
	publicKeyShares := []*shcrypto.EonPublicKeyShare{}
	epochID, _ := epochid.BigToEpochID(common.Big0)
	for i := uint64(0); i < tkg.NumKeypers; i++ {
		share := tkg.EonPublicKeyShare(epochID, i)
		publicKeyShares = append(publicKeyShares, share)
	}
	dkgResult := puredkg.Result{
		Eon:             eon,
		NumKeypers:      tkg.NumKeypers,
		Threshold:       tkg.Threshold,
		Keyper:          keyperIndex,
		SecretKeyShare:  tkg.EonSecretKeyShare(epochID, keyperIndex),
		PublicKey:       tkg.EonPublicKey(epochID),
		PublicKeyShares: publicKeyShares,
	}
	dkgResultEncoded, err := shdb.EncodePureDKGResult(&dkgResult)
	assert.NilError(t, err)

	err = db.InsertBatchConfig(ctx, kprdb.InsertBatchConfigParams{
		KeyperConfigIndex: 1,
		Height:            0,
		Keypers:           keypers,
		Threshold:         int32(tkg.Threshold),
	})
	assert.NilError(t, err)
	err = db.InsertEon(ctx, kprdb.InsertEonParams{
		Eon:                   0,
		Height:                0,
		ActivationBlockNumber: 0,
		KeyperConfigIndex:     1,
	})
	assert.NilError(t, err)
	err = db.InsertDKGResult(ctx, kprdb.InsertDKGResultParams{
		Eon:        0,
		Success:    true,
		Error:      sql.NullString{},
		PureResult: dkgResultEncoded,
	})
	assert.NilError(t, err)

	return tkg
}
