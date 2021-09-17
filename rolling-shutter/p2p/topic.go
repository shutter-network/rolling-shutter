package p2p

import (
	"context"
	"encoding/json"

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

	pubSub       *pubsub.PubSub
	Topic        *pubsub.Topic
	subscription *pubsub.Subscription

	topicName string
	Self      peer.ID
}

// Message gets converted to/from JSON and sent in the body of pubsub messages.
type Message struct {
	Message  string
	SenderID string
}

// Publish sends a message to the pubsub topic.
func (topicGossip *TopicGossip) Publish(ctx context.Context, message string) error {
	m := Message{
		Message:  message,
		SenderID: topicGossip.Self.Pretty(),
	}
	msgBytes, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return topicGossip.Topic.Publish(ctx, msgBytes)
}

func (topicGossip *TopicGossip) ListPeers() []peer.ID {
	return topicGossip.pubSub.ListPeers(topicGossip.topicName)
}

// readLoop pulls messages from the pubsub topic and pushes them onto the Messages channel.
func (topicGossip *TopicGossip) readLoop(ctx context.Context) error {
	defer func() {
		close(topicGossip.Messages)
	}()
	for {
		msg, err := topicGossip.subscription.Next(ctx)
		if err != nil {
			return err
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
		select {
		case topicGossip.Messages <- m:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
