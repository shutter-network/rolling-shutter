package p2p

import (
	"context"
	"encoding/json"

	"github.com/libp2p/go-libp2p-core/peer"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
)

// gossipRoom represents a subscription to a single PubSub topic. Messages
// can be published to the topic with gossipRoom.Publish, and received
// messages are pushed to the Messages channel.
type gossipRoom struct {
	pubSub       *pubsub.PubSub
	topic        *pubsub.Topic
	subscription *pubsub.Subscription
	topicName    string
	self         peer.ID
}

// Message gets converted to/from JSON and sent in the body of pubsub messages.
type Message struct {
	Topic    string
	Message  []byte
	SenderID string
}

// Publish sends a message to the pubsub topic.
func (room *gossipRoom) Publish(ctx context.Context, message []byte) error {
	m := Message{
		Topic:    room.topicName,
		Message:  message,
		SenderID: room.self.Pretty(),
	}
	msgBytes, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return room.topic.Publish(ctx, msgBytes)
}

func (room *gossipRoom) ListPeers() []peer.ID {
	return room.pubSub.ListPeers(room.topicName)
}

// readLoop pulls messages from the pubsub topic and pushes them onto the given messages channel.
func (room *gossipRoom) readLoop(ctx context.Context, messages chan *Message) error {
	for {
		msg, err := room.subscription.Next(ctx)
		if err != nil {
			return err
		}
		// only forward messages delivered by others
		if msg.ReceivedFrom == room.self {
			continue
		}
		m := new(Message)
		err = json.Unmarshal(msg.Data, m)
		if err != nil {
			continue
		}

		metricMessagesReceived.WithLabelValues(m.Topic).Inc()

		// send valid messages onto the Messages channel
		select {
		case messages <- m:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
