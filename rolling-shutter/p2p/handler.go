package p2p

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"strings"

	"github.com/libp2p/go-libp2p-core/peer"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/shmsg"
)

type (
	HandlerFuncStatic[M shmsg.P2PMessage] func(context.Context, M) ([]shmsg.P2PMessage, error)
	HandlerFunc                           func(context.Context, shmsg.P2PMessage) ([]shmsg.P2PMessage, error)
	HandlerRegistry                       map[protoreflect.FullName]HandlerFunc
	ValidatorFunc[M shmsg.P2PMessage]     func(context.Context, M) (bool, error)
	ValidatorRegistry                     map[string]pubsub.Validator
)

func GetMessageType(msg shmsg.P2PMessage) protoreflect.FullName {
	return msg.ProtoReflect().Type().Descriptor().FullName()
}

// AddValidator will add a validator-function to a P2PHandler instance:
// The passed in ValidatorFunc function takes a specific message of type M complying to the
// P2PMessage interface, processes it and returns wether it is valid or not (bool value).
// If the return value is false, the message is dropped and a potentially raised error is logged.
// If the validator is registered on the P2Phandler via the AddValidator function,
// the passed in validator will be called automatically when a message of type M is received
//
// For each message type M, there can only be one validator registered per P2PHandler.
func AddValidator[M shmsg.P2PMessage](handler *P2PHandler, valFunc ValidatorFunc[M]) pubsub.Validator {
	var messProto M
	topic := messProto.Topic()

	_, exists := handler.validatorRegistry[topic]
	if exists {
		// This is likely not intended and happens when different messages return the same P2PMessage.Topic().
		// Currently a topic is mapped 1 to 1 to a message type (instead of using an envelope for unmarshalling)

		// Instead of silently overwriting the old validator, rather panic.
		// (If feature needed, allow for chaining of successively registered validator functions per topic)
		panic(fmt.Sprintf("Can't register more than one validator per topic (topic: '%s', message-type: '%s')", topic, reflect.TypeOf(messProto)))
	}

	handleError := func(err error) {
		log.Printf("received invalid message on topic %s, cause: %s)", topic, err)
	}
	validate := func(ctx context.Context, _ peer.ID, libp2pMessage *pubsub.Message) bool {
		var (
			key M
			ok  bool
		)

		message := Message{
			Topic:    *libp2pMessage.Topic,
			Message:  libp2pMessage.Data,
			SenderID: libp2pMessage.GetFrom().Pretty(),
		}
		if strings.Compare(message.Topic, topic) != 0 {
			// This should not happen, if so then we registered the validator function on the wrong topic
			handleError(errors.Errorf("topic mismatch (message-topic: '%s')", message.Topic))
			return false
		}
		unmshl, err := message.Unmarshal()
		if err != nil {
			handleError(errors.Wrap(err, "error while unmarshalling message in validator"))
			return false
		}

		key, ok = unmshl.(M)
		if !ok {
			handleError(errors.Errorf("received message of unexpected type %s", reflect.TypeOf(unmshl)))
			return false
		}

		valid, err := valFunc(ctx, key)
		if err != nil {
			handleError(err)
		}
		return valid
	}
	handler.validatorRegistry[topic] = validate
	handler.AddGossipTopic(topic)
	return validate
}

// AddHandlerFunc will add a handler-function to a P2PHandler instance:
// The passed in handlerFunc function takes a specific message of type M complying to the
// P2PMessage interface, processes it and returns a slice of resulting P2PMessages.
// If the handler is registered on the P2Phandler via the AddHandlerFunc function,
// the passed in handler will be called automatically when a message of type M is received,
// AFTER it has been successefully validated by the ValidatorFunc, if one is registered on the P2PHandler
//
// For each message type M, there can only be one handler registered per P2PHandler.
func AddHandlerFunc[M shmsg.P2PMessage](handler *P2PHandler, handlerFunc HandlerFuncStatic[M]) HandlerFunc {
	var messProto M
	messageType := GetMessageType(messProto)

	_, exists := handler.handlerRegistry[messageType]
	if exists {
		panic(fmt.Sprintf("Can't register more than one handler per message-type (message-type: '%s')", messageType))
	}

	f := func(ctx context.Context, msg shmsg.P2PMessage) ([]shmsg.P2PMessage, error) {
		typedMsg, ok := msg.(M)
		if !ok {
			// this is programming error, when unmarshaling of the message did not
			// result in the expected schema struct / concrete implementation
			return []shmsg.P2PMessage{}, errors.New("Message type assertion mismatch")
		}
		return handlerFunc(ctx, typedMsg)
	}
	handler.handlerRegistry[messageType] = f
	handler.AddGossipTopic(messProto.Topic())
	return f
}

func New(config Config) *P2PHandler {
	h := &P2PHandler{
		P2P:               NewP2P(config),
		gossipTopicNames:  make(map[string]bool),
		handlerRegistry:   make(HandlerRegistry),
		validatorRegistry: make(ValidatorRegistry),
	}
	return h
}

type P2PHandler struct {
	P2P              *P2P
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

func (h *P2PHandler) Run(ctx context.Context) error {
	group, ctx := errgroup.WithContext(ctx)

	topics := make([]string, 0, len(h.gossipTopicNames))
	for topicName := range h.gossipTopicNames {
		topics = append(topics, topicName)
	}

	group.Go(func() error {
		return h.P2P.Run(ctx, topics, h.validatorRegistry)
	})
	if h.hasHandler() {
		group.Go(func() error {
			return h.runHandleMessages(ctx)
		})
	}
	return group.Wait()
}

func (h *P2PHandler) hasHandler() bool {
	return len(h.handlerRegistry) > 0
}

func (h *P2PHandler) runHandleMessages(ctx context.Context) error {
	// This will consume incoming messages and dispatch them to the registered handler functions
	// If the handler return messages, then they will be sent to the broadcast
	for {
		select {
		case msg, ok := <-h.P2P.GossipMessages:
			if !ok {
				return nil
			}
			if err := h.handle(ctx, msg); err != nil {
				log.Printf("error handling message %+v: %s", msg, err)
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (h *P2PHandler) handle(ctx context.Context, msg *Message) error {
	var msgsOut []shmsg.P2PMessage
	var err error

	m, err := msg.Unmarshal()
	if err != nil {
		return err
	}

	msgType := GetMessageType(m)
	handlerFunc, exists := h.handlerRegistry[msgType]
	if !exists {
		log.Printf("ignoring message received on topic '%s' (sender=%s)", msg.Topic, msg.SenderID)
		return nil
	}

	log.Printf("received message (msg: %s, topic=%s, sender=%s)", m.LogInfo(), msg.Topic, msg.SenderID)
	msgsOut, err = handlerFunc(ctx, m)
	if err != nil {
		return err
	}
	for _, msgOut := range msgsOut {
		if err := h.SendMessage(ctx, msgOut); err != nil {
			log.Printf("error sending message %+v: %s", msgOut, err)
			continue
		}
	}
	return nil
}

func (h *P2PHandler) SendMessage(ctx context.Context, msg shmsg.P2PMessage) error {
	msgBytes, err := proto.Marshal(msg)
	if err != nil {
		return errors.Wrap(err, "failed to marshal p2p message")
	}
	log.Printf("sending %s", msg.LogInfo())

	return h.P2P.Publish(ctx, msg.Topic(), msgBytes)
}
