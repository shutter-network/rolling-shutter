package collator

import (
	"context"
	"math"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/shutter-network/shutter/shlib/shcrypto"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/collator/batcher"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/db/cltrdb"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/epochid"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/retry"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/shmsg"
)

const (
	minRetryPollInterval time.Duration = 50 * time.Millisecond
	newDecryptionTrigger               = "new_decryption_trigger"
	newDecryptionKey                   = "new_decryption_key"
	newBatchtx                         = "new_batchtx"
)

var dbListenChannels []string = []string{
	newDecryptionTrigger,
	newDecryptionKey,
	newBatchtx,
}

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
	log.Info().Str("epoch-id", epochID.Hex()).Msg("inserted decryption key to database")
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

func (c *collator) sendDecryptionTriggers(ctx context.Context) error {
	triggers, err := c.getUnsentDecryptionTriggers(ctx, c.Config)
	if err != nil {
		return err
	}
	for _, msg := range triggers {
		err := c.p2p.SendMessage(ctx,
			msg,
			retry.Interval(time.Second),
			retry.ExponentialBackoff(),
			retry.NumberOfRetries(3),
			retry.LogIdentifier(msg.LogInfo()),
		)
		if err != nil {
			continue // continue sending other messages
		}
		err = c.dbpool.BeginFunc(ctx, func(dbtx pgx.Tx) error {
			return cltrdb.New(dbtx).UpdateDecryptionTriggerSent(ctx, msg.EpochID)
		})
		if err != nil {
			return err
		}
		log.Info().
			Str("msg", msg.LogInfo()).
			Str("commitment", hexutil.Encode(msg.TransactionsHash)).
			Msg("sent decryption trigger")
	}
	return nil
}

// closeBatchesTicker constantly tries to close the current batch after `interval` duration.
// Every time the `interval` has passed, closeBatchesTicker will first try to close the batch
// until successful.
// Then it will wait some time and try to initialize the chain state for the next batch
// in order to validate queued up transactions early on in the batch life-cycle.
func (c *collator) closeBatchesTicker(ctx context.Context, interval time.Duration) error {
	t := time.NewTicker(interval)
	assumedBatchProcessingDuration := time.Second
	retryPollInterval := interval / 5
	if minRetryPollInterval > retryPollInterval {
		retryPollInterval = minRetryPollInterval
	}
	for {
		select {
		case <-t.C:
			fnCloseBatch := func(ctx context.Context) (struct{}, error) {
				return struct{}{}, c.batcher.CloseBatch(ctx)
			}
			// retry indefinitely until successful or context canceled
			_, err := retry.FunctionCall(ctx,
				fnCloseBatch,
				retry.Interval(retryPollInterval),
				retry.NumberOfRetries(-1),
			)
			if err != nil {
				if ctx.Err() != nil {
					return err
				}
				continue
			}

			// give the keypers and sequencer some time, this is not strictly necessary
			time.Sleep(assumedBatchProcessingDuration)

			// try to validate already queued transactions as early as possible
			fnEnsureChainState := func(ctx context.Context) (struct{}, error) {
				return struct{}{}, c.batcher.EnsureChainState(ctx)
			}

			// retry indefinitely until successful, context canceled or
			// a severe error occurs
			_, err = retry.FunctionCall(ctx, fnEnsureChainState,
				retry.Interval(retryPollInterval),
				retry.NumberOfRetries(-1),
				retry.StopOnErrors(batcher.ErrBatchAlreadyExists),
			)
			if err == batcher.ErrBatchAlreadyExists {
				// something is seriously wrong
				return err
			}

		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
