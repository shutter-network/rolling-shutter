package epochkghandler

import (
	"context"
	"math"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/db/chainobsdb"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/db/kprdb"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/epochid"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2p"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2pmsg"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/shdb"
)

func NewDecryptionTriggerHandler(config Config, dbpool *pgxpool.Pool) p2p.MessageHandler {
	return &DecryptionTriggerHandler{config: config, dbpool: dbpool}
}

type DecryptionTriggerHandler struct {
	config Config
	dbpool *pgxpool.Pool
}

func (*DecryptionTriggerHandler) MessagePrototypes() []p2pmsg.Message {
	return []p2pmsg.Message{&p2pmsg.DecryptionTrigger{}}
}

func (handler *DecryptionTriggerHandler) ValidateMessage(ctx context.Context, msg p2pmsg.Message) (bool, error) {
	trigger := msg.(*p2pmsg.DecryptionTrigger)
	if trigger.GetInstanceID() != handler.config.GetInstanceID() {
		return false, errors.Errorf("instance ID mismatch (want=%d, have=%d)", handler.config.GetInstanceID(), trigger.GetInstanceID())
	}
	if _, err := epochid.BytesToEpochID(trigger.EpochID); err != nil {
		return false, errors.Wrapf(err, "invalid epoch id")
	}

	blk := trigger.BlockNumber
	if blk > math.MaxInt64 {
		return false, errors.Errorf("block number %d overflows int64", blk)
	}
	chainCollator, err := chainobsdb.New(handler.dbpool).GetChainCollator(ctx, int64(blk))
	if err == pgx.ErrNoRows {
		return false, errors.Errorf("got decryption trigger with no collator for given block number: %d", blk)
	}
	if err != nil {
		return false, errors.Wrapf(err, "error while getting collator from db for block number: %d", blk)
	}

	collator, err := shdb.DecodeAddress(chainCollator.Collator)
	if err != nil {
		return false, errors.Wrapf(err, "error while converting collator from string to address: %s", chainCollator.Collator)
	}

	signatureValid, err := p2pmsg.VerifySignature(trigger, collator)
	if err != nil {
		return false, errors.Wrapf(err, "error while verifying decryption trigger signature for epoch: %d", trigger.EpochID)
	}
	return signatureValid, nil
}

func (handler *DecryptionTriggerHandler) HandleMessage(ctx context.Context, m p2pmsg.Message) ([]p2pmsg.Message, error) {
	msg, ok := m.(*p2pmsg.DecryptionTrigger)
	if !ok {
		return nil, errors.New("Message type assertion mismatch")
	}
	log.Info().Str("message", msg.String()).Msg("received decryption trigger")
	epochID, err := epochid.BytesToEpochID(msg.EpochID)
	if err != nil {
		return nil, err
	}
	return SendDecryptionKeyShare(ctx, handler.config, kprdb.New(handler.dbpool), epochID, int64(msg.BlockNumber))
}
