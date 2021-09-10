package p2p

import (
	"context"
	"testing"
	"time"

	"github.com/multiformats/go-multiaddr"
)

// TestStartNetworkNode test that we can init two p2p nodes and make them send/receive messages.
func TestStartNetworkNodeIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	ctx := context.Background()
	gossipTopicNames := []string{"testTopic1", "testTopic2"}
	nodeShortAddress1 := "/ip4/127.0.0.1/tcp/2000"
	nodeAddress1, _ := multiaddr.NewMultiaddr(nodeShortAddress1)
	nodeAddress2, _ := multiaddr.NewMultiaddr("/ip4/127.0.0.1/tcp/3000")
	testMessage := "test message"

	p1 := NewP2p()
	p2 := NewP2p()

	nodeAddreses := []multiaddr.Multiaddr{nodeAddress1, nodeAddress2}
	for i, p := range []*P2P{p1, p2} {
		if err := p.CreateHost(ctx, nodeAddreses[i]); err != nil {
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

	// give time for nodes to connect to each other
	ticker := time.NewTicker(500 * time.Millisecond)
	for {
		timeElapsed := <-ticker.C
		if len(p2.ConnectedPeers()) != 0 {
			break
		}
		if timeElapsed.Second() >= 4 {
			t.Fatalf("Timeout while waiting for peers to connect")
		}
	}

	topicNameForMessage := gossipTopicNames[0]
	if err := p2.TopicGossips[topicNameForMessage].Publish(ctx, testMessage); err != nil {
		t.Fatalf("error while publishing message: %v", err)
	}
	message := <-p1.TopicGossips[topicNameForMessage].Messages
	if message.Message != testMessage {
		t.Fatalf("received wrong message: %s", message.Message)
	}
	if message.SenderID != p2.HostID() {
		t.Fatalf("received message with wrong sender: %s", message.SenderID)
	}
}
