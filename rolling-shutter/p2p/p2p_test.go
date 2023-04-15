package p2p

import (
	"bytes"
	"context"
	"crypto/rand"
	"fmt"
	"sync"
	"testing"
	"time"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	p2pcrypto "github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
	"github.com/rs/zerolog/log"
	"gotest.tools/assert"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/testlog"
)

func init() {
	testlog.Setup()
}

// TestStartNetworkNode test that we can init two p2p nodes and make them send/receive messages.
func TestStartNetworkNodeIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 1200*time.Millisecond)
	defer cancel()

	numBootstrappers := 2
	numPeers := 2
	numNodes := numPeers + numBootstrappers

	privKeys := []p2pcrypto.PrivKey{}
	listenAddrs := []multiaddr.Multiaddr{}
	bootStrapAddrs := []peer.AddrInfo{}
	isBootstrappers := []bool{}
	firstPort := 2000
	for i := 0; i < numNodes; i++ {
		privKey, _, err := p2pcrypto.GenerateEd25519Key(rand.Reader)
		assert.NilError(t, err)
		privKeys = append(privKeys, privKey)

		pid, err := peer.IDFromPrivateKey(privKey)
		assert.NilError(t, err)
		port := firstPort + i
		listenAddr, err := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/127.0.0.1/tcp/%d", port))
		listenAddrs = append(listenAddrs, listenAddr)
		assert.NilError(t, err)
		if i >= numNodes-numBootstrappers {
			isBootstrappers = append(isBootstrappers, true)
			nodeAddr, err := peer.AddrInfoFromString(fmt.Sprintf("/ip4/127.0.0.1/tcp/%d/p2p/%s", port, pid.String()))
			assert.NilError(t, err)
			bootStrapAddrs = append(bootStrapAddrs, *nodeAddr)
		} else {
			isBootstrappers = append(isBootstrappers, false)
		}
	}

	gossipTopicNames := []string{"testTopic1", "testTopic2"}
	testMessage := []byte("test message")

	runctx, stopRun := context.WithCancel(ctx)

	waitGroup := sync.WaitGroup{}
	p2ps := []*P2PNode{}
	for i := 0; i < numNodes; i++ {
		p := New(Config{
			ListenAddrs:     []multiaddr.Multiaddr{listenAddrs[i]},
			BootstrapPeers:  bootStrapAddrs,
			PrivKey:         privKeys[i],
			Environment:     Local,
			IsBootstrapNode: isBootstrappers[i],
		}).P2P
		p2ps = append(p2ps, p)
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			err := p.Run(runctx, gossipTopicNames, map[string]pubsub.ValidatorEx{})
			assert.Assert(t, err == context.Canceled)
		}()
	}
	defer func() {
		stopRun()
		waitGroup.Wait()
	}()
	// The following loop publishes the same message over and over. Even though we did call
	// ConnectToPeer, libp2p takes some time until the peer receives the first message.
	var message *pubsub.Message
	topicName := gossipTopicNames[0]
	for message == nil {
		if err := p2ps[1].Publish(ctx, topicName, testMessage); err != nil {
			t.Fatalf("error while publishing message: %v", err)
		}

		select {
		case message = <-p2ps[0].GossipMessages:
			log.Info().Interface("message", message).Msg("got message")
			if message == nil {
				t.Fatalf("channel closed unexpectedly")
			}
		case <-ctx.Done():
			t.Fatalf("waiting for message: %s", ctx.Err())
		case <-time.After(5 * time.Millisecond):
		}
	}
	assert.Equal(t, topicName, message.GetTopic(), "received message with wrong topic")
	assert.Check(t, bytes.Equal(testMessage, message.GetData()), "received wrong message")
	assert.Equal(t, p2ps[1].HostID(), message.GetFrom().String(), "received message with wrong sender")
}
