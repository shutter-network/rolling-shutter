package batchhandler

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/collator/database"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/identitypreimage"
)

// ComputeNextEpochID takes an epoch id as parameter and returns the id of the epoch following it.
func ComputeNextEpochID(identityPreimage identitypreimage.IdentityPreimage) identitypreimage.IdentityPreimage {
	n := identityPreimage.Big()
	return identitypreimage.BigToIdentityPreimage(n.Add(n, common.Big1))
}

// GetNextBatch gets the epochID and block number that will be used in the next batch.
func GetNextBatch(ctx context.Context, db *database.Queries) (identitypreimage.IdentityPreimage, uint64, error) {
	b, err := db.GetNextBatch(ctx)
	if err != nil {
		// There should already be an epochID in the database so not finding a row is an error
		return identitypreimage.IdentityPreimage{}, 0, err
	}
	identityPreimage := identitypreimage.IdentityPreimage(b.EpochID)
	if b.L1BlockNumber < 0 {
		return identitypreimage.IdentityPreimage{}, 0, errors.Errorf("negative l1 block number in db")
	}
	l1BlockNumber := uint64(b.L1BlockNumber)
	return identityPreimage, l1BlockNumber, nil
}
