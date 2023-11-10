package epochkghandler

import (
	"context"
	"crypto/ecdsa"
	"database/sql"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"
	"gotest.tools/assert"

	"github.com/shutter-network/shutter/shlib/puredkg"
	"github.com/shutter-network/shutter/shlib/shcrypto"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/db/chainobsdb"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/db/kprdb"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/identitypreimage"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/testkeygen"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/shdb"
)

func init() {
	var err error
	config.collatorKey, err = ethcrypto.GenerateKey()
	if err != nil {
		panic(errors.Wrap(err, "ethcrypto.GenerateKey failed"))
	}
}

type TestConfig struct {
	collatorKey *ecdsa.PrivateKey
}

var config = &TestConfig{}

func (TestConfig) GetAddress() common.Address {
	return common.HexToAddress("0x2222222222222222222222222222222222222222")
}

func (TestConfig) GetInstanceID() uint64 {
	return 55
}

func (TestConfig) GetEon() uint64 {
	return 22
}

func (c *TestConfig) GetCollatorKey() *ecdsa.PrivateKey {
	return config.collatorKey
}

func initializeEon(
	ctx context.Context,
	t *testing.T,
	dbpool *pgxpool.Pool,
	keyperIndex uint64, //nolint:unparam
) *testkeygen.TestKeyGenerator {
	t.Helper()
	db := kprdb.New(dbpool)
	keypers := []string{
		"0x0000000000000000000000000000000000000000",
		config.GetAddress().Hex(),
		"0x1111111111111111111111111111111111111111",
	}

	chdb := chainobsdb.New(dbpool)
	err := chdb.InsertChainCollator(ctx, chainobsdb.InsertChainCollatorParams{
		ActivationBlockNumber: 0,
		Collator:              shdb.EncodeAddress(ethcrypto.PubkeyToAddress(config.GetCollatorKey().PublicKey)),
	})
	assert.NilError(t, err)

	tkg := testkeygen.NewTestKeyGenerator(t, 3, 2, false)
	publicKeyShares := []*shcrypto.EonPublicKeyShare{}
	identityPreimage, _ := identitypreimage.BigToIdentityPreimage(common.Big0)
	for i := uint64(0); i < tkg.NumKeypers; i++ {
		share := tkg.EonPublicKeyShare(identityPreimage, i)
		publicKeyShares = append(publicKeyShares, share)
	}
	dkgResult := puredkg.Result{
		Eon:             config.GetEon(),
		NumKeypers:      tkg.NumKeypers,
		Threshold:       tkg.Threshold,
		Keyper:          keyperIndex,
		SecretKeyShare:  tkg.EonSecretKeyShare(identityPreimage, keyperIndex),
		PublicKey:       tkg.EonPublicKey(identityPreimage),
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
		Eon:                   int64(config.GetEon()),
		Height:                0,
		ActivationBlockNumber: 0,
		KeyperConfigIndex:     1,
	})
	assert.NilError(t, err)
	err = db.InsertDKGResult(ctx, kprdb.InsertDKGResultParams{
		Eon:        int64(config.GetEon()),
		Success:    true,
		Error:      sql.NullString{},
		PureResult: dkgResultEncoded,
	})
	assert.NilError(t, err)

	return tkg
}
