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

	// The one-time receival of the decryption key is the only event
	// that triggers the submitting of the batch-tx currently.
	// This call is not guaranteed to succeed and could fail e.g. due to networking issues.
	msgs, err := c.batchHandler.HandleDecryptionKey(ctx, msg.EpochID, msg.Key)
	if err != nil {
		// NOTE: If this fails then the batch will never be re-submitted to the sequencer
		// because we don't memorize the key and try to re-submit it later.
		return make([]shmsg.P2PMessage, 0), errors.Wrapf(err, "error while processing the batch (epoch %s)", epochid.LogInfo(msg.EpochID))
	}
	return msgs, nil
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
