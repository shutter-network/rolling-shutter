package mocknode

import (
	"context"
	"crypto/rand"
	"log"
	"math/big"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/proto"

	"github.com/shutter-network/shutter/shlib/shcrypto"
	"github.com/shutter-network/shutter/shlib/shcrypto/shbls"
	"github.com/shutter-network/shutter/shuttermint/p2p"
	"github.com/shutter-network/shutter/shuttermint/shmsg"
)

var gossipTopicNames = [4]string{
	"decryptionTrigger",
	"cipherBatch",
	"decryptionKey",
	"decryptionSignature",
}

type MockNode struct {
	Config Config

	p2p *p2p.P2P

	plainTxsSent  map[uint64][][]byte
	cipherTxsSent map[uint64][][]byte
}

func New(config Config) *MockNode {
	p2pConfig := p2p.Config{
		ListenAddr:     config.ListenAddress,
		PeerMultiaddrs: config.PeerMultiaddrs,
		PrivKey:        config.P2PKey,
	}
	p := p2p.New(p2pConfig)

	return &MockNode{
		Config: config,

		p2p: p,

		plainTxsSent:  make(map[uint64][][]byte),
		cipherTxsSent: make(map[uint64][][]byte),
	}
}

func (m *MockNode) Run(ctx context.Context) error {
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
			m.handleMessage(msg)
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (m *MockNode) handleMessage(plainMsg *p2p.Message) {
	switch plainMsg.Topic {
	case "decryptionSignature":
		msg := shmsg.AggregatedDecryptionSignature{}
		if err := proto.Unmarshal(plainMsg.Message, &msg); err != nil {
			log.Printf(
				"received invalid message on topic %s from %s: %X",
				plainMsg.Topic,
				plainMsg.SenderID,
				plainMsg.Message,
			)
		}
		m.handleDecryptionSignature(&msg, plainMsg.SenderID)
	default:
		log.Printf(
			"received message on topic %s from %s: %X",
			plainMsg.Topic,
			plainMsg.SenderID,
			plainMsg.Message,
		)
	}
}

func (m *MockNode) handleDecryptionSignature(msg *shmsg.AggregatedDecryptionSignature, senderID string) {
	sig := new(shbls.Signature)
	err := sig.Unmarshal(msg.AggregatedSignature)
	if err != nil {
		log.Printf(
			"received inunmarshalable decryption signature in epoch %d from %s: %+v",
			msg.EpochID,
			senderID,
			msg,
		)
		return
	}

	validSignature := shbls.Verify(
		sig,
		m.Config.DecryptorPublicKey,
		msg.SignedHash,
	)

	if !validSignature {
		log.Printf(
			"received invalid decryption signature in epoch %d from %s: %+v",
			msg.EpochID,
			senderID,
			msg,
		)
	} else {
		log.Printf("received valid decryption signature in epoch %d", msg.EpochID)
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

func computeKeys(epochID uint64) (*shcrypto.EonPublicKey, *shcrypto.EpochSecretKey, error) {
	epochIDG1 := shcrypto.ComputeEpochID(epochID)

	p, err := shcrypto.RandomPolynomial(rand.Reader, 0)
	if err != nil {
		return nil, nil, err
	}

	eonPublicKey := shcrypto.ComputeEonPublicKey([]*shcrypto.Gammas{p.Gammas()})

	v := p.EvalForKeyper(0)
	eonSecretKeyShare := shcrypto.ComputeEonSecretKeyShare([]*big.Int{v})
	epochSecretKeyShare := shcrypto.ComputeEpochSecretKeyShare(eonSecretKeyShare, epochIDG1)
	epochSecretKey, err := shcrypto.ComputeEpochSecretKey(
		[]int{0},
		[]*shcrypto.EpochSecretKeyShare{epochSecretKeyShare},
		1,
	)
	if err != nil {
		return nil, nil, err
	}

	return eonPublicKey, epochSecretKey, nil
}

func encryptRandomMessage(epochID uint64, eonPublicKey *shcrypto.EonPublicKey) ([]byte, []byte, error) {
	message := []byte("msgXXXXX")
	_, err := rand.Read(message[3:])
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to generate random batch data")
	}

	sigma, err := shcrypto.RandomSigma(rand.Reader)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to generate random sigma")
	}

	epochIDG1 := shcrypto.ComputeEpochID(epochID)
	encryptedMessage := shcrypto.Encrypt(message, eonPublicKey, epochIDG1, sigma)

	return message, encryptedMessage.Marshal(), nil
}

func (m *MockNode) sendMessagesForEpoch(ctx context.Context, epochID uint64) error {
	eonPublicKey, epochSecretKey, err := computeKeys(epochID)
	if err != nil {
		return errors.Wrap(err, "failed to generate key pair")
	}

	if m.Config.SendDecryptionTriggers {
		if err := m.sendDecryptionTrigger(ctx, epochID); err != nil {
			return err
		}
	}
	if m.Config.SendCipherBatches {
		if err := m.sendCipherBatchMessage(ctx, epochID, eonPublicKey); err != nil {
			return err
		}
	}
	if m.Config.SendDecryptionKeys {
		if err := m.sendDecryptionKey(ctx, epochID, epochSecretKey); err != nil {
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

func (m *MockNode) sendCipherBatchMessage(ctx context.Context, epochID uint64, eonPublicKey *shcrypto.EonPublicKey) error {
	if _, ok := m.plainTxsSent[epochID]; ok {
		return errors.Errorf("cipher batch for epoch %d already sent", epochID)
	}
	log.Printf("sending cipher batch for epoch %d", epochID)

	plainTxs := [][]byte{}
	cipherTxs := [][]byte{}
	for i := 0; i < 3; i++ {
		plainTx, cipherTx, err := encryptRandomMessage(epochID, eonPublicKey)
		if err != nil {
			return err
		}
		plainTxs = append(plainTxs, plainTx)
		cipherTxs = append(cipherTxs, cipherTx)
	}

	msg := &shmsg.CipherBatch{
		InstanceID:   m.Config.InstanceID,
		EpochID:      epochID,
		Transactions: cipherTxs,
	}
	msgBytes, err := proto.Marshal(msg)
	if err != nil {
		return err
	}

	if err := m.p2p.Publish(ctx, "cipherBatch", msgBytes); err != nil {
		return err
	}

	m.plainTxsSent[epochID] = plainTxs
	m.cipherTxsSent[epochID] = cipherTxs

	return nil
}

func (m *MockNode) sendDecryptionKey(ctx context.Context, epochID uint64, epochSecretKey *shcrypto.EpochSecretKey) error {
	log.Printf("sending decryption key for epoch %d", epochID)

	keyBytes, err := epochSecretKey.GobEncode()
	if err != nil {
		return err
	}

	msg := &shmsg.DecryptionKey{
		InstanceID: m.Config.InstanceID,
		EpochID:    epochID,
		Key:        keyBytes,
	}
	msgBytes, err := proto.Marshal(msg)
	if err != nil {
		return err
	}
	return m.p2p.Publish(ctx, "decryptionKey", msgBytes)
}
