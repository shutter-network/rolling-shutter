package gnosis

import (
	"bytes"
	"context"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/proto"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyperimpl/gnosis/database"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/identitypreimage"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/retry"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/service"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2p"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2pmsg"
)

type MessagingMiddleware struct {
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
		replacedMsgs = append(replacedMsgs, replacedMsg)
	}
	return replacedMsgs, nil
}

func NewMessagingMiddleware(messaging p2p.Messaging, dbpool *pgxpool.Pool) *MessagingMiddleware {
	return &MessagingMiddleware{messaging: messaging, dbpool: dbpool}
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
	msg, err := i.interceptMessage(ctx, msg)
	if err != nil {
		return err
	}
	if msg != nil {
		return i.messaging.SendMessage(ctx, msg, opts...)
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
	currentDecryptionTrigger, err := queries.GetCurrentDecryptionTrigger(ctx, int64(originalMsg.Eon))
	if err == pgx.ErrNoRows {
		log.Warn().
			Uint64("eon", originalMsg.Eon).
			Msg("intercepted decryption key shares message with unknown corresponding decryption trigger")
		return nil, nil
	} else if err != nil {
		return nil, errors.Wrapf(err, "failed to get current decryption trigger for eon %d", originalMsg.Eon)
	}
	identityPreimges := []identitypreimage.IdentityPreimage{}
	for _, share := range originalMsg.Shares {
		identityPreimges = append(identityPreimges, identitypreimage.IdentityPreimage(share.EpochID))
	}
	identitiesHash := computeIdentitiesHash(identityPreimges)
	if !bytes.Equal(identitiesHash, currentDecryptionTrigger.IdentitiesHash) {
		log.Warn().
			Uint64("eon", originalMsg.Eon).
			Hex("expectedIdentitiesHash", currentDecryptionTrigger.IdentitiesHash).
			Hex("actualIdentitiesHash", identitiesHash).
			Msg("intercepted decryption key shares message with unexpected identities hash")
		return nil, nil
	}

	msg := proto.Clone(originalMsg).(*p2pmsg.DecryptionKeyShares)
	msg.Extra = &p2pmsg.DecryptionKeyShares_Gnosis{
		Gnosis: &p2pmsg.GnosisDecryptionKeySharesExtra{
			Slot:      uint64(currentDecryptionTrigger.Block),
			TxPointer: uint64(currentDecryptionTrigger.TxPointer),
			Signature: []byte{},
		},
	}
	return msg, nil
}

func (i *MessagingMiddleware) interceptDecryptionKeys(
	_ context.Context,
	originalMsg *p2pmsg.DecryptionKeys,
) (p2pmsg.Message, error) {
	msg := proto.Clone(originalMsg).(*p2pmsg.DecryptionKeys)
	msg.Extra = &p2pmsg.DecryptionKeys_Gnosis{
		Gnosis: &p2pmsg.GnosisDecryptionKeysExtra{
			Slot:          0,
			TxPointer:     0,
			SignerIndices: []uint64{},
			Signatures:    [][]byte{},
		},
	}
	return msg, nil
}
