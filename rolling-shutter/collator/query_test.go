package collator

// we can't move this test to cltrdb, because medley/testdb imports cltrdb resulting in an import
// cycle

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v4"
	"gotest.tools/assert"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/collator/database"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/testsetup"
)

func TestFindEonPublicKeyForBlockIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	var err error
	ctx := context.Background()
	dbpool, dbclose := testsetup.NewTestDBPool(ctx, t, database.Definition)
	t.Cleanup(dbclose)
	db := database.New(dbpool)

	err = db.InsertEonPublicKeyCandidate(ctx, database.InsertEonPublicKeyCandidateParams{
		Hash:                  []byte{1, 1},
		EonPublicKey:          []byte{1, 1, 1},
		ActivationBlockNumber: 50,
		KeyperConfigIndex:     21,
		Eon:                   3,
	})
	assert.NilError(t, err)

	err = db.InsertEonPublicKeyCandidate(ctx, database.InsertEonPublicKeyCandidateParams{
		Hash:                  []byte{2, 2},
		EonPublicKey:          []byte{2, 2, 2},
		ActivationBlockNumber: 100,
		KeyperConfigIndex:     22,
		Eon:                   4,
	})
	assert.NilError(t, err)

	err = db.InsertEonPublicKeyCandidate(ctx, database.InsertEonPublicKeyCandidateParams{
		Hash:                  []byte{3, 3},
		EonPublicKey:          []byte{3, 3, 3},
		ActivationBlockNumber: 100,
		KeyperConfigIndex:     23,
		Eon:                   3,
	})
	assert.NilError(t, err)

	// nothing is confirmed yet, so we should not find anything
	_, err = db.FindEonPublicKeyForBlock(ctx, 150)
	assert.Equal(t, err, pgx.ErrNoRows)

	err = db.ConfirmEonPublicKey(ctx, []byte{1, 1})
	assert.NilError(t, err)

	// nothing is confirmed yet, so we should not find anything
	eonPubKey, err := db.FindEonPublicKeyForBlock(ctx, 150)
	assert.NilError(t, err)
	assert.DeepEqual(t, eonPubKey, database.EonPublicKeyCandidate{
		Hash:                  []byte{1, 1},
		EonPublicKey:          []byte{1, 1, 1},
		ActivationBlockNumber: 50,
		KeyperConfigIndex:     21,
		Eon:                   3,
		Confirmed:             true,
	})

	err = db.ConfirmEonPublicKey(ctx, []byte{2, 2})
	assert.NilError(t, err)

	err = db.ConfirmEonPublicKey(ctx, []byte{3, 3})
	assert.NilError(t, err)

	// we should still find the old candidate with block number less than the activation block
	// number of the later keys
	eonPubKey, err = db.FindEonPublicKeyForBlock(ctx, 99)
	assert.NilError(t, err)
	assert.DeepEqual(t, eonPubKey, database.EonPublicKeyCandidate{
		Hash:                  []byte{1, 1},
		EonPublicKey:          []byte{1, 1, 1},
		ActivationBlockNumber: 50,
		KeyperConfigIndex:     21,
		Eon:                   3,
		Confirmed:             true,
	})

	eonPubKey, err = db.FindEonPublicKeyForBlock(ctx, 100)
	assert.NilError(t, err)
	assert.DeepEqual(t, eonPubKey, database.EonPublicKeyCandidate{
		Hash:                  []byte{3, 3},
		EonPublicKey:          []byte{3, 3, 3},
		ActivationBlockNumber: 100,
		KeyperConfigIndex:     23,
		Eon:                   3,
		Confirmed:             true,
	})
}
