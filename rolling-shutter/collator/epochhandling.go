package collator

import (
	"context"
	"log"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"

	"github.com/shutter-network/shutter/shlib/shcrypto"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/collator/cltrdb"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/epochid"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/shdb"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/shmsg"
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
	ctx context.Context, cfg config.Config, db *cltrdb.Queries, currentBlockNumber uint32,
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
		cfg.InstanceID, epochID, shmsg.HashTransactions(transactions), cfg.EthereumKey,
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

	nextEpochID := computeNextEpochID(epochID, currentBlockNumber, cfg.ExecutionBlockDelay)
	err = db.SetNextEpochID(ctx, shdb.EncodeUint64(nextEpochID))
	if err != nil {
		return nil, err
	}

	return []shmsg.P2PMessage{trigger}, nil
}

func (c *collator) handleDecryptionKey(ctx context.Context, msg *shmsg.DecryptionKey) ([]shmsg.P2PMessage, error) {
	err := c.dbpool.BeginFunc(ctx, func(tx pgx.Tx) error {
		db := cltrdb.New(tx)
		_, err := db.InsertDecryptionKey(ctx, cltrdb.InsertDecryptionKeyParams{
			EpochID:       shdb.EncodeUint64(msg.EpochID),
			DecryptionKey: msg.Key,
		})
		if err != nil {
			return errors.Wrapf(err, "failed to insert decryption key for epoch %s", epochid.LogInfo(msg.EpochID))
		}
		return nil
	})
	if err != nil {
		return make([]shmsg.P2PMessage, 0), errors.Wrapf(err, "error while inserting decryption key for epoch %s", epochid.LogInfo(msg.EpochID))
	}
	log.Printf(
		"inserted decryption key for epoch %s to database",
		epochid.LogInfo(msg.EpochID),
	)
	return make([]shmsg.P2PMessage, 0), nil
}

func (c *collator) validateDecryptionKey(ctx context.Context, key *shmsg.DecryptionKey) (bool, error) {
	var eonPublicKey shcrypto.EonPublicKey
	if key.GetInstanceID() != c.Config.InstanceID {
		return false, errors.Errorf("instance ID mismatch (want=%d, have=%d)", c.Config.InstanceID, key.GetInstanceID())
	}

	err := c.dbpool.BeginFunc(ctx, func(tx pgx.Tx) error {
		db := cltrdb.New(tx)
		msgActivationBlock := int64(epochid.BlockNumber(key.EpochID))
		eonPub, err := db.FindEonPublicKeyForBlock(ctx, msgActivationBlock)
		if err != nil {
			return errors.Wrap(err, "failed to retrieve EonPublicKey from DB")
		}

		err = eonPublicKey.GobDecode(eonPub.EonPublicKey)
		if err != nil {
			return errors.Wrap(err, "failed to decode persisted EonPublicKey")
		}
		return nil
	})
	if err != nil {
		return false, err
	}
	epochSecretKey, err := key.GetEpochSecretKey()
	if err != nil {
		return false, err
	}

	ok, err := shcrypto.VerifyEpochSecretKey(epochSecretKey, &eonPublicKey, key.EpochID)
	if err != nil {
		return false, err
	}
	if !ok {
		return false, errors.Errorf("recovery of epoch secret key failed for epoch %v", key.EpochID)
	}
	return true, nil
}
