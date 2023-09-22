package collator

import (
	"context"
	"math"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/shutter-network/shutter/shlib/shcrypto"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/collator/batcher"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/collator/config"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/collator/database"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/identitypreimage"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/retry"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2pmsg"
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

type decryptionKeyHandler struct {
	Config *config.Config
	dbpool *pgxpool.Pool
}

func (*decryptionKeyHandler) MessagePrototypes() []p2pmsg.Message {
	return []p2pmsg.Message{&p2pmsg.DecryptionKey{}}
}

func (handler *decryptionKeyHandler) HandleMessage(
	ctx context.Context,
	m p2pmsg.Message,
) ([]p2pmsg.Message, error) {
	msg := m.(*p2pmsg.DecryptionKey)
	identityPreimage := identitypreimage.IdentityPreimage(msg.EpochID)

	err := handler.dbpool.BeginFunc(ctx, func(tx pgx.Tx) error {
		db := database.New(tx)
		_, err := db.InsertDecryptionKey(ctx, database.InsertDecryptionKeyParams{
			EpochID:       identityPreimage.Bytes(),
			DecryptionKey: msg.Key,
		})
		return err
	})
	if err != nil {
		return nil, errors.Wrapf(err, "error while inserting decryption key for epoch %s", identityPreimage)
	}
	log.Info().Str("epoch-id", identityPreimage.Hex()).Msg("inserted decryption key to database")
	return []p2pmsg.Message{}, nil
}

func (handler *decryptionKeyHandler) ValidateMessage(
	ctx context.Context,
	k p2pmsg.Message,
) (bool, error) {
	key := k.(*p2pmsg.DecryptionKey)

	var eonPublicKey shcrypto.EonPublicKey
	if key.GetInstanceID() != handler.Config.InstanceID {
		return false, errors.Errorf(
			"instance ID mismatch (want=%d, have=%d)",
			handler.Config.InstanceID,
			key.GetInstanceID(),
		)
	}
	if key.Eon > math.MaxInt64 {
		return false, errors.Errorf("eon %d overflows int64", key.Eon)
	}
	identityPreimage := identitypreimage.IdentityPreimage(key.EpochID)

	err := handler.dbpool.BeginFunc(ctx, func(tx pgx.Tx) error {
		db := database.New(tx)
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

	ok, err := shcrypto.VerifyEpochSecretKey(epochSecretKey, &eonPublicKey, identityPreimage.Bytes())
	if err != nil {
		return false, err
	}
	if !ok {
		return false, errors.Errorf("recovery of epoch secret key failed for epoch %s", identityPreimage)
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
			return database.New(dbtx).UpdateDecryptionTriggerSent(ctx, msg.EpochID)
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
