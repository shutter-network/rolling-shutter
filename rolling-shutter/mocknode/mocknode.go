package mocknode

import (
	"context"
	"crypto/rand"
	"log"
	"time"

	bn256 "github.com/ethereum/go-ethereum/crypto/bn256/cloudflare"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/proto"

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

func (m *MockNode) Run(ctx context.Context) error {
	p2pConfig := p2p.Config{
		ListenAddr:     m.Config.ListenAddress,
		PeerMultiaddrs: m.Config.PeerMultiaddrs,
		PrivKey:        m.Config.P2PKey,
	}
	m.p2p = p2p.NewP2P(p2pConfig)

	g, errctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		return m.p2p.Run(errctx, gossipTopicNames[:])
	})
	g.Go(func() error {
		return m.listen(errctx)
	})
	g.Go(func() error {
		return m.sendMessages(errctx)
	})
	return g.Wait()
}

func (m *MockNode) listen(ctx context.Context) error {
	for {
		select {
		case msg := <-m.p2p.GossipMessages:
			log.Printf("received message on topic %s from %s: %X", msg.Topic, msg.SenderID, msg.Message)
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
	msg := &shmsg.DecryptionTrigger{
		InstanceID: m.Config.InstanceID,
		EpochID:    epochID,
	}
	msgBytes, err := proto.Marshal(msg)
	if err != nil {
		return err
	}
	return m.p2p.Publish(ctx, "decryptionTrigger", msgBytes)
}

func (m *MockNode) sendCipherBatchMessage(ctx context.Context, epochID uint64) error {
	log.Printf("sending cipher batch for epoch %d", epochID)
	data := make([]byte, 8)
	_, err := rand.Read(data)
	if err != nil {
		return errors.Wrapf(err, "failed to generate random batch data")
	}
	msg := &shmsg.CipherBatch{
		InstanceID: m.Config.InstanceID,
		EpochID:    epochID,
		Data:       data,
	}
	msgBytes, err := proto.Marshal(msg)
	if err != nil {
		return err
	}
	return m.p2p.Publish(ctx, "cipherBatch", msgBytes)
}

func (m *MockNode) sendDecryptionKey(ctx context.Context, epochID uint64) error {
	log.Printf("sending decryption key for epoch %d", epochID)
	_, g1, err := bn256.RandomG1(rand.Reader)
	if err != nil {
		return errors.Wrapf(err, "failed to generate random decryption key")
	}
	msg := &shmsg.DecryptionKey{
		InstanceID: m.Config.InstanceID,
		EpochID:    epochID,
		Key:        g1.Marshal(),
	}
	msgBytes, err := proto.Marshal(msg)
	if err != nil {
		return err
	}
	return m.p2p.Publish(ctx, "decryptionKey", msgBytes)
}
