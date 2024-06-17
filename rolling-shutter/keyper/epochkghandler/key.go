package epochkghandler

import (
	"bytes"
	"context"

	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/shutter-network/shutter/shlib/shcrypto"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/database"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2p"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2pmsg"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/shdb"
)

func NewDecryptionKeyHandler(config Config, dbpool *pgxpool.Pool) p2p.MessageHandler {
	// Not catching the error as it only can happen if non-positive size was applied
	cache, _ := lru.New[shcrypto.EpochSecretKey, []byte](1024)
	return &DecryptionKeyHandler{config: config, dbpool: dbpool, cache: cache}
}

type DecryptionKeyHandler struct {
	config Config
	dbpool *pgxpool.Pool
	// keep 1024 verified keys in Cache to skip additional verifications
	cache *lru.Cache[shcrypto.EpochSecretKey, []byte]
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
	eon, err := medley.Uint64ToInt64Safe(key.Eon)
	if err != nil {
		return pubsub.ValidationReject, errors.Wrapf(err, "overflow error while converting eon to int64 %d", eon)
	}

	queries := database.New(handler.dbpool)
	_, isKeyper, err := queries.GetKeyperIndex(ctx, eon, handler.config.GetAddress())
	if err != nil {
		return pubsub.ValidationReject, err
	}
	if !isKeyper {
		log.Debug().Uint64("eon", key.Eon).Msg("Ignoring decryptionKey for eon; we're not a Keyper")
		return pubsub.ValidationReject, nil
	}
	dkgResultDB, err := queries.GetDKGResultForKeyperConfigIndex(ctx, eon)
	if errors.Is(err, pgx.ErrNoRows) {
		return pubsub.ValidationReject, errors.Errorf("no DKG result found for eon %d", eon)
	}
	if err != nil {
		return pubsub.ValidationReject, errors.Wrapf(err, "failed to get dkg result for eon %d from db", eon)
	}
	if !dkgResultDB.Success {
		return pubsub.ValidationReject, errors.Errorf("no successful DKG result found for eon %d", eon)
	}
	pureDKGResult, err := shdb.DecodePureDKGResult(dkgResultDB.PureResult)
	if err != nil {
		return pubsub.ValidationReject, errors.Wrapf(err, "error while decoding pure DKG result for eon %d", eon)
	}

	if len(key.Keys) == 0 {
		return pubsub.ValidationReject, errors.New("no keys in message")
	}
	if len(key.Keys) > int(handler.config.GetMaxNumKeysPerMessage()) {
		return pubsub.ValidationReject, errors.Errorf("too many keys in message (%d > %d)", len(key.Keys), handler.config.GetMaxNumKeysPerMessage())
	}

	for i, k := range key.Keys {
		epochSecretKey, err := k.GetEpochSecretKey()
		if err != nil {
			return pubsub.ValidationReject, err
		}
		identity, exists := handler.cache.Get(*epochSecretKey)
		if exists {
			if bytes.Equal(k.Identity, identity) {
				continue
			}
			return pubsub.ValidationReject, errors.Errorf("epoch secret key for identity %x is not valid", k.Identity)
		}
		ok, err := shcrypto.VerifyEpochSecretKey(epochSecretKey, pureDKGResult.PublicKey, k.Identity)
		if err != nil {
			return pubsub.ValidationReject, errors.Wrapf(err, "error while checking epoch secret key for identity %x", k.Identity)
		}
		if !ok {
			return pubsub.ValidationReject, errors.Errorf("epoch secret key for identity %x is not valid", k.Identity)
		}
		if i > 0 && bytes.Compare(k.Identity, key.Keys[i-1].Identity) < 0 {
			return pubsub.ValidationReject, errors.Errorf("keys not ordered")
		}
	}
	return pubsub.ValidationAccept, nil
}

func (handler *DecryptionKeyHandler) HandleMessage(ctx context.Context, msg p2pmsg.Message) ([]p2pmsg.Message, error) {
	metricsEpochKGDecryptionKeysReceived.Inc()
	key := msg.(*p2pmsg.DecryptionKeys)
	// We assume that it's valid as it already passed the libp2p validator.
	// Insert the key into the cache.
	for _, k := range key.Keys {
		epochSecretKey, err := k.GetEpochSecretKey()
		if err != nil {
			return nil, err
		}
		handler.cache.Add(*epochSecretKey, k.Identity)
	}
	// Insert the key into the db.
	return nil, database.New(handler.dbpool).InsertDecryptionKeysMsg(ctx, key)
}
