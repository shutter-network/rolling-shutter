package collator

import (
	"context"
	"math"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
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
	MaxNumKeysPerMessage               = 128
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
	return []p2pmsg.Message{&p2pmsg.DecryptionKeys{}}
}

func (handler *decryptionKeyHandler) HandleMessage(
	ctx context.Context,
	m p2pmsg.Message,
) ([]p2pmsg.Message, error) {
	msg := m.(*p2pmsg.DecryptionKeys)

	err := handler.dbpool.BeginFunc(ctx, func(tx pgx.Tx) error {
		db := database.New(tx)
		for _, key := range msg.Keys {
			identityPreimage := identitypreimage.IdentityPreimage(key.Identity)
			_, err := db.InsertDecryptionKey(ctx, database.InsertDecryptionKeyParams{
				EpochID:       identityPreimage.Bytes(),
				DecryptionKey: key.Key,
			})
			if err != nil {
				return errors.Wrapf(err, "error while inserting decryption key for epoch %s", identityPreimage)
			}
			log.Info().Str("epoch-id", identityPreimage.Hex()).Msg("inserted decryption key to database")
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return []p2pmsg.Message{}, nil
}

func (handler *decryptionKeyHandler) ValidateMessage(
	ctx context.Context,
	k p2pmsg.Message,
) (pubsub.ValidationResult, error) {
	keys := k.(*p2pmsg.DecryptionKeys)

	if keys.GetInstanceID() != handler.Config.InstanceID {
		return pubsub.ValidationReject, errors.Errorf(
			"instance ID mismatch (want=%d, have=%d)",
			handler.Config.InstanceID,
			keys.GetInstanceID(),
		)
	}
	if keys.Eon > math.MaxInt64 {
		return pubsub.ValidationReject, errors.Errorf("eon %d overflows int64", keys.Eon)
	}

	var eonPublicKey shcrypto.EonPublicKey
	err := handler.dbpool.BeginFunc(ctx, func(tx pgx.Tx) error {
		db := database.New(tx)
		eonPub, err := db.GetEonPublicKey(ctx, int64(keys.Eon))
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
		return pubsub.ValidationReject, err
	}

	if len(keys.Keys) == 0 {
		return pubsub.ValidationReject, errors.Errorf("no keys in message")
	}
	if len(keys.Keys) > MaxNumKeysPerMessage {
		return pubsub.ValidationReject, errors.Errorf("too many keys in message (%d > %d)", len(keys.Keys), MaxNumKeysPerMessage)
	}
	for _, key := range keys.Keys {
		identityPreimage := identitypreimage.IdentityPreimage(key.Identity)
		epochSecretKey, err := key.GetEpochSecretKey()
		if err != nil {
			return pubsub.ValidationReject, err
		}

		ok, err := shcrypto.VerifyEpochSecretKey(epochSecretKey, &eonPublicKey, identityPreimage.Bytes())
		if err != nil {
			return pubsub.ValidationReject, err
		}
		if !ok {
			return pubsub.ValidationReject, errors.Errorf("recovery of epoch secret key failed for epoch %s", identityPreimage)
		}
	}
	return pubsub.ValidationAccept, nil
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
			retry.ExponentialBackoff(nil),
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
