package shutterservice

import (
	"bytes"
	"context"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/proto"

	obskeyperdatabase "github.com/shutter-network/rolling-shutter/rolling-shutter/chainobserver/db/keyper"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyperimpl/shutterservice/database"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyperimpl/shutterservice/serviceztypes"
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

func (i *MessagingMiddleware) interceptDecryptionKeyShares(
	ctx context.Context,
	originalMsg *p2pmsg.DecryptionKeyShares,
) (p2pmsg.Message, error) {
	queries := database.New(i.dbpool)

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

	decryptionSignatureData, err := serviceztypes.NewDecryptionSignatureData(
		i.config.InstanceID,
		originalMsg.Eon,
		identityPreimages,
	)
	if err != nil {
		return nil, err
	}
	signature, err := decryptionSignatureData.ComputeSignature(i.config.Chain.Node.PrivateKey.Key)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to compute decryption signature")
	}

	err = queries.InsertDecryptionSignature(ctx, database.InsertDecryptionSignatureParams{
		Eon:            int64(originalMsg.Eon),
		KeyperIndex:    int64(originalMsg.KeyperIndex),
		IdentitiesHash: identitiesHash,
		Signature:      signature,
	})
	if err != nil {
		return nil, errors.Wrapf(err,
			"failed to insert decryption signature for eon %d and keyperIndex %d",
			originalMsg.Eon,
			originalMsg.KeyperIndex,
		)
	}

	extra := &p2pmsg.ShutterServiceDecryptionKeySharesExtra{
		Signature: signature,
	}
	msg := proto.Clone(originalMsg).(*p2pmsg.DecryptionKeyShares)
	msg.Extra = &p2pmsg.DecryptionKeyShares_Service{Service: extra}
	return msg, nil
}

func (i *MessagingMiddleware) interceptDecryptionKeys(
	ctx context.Context,
	originalMsg *p2pmsg.DecryptionKeys,
) (p2pmsg.Message, error) {
	// TODO: update flag in event table to notify the decryption is already done
	if originalMsg.Extra != nil {
		return originalMsg, nil
	}

	serviceDB := database.New(i.dbpool)
	obsKeyperDB := obskeyperdatabase.New(i.dbpool)
	trigger, err := serviceDB.GetCurrentDecryptionTrigger(ctx, int64(originalMsg.Eon))
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

	signatures, err := serviceDB.GetDecryptionSignatures(ctx, database.GetDecryptionSignaturesParams{
		Eon:            int64(originalMsg.Eon),
		IdentitiesHash: trigger.IdentitiesHash,
		Limit:          keyperSet.Threshold,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to count decryption signatures for eon %d and keyperConfigIndex %d", originalMsg.Eon, keyperSet.KeyperConfigIndex)
	}

	if len(signatures) < int(keyperSet.Threshold) {
		log.Debug().
			Uint64("eon", originalMsg.Eon).
			Hex("identities-hash", trigger.IdentitiesHash).
			Int32("threshold", keyperSet.Threshold).
			Int("num-signatures", len(signatures)).
			Msg("dropping intercepted keys message as signature count is not high enough yet")
		return nil, nil
	}

	signerIndices := []uint64{}
	signaturesCum := [][]byte{}
	for _, signature := range signatures {
		signerIndices = append(signerIndices, uint64(signature.KeyperIndex))
		signaturesCum = append(signaturesCum, signature.Signature)
	}
	msg := proto.Clone(originalMsg).(*p2pmsg.DecryptionKeys)
	extra := &p2pmsg.ShutterServiceDecryptionKeysExtra{
		SignerIndices: signerIndices,
		Signature:     signaturesCum,
	}
	msg.Extra = &p2pmsg.DecryptionKeys_Service{Service: extra}

	err = updateEventFlag(ctx, serviceDB, originalMsg)
	if err != nil {
		log.Warn().
			Err(err).
			Msg("failed to update events for decryption keys released")
	}

	log.Info().
		Uint64("eon", originalMsg.Eon).
		Hex("identities-hash", trigger.IdentitiesHash).
		Int("num-signatures", len(signatures)).
		Int("num-keys", len(msg.Keys)).
		Msg("sending keys")
	return msg, nil
}

func updateEventFlag(ctx context.Context, serviceDB *database.Queries, keys *p2pmsg.DecryptionKeys) error {
	column1 := make([]int64, 0)
	column2 := make([][]byte, 0)
	for _, key := range keys.Keys {
		column1 = append(column1, int64(keys.Eon))
		column2 = append(column2, key.IdentityPreimage)
	}

	err := serviceDB.UpdateDecryptedFlag(ctx, database.UpdateDecryptedFlagParams{
		Column1: column1,
		Column2: column2,
	})
	if err != nil {
		return errors.Wrap(err, "failed to update decrypted flag")
	}
	return nil
}
