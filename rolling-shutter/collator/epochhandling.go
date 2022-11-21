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

	"github.com/shutter-network/rolling-shutter/rolling-shutter/collator/cltrdb"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/epochid"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/retry"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/shmsg"
)

const dBPollTime = 500 * time.Millisecond

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

func (c *collator) listenNewDecryptionTrigger(ctx context.Context) <-chan time.Time {
	chann := make(chan time.Time, 1)

	go func() {
		defer close(chann)
		defer log.Debug().Msg("stop listening for new_decryption_trigger")

		conn, err := c.dbpool.Acquire(ctx)
		defer conn.Release()
		if err != nil {
			log.Error().Msg("error acquiring connection")
			return
		}

		_, err = conn.Exec(ctx, "listen new_decryption_trigger")
		if err != nil {
			log.Error().Msg("error listening to channel")
			return
		}

		for {
			notification, err := conn.Conn().WaitForNotification(ctx)
			select {
			case <-ctx.Done():
				return
			default:
				if err != nil {
					log.Error().Err(err).Msg("error waiting for notification")
					continue
				}
				log.Info().Str("channel", notification.Channel).Msg("database notification received")
				chann <- time.Now()
			}
		}
	}()
	return chann
}

func (c *collator) handleNewDecryptionTrigger(ctx context.Context) error {
	newTrigger := c.listenNewDecryptionTrigger(ctx)
	for {
		select {
		case <-newTrigger:
			triggers, err := c.getUnsentDecryptionTriggers(ctx, c.Config)
			if err != nil {
				return err
			}
			if len(triggers) == 0 {
				// no unsent triggers, continue and wait for next tick
				continue
			}
			for _, msg := range triggers {
				err := c.p2p.SendMessage(ctx,
					msg,
					retry.Interval(time.Second),
					retry.ExponentialBackoff(),
					retry.NumberOfRetries(3),
					retry.LogIdentifier(msg.LogInfo()),
					retry.LogCaptureStackFrameContext(),
				)
				if err != nil {
					continue // continue sending other messages
				}
				err = c.dbpool.BeginFunc(ctx, func(dbtx pgx.Tx) error {
					db := cltrdb.New(dbtx)
					return db.UpdateDecryptionTriggerSent(ctx, msg.EpochID)
				})
				if err != nil {
					return err
				}
				log.Info().
					Str("msg", msg.LogInfo()).
					Str("commitment", hexutil.Encode(msg.TransactionsHash)).
					Msg("sent decryption trigger")
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
