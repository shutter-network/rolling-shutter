package collator

import (
	"context"

	"github.com/jackc/pgx/v4"

	"github.com/shutter-network/shutter/shuttermint/collator/cltrdb"
	"github.com/shutter-network/shutter/shuttermint/shdb"
	"github.com/shutter-network/shutter/shuttermint/shmsg"
)

func handleEpoch(ctx context.Context, config Config, db *cltrdb.Queries) ([]shmsg.P2PMessage, error) {
	cipherBatch, err := makeBatch(ctx, config, db)
	if err != nil {
		return nil, err
	}
	decryptionTrigger, err := makeDecryptionTrigger(ctx, config, db, cipherBatch)
	if err != nil {
		return nil, err
	}
	return []shmsg.P2PMessage{cipherBatch, decryptionTrigger}, nil
}

func makeBatch(ctx context.Context, config Config, db *cltrdb.Queries) (*shmsg.CipherBatch, error) {
	lastTrigger, err := db.GetLastTrigger(ctx)
	nextEpochID := uint64(0)
	if err == nil {
		nextEpochID = shdb.DecodeUint64(lastTrigger.EpochID) + 1
	} else if err != pgx.ErrNoRows {
		return nil, err
	}

	// TODO: fill batch with transactions

	batch := &shmsg.CipherBatch{
		InstanceID:   config.InstanceID,
		EpochID:      nextEpochID,
		Transactions: [][]byte{},
	}

	err = db.InsertBatch(ctx, cltrdb.InsertBatchParams{EpochID: shdb.EncodeUint64(batch.EpochID), Transactions: batch.Transactions})
	if err != nil {
		return nil, err
	}

	return batch, nil
}

func makeDecryptionTrigger(
	ctx context.Context, config Config, db *cltrdb.Queries, cipherBatch *shmsg.CipherBatch) (*shmsg.DecryptionTrigger, error) {

	trigger, err := shmsg.NewSignedDecryptionTrigger(config.InstanceID, cipherBatch.EpochID, cipherBatch.Transactions, config.EthereumKey)
	if err != nil {
		return nil, err
	}

	err = db.InsertTrigger(ctx, cltrdb.InsertTriggerParams{
		EpochID:   shdb.EncodeUint64(trigger.EpochID),
		BatchHash: trigger.TransactionsHash,
	})
	if err != nil {
		return nil, err
	}

	return trigger, nil
}
