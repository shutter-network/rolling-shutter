package testsetup

import (
	"context"
	"crypto/ecdsa"
	"database/sql"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/jackc/pgx/v4/pgxpool"
	"gotest.tools/assert"

	"github.com/shutter-network/shutter/shlib/puredkg"
	"github.com/shutter-network/shutter/shlib/shcrypto"

	chainobsdb "github.com/shutter-network/rolling-shutter/rolling-shutter/chainobserver/db/collator"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/database"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/db"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/identitypreimage"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/testkeygen"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/shdb"
)

type TestConfig interface {
	GetAddress() common.Address
	GetInstanceID() uint64
	GetEon() uint64
	GetCollatorKey() *ecdsa.PrivateKey
}

func InitializeEon(
	ctx context.Context,
	t *testing.T,
	dbpool *pgxpool.Pool,
	config TestConfig,
	keyperIndex uint64,
) *testkeygen.TestKeyGenerator {
	t.Helper()

	err := dbpool.BeginFunc(db.WrapContext(ctx, database.Definition.Validate))
	assert.NilError(t, err)

	keyperDB := database.New(dbpool)
	keypers := []string{
		"0x0000000000000000000000000000000000000000",
		config.GetAddress().Hex(),
		"0x1111111111111111111111111111111111111111",
	}

	collatorKey := config.GetCollatorKey()
	if collatorKey != nil {
		err := dbpool.BeginFunc(db.WrapContext(ctx, chainobsdb.Definition.Validate))
		assert.NilError(t, err)
		chdb := chainobsdb.New(dbpool)
		err = chdb.InsertChainCollator(ctx, chainobsdb.InsertChainCollatorParams{
			ActivationBlockNumber: 0,
			Collator:              shdb.EncodeAddress(ethcrypto.PubkeyToAddress(config.GetCollatorKey().PublicKey)),
		})
		assert.NilError(t, err)
	}

	tkg := testkeygen.NewTestKeyGenerator(t, 3, 2, false)
	publicKeyShares := []*shcrypto.EonPublicKeyShare{}
	identityPreimage := identitypreimage.BigToIdentityPreimage(common.Big0)
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

	err = keyperDB.InsertBatchConfig(ctx, database.InsertBatchConfigParams{
		KeyperConfigIndex: 1,
		Height:            0,
		Keypers:           keypers,
		Threshold:         int32(tkg.Threshold),
	})
	assert.NilError(t, err)
	err = keyperDB.InsertEon(ctx, database.InsertEonParams{
		Eon:                   int64(config.GetEon()),
		Height:                0,
		ActivationBlockNumber: 0,
		KeyperConfigIndex:     1,
	})
	assert.NilError(t, err)
	err = keyperDB.InsertDKGResult(ctx, database.InsertDKGResultParams{
		Eon:        int64(config.GetEon()),
		Success:    true,
		Error:      sql.NullString{},
		PureResult: dkgResultEncoded,
	})
	assert.NilError(t, err)

	return tkg
}
