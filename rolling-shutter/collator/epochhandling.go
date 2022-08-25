package collator

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common"

	"github.com/shutter-network/shutter/shuttermint/collator/cltrdb"
	"github.com/shutter-network/shutter/shuttermint/medley/epochid"
	"github.com/shutter-network/shutter/shuttermint/shmsg"
)

// computeNextEpochID takes an epoch id as parameter and returns the id of the epoch following it.
func computeNextEpochID(epochID epochid.EpochID) (epochid.EpochID, error) {
	n := epochID.Big()
	nextN := new(big.Int).Add(n, common.Big1)
	return epochid.BigToEpochID(nextN)
}

func startNextEpoch(
	ctx context.Context, config Config, db *cltrdb.Queries, currentBlockNumber uint32,
) ([]shmsg.P2PMessage, error) {
	epochID, blockNumber, err := getNextEpochID(ctx, db)
	if err != nil {
		return nil, err
	}

	transactions, err := db.GetTransactionsByEpoch(ctx, epochID.Bytes())
	if err != nil {
		return nil, err
	}

	trigger, err := shmsg.NewSignedDecryptionTrigger(
		config.InstanceID, epochID, uint64(currentBlockNumber), transactions, config.EthereumKey,
	)
	if err != nil {
		return nil, err
	}

	// Write back the generated trigger to the database
	err = db.InsertTrigger(ctx, cltrdb.InsertTriggerParams{
		EpochID:   trigger.EpochID,
		BatchHash: trigger.TransactionsHash,
	})
	if err != nil {
		return nil, err
	}

	nextEpochID, err := computeNextEpochID(epochID)
	if err != nil {
		return nil, err
	}
	err = db.SetNextEpochID(ctx, cltrdb.SetNextEpochIDParams{
		EpochID:     nextEpochID.Bytes(),
		BlockNumber: int64(blockNumber),
	})
	if err != nil {
		return nil, err
	}
	batch := &shmsg.CipherBatch{
		DecryptionTrigger: trigger,
		Transactions:      transactions,
	}

	return []shmsg.P2PMessage{batch, trigger}, nil
}
