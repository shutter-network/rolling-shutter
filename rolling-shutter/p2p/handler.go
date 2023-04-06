package p2p

import (
	"context"
	"reflect"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/retry"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/service"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2pmsg"
)

type (
	HandlerFuncStatic[M p2pmsg.Message] func(context.Context, M) ([]p2pmsg.Message, error)
	HandlerFunc                         func(context.Context, p2pmsg.Message) ([]p2pmsg.Message, error)
	HandlerRegistry                     map[protoreflect.FullName]HandlerFunc
	ValidatorFunc[M p2pmsg.Message]     func(context.Context, M) (bool, error)
	ValidatorRegistry                   map[string]pubsub.ValidatorEx

	ValidatorOption func(*validator) error
	validator       struct {
		allowTraceContext bool
		invalidResultType pubsub.ValidationResult
	}
)

func defaults() ValidatorOption {
	return func(v *validator) error {
		v.allowTraceContext = false
		v.invalidResultType = pubsub.ValidationReject
		return nil
	}
}

// WithTraceContextPropagation option allows for
// cross-cutting trace-context to be propagated and accepted.
// If this is not set, if a peer sends a trace-context within the
// message envelope, the message will be rejected and the
// sender will be punished (as per local peer score).
func WithTraceContextPropagation() ValidatorOption {
	return func(v *validator) error {
		v.allowTraceContext = true
		return nil
	}
}

// WithGracefulIgnore option allows to "upgrade" all messages
// that would have been rejected to a ignore only.
// This means that the peer will not get punished (as per local
// peer score and can continue to send messages.
func WithGracefulIgnore() ValidatorOption {
	return func(v *validator) error {
		v.invalidResultType = pubsub.ValidationIgnore
		return nil
	}
}

func GetMessageType(msg protoreflect.ProtoMessage) protoreflect.FullName {
	return msg.ProtoReflect().Type().Descriptor().FullName()
}

// AddValidator will add a validator-function to a P2PHandler instance:
// The passed in ValidatorFunc function takes a specific message of type M complying to the
// P2PMessage interface, processes it and returns whether it is valid or not (bool value).
// If the return value is false, the message is dropped and a potentially raised error is logged.
// If the validator is registered on the P2Phandler via the AddValidator function,
// the passed in validator will be called automatically when a message of type M is received
//
// For each message type M, there can only be one validator registered per P2PHandler.
func AddValidator[M p2pmsg.Message](handler *P2PHandler, valFunc ValidatorFunc[M], opts ...ValidatorOption) error {
	var messProto M
	topic := messProto.Topic()

	_, exists := handler.validatorRegistry[topic]
	if exists {
		// This is likely not intended and happens when different messages return the same P2PMessage.Topic().
		// Currently a topic is mapped 1 to 1 to a message type (instead of using an envelope for unmarshalling)
		// (If feature needed, allow for chaining of successively registered validator functions per topic)
		return errors.Errorf(
			"can't register more than one validator per topic (topic: '%s', message-type: '%s')",
			topic,
			reflect.TypeOf(messProto))
	}

	handleError := func(err error) {
		log.Info().Str("topic", topic).Err(err).Msg("received invalid message)")
	}
	val := &validator{
		allowTraceContext: false,
	}
	opts = append([]ValidatorOption{defaults()}, opts...)
	for _, opt := range opts {
		if err := opt(val); err != nil {
			return errors.Wrap(err, "invalid validator option")
		}
	}
	validate := func(ctx context.Context, sender peer.ID, libp2pMessage *pubsub.Message) pubsub.ValidationResult {
		var (
			key M
			ok  bool
		)

		message := Message{
			Topic:        *libp2pMessage.Topic,
			Message:      libp2pMessage.Data,
			Sender:       libp2pMessage.GetFrom(),
			ReceivedFrom: libp2pMessage.ReceivedFrom,
			ID:           libp2pMessage.ID,
		}
		if message.Topic != topic {
			handleError(errors.Errorf("topic mismatch (message-topic: '%s')", message.Topic))
			return val.invalidResultType
		}
		unmshl, traceContext, err := message.Unmarshal()
		if err != nil {
			handleError(errors.Wrap(err, "error while unmarshalling message in validator"))
			return val.invalidResultType
		}

		if traceContext != nil && !val.allowTraceContext {
			handleError(errors.New("received non-empty trace-context"))
			return val.invalidResultType
		}

		key, ok = unmshl.(M)
		if !ok {
			handleError(errors.Errorf("received message of unexpected type %s", reflect.TypeOf(unmshl)))
			return val.invalidResultType
		}

		valid, err := valFunc(ctx, key)
		if err != nil {
			handleError(err)
		}
		if !valid {
			return val.invalidResultType
		}
		return pubsub.ValidationAccept
	}
	handler.validatorRegistry[topic] = validate
	handler.AddGossipTopic(topic)
	return nil
}

// AddHandlerFunc will add a handler-function to a P2PHandler instance:
// The passed in handlerFunc function takes a specific message of type M complying to the
// P2PMessage interface, processes it and returns a slice of resulting P2PMessages.
// If the handler is registered on the P2Phandler via the AddHandlerFunc function,
// the passed in handler will be called automatically when a message of type M is received,
// AFTER it has been successefully validated by the ValidatorFunc, if one is registered on the P2PHandler
//
// For each message type M, there can only be one handler registered per P2PHandler.
func AddHandlerFunc[M p2pmsg.Message](handler *P2PHandler, handlerFunc HandlerFuncStatic[M]) error {
	var messProto M
	messageType := GetMessageType(messProto)

	_, exists := handler.handlerRegistry[messageType]
	if exists {
		return errors.Errorf("Can't register more than one handler per message-type (message-type: '%s')", messageType)
	}

	f := func(ctx context.Context, msg p2pmsg.Message) ([]p2pmsg.Message, error) {
		typedMsg, ok := msg.(M)
		if !ok {
			// this is programming error, when unmarshaling of the message did not
			// result in the expected schema struct / concrete implementation
			return []p2pmsg.Message{}, errors.New("Message type assertion mismatch")
		}
		return handlerFunc(ctx, typedMsg)
	}
	handler.handlerRegistry[messageType] = f
	handler.AddGossipTopic(messProto.Topic())
	return nil
}

func New(config Config) *P2PHandler {
	bootstrapPeers := config.BootstrapPeers
	if len(bootstrapPeers) == 0 && config.Environment == Production {
		bootstrapPeers = DefaultBootstrapPeers
	}
	// exclude one's own address from the bootstrap list,
	// in case we are a bootstrap node
	bstrpPeersWithoutSelf := []peer.AddrInfo{}
	for _, bs := range bootstrapPeers {
		if !bs.ID.MatchesPrivateKey(config.PrivKey) {
			bstrpPeersWithoutSelf = append(bstrpPeersWithoutSelf, bs)
		}
	}
	config.BootstrapPeers = bstrpPeersWithoutSelf
	h := &P2PHandler{
		P2P:               NewP2PNode(config),
		gossipTopicNames:  make(map[string]bool),
		handlerRegistry:   make(HandlerRegistry),
		validatorRegistry: make(ValidatorRegistry),
	}
	return h
}

type P2PHandler struct {
	P2P              *P2PNode
	gossipTopicNames map[string]bool

	handlerRegistry   HandlerRegistry
	validatorRegistry ValidatorRegistry
}

// AddGossipTopic will subscribe to a specific topic on the
// gossip p2p-messaging layer.
// This is only necessary to call manually when we want to
// join a topic for which no handlers or validators are registered
// with the AddHandlerFunc() and AddValidator() functions
// (e.g. for a publish only scenario for the topic).
func (h *P2PHandler) AddGossipTopic(topic string) {
	h.gossipTopicNames[topic] = true
}

func (h *P2PHandler) Start(ctx context.Context, runner service.Runner) error { //nolint:unparam
	runner.Go(func() error {
		return h.P2P.Run(ctx, h.topics(), h.validatorRegistry)
	})
	if h.hasHandler() {
		runner.Go(func() error {
			return h.runHandleMessages(ctx)
		})
	}

	return nil
}

func (h *P2PHandler) topics() []string {
	topics := make([]string, 0, len(h.gossipTopicNames))
	for topicName := range h.gossipTopicNames {
		topics = append(topics, topicName)
	}
	return topics
}

func (h *P2PHandler) hasHandler() bool {
	return len(h.handlerRegistry) > 0
}

func (h *P2PHandler) runHandleMessages(ctx context.Context) error {
	// This will consume incoming messages and dispatch them to the registered handler functions
	// If the handler returns messages, then they will be sent to the broadcast
	for {
		select {
		case msg, ok := <-h.P2P.GossipMessages:
			if !ok {
				return nil
			}
			if err := h.handle(ctx, msg); err != nil {
				log.Info().Err(err).Str("topic", msg.Topic).Str("sender-id", msg.Sender.String()).
					Msg("failed to handle message")
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (h *P2PHandler) handle(ctx context.Context, msg *Message) error {
	var msgsOut []p2pmsg.Message
	var err error

	m, traceContext, err := msg.Unmarshal()
	if err != nil {
		return err
	}

	ctx, span, reportError := newSpanForReceive(ctx, h.P2P, traceContext, msg, m)
	defer span.End()

	handlerFunc, exists := h.handlerRegistry[proto.MessageName(m)]
	if !exists {
		log.Info().Str("message", m.LogInfo()).Str("topic", msg.Topic).Str("sender-id", msg.Sender.String()).
			Msg("ignoring message, no handler registered for topic")
		return nil
	}

	log.Info().Str("message", m.LogInfo()).Str("topic", msg.Topic).Str("sender-id", msg.Sender.String()).
		Msg("received message")
	msgsOut, err = handlerFunc(ctx, m)
	if err != nil {
		return reportError(err)
	}
	for _, msgOut := range msgsOut {
		if err := h.SendMessage(ctx, msgOut); err != nil {
			log.Info().Err(err).Str("message", msgOut.LogInfo()).Str("topic", msgOut.Topic()).
				Msg("failed to send message")
			continue
		}
	}
	return nil
}

func (h *P2PHandler) SendMessage(ctx context.Context, msg p2pmsg.Message, retryOpts ...retry.Option) error {
	var traceContext *p2pmsg.TraceContext
	ctx, span, reportError := newSpanForPublish(ctx, h.P2P, traceContext, msg)
	defer span.End()

	msgBytes, err := p2pmsg.Marshal(msg, traceContext)
	if err != nil {
		return reportError(errors.Wrap(err, "failed to marshal p2p message"))
	}

	// if no retry options are passed, don't do any retries!
	if len(retryOpts) == 0 {
		retryOpts = []retry.Option{retry.NumberOfRetries(0)}
	}

	log.Info().Str("message", msg.LogInfo()).Str("topic", msg.Topic()).
		Msg("sending message")
	_, callErr := retry.FunctionCall(
		ctx,
		func(ctx context.Context) (struct{}, error) {
			return struct{}{}, h.P2P.Publish(ctx, msg.Topic(), msgBytes)
		},
		retryOpts...,
	)
	return reportError(callErr)
}
