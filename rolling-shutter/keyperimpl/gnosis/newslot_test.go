package gnosis

import (
	"context"
	"database/sql"
	"testing"

	"gotest.tools/assert"

	gnosisDatabase "github.com/shutter-network/rolling-shutter/rolling-shutter/keyperimpl/gnosis/database"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/testsetup"
)

func TestGetTxPointerBasicIntegration(t *testing.T) {
	ctx := context.Background()
	dbpool, dbclose := testsetup.NewTestDBPool(ctx, t, gnosisDatabase.Definition)
	t.Cleanup(dbclose)
	db := gnosisDatabase.New(dbpool)

	err := db.SetTxPointer(ctx, gnosisDatabase.SetTxPointerParams{
		Eon: 2,
		Age: sql.NullInt64{
			Int64: maxTxPointerAge,
			Valid: true,
		},
		Value: 5,
	})
	assert.NilError(t, err)

	txPointer, err := getTxPointer(ctx, dbpool, 2)
	assert.NilError(t, err)
	assert.Equal(t, txPointer, int64(5))
}

func TestGetTxPointerInfiniteIntegration(t *testing.T) {
	ctx := context.Background()
	dbpool, dbclose := testsetup.NewTestDBPool(ctx, t, gnosisDatabase.Definition)
	t.Cleanup(dbclose)
	db := gnosisDatabase.New(dbpool)

	err := db.SetTxPointer(ctx, gnosisDatabase.SetTxPointerParams{
		Eon: 2,
		Age: sql.NullInt64{
			Int64: 0,
			Valid: false,
		},
		Value: 5,
	})
	assert.NilError(t, err)
	err = db.SetTransactionSubmittedEventCount(ctx, gnosisDatabase.SetTransactionSubmittedEventCountParams{
		Eon:        2,
		EventCount: 10,
	})
	assert.NilError(t, err)

	txPointer, err := getTxPointer(ctx, dbpool, 2)
	assert.NilError(t, err)
	assert.Equal(t, txPointer, int64(10))
}

func TestGetTxPointerOutdatedIntegration(t *testing.T) {
	ctx := context.Background()
	dbpool, dbclose := testsetup.NewTestDBPool(ctx, t, gnosisDatabase.Definition)
	t.Cleanup(dbclose)
	db := gnosisDatabase.New(dbpool)

	err := db.SetTxPointer(ctx, gnosisDatabase.SetTxPointerParams{
		Eon: 2,
		Age: sql.NullInt64{
			Int64: maxTxPointerAge + 1,
			Valid: false,
		},
		Value: 5,
	})
	assert.NilError(t, err)
	err = db.SetTransactionSubmittedEventCount(ctx, gnosisDatabase.SetTransactionSubmittedEventCountParams{
		Eon:        2,
		EventCount: 10,
	})
	assert.NilError(t, err)

	txPointer, err := getTxPointer(ctx, dbpool, 2)
	assert.NilError(t, err)
	assert.Equal(t, txPointer, int64(10))
}

func TestGetTxPointerMissingIntegration(t *testing.T) {
	ctx := context.Background()
	dbpool, dbclose := testsetup.NewTestDBPool(ctx, t, gnosisDatabase.Definition)
	t.Cleanup(dbclose)

	txPointer, err := getTxPointer(ctx, dbpool, 2)
	assert.NilError(t, err)
	assert.Equal(t, txPointer, int64(0))

	db := gnosisDatabase.New(dbpool)
	txPointerDB, err := db.GetTxPointer(ctx, 2)
	assert.NilError(t, err)
	assert.Equal(t, txPointerDB.Eon, int64(2))
	assert.Equal(t, txPointerDB.Age.Int64, int64(0))
	assert.Equal(t, txPointerDB.Age.Valid, true)
	assert.Equal(t, txPointerDB.Value, int64(0))
}
