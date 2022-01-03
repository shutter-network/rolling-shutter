package decryptor

import (
	"bytes"
	"context"
	"testing"

	"github.com/jackc/pgx/v4"
	"gotest.tools/v3/assert"

	"github.com/shutter-network/shutter/shuttermint/decryptor/dcrdb"
	"github.com/shutter-network/shutter/shuttermint/medley/testdb"
)

func TestGetDecryptorSetIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	ctx := context.Background()
	db, closedb := testdb.NewDecryptorTestDB(ctx, t)
	defer closedb()

	addresses := []string{"address1", "address2", "address3"}
	keys := [][]byte{[]byte("key1"), []byte("key2"), []byte("key3")}
	activationBlockNumbers := [][]int64{
		{0, 100},
		{0},
		{100},
	}
	setIndices := [][]int{
		{0, 0},
		{1},
		{1},
	}
	for i := 0; i < len(addresses); i++ {
		valid := i%2 == 0
		var key []byte
		if valid {
			key = keys[i]
		} else {
			key = []byte{}
		}
		err := db.InsertDecryptorIdentity(ctx, dcrdb.InsertDecryptorIdentityParams{
			Address:        addresses[i],
			BlsPublicKey:   key,
			BlsSignature:   []byte{},
			SignatureValid: valid,
		})
		assert.NilError(t, err)

		for j := 0; j < len(activationBlockNumbers[i]); j++ {
			err = db.InsertDecryptorSetMember(ctx, dcrdb.InsertDecryptorSetMemberParams{
				ActivationBlockNumber: activationBlockNumbers[i][j],
				Index:                 int32(setIndices[i][j]),
				Address:               addresses[i],
			})
			assert.NilError(t, err)
		}
	}

	rows, err := db.GetDecryptorSet(ctx, 0)
	assert.NilError(t, err)
	assert.Check(t, len(rows) == 2)
	assert.DeepEqual(t, rows, []dcrdb.GetDecryptorSetRow{
		{
			ActivationBlockNumber: 0,
			Index:                 0,
			Address:               addresses[0],
			BlsPublicKey:          keys[0],
			BlsSignature:          []byte{},
			SignatureValid:        true,
		},
		{
			ActivationBlockNumber: 0,
			Index:                 1,
			Address:               addresses[1],
			BlsPublicKey:          []byte{},
			BlsSignature:          []byte{},
			SignatureValid:        false,
		},
	})

	rows, err = db.GetDecryptorSet(ctx, 100)
	assert.NilError(t, err)
	assert.Check(t, len(rows) == 2)
	assert.DeepEqual(t, rows, []dcrdb.GetDecryptorSetRow{
		{
			ActivationBlockNumber: 100,
			Index:                 0,
			Address:               addresses[0],
			BlsPublicKey:          keys[0],
			BlsSignature:          []byte{},
			SignatureValid:        true,
		},
		{
			ActivationBlockNumber: 100,
			Index:                 1,
			Address:               addresses[2],
			BlsPublicKey:          keys[2],
			BlsSignature:          []byte{},
			SignatureValid:        true,
		},
	})
}

func TestEonPublicKeyIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	ctx := context.Background()
	db, closedb := testdb.NewDecryptorTestDB(ctx, t)
	defer closedb()

	key1 := []byte("key1")
	key2 := []byte("key2")

	err := db.InsertEonPublicKey(ctx, dcrdb.InsertEonPublicKeyParams{
		ActivationBlockNumber: 10,
		EonPublicKey:          key1,
	})
	assert.NilError(t, err)
	err = db.InsertEonPublicKey(ctx, dcrdb.InsertEonPublicKeyParams{
		ActivationBlockNumber: 20,
		EonPublicKey:          key2,
	})
	assert.NilError(t, err)

	blockNumbers := []int64{5, 9, 10, 11, 19, 20, 21, 25}
	keys := [][]byte{nil, nil, key1, key1, key1, key2, key2, key2}

	for i := 0; i < len(blockNumbers); i++ {
		expectedKey := keys[i]
		key, err := db.GetEonPublicKey(ctx, blockNumbers[i])
		if expectedKey == nil {
			assert.Check(t, err == pgx.ErrNoRows)
		} else {
			assert.NilError(t, err)
			assert.Check(t, bytes.Equal(key, expectedKey))
		}
	}
}
