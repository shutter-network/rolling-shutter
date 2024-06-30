package p2p

import (
	"context"
	"reflect"
	"time"

	"github.com/hashicorp/go-multierror"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	oteltrace "go.opentelemetry.io/otel/trace"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/encodeable/env"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/retry"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/service"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2pmsg"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/trace"
)

type (
	HandlerFunc       func(context.Context, p2pmsg.Message) ([]p2pmsg.Message, error)
	HandlerRegistry   map[protoreflect.FullName][]HandlerFunc
	ValidatorFunc     func(context.Context, p2pmsg.Message) (pubsub.ValidationResult, error)
	ValidatorRegistry map[string][]pubsub.ValidatorEx
)

func (r *ValidatorRegistry) GetCombinedValidator(topic string) pubsub.ValidatorEx {
	validate := func(ctx context.Context, sender peer.ID, message *pubsub.Message) pubsub.ValidationResult {
		startTime := time.Now()
		defer func() {
			elapsedTime := time.Since(startTime)
			metricsP2PMessageValidationTime.WithLabelValues(topic).Observe(elapsedTime.Seconds())
		}()

		ignored := false
		for _, valFunc := range (*r)[topic] {
			res := valFunc(ctx, sender, message)
			switch res {
			case pubsub.ValidationAccept:
				continue
			case pubsub.ValidationReject:
				return pubsub.ValidationReject
			case pubsub.ValidationIgnore:
				ignored = true
			default:
				log.Warn().Str("topic", topic).Msg("unknown validation result %d, treating as reject")
				return pubsub.ValidationReject
			}
		}
		if ignored {
			return pubsub.ValidationIgnore
		}
		return pubsub.ValidationAccept
	}
	return validate
}

const (
	allowTraceContext = true // whether we allow the trace field to be set in the message envelope
	invalidResultType = pubsub.ValidationReject
)

type Messaging interface {
	service.Service
	SendMessage(context.Context, p2pmsg.Message, ...retry.Option) error
	AddValidator(valFunc ValidatorFunc, protos ...p2pmsg.Message)
	AddMessageHandler(mhs ...MessageHandler)
}

type MessageHandler interface {
	ValidateMessage(context.Context, p2pmsg.Message) (pubsub.ValidationResult, error)
	HandleMessage(context.Context, p2pmsg.Message) ([]p2pmsg.Message, error)
	MessagePrototypes() []p2pmsg.Message
}

func New(config *Config) (*P2PMessaging, error) {
	peerID, err := config.P2PKey.PeerID()
	if err != nil {
		return nil, err
	}

	listenAddresses := []multiaddr.Multiaddr{}
	for _, addr := range config.ListenAddresses {
		listenAddresses = append(listenAddresses, addr.Multiaddr)
	}
	cfg := &p2pNodeConfig{
		ListenAddrs:        listenAddresses,
		PrivKey:            *config.P2PKey,
		Environment:        config.Environment,
		DiscoveryNamespace: config.DiscoveryNamespace,
	}

	bootstrapAddresses := config.CustomBootstrapAddresses
	if len(bootstrapAddresses) == 0 && config.Environment == env.EnvironmentProduction {
		bootstrapAddresses = DefaultBootstrapPeers
	}
	// exclude one's own address from the bootstrap list,
	// in case we are a bootstrap node
	for _, addr := range bootstrapAddresses {
		bootstrapID, err := addr.Identifier()
		if err != nil {
			log.Warn().
				Err(err).
				Msg("invalid bootstrap peer address, dropping")
			continue
		}
		if bootstrapID.Equal(peerID) {
			cfg.IsBootstrapNode = true
		} else {
			ai, err := peer.AddrInfoFromP2pAddr(addr.Multiaddr)
			if err != nil {
				log.Warn().
					Err(err).
					Str("address", addr.String()).
					Msg("invalid bootstrap peer address, dropping")
				continue
			}
			cfg.BootstrapPeers = append(cfg.BootstrapPeers, *ai)
		}
	}
	if len(cfg.BootstrapPeers) < 1 && !cfg.IsBootstrapNode {
		return nil, errors.New("no bootstrap peers configured")
	}

	return &P2PMessaging{
		P2P:               NewP2PNode(*cfg),
		gossipTopicNames:  make(map[string]struct{}),
		handlerRegistry:   make(HandlerRegistry),
		validatorRegistry: make(ValidatorRegistry),
	}, nil
}

type P2PMessaging struct {
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
func (m *P2PMessaging) AddHandlerFunc(handlerFunc HandlerFunc, protos ...p2pmsg.Message) {
	for _, p := range protos {
		messageType := proto.MessageName(p)
		fns, exists := m.handlerRegistry[messageType]
		if !exists {
			fns = []HandlerFunc{}
		}
		m.handlerRegistry[messageType] = append(fns, handlerFunc)
		m.AddGossipTopic(p.Topic())
	}
}

func (m *P2PMessaging) addValidatorImpl(valFunc ValidatorFunc, messProto p2pmsg.Message) {
	topic := messProto.Topic()
	handleError := func(err error) {
		log.Info().Str("topic", topic).Err(err).Msg("received invalid message")
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
			handleError(
				errors.Errorf("received message of unexpected type %s", reflect.TypeOf(unmshl)),
			)
			return invalidResultType
		}

		valid, err := valFunc(ctx, unmshl)
		if err != nil {
			handleError(err)
		}
		return valid
	}

	_, exists := m.validatorRegistry[topic]
	if !exists {
		m.AddGossipTopic(topic)
	}
	m.validatorRegistry[topic] = append(m.validatorRegistry[topic], validate)
}

// AddValidator will add a validator-function to a P2PHandler instance:
// The passed in ValidatorFunc function takes a specific message of type M complying to the
// P2PMessage interface, processes it and returns whether it is valid or not (bool value).
// If the return value is false, the message is dropped and a potentially raised error is logged.
// If the validator is registered on the P2Phandler via the AddValidator function,
// the passed in validator will be called automatically when a message of type M is received
//
// For each message type M, there can only be one validator registered per P2PHandler.
func (m *P2PMessaging) AddValidator(valFunc ValidatorFunc, protos ...p2pmsg.Message) {
	for _, p := range protos {
		m.addValidatorImpl(valFunc, p)
	}
}

func (m *P2PMessaging) AddMessageHandler(mhs ...MessageHandler) {
	for _, mh := range mhs {
		protos := mh.MessagePrototypes()
		m.AddHandlerFunc(mh.HandleMessage, protos...)
		m.AddValidator(mh.ValidateMessage, protos...)
	}
}

// AddGossipTopic will subscribe to a specific topic on the
// gossip p2p-messaging layer.
// This is only necessary to call manually when we want to
// join a topic for which no handlers or validators are registered
// with the AddHandlerFunc() and AddValidator() functions
// (e.g. for a publish only scenario for the topic).
func (m *P2PMessaging) AddGossipTopic(topic string) {
	m.gossipTopicNames[topic] = struct{}{}
}

func (m *P2PMessaging) Start(
	ctx context.Context,
	runner service.Runner,
) error { //nolint:unparam
	runner.Go(func() error {
		return m.P2P.Run(ctx, runner, m.topics(), m.validatorRegistry)
	})
	if m.hasHandler() {
		runner.Go(func() error {
			return m.runHandleMessages(ctx)
		})
	}

	return nil
}

func (m *P2PMessaging) topics() []string {
	topics := make([]string, 0, len(m.gossipTopicNames))
	for topicName := range m.gossipTopicNames {
		topics = append(topics, topicName)
	}
	return topics
}

func (m *P2PMessaging) hasHandler() bool {
	return len(m.handlerRegistry) > 0
}

func (m *P2PMessaging) runHandleMessages(ctx context.Context) error {
	// This will consume incoming messages and dispatch them to the registered handler functions
	// If the handler returns messages, then they will be sent to the broadcast
	for {
		select {
		case msg, ok := <-m.P2P.GossipMessages:
			if !ok {
				return nil
			}
			startTime := time.Now()
			if err := m.handle(ctx, msg); err != nil {
				log.Info().
					Err(err).
					Str("topic", msg.GetTopic()).
					Str("sender-id", msg.GetFrom().String()).
					Msg("failed to handle message")
			}
			elapsedTime := time.Since(startTime)
			metricsP2PMessageHandlingTime.WithLabelValues(msg.GetTopic()).Observe(elapsedTime.Seconds())
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

var (
	ErrNoMessageHandler = errors.New("No message handler registered")
	ErrFailedHandler    = errors.New("Error during handler execution")
)

func (m *P2PMessaging) Handle(ctx context.Context, msg p2pmsg.Message) ([]p2pmsg.Message, error) {
	startTime := time.Now()
	messageName := proto.MessageName(msg)
	defer func() {
		elapsedTime := time.Since(startTime)
		log.Debug().
			Str("message-name", string(messageName)).
			Str("duration", elapsedTime.String()).
			Msg("handled message")
	}()

	var (
		msgsOut   []p2pmsg.Message
		errResult error
	)
	fns, exists := m.handlerRegistry[messageName]
	if !exists {
		return nil, ErrNoMessageHandler
	}
	for _, handlerFunc := range fns {
		msgs, err := handlerFunc(ctx, msg)
		if err != nil {
			errResult = multierror.Append(errResult, errors.Wrap(err, ErrFailedHandler.Error()))
			continue
		}
		msgsOut = append(msgsOut, msgs...)
	}
	return msgsOut, errResult
}

func (m *P2PMessaging) handle(ctx context.Context, msg *pubsub.Message) error {
	var (
		err         error
		span        oteltrace.Span
		reportError trace.ErrorWrapper
	)
	p2pMsg, traceContext, err := UnmarshalPubsubMessage(msg)
	if err != nil {
		return err
	}

	if m.P2P != nil {
		ctx, span, reportError = newSpanForReceive(ctx, m.P2P, traceContext, msg, p2pMsg)
		defer span.End()
	} else {
		reportError = func(err error) error {
			return err
		}
	}
	outMsgs, err := m.Handle(ctx, p2pMsg)
	if errors.Is(err, ErrNoMessageHandler) {
		log.Info().
			Str("message", p2pMsg.LogInfo()).
			Str("topic", msg.GetTopic()).
			Str("sender-id", msg.GetFrom().String()).
			Msg("ignoring message, no handler registered for topic")
		return nil
	}
	if err != nil {
		return reportError(err)
	}
	for _, msgOut := range outMsgs {
		if err := m.SendMessage(ctx, msgOut); err != nil {
			log.Info().Err(err).Str("message", msgOut.LogInfo()).Str("topic", msgOut.Topic()).
				Msg("failed to send message")
			continue
		}
	}
	log.Info().
		Str("message", p2pMsg.LogInfo()).
		Str("topic", msg.GetTopic()).
		Str("sender-id", msg.GetFrom().String()).
		Msg("received message")
	return nil
}

func (m *P2PMessaging) SendMessage(
	ctx context.Context,
	msg p2pmsg.Message,
	retryOpts ...retry.Option,
) error {
	var traceContext *p2pmsg.TraceContext
	ctx, span, reportError := newSpanForPublish(ctx, m.P2P, traceContext, msg)
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
			return struct{}{}, m.P2P.Publish(ctx, msg.Topic(), msgBytes)
		},
		retryOpts...,
	)
	return reportError(callErr)
}
