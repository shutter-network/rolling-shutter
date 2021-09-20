package p2p

import (
	"context"
	"crypto/rand"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p-core/crypto"
	peer "github.com/libp2p/go-libp2p-peer"
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

	numPeers := 2
	privKeys := []crypto.PrivKey{}
	listenAddrs := []multiaddr.Multiaddr{}
	nodeAddrs := []multiaddr.Multiaddr{}
	firstPort := 2000
	for i := 0; i < numPeers; i++ {
		privKey, _, err := crypto.GenerateEd25519Key(rand.Reader)
		assert.NilError(t, err)
		privKeys = append(privKeys, privKey)

		pid, err := peer.IDFromPrivateKey(privKey)
		assert.NilError(t, err)
		port := firstPort + i
		listenAddr, err := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/127.0.0.1/tcp/%d", port))
		listenAddrs = append(listenAddrs, listenAddr)
		assert.NilError(t, err)
		nodeAddr, err := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/127.0.0.1/tcp/%d/p2p/%s", port, pid.String()))
		assert.NilError(t, err)
		nodeAddrs = append(nodeAddrs, nodeAddr)
	}

	gossipTopicNames := []string{"testTopic1", "testTopic2"}
	testMessage := "test message"

	p2ps := []*P2P{}
	for i := 0; i < numPeers; i++ {
		p := NewP2P(Config{
			ListenAddr:     listenAddrs[i],
			PeerMultiaddrs: nodeAddrs,
			PrivKey:        privKeys[i],
		})
		p2ps = append(p2ps, p)

		go func() {
			err := p.Run(ctx, gossipTopicNames)
			assert.NilError(t, err)
		}()
	}

	// The following loop publishes the same message over and over. Even though we did call
	// ConnectToPeer, libp2p takes some time until the peer receives the first message.
	var message *Message
	topicName := gossipTopicNames[0]
	for message == nil {
		if err := p2ps[1].Publish(ctx, topicName, testMessage); err != nil {
			t.Fatalf("error while publishing message: %v", err)
		}

		select {
		case message = <-p2ps[0].GossipMessages:
			log.Println("got", message)
			if message == nil {
				t.Fatalf("channel closed unexpectedly")
			}
		case <-ctx.Done():
			t.Fatalf("waiting for message: %s", ctx.Err())
		case <-time.After(5 * time.Millisecond):
		}
	}
	assert.Equal(t, testMessage, message.Message, "received wrong message")
	assert.Equal(t, p2ps[1].HostID(), message.SenderID, "received message with wrong sender")
}
