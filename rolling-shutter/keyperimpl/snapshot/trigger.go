package snapshot

import (
	"context"
	"math"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	chainobscolldb "github.com/shutter-network/rolling-shutter/rolling-shutter/chainobserver/db/collator"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/epochkghandler"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/broker"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/identitypreimage"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2p"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2pmsg"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/shdb"
)

func NewDecryptionTriggerHandler(
	config Config,
	dbpool *pgxpool.Pool,
	trigger chan<- *broker.Event[*epochkghandler.DecryptionTrigger],
) p2p.MessageHandler {
	return &DecryptionTriggerHandler{
		trigger: trigger,
		config:  config,
		dbpool:  dbpool,
	}
}

type DecryptionTriggerHandler struct {
	trigger chan<- *broker.Event[*epochkghandler.DecryptionTrigger]
	config  Config
	dbpool  *pgxpool.Pool
}

func (*DecryptionTriggerHandler) MessagePrototypes() []p2pmsg.Message {
	return []p2pmsg.Message{&p2pmsg.DecryptionTrigger{}}
}

func (handler *DecryptionTriggerHandler) ValidateMessage(ctx context.Context, msg p2pmsg.Message) (pubsub.ValidationResult, error) {
	trigger := msg.(*p2pmsg.DecryptionTrigger)
	if trigger.GetInstanceId() != handler.config.InstanceID {
		return pubsub.ValidationReject,
			errors.Errorf("instance ID mismatch (want=%d, have=%d)", handler.config.InstanceID, trigger.GetInstanceId())
	}

	blk := trigger.BlockNumber
	if blk > math.MaxInt64 {
		return pubsub.ValidationReject, errors.Errorf("block number %d overflows int64", blk)
	}
	chainCollator, err := chainobscolldb.New(handler.dbpool).GetChainCollator(ctx, int64(blk))
	if err == pgx.ErrNoRows {
		return pubsub.ValidationReject, errors.Errorf("got decryption trigger with no collator for given block number: %d", blk)
	}
	if err != nil {
		return pubsub.ValidationReject, errors.Wrapf(err, "error while getting collator from db for block number: %d", blk)
	}

	collator, err := shdb.DecodeAddress(chainCollator.Collator)
	if err != nil {
		return pubsub.ValidationReject, errors.Wrapf(err, "error while converting collator from string to address: %s", chainCollator.Collator)
	}

	signatureValid, err := p2pmsg.VerifySignature(trigger, collator)
	if err != nil {
		return pubsub.ValidationReject, errors.Wrapf(err, "error while verifying decryption trigger signature for epoch: %x", trigger.EpochId)
	}
	if !signatureValid {
		return pubsub.ValidationReject, errors.Errorf("decryption trigger signature invalid for epoch: %x", trigger.EpochId)
	}
	return pubsub.ValidationAccept, nil
}

func (handler *DecryptionTriggerHandler) HandleMessage(ctx context.Context, m p2pmsg.Message) ([]p2pmsg.Message, error) {
	msg, ok := m.(*p2pmsg.DecryptionTrigger)
	if !ok {
		return nil, errors.New("Message type assertion mismatch")
	}
	log.Info().Str("message", msg.LogInfo()).Msg("received decryption trigger")
	identityPreimage := identitypreimage.IdentityPreimage(msg.EpochId)

	trig := &epochkghandler.DecryptionTrigger{
		BlockNumber:       msg.BlockNumber,
		IdentityPreimages: []identitypreimage.IdentityPreimage{identityPreimage},
	}

	select {
	case handler.trigger <- broker.NewEvent(trig):
		return nil, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}
