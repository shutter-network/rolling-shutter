package decryptor

import (
	"bytes"
	"context"
	"database/sql"
	"testing"

	"github.com/jackc/pgx/v4"
	"gotest.tools/v3/assert"

	"github.com/shutter-network/shutter/shuttermint/decryptor/dcrdb"
	"github.com/shutter-network/shutter/shuttermint/medley"
)

func TestGetDecryptorSet(t *testing.T) {
	ctx := context.Background()
	db, closedb := medley.NewDecryptorTestDB(ctx, t)
	defer closedb()

	addresses := []string{"address1", "address2", "address3"}
	keys := [][]byte{[]byte("key1"), []byte("key2"), []byte("key3")}
	startEpochs := [][]int{
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
		err := db.InsertDecryptorIdentity(ctx, dcrdb.InsertDecryptorIdentityParams{
			Address:      addresses[i],
			BlsPublicKey: keys[i],
		})
		assert.NilError(t, err)

		for j := 0; j < len(startEpochs[i]); j++ {
			err = db.InsertDecryptorSetMember(ctx, dcrdb.InsertDecryptorSetMemberParams{
				StartEpochID: medley.Uint64EpochIDToBytes(uint64(startEpochs[i][j])),
				Index:        int32(setIndices[i][j]),
				Address:      sql.NullString{String: addresses[i], Valid: true},
			})
			assert.NilError(t, err)
		}
	}

	rows, err := db.GetDecryptorSet(ctx, medley.Uint64EpochIDToBytes(uint64(0)))
	assert.NilError(t, err)
	assert.Check(t, len(rows) == 2)
	assert.DeepEqual(t, rows, []dcrdb.GetDecryptorSetRow{
		{
			StartEpochID: medley.Uint64EpochIDToBytes(uint64(0)),
			Index:        0,
			Address:      sql.NullString{String: addresses[0], Valid: true},
			BlsPublicKey: keys[0],
		},
		{
			StartEpochID: medley.Uint64EpochIDToBytes(uint64(0)),
			Index:        1,
			Address:      sql.NullString{String: addresses[1], Valid: true},
			BlsPublicKey: keys[1],
		},
	})

	rows, err = db.GetDecryptorSet(ctx, medley.Uint64EpochIDToBytes(uint64(100)))
	assert.NilError(t, err)
	assert.Check(t, len(rows) == 2)
	assert.DeepEqual(t, rows, []dcrdb.GetDecryptorSetRow{
		{
			StartEpochID: medley.Uint64EpochIDToBytes(uint64(100)),
			Index:        0,
			Address:      sql.NullString{String: addresses[0], Valid: true},
			BlsPublicKey: keys[0],
		},
		{
			StartEpochID: medley.Uint64EpochIDToBytes(uint64(100)),
			Index:        1,
			Address:      sql.NullString{String: addresses[2], Valid: true},
			BlsPublicKey: keys[2],
		},
	})
}

func TestEonPublicKey(t *testing.T) {
	ctx := context.Background()
	db, closedb := medley.NewDecryptorTestDB(ctx, t)
	defer closedb()

	key1 := []byte("key1")
	key2 := []byte("key2")

	err := db.InsertEonPublicKey(ctx, dcrdb.InsertEonPublicKeyParams{
		StartEpochID: medley.Uint64EpochIDToBytes(uint64(10)),
		EonPublicKey: key1,
	})
	assert.NilError(t, err)
	err = db.InsertEonPublicKey(ctx, dcrdb.InsertEonPublicKeyParams{
		StartEpochID: medley.Uint64EpochIDToBytes(uint64(20)),
		EonPublicKey: key2,
	})
	assert.NilError(t, err)

	epochIDs := []uint64{5, 9, 10, 15, 19, 20, 21, 25}
	keys := [][]byte{nil, nil, key1, key1, key1, key2, key2, key2}

	for i := 0; i < len(epochIDs); i++ {
		epochIDBytes := medley.Uint64EpochIDToBytes(epochIDs[i])
		expectedKey := keys[i]
		key, err := db.GetEonPublicKey(ctx, epochIDBytes)
		if expectedKey == nil {
			assert.Check(t, err == pgx.ErrNoRows)
		} else {
			assert.NilError(t, err)
			assert.Check(t, bytes.Equal(key, expectedKey))
		}
	}
}
