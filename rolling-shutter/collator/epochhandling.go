package collator

import (
	"context"

	"github.com/jackc/pgx/v4"

	"github.com/shutter-network/shutter/shuttermint/collator/cltrdb"
	"github.com/shutter-network/shutter/shuttermint/shdb"
	"github.com/shutter-network/shutter/shuttermint/shmsg"
)

func startNextEpoch(ctx context.Context, config Config, db *cltrdb.Queries) ([]shmsg.P2PMessage, error) {
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

	batch := &shmsg.CipherBatch{
		InstanceID:   config.InstanceID,
		EpochID:      epochID,
		Transactions: transactions,
	}

	return []shmsg.P2PMessage{batch, trigger}, nil
}

func getNextEpochID(ctx context.Context, db *cltrdb.Queries) (uint64, error) {
	lastEpochID, err := db.GetLastBatchEpochID(ctx)
	if err == pgx.ErrNoRows {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return shdb.DecodeUint64(lastEpochID) + 1, nil
}
