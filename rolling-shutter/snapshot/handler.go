package snapshot

import (
	"context"

	"github.com/jackc/pgx/v4/pgxpool"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/shutter-network/shutter/shlib/shcrypto"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2p"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2pmsg"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/snapshot/database"
)

const MaxNumKeysPerMessage = 128

func NewDecryptionKeyHandler(config *Config, snapshot *Snapshot) p2p.MessageHandler {
	return &DecryptionKeyHandler{config: config, snapshot: snapshot, dbpool: snapshot.dbpool}
}

type DecryptionKeyHandler struct {
	config   *Config
	snapshot *Snapshot
	dbpool   *pgxpool.Pool
}

func NewEonPublicKeyHandler(config *Config, snapshot *Snapshot) p2p.MessageHandler {
	return &EonPublicKeyHandler{config: config, snapshot: snapshot, dbpool: snapshot.dbpool}
}

type EonPublicKeyHandler struct {
	config   *Config
	snapshot *Snapshot
	dbpool   *pgxpool.Pool
}

func NewDecryptionTriggerHandler() p2p.MessageHandler {
	return &DecryptionTriggerHandler{}
}

type DecryptionTriggerHandler struct{}

func (*DecryptionKeyHandler) MessagePrototypes() []p2pmsg.Message {
	return []p2pmsg.Message{&p2pmsg.DecryptionKeys{}}
}

func (*EonPublicKeyHandler) MessagePrototypes() []p2pmsg.Message {
	return []p2pmsg.Message{&p2pmsg.EonPublicKey{}}
}

func (d *DecryptionTriggerHandler) MessagePrototypes() []p2pmsg.Message {
	return []p2pmsg.Message{&p2pmsg.DecryptionTrigger{}}
}

func (handler *DecryptionKeyHandler) ValidateMessage(ctx context.Context, msg p2pmsg.Message) (pubsub.ValidationResult, error) {
	var eonPublicKey shcrypto.EonPublicKey

	keys := msg.(*p2pmsg.DecryptionKeys)
	// FIXME: check snapshot business logic for decryptionKeyMsg validation
	if keys.GetInstanceId() != handler.config.InstanceID {
		return pubsub.ValidationReject,
			errors.Errorf("instance ID mismatch (want=%d, have=%d)", handler.config.InstanceID, keys.GetInstanceId())
	}

	if len(keys.Keys) == 0 {
		return pubsub.ValidationReject, errors.Errorf("no keys in message")
	}
	if len(keys.Keys) > MaxNumKeysPerMessage {
		return pubsub.ValidationReject, errors.Errorf("too many keys in message (%d > %d)", len(keys.Keys), MaxNumKeysPerMessage)
	}

	eonID, err := medley.Uint64ToInt64Safe(keys.GetEon())
	if err != nil {
		return pubsub.ValidationReject, errors.Wrap(err, "can't cast eon to int64")
	}

	eon, err := handler.snapshot.db.GetEonPublicKey(ctx, eonID)
	if err != nil {
		return pubsub.ValidationReject, errors.Wrap(err, "failed to retrieve eon for decryption key")
	}

	err = eonPublicKey.GobDecode(eon)
	if err != nil {
		return pubsub.ValidationReject, errors.Wrap(err, "failed to retrieve eon for decryption key")
	}

	for _, key := range keys.Keys {
		k, err := key.GetEpochSecretKey()
		if err != nil {
			return pubsub.ValidationReject, errors.Wrapf(err, "error getting epochSecretKey for identity: %d", key.IdentityPreimage)
		}
		ok, err := shcrypto.VerifyEpochSecretKey(k, &eonPublicKey, key.IdentityPreimage)
		if err != nil {
			return pubsub.ValidationReject, err
		}
		if !ok {
			return pubsub.ValidationReject, errors.Errorf("recovery of epoch secret key failed for identity %s", key.IdentityPreimage)
		}
	}

	return pubsub.ValidationAccept, nil
}

func (handler *EonPublicKeyHandler) ValidateMessage(_ context.Context, msg p2pmsg.Message) (pubsub.ValidationResult, error) {
	eonKeyMsg := msg.(*p2pmsg.EonPublicKey)
	if eonKeyMsg.GetInstanceId() != handler.config.InstanceID {
		return pubsub.ValidationReject,
			errors.Errorf("instance ID mismatch (want=%d, have=%d)", handler.config.InstanceID, eonKeyMsg.GetInstanceId())
	}
	eon := eonKeyMsg.GetEon()
	if eon == 0 {
		return pubsub.ValidationReject, errors.Errorf("failed to get eon public key from P2P message")
	}
	return pubsub.ValidationAccept, nil
}

func (handler *DecryptionKeyHandler) HandleMessage(ctx context.Context, m p2pmsg.Message) ([]p2pmsg.Message, error) {
	var result []p2pmsg.Message
	keys := m.(*p2pmsg.DecryptionKeys)
	db := database.New(handler.dbpool)

	newKeys := []*p2pmsg.Key{}
	for _, key := range keys.Keys {
		rows, err := db.InsertDecryptionKey(
			ctx, database.InsertDecryptionKeyParams{
				EpochID: key.IdentityPreimage,
				Key:     key.Key,
			},
		)
		if err != nil {
			return result, err
		}
		// not yet seen
		if rows != 0 {
			newKeys = append(newKeys, key)
			metricKeysGenerated.Inc()
		}
	}

	for _, key := range newKeys {
		log.Printf("Sending key %X for proposal %X to hub", key.Key, key.IdentityPreimage)
		err := handler.snapshot.hubapi.SubmitProposalKey(key.IdentityPreimage, key.Key)
		if err != nil {
			return result, err
		}
	}

	return result, nil
}

func (handler *EonPublicKeyHandler) HandleMessage(ctx context.Context, m p2pmsg.Message) ([]p2pmsg.Message, error) {
	eonPubKeyMsg := m.(*p2pmsg.EonPublicKey)

	eonID := eonPubKeyMsg.GetEon()
	key := eonPubKeyMsg.GetPublicKey()
	db := database.New(handler.dbpool)
	rows, err := db.InsertEonPublicKey(
		ctx, database.InsertEonPublicKeyParams{
			EonID:        int64(eonID),
			EonPublicKey: key,
		},
	)
	if err != nil {
		return nil, err
	}
	// we have already seen the eon
	if rows == 0 {
		return nil, nil
	}

	metricEons.Inc()

	log.Printf("Sending Eon %d public key to hub", eonID)
	err = handler.snapshot.hubapi.SubmitEonKey(eonID, key)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (d *DecryptionTriggerHandler) ValidateMessage(_ context.Context, _ p2pmsg.Message) (pubsub.ValidationResult, error) {
	log.Printf("Validating decryptionTrigger")
	return pubsub.ValidationAccept, nil
}

func (d *DecryptionTriggerHandler) HandleMessage(_ context.Context, _ p2pmsg.Message) ([]p2pmsg.Message, error) {
	log.Printf("Ignoring decryptionTrigger")
	return nil, nil
}
