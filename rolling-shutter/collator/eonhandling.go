package collator

import (
	"context"
	"fmt"
	"log"
	"math"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/collator/cltrdb"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/commondb"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/kprdb"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/shdb"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/shmsg"
)

// ensureNoIntegerOverflowsInEonPublicKey checks that the uint64 fields declared in the
// EonPublicKey do not overflow when converted to an int64. It returns an error if there is an
// overflow.
func ensureNoIntegerOverflowsInEonPublicKey(key *shmsg.EonPublicKey) error {
	if key.KeyperConfigIndex > math.MaxInt64 {
		return errors.New("int64 overflow for msg.KeyperConfigIndex")
	}
	if key.ActivationBlock > math.MaxInt64 {
		return errors.New("int64 overflow for msg.ActivationBlock")
	}
	return nil
}

// ensureEonPublicKeyMatchesKeyperSet checks that the information stored in the EonPublicKey
// matches the commondb.KeyperSet stored in the database. It returns an error if there is a
// mismatch.
func ensureEonPublicKeyMatchesKeyperSet(keyperSet commondb.KeyperSet, key *shmsg.EonPublicKey) error {
	activationBlock := int64(key.ActivationBlock)

	// Ensure that the information in the keyperSet matches the information stored in the EonPublicKey
	if keyperSet.ActivationBlockNumber != activationBlock {
		// Can also happen when the Keyper is dishonest.
		return errors.Errorf("eonPublicKey message's activation-block (%d) does not match the expected"+
			"activation-block on-chain (%d)", activationBlock, keyperSet.ActivationBlockNumber)
	}

	recoveredAddress, err := shmsg.RecoverAddress(key)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Validation: Error while recovering signature for EonPublicKey "+
			"(activation-block=%d)", activationBlock))
	}
	_, ok := kprdb.GetKeyperIndex(recoveredAddress, keyperSet.Keypers)

	if !ok {
		return errors.Errorf(
			"eonPublicKey's recovered address %s not found in on-chain addresses",
			recoveredAddress.Hex(),
		)
	}
	return nil
}

// validateEonPublicKey is a libp2p validator for incoming EonPublicKey messages.
func (c *collator) validateEonPublicKey(ctx context.Context, key *shmsg.EonPublicKey) (bool, error) {
	if err := ensureNoIntegerOverflowsInEonPublicKey(key); err != nil {
		return false, err
	}

	if c.Config.InstanceID != key.GetInstanceID() {
		return false, errors.Errorf("eonPublicKey has wrong InstanceID (expected=%d, have=%d)",
			c.Config.InstanceID, key.GetInstanceID())
	}

	// Theoretically, there could be a race condition, where we learn of the EonPublicKey
	// broadcast message before we notice the new keyper-config, and would ignore it
	// because of that.
	// In practice however, this won't play a role since the DKG of the keypers takes
	// place later in wall-time
	var keyperSet commondb.KeyperSet
	if err := c.dbpool.BeginFunc(ctx, func(tx pgx.Tx) error {
		var err error
		keyperSet, err = commondb.New(tx).GetKeyperSetByKeyperConfigIndex(
			ctx, int64(key.KeyperConfigIndex),
		)
		return err
	}); err != nil {
		return false, errors.Wrap(err, "failed to retrieve keyper set from db")
	}

	if err := ensureEonPublicKeyMatchesKeyperSet(keyperSet, key); err != nil {
		return false, err
	}
	return true, nil
}

func (c *collator) handleEonPublicKey(ctx context.Context, key *shmsg.EonPublicKey) ([]shmsg.P2PMessage, error) {
	recoveredAddress, err := shmsg.RecoverAddress(key)
	if err != nil {
		return make([]shmsg.P2PMessage, 0), err
	}
	err = c.dbpool.BeginFunc(ctx, func(tx pgx.Tx) error {
		var err error

		db := cltrdb.New(tx)
		keyperSet, err := commondb.New(tx).GetKeyperSetByKeyperConfigIndex(
			ctx, int64(key.KeyperConfigIndex),
		)
		if err != nil {
			return errors.Wrap(err, "failed to retrieve keyper set from db")
		}
		hash := key.Hash()
		err = db.InsertEonPublicKeyCandidate(ctx, cltrdb.InsertEonPublicKeyCandidateParams{
			Hash:                  hash,
			EonPublicKey:          key.PublicKey,
			ActivationBlockNumber: int64(key.ActivationBlock),
			KeyperConfigIndex:     int64(key.KeyperConfigIndex),
			Eon:                   int64(key.Eon),
		})
		if err != nil {
			return err
		}
		insertEonPublicKeyVoteParam := cltrdb.InsertEonPublicKeyVoteParams{
			Hash:              hash,
			Sender:            shdb.EncodeAddress(recoveredAddress),
			Signature:         key.Signature,
			Eon:               int64(key.Eon),
			KeyperConfigIndex: int64(key.KeyperConfigIndex),
		}
		err = db.InsertEonPublicKeyVote(ctx, insertEonPublicKeyVoteParam)
		if err != nil {
			return err
		}
		count, err := db.CountEonPublicKeyVotes(ctx, hash)
		if err != nil {
			return err
		}
		if count == int64(keyperSet.Threshold) {
			err = db.ConfirmEonPublicKey(ctx, hash)
			if err != nil {
				return err
			}
			log.Printf("Confirmed eon public key for keyper config index=%d, eon=%d",
				key.KeyperConfigIndex,
				key.Eon,
			)
		}
		return nil
	})
	if err != nil {
		return make([]shmsg.P2PMessage, 0), err
	}
	return make([]shmsg.P2PMessage, 0), nil
}
