package collator

import (
	"context"

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
		DecryptionTrigger: trigger,
		Transactions:      transactions,
	}

	return []shmsg.P2PMessage{batch, trigger}, nil
}
