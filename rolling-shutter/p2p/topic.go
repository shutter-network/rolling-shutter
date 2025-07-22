package p2p

import (
	"context"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/rs/zerolog/log"
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

// Publish sends a message to the pubsub topic.
func (room *gossipRoom) Publish(ctx context.Context, message []byte) error {
	return room.topic.Publish(ctx, message)
}

func (room *gossipRoom) ListPeers() []peer.ID {
	return room.pubSub.ListPeers(room.topicName)
}

// readLoop pulls messages from the pubsub topic and pushes them onto the given messages channel.
func (room *gossipRoom) readLoop(ctx context.Context, messages chan *pubsub.Message) error {
	for {
		msg, err := room.subscription.Next(ctx)
		if err != nil {
			return err
		}
		// only forward messages delivered by others
		if msg.ReceivedFrom == room.self {
			continue
		}
		// send valid messages onto the Messages channel
		select {
		case messages <- msg:
		case <-ctx.Done():
			log.Debug().Msg("subscription canceled, closing read loop")
			room.subscription.Cancel()
			if err := room.topic.Close(); err != nil {
				return err
			}
			return ctx.Err()
		}
	}
}
