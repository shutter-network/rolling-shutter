package p2p

import (
	"bytes"
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/rs/zerolog/log"
	"gotest.tools/assert"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/encodeable"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/encodeable/address"
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
	configs := []*Config{}
	bootstrapAddrs := []*address.P2PAddress{}

	firstPort := 2000
	for i := 0; i < numBootstrappers; i++ {
		cfg := NewConfig()
		err := cfg.SetExampleValues()
		assert.NilError(t, err)

		port := firstPort + i

		addr := &address.P2PAddress{}
		err = encodeable.FromString(addr, fmt.Sprintf("/ip4/127.0.0.1/tcp/%d", port))
		assert.NilError(t, err)
		cfg.ListenAddresses = []*address.P2PAddress{addr}

		pid, err := cfg.P2PKey.PeerID()
		assert.NilError(t, err)
		externalAddr := &address.P2PAddress{}
		err = encodeable.FromString(
			externalAddr,
			fmt.Sprintf("/ip4/127.0.0.1/tcp/%d/p2p/%s", port, pid),
		)
		assert.NilError(t, err)
		bootstrapAddrs = append(bootstrapAddrs, externalAddr)
		configs = append(configs, cfg)
	}

	for _, cfg := range configs {
		cfg.CustomBootstrapAddresses = bootstrapAddrs
	}

	for i := 0; i < numPeers; i++ {
		cfg := NewConfig()
		err := cfg.SetExampleValues()
		assert.NilError(t, err)

		port := firstPort + numBootstrappers + i

		addr := &address.P2PAddress{}
		err = encodeable.FromString(addr, fmt.Sprintf("/ip4/127.0.0.1/tcp/%d", port))
		assert.NilError(t, err)
		cfg.ListenAddresses = []*address.P2PAddress{addr}
		cfg.CustomBootstrapAddresses = bootstrapAddrs

		configs = append(configs, cfg)
	}

	gossipTopicNames := []string{"testTopic1", "testTopic2"}
	testMessage := []byte("test message")

	runctx, stopRun := context.WithCancel(ctx)

	waitGroup := sync.WaitGroup{}
	p2ps := []*P2PNode{}
	for _, cfg := range configs {
		p2pHandler, err := New(cfg)
		assert.NilError(t, err)
		p2ps = append(p2ps, p2pHandler.P2P)
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			err := p2pHandler.P2P.Run(runctx, gossipTopicNames, map[string]pubsub.ValidatorEx{})
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
	assert.Equal(
		t,
		p2ps[1].HostID(),
		message.GetFrom().String(),
		"received message with wrong sender",
	)
}
