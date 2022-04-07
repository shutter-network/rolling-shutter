package collator

import (
	"context"

	"github.com/shutter-network/shutter/shuttermint/collator/cltrdb"
	"github.com/shutter-network/shutter/shuttermint/medley/epochid"
	"github.com/shutter-network/shutter/shuttermint/shdb"
	"github.com/shutter-network/shutter/shuttermint/shmsg"
)

// computeNextEpochID takes an epoch id as parameter and returns the id of the epoch following it.
// The function also depends on the current mainchain block number and the configured execution
// block delay. The result will encode a block number and a sequence number. The sequence number
// will be the sequence number of the previous epoch id plus one. The block number will be
// max(current block number - execution block delay, block number encoded in previous epoch id, 0).
func computeNextEpochID(epochID uint64, currentBlockNumber uint32, executionBlockDelay uint32) uint64 {
	executionBlockNumber := uint32(0)
	if currentBlockNumber >= executionBlockDelay {
		executionBlockNumber = currentBlockNumber - executionBlockDelay
	}

	previousExecutionBlockNumber := epochid.BlockNumber(epochID)
	if executionBlockNumber < previousExecutionBlockNumber {
		executionBlockNumber = previousExecutionBlockNumber
	}

	sequenceNumber := epochid.SequenceNumber(epochID)
	return epochid.New(sequenceNumber+1, executionBlockNumber)
}

func startNextEpoch(
	ctx context.Context, config Config, db *cltrdb.Queries, currentBlockNumber uint32,
) ([]shmsg.P2PMessage, error) {
	epochID, err := getNextEpochID(ctx, db)
	if err != nil {
		return nil, err
	}

	transactions, err := db.GetTransactionsByEpoch(ctx, shdb.EncodeUint64(epochID))
	if err != nil {
		return nil, err
	}

	trigger, err := shmsg.NewSignedDecryptionTrigger(
		config.InstanceID, epochID, transactions, config.EthereumKey,
	)
	if err != nil {
		return nil, err
	}

	// Write back the generated trigger to the database
	err = db.InsertTrigger(ctx, cltrdb.InsertTriggerParams{
		EpochID:   shdb.EncodeUint64(trigger.EpochID),
		BatchHash: trigger.TransactionsHash,
	})
	if err != nil {
		return nil, err
	}

	nextEpochID := computeNextEpochID(epochID, currentBlockNumber, config.ExecutionBlockDelay)
	err = db.SetNextEpochID(ctx, shdb.EncodeUint64(nextEpochID))
	if err != nil {
		return nil, err
	}

	return []shmsg.P2PMessage{trigger}, nil
}
