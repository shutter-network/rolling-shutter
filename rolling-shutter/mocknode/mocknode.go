package mocknode

import (
	"context"
	"crypto/rand"
	"log"
	"time"

	crypto "github.com/libp2p/go-libp2p-crypto"
	"github.com/multiformats/go-multiaddr"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bn256"
	"golang.org/x/sync/errgroup"

	"github.com/shutter-network/shutter/shuttermint/p2p"
	"github.com/shutter-network/shutter/shuttermint/shmsg"
)

var gossipTopicNames = [3]string{
	"decryptionTrigger",
	"cipherBatch",
	"decryptionKey",
}

type MockNode struct {
	Config Config

	p2p *p2p.P2P
}

type Config struct {
	ListenAddress  multiaddr.Multiaddr
	PeerMultiaddrs []multiaddr.Multiaddr
	P2PKey         crypto.PrivKey

	Rate                   float64
	SendDecryptionTriggers bool
	SendCipherBatches      bool
	SendDecryptionKeys     bool
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
			log.Printf("received message from %s: %s", msg.SenderID, msg.Message)
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (m *MockNode) sendMessages(ctx context.Context) error {
	sleepDuration := time.Duration(1000/m.Config.Rate) * time.Millisecond

	epochID := uint64(0)
	for {
		select {
		case <-time.After(sleepDuration):
			if err := m.sendMessagesForEpoch(ctx, epochID); err != nil {
				return err
			}
			epochID++
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (m *MockNode) sendMessagesForEpoch(ctx context.Context, epochID uint64) error {
	if m.Config.SendDecryptionTriggers {
		if err := m.sendDecryptionTrigger(ctx, epochID); err != nil {
			return err
		}
	}
	if m.Config.SendCipherBatches {
		if err := m.sendCipherBatchMessage(ctx, epochID); err != nil {
			return err
		}
	}
	if m.Config.SendDecryptionKeys {
		if err := m.sendDecryptionKey(ctx, epochID); err != nil {
			return err
		}
	}
	return nil
}

func (m *MockNode) sendDecryptionTrigger(ctx context.Context, epochID uint64) error {
	log.Printf("sending decryption trigger for epoch %d", epochID)
	msg := shmsg.DecryptionTrigger{
		InstanceID: 0,
		EpochID:    epochID,
	}
	return m.p2p.TopicGossips["decryptionTrigger"].Publish(ctx, msg.String())
}

func (m *MockNode) sendCipherBatchMessage(ctx context.Context, epochID uint64) error {
	log.Printf("sending cipher batch for epoch %d", epochID)
	data := make([]byte, 8)
	rand.Read(data)
	msg := shmsg.CipherBatch{
		InstanceID: 0,
		EpochID:    epochID,
		Data:       data,
	}
	return m.p2p.TopicGossips["cipherBatch"].Publish(ctx, msg.String())
}

func (m *MockNode) sendDecryptionKey(ctx context.Context, epochID uint64) error {
	log.Printf("sending decryption key for epoch %d", epochID)
	_, g1, err := bn256.RandomG1(rand.Reader)
	if err != nil {
		return errors.Wrapf(err, "failed to generate random decryption key")
	}
	msg := shmsg.DecryptionKey{
		InstanceID: 0,
		EpochID:    epochID,
		Key:        g1.Marshal(),
	}
	return m.p2p.TopicGossips["decryptionKey"].Publish(ctx, msg.String())
}
