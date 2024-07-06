package epochkghandler

import (
	"bytes"
	"context"
	"math"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/shutter-network/shutter/shlib/puredkg"
	"github.com/shutter-network/shutter/shlib/shcrypto"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/database"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2p"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2pmsg"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/shdb"
)

func NewDecryptionKeyHandler(config Config, dbpool *pgxpool.Pool) p2p.MessageHandler {
	return &DecryptionKeyHandler{config: config, dbpool: dbpool}
}

type DecryptionKeyHandler struct {
	config Config
	dbpool *pgxpool.Pool
}

func (*DecryptionKeyHandler) MessagePrototypes() []p2pmsg.Message {
	return []p2pmsg.Message{&p2pmsg.DecryptionKeys{}}
}

func (handler *DecryptionKeyHandler) ValidateMessage(ctx context.Context, msg p2pmsg.Message) (pubsub.ValidationResult, error) {
	key := msg.(*p2pmsg.DecryptionKeys)
	if key.GetInstanceID() != handler.config.GetInstanceID() {
		return pubsub.ValidationReject,
			errors.Errorf("instance ID mismatch (want=%d, have=%d)", handler.config.GetInstanceID(), key.GetInstanceID())
	}
	if key.Eon > math.MaxInt64 {
		return pubsub.ValidationReject, errors.Errorf("eon %d overflows int64", key.Eon)
	}

	queries := database.New(handler.dbpool)

	_, isKeyper, err := queries.GetKeyperIndex(ctx, int64(key.Eon), handler.config.GetAddress())
	if err != nil {
		return pubsub.ValidationReject, err
	}
	if !isKeyper {
		log.Debug().Uint64("eon", key.Eon).Msg("Ignoring decryptionKey for eon; we're not a Keyper")
		return pubsub.ValidationReject, nil
	}

	dkgResultDB, err := queries.GetDKGResultForKeyperConfigIndex(ctx, int64(key.Eon))
	if err == pgx.ErrNoRows {
		return pubsub.ValidationReject, errors.Errorf("no DKG result found for eon %d", key.Eon)
	}
	if err != nil {
		return pubsub.ValidationReject, errors.Wrapf(err, "failed to get dkg result for eon %d from db", key.Eon)
	}
	if !dkgResultDB.Success {
		return pubsub.ValidationReject, errors.Errorf("no successful DKG result found for eon %d", key.Eon)
	}
	pureDKGResult, err := shdb.DecodePureDKGResult(dkgResultDB.PureResult)
	if err != nil {
		return pubsub.ValidationReject, errors.Wrapf(err, "error while decoding pure DKG result for eon %d", key.Eon)
	}

	if len(key.Keys) == 0 {
		return pubsub.ValidationReject, errors.New("no keys in message")
	}
	if len(key.Keys) > int(handler.config.GetMaxNumKeysPerMessage()) {
		return pubsub.ValidationReject, errors.Errorf(
			"too many keys in message (%d > %d)",
			len(key.Keys),
			handler.config.GetMaxNumKeysPerMessage(),
		)
	}

	validationResult, err := checkKeysErrors(key.Keys, pureDKGResult)
	return validationResult, err
}

func checkKeysErrors(keys []*p2pmsg.Key, pureDKGResult *puredkg.Result) (pubsub.ValidationResult, error) {
	for i, k := range keys {
		epochSecretKey, err := k.GetEpochSecretKey()
		if err != nil {
			return pubsub.ValidationReject, err
		}
		ok, err := shcrypto.VerifyEpochSecretKey(epochSecretKey, pureDKGResult.PublicKey, k.Identity)
		if err != nil {
			return pubsub.ValidationReject, errors.Wrapf(err, "error while checking epoch secret key for identity %x", k.Identity)
		}
		if !ok {
			return pubsub.ValidationReject, errors.Errorf("epoch secret key for identity %x is not valid", k.Identity)
		}

		if i > 0 && bytes.Compare(k.Identity, keys[i-1].Identity) < 0 {
			return pubsub.ValidationReject, errors.Errorf("keys not ordered")
		}
	}
	return pubsub.ValidationAccept, nil
}

func (handler *DecryptionKeyHandler) HandleMessage(ctx context.Context, msg p2pmsg.Message) ([]p2pmsg.Message, error) {
	metricsEpochKGDecryptionKeysReceived.Inc()
	key := msg.(*p2pmsg.DecryptionKeys)
	// Insert the key into the db. We assume that it's valid as it already passed the libp2p
	// validator.
	log.Debug().
		Uint64("keyper-set-index", key.Eon).
		Str("msg", msg.String()).
		Msg("handling decryption key")
	return nil, database.New(handler.dbpool).InsertDecryptionKeysMsg(ctx, key)
}
