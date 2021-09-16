package mocknode

import (
	"context"
	"log"
	"time"

	crypto "github.com/libp2p/go-libp2p-crypto"
	"github.com/multiformats/go-multiaddr"
	"golang.org/x/sync/errgroup"

	"github.com/shutter-network/shutter/shuttermint/p2p"
)

var gossipTopicNames = [1]string{"cipherBatch"}

type MockNode struct {
	Config Config

	p2p *p2p.P2P
}

type Config struct {
	ListenAddress  multiaddr.Multiaddr
	PeerMultiaddrs []multiaddr.Multiaddr
	P2PKey         crypto.PrivKey

	Rate              float64
	SendCipherBatches bool
}

func (m *MockNode) Run(ctx context.Context) error {
	m.p2p = p2p.NewP2PWithKey(m.Config.P2PKey)

	if err := m.p2p.CreateHost(ctx, m.Config.ListenAddress); err != nil {
		return err
	}
	if err := m.p2p.JoinTopics(ctx, gossipTopicNames[:]); err != nil {
		return err
	}
	if err := m.p2p.ConnectToPeers(ctx, m.Config.PeerMultiaddrs); err != nil {
		return err
	}

	g, errctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		return m.listen(errctx)
	})
	g.Go(func() error {
		return m.sendMessages(errctx)
	})
	return g.Wait()
}

func (m *MockNode) listen(ctx context.Context) error {
	messages := make(chan *p2p.Message)
	for _, topic := range m.p2p.TopicGossips {
		go func(t *p2p.TopicGossip) {
			for {
				select {
				case msg := <-t.Messages:
					messages <- msg
				case <-ctx.Done():
					return
				}
			}
		}(topic)
	}

	for {
		select {
		case msg := <-messages:
			log.Println("received message", msg.Message, msg.SenderID)
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (m *MockNode) sendMessages(ctx context.Context) error {
	sleepDuration := time.Duration(1000/m.Config.Rate) * time.Millisecond

	for {
		select {
		case <-time.After(sleepDuration):
			log.Println("sending message")
			err := m.p2p.TopicGossips["cipherBatch"].Publish(ctx, "test cipher batch")
			if err != nil {
				return err
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
