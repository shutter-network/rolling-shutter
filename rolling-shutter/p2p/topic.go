package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/libp2p/go-libp2p-core/peer"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
)

// MessagesBufSize is the number of incoming messages to buffer for each topic.
const MessagesBufSize = 128

// TopicGossip represents a subscription to a single PubSub topic. Messages
// can be published to the topic with TopicGossip.Publish, and received
// messages are pushed to the Messages channel.
type TopicGossip struct {
	// Messages is a channel of messages received from other peers in the chat room
	Messages chan *Message

	ctx   context.Context
	ps    *pubsub.PubSub
	Topic *pubsub.Topic
	sub   *pubsub.Subscription

	topicName string
	Self      peer.ID
}

// Message gets converted to/from JSON and sent in the body of pubsub messages.
type Message struct {
	Message  string
	SenderID string
}

// JoinTopic tries to subscribe to the PubSub topic returning a TopicGossip on success.
func JoinTopic(ctx context.Context, ps *pubsub.PubSub, selfID peer.ID, topicName string) (*TopicGossip, error) {
	// join the pubsub topic
	topic, err := ps.Join(topicName)
	if err != nil {
		return nil, err
	}

	// and subscribe to it
	sub, err := topic.Subscribe()
	if err != nil {
		return nil, err
	}

	topicGossip := &TopicGossip{
		ctx:       ctx,
		ps:        ps,
		Topic:     topic,
		sub:       sub,
		Self:      selfID,
		topicName: topicName,
		Messages:  make(chan *Message, MessagesBufSize),
	}

	// start reading messages from the subscription in a loop
	go topicGossip.readLoop()
	return topicGossip, nil
}

// Publish sends a message to the pubsub topic.
func (topicGossip *TopicGossip) Publish(message string) error {
	m := Message{
		Message:  message,
		SenderID: topicGossip.Self.Pretty(),
	}
	msgBytes, err := json.Marshal(m)
	if err != nil {
		return err
	}
	fmt.Println("Publishing message: ", message)
	return topicGossip.Topic.Publish(topicGossip.ctx, msgBytes)
}

func (topicGossip *TopicGossip) ListPeers() []peer.ID {
	return topicGossip.ps.ListPeers(topicGossip.topicName)
}

// readLoop pulls messages from the pubsub topic and pushes them onto the Messages channel.
func (topicGossip *TopicGossip) readLoop() {
	for {
		msg, err := topicGossip.sub.Next(topicGossip.ctx)
		if err != nil {
			close(topicGossip.Messages)
			return
		}
		// only forward messages delivered by others
		if msg.ReceivedFrom == topicGossip.Self {
			continue
		}
		m := new(Message)
		err = json.Unmarshal(msg.Data, m)
		if err != nil {
			continue
		}
		// send valid messages onto the Messages channel
		topicGossip.Messages <- m
	}
}
