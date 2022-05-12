package collator

import (
	"context"
	"fmt"
	"log"
	"math"

	"github.com/ethereum/go-ethereum/common"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/collator/cltrdb"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/commondb"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/shmsg"
)

// ensureNoIntegerOverflowsInEonPublicKey checks that the uint64 fields declared in the
// EonPublicKey do not overflow when converted to an int64. It returns an error if there is an
// overflow.
func ensureNoIntegerOverflowsInEonPublicKey(key *shmsg.EonPublicKey) error {
	if key.Candidate.KeyperConfigIndex > math.MaxInt64 {
		return errors.New("int64 overflow for msg.KeyperConfigIndex")
	}
	if key.Candidate.ActivationBlock > math.MaxInt64 {
		return errors.New("int64 overflow for msg.ActivationBlock")
	}
	if key.KeyperIndex > math.MaxInt64 {
		return errors.New("int64 overflow for msg.KeyperIndex")
	}
	return nil
}

// ensureEonPublicKeyMatchesKeyperSet checks that the information stored in the EonPublicKey
// matches the commondb.KeyperSet stored in the database. It returns an error if there is a
// mismatch.
func ensureEonPublicKeyMatchesKeyperSet(keyperSet commondb.KeyperSet, key *shmsg.EonPublicKey) error {
	activationBlock := int64(key.Candidate.ActivationBlock)
	keyperIndex := int64(key.KeyperIndex)

	// Ensure that the information in the keyperSet matches the information stored in the EonPublicKey
	if keyperSet.ActivationBlockNumber != activationBlock {
		// Can also happen when the Keyper is dishonest.
		return errors.Errorf("eonPublicKey message's activation-block (%d) does not match the expected"+
			"activation-block on-chain (%d)", activationBlock, keyperSet.ActivationBlockNumber)
	}
	if keyperIndex >= int64(len(keyperSet.Keypers)) {
		// Can also happen when the Keyper is dishonest.
		return errors.Errorf("keyper index out of bounds for keyper set. "+
			"(activation-block=%d, keyper-index=%d)", activationBlock, keyperIndex)
	}
	expectedAddress := keyperSet.Keypers[keyperIndex]

	ok, err := key.VerifySignature(common.HexToAddress(expectedAddress))
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Validation: Error while recovering signature for EonPublicKey "+
			"(activation-block=%d, keyper-index=%d)", activationBlock, keyperIndex))
	}
	if !ok {
		return errors.Errorf("eonPublicKey's recovered address does not match on-chain address (%s) for keyper-index (%d)",
			common.HexToAddress(expectedAddress), keyperIndex)
	}
	return nil
}

// validateEonPublicKey is a libp2p validator for incoming EonPublicKey messages.
func (c *collator) validateEonPublicKey(ctx context.Context, key *shmsg.EonPublicKey) (bool, error) {
	if err := ensureNoIntegerOverflowsInEonPublicKey(key); err != nil {
		return false, err
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
			ctx, int64(key.Candidate.KeyperConfigIndex),
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
	err := c.dbpool.BeginFunc(ctx, func(tx pgx.Tx) error {
		var (
			err      error
			msgBytes []byte
		)

		activationBlock := int64(key.Candidate.ActivationBlock)

		db := cltrdb.New(tx)
		keyperSet, err := commondb.New(tx).GetKeyperSetByKeyperConfigIndex(
			ctx, int64(key.Candidate.KeyperConfigIndex),
		)
		if err != nil {
			return errors.Wrap(err, "failed to retrieve keyper set from db")
		}
		if true {
			hash := key.Candidate.Hash()
			err = db.InsertEonPublicKeyCandidate(ctx, cltrdb.InsertEonPublicKeyCandidateParams{
				Hash:                  hash,
				EonPublicKey:          key.Candidate.PublicKey,
				ActivationBlockNumber: int64(key.Candidate.ActivationBlock),
				KeyperConfigIndex:     int64(key.Candidate.KeyperConfigIndex),
				Eon:                   int64(key.Candidate.Eon),
			})
			if err != nil {
				return err
			}
			insertEonPublicKeyVoteParam := cltrdb.InsertEonPublicKeyVoteParams{
				Hash:              hash,
				Sender:            keyperSet.Keypers[key.KeyperIndex],
				Signature:         key.Signature,
				Eon:               int64(key.Candidate.Eon),
				KeyperConfigIndex: int64(key.Candidate.KeyperConfigIndex),
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
					key.Candidate.KeyperConfigIndex,
					key.Candidate.Eon,
				)
			}
		}

		err = db.InsertCandidateEonIfNotExists(ctx, cltrdb.InsertCandidateEonIfNotExistsParams{
			ActivationBlockNumber: activationBlock,
			EonPublicKey:          key.Candidate.PublicKey,
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
			EonPublicKey:          key.Candidate.PublicKey,
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
