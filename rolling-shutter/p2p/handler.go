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
	HandlerFunc       func(context.Context, p2pmsg.Message) ([]p2pmsg.Message, error)
	HandlerRegistry   map[protoreflect.FullName]HandlerFunc
	ValidatorFunc     func(context.Context, p2pmsg.Message) (bool, error)
	ValidatorRegistry map[string]pubsub.ValidatorEx
)

const (
	allowTraceContext = true // whether we allow the trace field to be set in the message envelope
	invalidResultType = pubsub.ValidationReject
)

type MessageHandler interface {
	ValidateMessage(context.Context, p2pmsg.Message) (bool, error)
	HandleMessage(context.Context, p2pmsg.Message) ([]p2pmsg.Message, error)
	MessagePrototypes() []p2pmsg.Message
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
	return &P2PHandler{
		P2P:               NewP2PNode(config),
		gossipTopicNames:  make(map[string]struct{}),
		handlerRegistry:   make(HandlerRegistry),
		validatorRegistry: make(ValidatorRegistry),
	}
}

type P2PHandler struct {
	P2P              *P2PNode
	gossipTopicNames map[string]struct{}

	handlerRegistry   HandlerRegistry
	validatorRegistry ValidatorRegistry
}

// AddHandlerFunc will add a handler-function to a P2PHandler instance:
// The passed in handlerFunc function takes a specific message of type M complying to the
// P2PMessage interface, processes it and returns a slice of resulting P2PMessages.
// If the handler is registered on the P2Phandler via the AddHandlerFunc function,
// the passed in handler will be called automatically when a message of type M is received,
// AFTER it has been successefully validated by the ValidatorFunc, if one is registered on the P2PHandler
//
// For each message type M, there can only be one handler registered per P2PHandler.
func (handler *P2PHandler) AddHandlerFunc(handlerFunc HandlerFunc, protos ...p2pmsg.Message) {
	for _, p := range protos {
		messageType := proto.MessageName(p)
		_, exists := handler.handlerRegistry[messageType]
		if exists {
			panic(errors.Errorf(
				"handler already registered: message-type=%s", messageType))
		}
		handler.handlerRegistry[messageType] = handlerFunc
		handler.AddGossipTopic(p.Topic())
	}
}

func (handler *P2PHandler) addValidatorImpl(valFunc ValidatorFunc, messProto p2pmsg.Message) {
	topic := messProto.Topic()
	_, exists := handler.validatorRegistry[topic]
	if exists {
		// This is likely not intended and happens when different messages return the same P2PMessage.Topic().
		// Currently a topic is mapped 1 to 1 to a message type (instead of using an envelope for unmarshalling)
		// (If feature needed, allow for chaining of successively registered validator functions per topic)
		panic(errors.Errorf(
			"can't register more than one validator per topic (topic: '%s', message-type: '%s')",
			topic,
			reflect.TypeOf(messProto)))
	}
	handleError := func(err error) {
		log.Info().Str("topic", topic).Err(err).Msg("received invalid message)")
	}
	validate := func(ctx context.Context, sender peer.ID, message *pubsub.Message) pubsub.ValidationResult {
		if message.GetTopic() != topic {
			handleError(errors.Errorf("topic mismatch (message-topic: '%s')", message.GetTopic()))
			return invalidResultType
		}
		unmshl, traceContext, err := UnmarshalPubsubMessage(message)
		if err != nil {
			handleError(errors.Wrap(err, "error while unmarshalling message in validator"))
			return invalidResultType
		}

		if traceContext != nil && !allowTraceContext {
			handleError(errors.New("received non-empty trace-context"))
			return invalidResultType
		}

		if reflect.TypeOf(unmshl) != reflect.TypeOf(messProto) {
			handleError(errors.Errorf("received message of unexpected type %s", reflect.TypeOf(unmshl)))
			return invalidResultType
		}

		valid, err := valFunc(ctx, unmshl)
		if err != nil {
			handleError(err)
		}
		if !valid {
			return invalidResultType
		}
		return pubsub.ValidationAccept
	}
	handler.validatorRegistry[topic] = validate
	handler.AddGossipTopic(topic)
}

// AddValidator will add a validator-function to a P2PHandler instance:
// The passed in ValidatorFunc function takes a specific message of type M complying to the
// P2PMessage interface, processes it and returns whether it is valid or not (bool value).
// If the return value is false, the message is dropped and a potentially raised error is logged.
// If the validator is registered on the P2Phandler via the AddValidator function,
// the passed in validator will be called automatically when a message of type M is received
//
// For each message type M, there can only be one validator registered per P2PHandler.
func (handler *P2PHandler) AddValidator(valFunc ValidatorFunc, protos ...p2pmsg.Message) {
	for _, p := range protos {
		handler.addValidatorImpl(valFunc, p)
	}
}

func (handler *P2PHandler) AddMessageHandler(mhs ...MessageHandler) {
	for _, mh := range mhs {
		protos := mh.MessagePrototypes()
		handler.AddHandlerFunc(mh.HandleMessage, protos...)
		handler.AddValidator(mh.ValidateMessage, protos...)
	}
}

// AddGossipTopic will subscribe to a specific topic on the
// gossip p2p-messaging layer.
// This is only necessary to call manually when we want to
// join a topic for which no handlers or validators are registered
// with the AddHandlerFunc() and AddValidator() functions
// (e.g. for a publish only scenario for the topic).
func (handler *P2PHandler) AddGossipTopic(topic string) {
	handler.gossipTopicNames[topic] = struct{}{}
}

func (handler *P2PHandler) Start(ctx context.Context, runner service.Runner) error { //nolint:unparam
	runner.Go(func() error {
		return handler.P2P.Run(ctx, handler.topics(), handler.validatorRegistry)
	})
	if handler.hasHandler() {
		runner.Go(func() error {
			return handler.runHandleMessages(ctx)
		})
	}

	return nil
}

func (handler *P2PHandler) topics() []string {
	topics := make([]string, 0, len(handler.gossipTopicNames))
	for topicName := range handler.gossipTopicNames {
		topics = append(topics, topicName)
	}
	return topics
}

func (handler *P2PHandler) hasHandler() bool {
	return len(handler.handlerRegistry) > 0
}

func (handler *P2PHandler) runHandleMessages(ctx context.Context) error {
	// This will consume incoming messages and dispatch them to the registered handler functions
	// If the handler returns messages, then they will be sent to the broadcast
	for {
		select {
		case msg, ok := <-handler.P2P.GossipMessages:
			if !ok {
				return nil
			}
			if err := handler.handle(ctx, msg); err != nil {
				log.Info().Err(err).Str("topic", msg.GetTopic()).Str("sender-id", msg.GetFrom().String()).
					Msg("failed to handle message")
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (handler *P2PHandler) handle(ctx context.Context, msg *pubsub.Message) error {
	var msgsOut []p2pmsg.Message
	var err error

	m, traceContext, err := UnmarshalPubsubMessage(msg)
	if err != nil {
		return err
	}

	ctx, span, reportError := newSpanForReceive(ctx, handler.P2P, traceContext, msg, m)
	defer span.End()

	handlerFunc, exists := handler.handlerRegistry[proto.MessageName(m)]
	if !exists {
		log.Info().Str("message", m.LogInfo()).Str("topic", msg.GetTopic()).Str("sender-id", msg.GetFrom().String()).
			Msg("ignoring message, no handler registered for topic")
		return nil
	}

	log.Info().Str("message", m.LogInfo()).Str("topic", msg.GetTopic()).Str("sender-id", msg.GetFrom().String()).
		Msg("received message")
	msgsOut, err = handlerFunc(ctx, m)
	if err != nil {
		return reportError(err)
	}
	for _, msgOut := range msgsOut {
		if err := handler.SendMessage(ctx, msgOut); err != nil {
			log.Info().Err(err).Str("message", msgOut.LogInfo()).Str("topic", msgOut.Topic()).
				Msg("failed to send message")
			continue
		}
	}
	return nil
}

func (handler *P2PHandler) SendMessage(ctx context.Context, msg p2pmsg.Message, retryOpts ...retry.Option) error {
	var traceContext *p2pmsg.TraceContext
	ctx, span, reportError := newSpanForPublish(ctx, handler.P2P, traceContext, msg)
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
			return struct{}{}, handler.P2P.Publish(ctx, msg.Topic(), msgBytes)
		},
		retryOpts...,
	)
	return reportError(callErr)
}
