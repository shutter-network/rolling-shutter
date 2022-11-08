package collator

import (
	"context"
	"log"
	"math"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"

	"github.com/shutter-network/shutter/shlib/shcrypto"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/collator/cltrdb"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/epochid"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/shmsg"
)

func (c *collator) handleDecryptionKey(ctx context.Context, msg *shmsg.DecryptionKey) ([]shmsg.P2PMessage, error) {
	epochID, err := epochid.BytesToEpochID(msg.EpochID)
	if err != nil {
		return nil, err
	}

	err = c.dbpool.BeginFunc(ctx, func(tx pgx.Tx) error {
		db := cltrdb.New(tx)
		_, err := db.InsertDecryptionKey(ctx, cltrdb.InsertDecryptionKeyParams{
			EpochID:       epochID.Bytes(),
			DecryptionKey: msg.Key,
		})
		return err
	})
	if err != nil {
		return nil, errors.Wrapf(err, "error while inserting decryption key for epoch %s", epochID)
	}
	log.Printf("inserted decryption key for epoch %s to database", epochID)
	return []shmsg.P2PMessage{}, nil
}

func (c *collator) validateDecryptionKey(ctx context.Context, key *shmsg.DecryptionKey) (bool, error) {
	var eonPublicKey shcrypto.EonPublicKey
	if key.GetInstanceID() != c.Config.InstanceID {
		return false, errors.Errorf("instance ID mismatch (want=%d, have=%d)", c.Config.InstanceID, key.GetInstanceID())
	}
	if key.Eon > math.MaxInt64 {
		return false, errors.Errorf("eon %d overflows int64", key.Eon)
	}
	epochID, err := epochid.BytesToEpochID(key.EpochID)
	if err != nil {
		return false, errors.Wrapf(err, "invalid epoch id")
	}

	err = c.dbpool.BeginFunc(ctx, func(tx pgx.Tx) error {
		db := cltrdb.New(tx)
		eonPub, err := db.GetEonPublicKey(ctx, int64(key.Eon))
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

	ok, err := shcrypto.VerifyEpochSecretKey(epochSecretKey, &eonPublicKey, epochID.Bytes())
	if err != nil {
		return false, err
	}
	if !ok {
		return false, errors.Errorf("recovery of epoch secret key failed for epoch %s", epochID)
	}
	return true, nil
}
