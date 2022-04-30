package collator

import (
	"context"
	"fmt"
	"math"

	"github.com/ethereum/go-ethereum/common"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/collator/cltrdb"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/commondb"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/shmsg"
)

func (c *collator) getKeyperSet(ctx context.Context, activationBlock int64) (commondb.KeyperSet, error) {
	var keyperSet commondb.KeyperSet

	err := c.dbpool.BeginFunc(ctx, func(tx pgx.Tx) error {
		var err error
		db := commondb.New(tx)

		// only using activationBlock as identifier could be ambiguous (see issue #238)
		keyperSet, err = db.GetKeyperSet(ctx, activationBlock)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return keyperSet, errors.Wrap(err, "failed to retrieve keyper set from db")
	}
	return keyperSet, nil
}

func (c *collator) validateEonPublicKey(ctx context.Context, key *shmsg.EonPublicKey) (bool, error) {
	var (
		ok  bool
		ks  commondb.KeyperSet
		err error
	)

	if key.ActivationBlock > math.MaxInt64 {
		return false, errors.New("int64 overflow for msg.ActivationBlock")
	}
	activationBlock := int64(key.ActivationBlock)
	if key.KeyperIndex > math.MaxInt64 {
		return false, errors.New("int64 overflow for msg.KeyperIndex")
	}
	keyperIndex := int64(key.KeyperIndex)
	// Theoretically, there could be a race condition, where we learn of the EonPublicKey
	// broadcast message before we notice the new keyper-config, and would ignore it
	// because of that.
	// In practice however, this won't play a role since the DKG of the keypers takes
	// place later in wall-time

	// If here we inserted two different keyper sets for the same
	// activation-block, then this is ambiguous (there is no sorting on DB-insert order or similar) (see #238)
	ks, err = c.getKeyperSet(ctx, activationBlock)
	if err != nil {
		return false, errors.Wrap(err, fmt.Sprintf("error while retrieving on-chain Keyper set for "+
			"EonPublicKey (activation-block=%d)", activationBlock))
	}

	if ks.ActivationBlockNumber != activationBlock {
		// This is the case when either there were ambiguities in
		// retrieving the correct keyper set for this activationBlock (see issue #238)
		// Can also happen when the Keyper is dishonest.

		return false, errors.Errorf("eonPublicKey message's activation-block (%d) does not match the expected"+
			"activation-block on-chain (%d)", activationBlock, ks.ActivationBlockNumber)
	}
	if int64(len(ks.Keypers)) < keyperIndex+1 {
		// Using the wrong keyper set (e.g. due to #238) will result in out-of-bounds error,
		// Can also happen when the Keyper is dishonest.
		return false, errors.Wrapf(err, "keyper index out of bounds for keyper set. "+
			"(activation-block=%d, keyper-index=%d)", activationBlock, keyperIndex)
	}
	// This could be susceptible to replay attacks, when the keyper index in an older keyper-config maps to the same
	// address as in this keyper-config(see issue #238)
	expectedAddress := ks.Keypers[keyperIndex]

	ok, err = key.VerifySignature(common.HexToAddress(expectedAddress))
	if err != nil {
		return false, errors.Wrap(err, fmt.Sprintf("Validation: Error while recovering signature for EonPublicKey "+
			"(activation-block=%d, keyper-index=%d)", activationBlock, keyperIndex))
	}
	if !ok {
		return false, errors.Errorf("eonPublicKey's recovered address does not match on-chain address (%s) for keyper-index (%d)",
			common.HexToAddress(expectedAddress), keyperIndex)
	}
	return true, nil
}

func (c *collator) handleEonPublicKey(ctx context.Context, key *shmsg.EonPublicKey) ([]shmsg.P2PMessage, error) {
	activationBlock := int64(key.ActivationBlock)
	keyperSet, err := c.getKeyperSet(ctx, activationBlock)
	if err != nil {
		return make([]shmsg.P2PMessage, 0), err
	}
	err = c.dbpool.BeginFunc(ctx, func(tx pgx.Tx) error {
		var (
			err      error
			msgBytes []byte
		)
		db := cltrdb.New(tx)

		err = db.InsertCandidateEonIfNotExists(ctx, cltrdb.InsertCandidateEonIfNotExistsParams{
			ActivationBlockNumber: activationBlock,
			EonPublicKey:          key.PublicKey,
			Threshold:             int64(keyperSet.Threshold),
		})
		if err != nil {
			return err
		}

		// inefficient to marshal again after the message has just been umarshaled and
		// passed to the handler function
		// for optimisation: also pass the raw msg bytes to the handler functions and ignore
		// the argument if not needed
		msgBytes, err = proto.Marshal(key)
		if err != nil {
			return err
		}
		err = db.InsertEonPublicKeyMessage(ctx, cltrdb.InsertEonPublicKeyMessageParams{
			EonPublicKey:          key.PublicKey,
			ActivationBlockNumber: activationBlock,
			KeyperIndex:           int64(key.KeyperIndex),
			MsgBytes:              msgBytes,
		})
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return make([]shmsg.P2PMessage, 0), err
	}
	return make([]shmsg.P2PMessage, 0), nil
}
