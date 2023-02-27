package collator

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v4"
	"gotest.tools/assert"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/db/cltrdb"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/epochid"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/testdb"
)

// TestDBSubmissionIntegration tests the basic database query functions that we've implemented.
func TestDBSubmissionIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	db, _, closedb := testdb.NewCollatorTestDB(ctx, t)
	defer closedb()

	// Trying to get the unsubmitted batch should give us an error
	_, err := db.GetUnsubmittedBatchTx(ctx)
	assert.Equal(t, err, pgx.ErrNoRows)

	// Insert a batchtx
	epoch := epochid.Uint64ToEpochID(1).Bytes()
	err = db.InsertBatchTx(ctx, cltrdb.InsertBatchTxParams{
		EpochID:   epoch,
		Marshaled: []byte{1, 2, 3},
	})
	assert.NilError(t, err)

	// We should now have a unsubmitted batchtx
	unsubmitted, err := db.GetUnsubmittedBatchTx(ctx)
	assert.NilError(t, err)
	assert.DeepEqual(t, unsubmitted, cltrdb.Batchtx{
		EpochID:   epoch,
		Marshaled: []byte{1, 2, 3},
	})

	epoch2 := epochid.Uint64ToEpochID(2).Bytes()
	// We should not be able to add a second batchtx
	err = db.InsertBatchTx(ctx, cltrdb.InsertBatchTxParams{
		EpochID:   epoch2,
		Marshaled: []byte{1, 2, 3, 4},
	})
	assert.ErrorContains(t, err, "duplicate key")

	err = db.SetBatchSubmitted(ctx)
	assert.NilError(t, err)

	// Trying to get the unsubmitted batch should give us an error again
	_, err = db.GetUnsubmittedBatchTx(ctx)
	assert.Equal(t, err, pgx.ErrNoRows)

	// We should now be able to add a second batchtx
	err = db.InsertBatchTx(ctx, cltrdb.InsertBatchTxParams{
		EpochID:   epoch2,
		Marshaled: []byte{1, 2, 3, 4},
	})
	assert.NilError(t, err)

	// And we should have a new unsubmitted batchtx
	unsubmitted, err = db.GetUnsubmittedBatchTx(ctx)
	assert.NilError(t, err)
	assert.DeepEqual(t, unsubmitted, cltrdb.Batchtx{
		EpochID:   epoch2,
		Marshaled: []byte{1, 2, 3, 4},
	})
}
