package keyper

import (
	"context"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/database"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/kprconfig"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/service"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2p"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2pmsg"
)

var eonPubkeyTickerTime = 2 * time.Second

type EonPublicKey struct {
	PublicKey         []byte
	ActivationBlock   uint64
	KeyperConfigIndex uint64
	Eon               uint64
}

func newEonPubKeyHandler(keyper *KeyperCore) *eonPubKeyHandler {
	return &eonPubKeyHandler{
		dbpool:             keyper.dbpool,
		config:             keyper.config,
		messaging:          keyper.messaging,
		eonPubkeyHandler:   keyper.opts.eonPubkeyHandler,
		broadcastEonPubKey: keyper.opts.broadcastEonPubKey,
		stopOnErrors:       false,
	}
}

type eonPubKeyHandler struct {
	dbpool    *pgxpool.Pool
	config    *kprconfig.Config
	messaging p2p.Messaging

	eonPubkeyHandler   EonPublicKeyHandlerFunc
	broadcastEonPubKey bool

	stopOnErrors bool
}

func (pkh *eonPubKeyHandler) Start(ctx context.Context, runner service.Runner) error {
	runner.Go(func() error {
		return pkh.loop(ctx)
	})
	return nil
}

func (pkh *eonPubKeyHandler) loop(ctx context.Context) error {
	t := time.NewTicker(eonPubkeyTickerTime)
	for {
		err := pkh.queryAndHandleNewEonPubKeys(ctx)
		if err != nil {
			if pkh.stopOnErrors {
				return err
			}
			log.Error().Err(err).Msg("error during handling of new eon public keys")
		}
		select {
		case <-ctx.Done():
			t.Stop()
			return ctx.Err()
		case <-t.C:
		}
	}
}

func (pkh *eonPubKeyHandler) broadcastEonPublicKey(ctx context.Context, eonPubKey EonPublicKey) error {
	msg, err := p2pmsg.NewSignedEonPublicKey(
		pkh.config.InstanceID,
		eonPubKey.PublicKey,
		eonPubKey.ActivationBlock,
		eonPubKey.KeyperConfigIndex,
		eonPubKey.Eon,
		pkh.config.Ethereum.PrivateKey.Key,
	)
	if err != nil {
		return errors.Wrap(err, "error while signing EonPublicKey")
	}

	err = pkh.messaging.SendMessage(ctx, msg)
	if err != nil {
		return errors.Wrap(err, "error while broadcasting EonPublicKey")
	}
	return nil
}

func (pkh *eonPubKeyHandler) queryAndHandleNewEonPubKeys(ctx context.Context) error {
	eonPublicKeys, err := database.New(pkh.dbpool).GetAndDeleteEonPublicKeys(ctx)
	if err != nil {
		return err
	}
	for _, eonPublicKey := range eonPublicKeys {
		_, exists := database.GetKeyperIndex(pkh.config.GetAddress(), eonPublicKey.Keypers)
		if !exists {
			return errors.Errorf("own keyper index not found for Eon=%d", eonPublicKey.Eon)
		}
		activationBlock, err := medley.Int64ToUint64Safe(eonPublicKey.ActivationBlockNumber)
		if err != nil {
			return errors.Wrap(err, "failed safe int cast")
		}
		keyperIndex, err := medley.Int32ToUint64Safe(eonPublicKey.KeyperConfigIndex)
		if err != nil {
			return errors.Wrap(err, "failed safe int cast")
		}
		eon, err := medley.Int64ToUint64Safe(eonPublicKey.Eon)
		if err != nil {
			return errors.Wrap(err, "failed safe int cast")
		}
		eonPubKey := EonPublicKey{
			PublicKey:         eonPublicKey.EonPublicKey,
			ActivationBlock:   activationBlock,
			KeyperConfigIndex: keyperIndex,
			Eon:               eon,
		}
		if pkh.broadcastEonPubKey {
			err := pkh.broadcastEonPublicKey(ctx, eonPubKey)
			return errors.Wrap(err, "failed to broadcast eon public key")
		}
		if pkh.eonPubkeyHandler != nil {
			err := pkh.eonPubkeyHandler(ctx, eonPubKey)
			return errors.Wrap(err, "failed to handle eon public key")
		}
	}
	return nil
}
