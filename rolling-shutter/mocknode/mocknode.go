package mocknode

import (
	"context"
	"time"

	crypto "github.com/libp2p/go-libp2p-crypto"
	"github.com/multiformats/go-multiaddr"

	"github.com/shutter-network/shutter/shuttermint/p2p"
)

var gossipTopicNames = [3]string{"cipherBatch"}

type MockNode struct {
	Config Config
}

type Config struct {
	ListenAddress  multiaddr.Multiaddr
	PeerMultiaddrs []multiaddr.Multiaddr
	P2PKey         crypto.PrivKey

	Rate              float64
	SendCipherBatches bool
}

func (m *MockNode) Run(ctx context.Context) error {
	p2p := p2p.NewP2PWithKey(m.Config.P2PKey)

	if err := p2p.CreateHost(ctx, m.Config.ListenAddress); err != nil {
		return err
	}
	if err := p2p.JoinTopics(ctx, gossipTopicNames[:]); err != nil {
		return err
	}
	if err := p2p.ConnectToPeers(ctx, m.Config.PeerMultiaddrs); err != nil {
		return err
	}

	for {
		p2p.TopicGossips["cipherBatch"].Publish(ctx, "test cipher batch")
		time.Sleep(time.Duration(1000/m.Config.Rate) * time.Millisecond)
	}
}
