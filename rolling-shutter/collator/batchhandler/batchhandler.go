package batchhandler

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/collator/cltrdb"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/epochid"
)

// ComputeNextEpochID takes an epoch id as parameter and returns the id of the epoch following it.
func ComputeNextEpochID(epochID epochid.EpochID) (epochid.EpochID, error) {
	n := epochID.Big()
	return epochid.BigToEpochID(n.Add(n, common.Big1))
}

// GetNextBatch gets the epochID and block number that will be used in the next batch.
func GetNextBatch(ctx context.Context, db *cltrdb.Queries) (epochid.EpochID, uint64, error) {
	b, err := db.GetNextBatch(ctx)
	if err != nil {
		// There should already be an epochID in the database so not finding a row is an error
		return epochid.EpochID{}, 0, err
	}
	epochID, err := epochid.BytesToEpochID(b.EpochID)
	if err != nil {
		return epochid.EpochID{}, 0, err
	}
	if b.L1BlockNumber < 0 {
		return epochid.EpochID{}, 0, errors.Errorf("negative l1 block number in db")
	}
	l1BlockNumber := uint64(b.L1BlockNumber)
	return epochID, l1BlockNumber, nil
}
