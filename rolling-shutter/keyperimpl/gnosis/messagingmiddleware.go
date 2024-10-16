package gnosis

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/proto"

	obskeyperdatabase "github.com/shutter-network/rolling-shutter/rolling-shutter/chainobserver/db/keyper"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyperimpl/gnosis/database"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyperimpl/gnosis/gnosisssztypes"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/identitypreimage"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/retry"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/service"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2p"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2pmsg"
)

type MessagingMiddleware struct {
	config    *Config
	messaging p2p.Messaging
	dbpool    *pgxpool.Pool
}

type WrappedMessageHandler struct {
	handler    p2p.MessageHandler
	middleware *MessagingMiddleware
}

func (h *WrappedMessageHandler) MessagePrototypes() []p2pmsg.Message {
	return h.handler.MessagePrototypes()
}

func (h *WrappedMessageHandler) ValidateMessage(ctx context.Context, msg p2pmsg.Message) (pubsub.ValidationResult, error) {
	return h.handler.ValidateMessage(ctx, msg)
}

func (h *WrappedMessageHandler) HandleMessage(ctx context.Context, msg p2pmsg.Message) ([]p2pmsg.Message, error) {
	msgs, err := h.handler.HandleMessage(ctx, msg)
	if err != nil {
		return []p2pmsg.Message{}, err
	}
	replacedMsgs := []p2pmsg.Message{}
	for _, msg := range msgs {
		replacedMsg, err := h.middleware.interceptMessage(ctx, msg)
		if err != nil {
			return []p2pmsg.Message{}, err
		}
		if replacedMsg != nil {
			replacedMsgs = append(replacedMsgs, replacedMsg)
		}
	}
	return replacedMsgs, nil
}

func NewMessagingMiddleware(messaging p2p.Messaging, dbpool *pgxpool.Pool, config *Config) *MessagingMiddleware {
	return &MessagingMiddleware{messaging: messaging, dbpool: dbpool, config: config}
}

func (i *MessagingMiddleware) Start(_ context.Context, runner service.Runner) error {
	return runner.StartService(i.messaging)
}

func (i *MessagingMiddleware) interceptMessage(ctx context.Context, msg p2pmsg.Message) (p2pmsg.Message, error) {
	switch msg := msg.(type) {
	case *p2pmsg.DecryptionKeyShares:
		return i.interceptDecryptionKeyShares(ctx, msg)
	case *p2pmsg.DecryptionKeys:
		return i.interceptDecryptionKeys(ctx, msg)
	default:
		return msg, nil
	}
}

func (i *MessagingMiddleware) SendMessage(ctx context.Context, msg p2pmsg.Message, opts ...retry.Option) error {
	msgOut, err := i.interceptMessage(ctx, msg)
	if err != nil {
		return err
	}
	if msgOut != nil {
		return i.messaging.SendMessage(ctx, msgOut, opts...)
	}
	return nil
}

func (i *MessagingMiddleware) AddValidator(ctx p2p.ValidatorFunc, protos ...p2pmsg.Message) {
	i.messaging.AddValidator(ctx, protos...)
}

func (i *MessagingMiddleware) AddMessageHandler(mhs ...p2p.MessageHandler) {
	for _, mh := range mhs {
		wmh := &WrappedMessageHandler{handler: mh, middleware: i}
		i.messaging.AddMessageHandler(wmh)
	}
}

func (i *MessagingMiddleware) interceptDecryptionKeyShares(
	ctx context.Context,
	originalMsg *p2pmsg.DecryptionKeyShares,
) (p2pmsg.Message, error) {
	queries := database.New(i.dbpool)

	// We have to populate the outgoing message with slot and tx pointer information. We fetch
	// this information from the database. It should have been inserted when the decryption
	// trigger was produced. If creating the message takes unexpectedly long, it is possible that
	// it was overridden with the following trigger. In this case, we drop the message.
	currentDecryptionTrigger, err := queries.GetCurrentDecryptionTrigger(ctx, int64(originalMsg.Eon))
	if err == pgx.ErrNoRows {
		log.Warn().
			Uint64("eon", originalMsg.Eon).
			Msg("intercepted decryption key shares message with unknown corresponding decryption trigger")
		return nil, nil
	} else if err != nil {
		return nil, errors.Wrapf(err, "failed to get current decryption trigger for eon %d", originalMsg.Eon)
	}
	if originalMsg.Eon != uint64(currentDecryptionTrigger.Eon) {
		log.Warn().
			Uint64("eon-got", originalMsg.Eon).
			Int64("eon-expected", currentDecryptionTrigger.Eon).
			Msg("intercepted decryption key shares message with unexpected eon")
		return nil, nil
	}
	identitiesHash := computeIdentitiesHashFromShares(originalMsg.Shares)
	if !bytes.Equal(identitiesHash, currentDecryptionTrigger.IdentitiesHash) {
		log.Warn().
			Uint64("eon", originalMsg.Eon).
			Hex("expectedIdentitiesHash", currentDecryptionTrigger.IdentitiesHash).
			Hex("actualIdentitiesHash", identitiesHash).
			Msg("intercepted decryption key shares message with unexpected identities hash")
		return nil, nil
	}

	identityPreimages := []identitypreimage.IdentityPreimage{}
	for _, share := range originalMsg.Shares {
		identityPreimages = append(identityPreimages, identitypreimage.IdentityPreimage(share.IdentityPreimage))
	}
	slotDecryptionSignatureData, err := gnosisssztypes.NewSlotDecryptionSignatureData(
		i.config.InstanceID,
		originalMsg.Eon,
		uint64(currentDecryptionTrigger.Slot),
		uint64(currentDecryptionTrigger.TxPointer),
		identityPreimages,
	)
	if err != nil {
		return nil, err
	}
	signature, err := slotDecryptionSignatureData.ComputeSignature(i.config.Gnosis.Node.PrivateKey.Key)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to compute slot decryption signature")
	}

	err = queries.InsertSlotDecryptionSignature(ctx, database.InsertSlotDecryptionSignatureParams{
		Eon:            currentDecryptionTrigger.Eon,
		Slot:           currentDecryptionTrigger.Slot,
		KeyperIndex:    int64(originalMsg.KeyperIndex),
		TxPointer:      currentDecryptionTrigger.TxPointer,
		IdentitiesHash: identitiesHash,
		Signature:      signature,
	})
	if err != nil {
		return nil, errors.Wrapf(err,
			"failed to insert slot decryption signature for eon %d and slot %d",
			originalMsg.Eon,
			currentDecryptionTrigger.Slot,
		)
	}

	extra := &p2pmsg.GnosisDecryptionKeySharesExtra{
		Slot:      uint64(currentDecryptionTrigger.Slot),
		TxPointer: uint64(currentDecryptionTrigger.TxPointer),
		Signature: signature,
	}
	msg := proto.Clone(originalMsg).(*p2pmsg.DecryptionKeyShares)
	msg.Extra = &p2pmsg.DecryptionKeyShares_Gnosis{Gnosis: extra}
	slotStartTimestamp := medley.SlotToTimestamp(
		extra.Slot,
		i.config.Gnosis.GenesisSlotTimestamp,
		i.config.Gnosis.SecondsPerSlot,
	)
	slotStartTime := time.Unix(int64(slotStartTimestamp), 0)
	delta := time.Since(slotStartTime)
	metricsKeySharesSentTimeDelta.WithLabelValues(fmt.Sprint(originalMsg.Eon)).Observe(delta.Seconds())
	return msg, nil
}

func (i *MessagingMiddleware) interceptDecryptionKeys(
	ctx context.Context,
	originalMsg *p2pmsg.DecryptionKeys,
) (p2pmsg.Message, error) {
	if originalMsg.Extra != nil {
		err := i.advanceTxPointer(ctx, originalMsg)
		if err != nil {
			return nil, err
		}
		return originalMsg, nil
	}

	gnosisDB := database.New(i.dbpool)
	obsKeyperDB := obskeyperdatabase.New(i.dbpool)

	trigger, err := gnosisDB.GetCurrentDecryptionTrigger(ctx, int64(originalMsg.Eon))
	if err == pgx.ErrNoRows {
		log.Warn().
			Uint64("eon", originalMsg.Eon).
			Msg("unknown decryption trigger for intercepted keys message")
		return nil, nil
	}
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get current decryption trigger for eon %d", originalMsg.Eon)
	}

	keyperSet, err := obsKeyperDB.GetKeyperSetByKeyperConfigIndex(ctx, int64(originalMsg.Eon))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get keyper set from database for eon %d", originalMsg.Eon)
	}

	signaturesDB, err := gnosisDB.GetSlotDecryptionSignatures(ctx, database.GetSlotDecryptionSignaturesParams{
		Eon:            int64(originalMsg.Eon),
		Slot:           trigger.Slot,
		TxPointer:      trigger.TxPointer,
		IdentitiesHash: trigger.IdentitiesHash,
		Limit:          keyperSet.Threshold,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to count slot decryption signatures for eon %d and slot %d", originalMsg.Eon, trigger.Slot)
	}
	if len(signaturesDB) < int(keyperSet.Threshold) {
		log.Debug().
			Uint64("eon", originalMsg.Eon).
			Int64("slot", trigger.Slot).
			Int64("tx-pointer", trigger.TxPointer).
			Hex("identities-hash", trigger.IdentitiesHash).
			Int32("threshold", keyperSet.Threshold).
			Int("num-signatures", len(signaturesDB)).
			Msg("dropping intercepted keys message as signature count is not high enough yet")
		return nil, nil
	}

	signerIndices := []uint64{}
	signatures := [][]byte{}
	for _, signature := range signaturesDB {
		signerIndices = append(signerIndices, uint64(signature.KeyperIndex))
		signatures = append(signatures, signature.Signature)
	}
	msg := proto.Clone(originalMsg).(*p2pmsg.DecryptionKeys)
	extra := &p2pmsg.GnosisDecryptionKeysExtra{
		Slot:          uint64(trigger.Slot),
		TxPointer:     uint64(trigger.TxPointer),
		SignerIndices: signerIndices,
		Signatures:    signatures,
	}
	msg.Extra = &p2pmsg.DecryptionKeys_Gnosis{Gnosis: extra}
	err = i.advanceTxPointer(ctx, msg)
	if err != nil {
		return nil, err
	}

	log.Info().
		Uint64("slot", extra.Slot).
		Uint64("tx-pointer", extra.TxPointer).
		Int("num-signatures", len(signaturesDB)).
		Int("num-keys", len(msg.Keys)).
		Msg("sending keys")

	slotStartTimestamp := medley.SlotToTimestamp(
		extra.Slot,
		i.config.Gnosis.GenesisSlotTimestamp,
		i.config.Gnosis.SecondsPerSlot,
	)
	slotStartTime := time.Unix(int64(slotStartTimestamp), 0)
	delta := time.Since(slotStartTime)
	metricsKeysSentTimeDelta.WithLabelValues(fmt.Sprint(originalMsg.Eon)).Observe(delta.Seconds())
	return msg, nil
}

// advanceTxPointer updates the tx pointer in the database such that decryption will continue with
// the next transaction. Panics if the message does not have Gnosis extra data.
func (i *MessagingMiddleware) advanceTxPointer(ctx context.Context, msg *p2pmsg.DecryptionKeys) error {
	extra := msg.Extra.(*p2pmsg.DecryptionKeys_Gnosis).Gnosis

	gnosisDB := database.New(i.dbpool)
	newTxPointer := int64(extra.TxPointer) + int64(len(msg.Keys)) - 1
	log.Debug().
		Uint64("eon", msg.Eon).
		Uint64("slot", extra.Slot).
		Uint64("tx-pointer-msg", extra.TxPointer).
		Int("num-keys", len(msg.Keys)).
		Int64("tx-pointer-updated", newTxPointer).
		Msg("updating tx pointer")
	err := gnosisDB.SetTxPointer(ctx, database.SetTxPointerParams{
		Eon: int64(msg.Eon),
		Age: sql.NullInt64{
			Int64: 0,
			Valid: true,
		},
		Value: newTxPointer,
	})
	if err != nil {
		return errors.Wrap(err, "failed to set tx pointer")
	}
	eonString := fmt.Sprint(msg.Eon)
	metricsTxPointer.WithLabelValues(eonString).Set(float64(newTxPointer))
	metricsTxPointerAge.WithLabelValues(eonString).Set(0)
	return nil
}
