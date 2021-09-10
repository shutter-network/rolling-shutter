package p2p

import (
	"context"
	"testing"
	"time"

	"github.com/multiformats/go-multiaddr"
	"gotest.tools/assert"
)

// TestStartNetworkNode test that we can init two p2p nodes and make them send/receive messages.
func TestStartNetworkNodeIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 1200*time.Millisecond)
	defer cancel()
	gossipTopicNames := []string{"testTopic1", "testTopic2"}
	nodeAddress1, _ := multiaddr.NewMultiaddr("/ip4/127.0.0.1/tcp/2000")
	nodeAddress2, _ := multiaddr.NewMultiaddr("/ip4/127.0.0.1/tcp/3000")
	testMessage := "test message"

	p1 := NewP2p()
	p2 := NewP2p()

	nodeAddresses := []multiaddr.Multiaddr{nodeAddress1, nodeAddress2}
	for i, p := range []*P2P{p1, p2} {
		if err := p.CreateHost(ctx, nodeAddresses[i]); err != nil {
			t.Fatalf("Error while creating node %d: %v", i, err)
		}
		if err := p.JoinTopics(ctx, gossipTopicNames); err != nil {
			t.Fatalf("Error while joining topics for node %d: %v", i, err)
		}
	}

	identityMultiAddress1, err := p1.GetMultiaddr()
	if err != nil {
		t.Fatalf("Error while getting node 1 multiaddr: %v", err)
	}

	if err := p2.ConnectToPeer(ctx, identityMultiAddress1); err != nil {
		t.Fatalf("Error while connecting to node 1: %v", err)
	}

	// The following loop publishes the same message over and over. Even though we did call
	// ConnectToPeer, libp2p takes some time until the peer receives the first message.
	var message *Message
	topicName := gossipTopicNames[0]
	for message == nil {
		if err := p2.TopicGossips[topicName].Publish(ctx, testMessage); err != nil {
			t.Fatalf("error while publishing message: %v", err)
		}

		select {
		case message = <-p1.TopicGossips[topicName].Messages:
			if message == nil {
				t.Fatalf("channel closed unexpectedly")
			}
		case <-ctx.Done():
			t.Fatalf("waiting for message: %s", ctx.Err())
		case <-time.After(5 * time.Millisecond):
		}
	}
	assert.Equal(t, testMessage, message.Message, "received wrong message")
	assert.Equal(t, p2.HostID(), message.SenderID, "received message with wrong sender")
}
