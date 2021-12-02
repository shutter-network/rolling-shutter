package mocknode

import (
	"bytes"
	"context"
	"log"
	"math/big"
	"math/rand"
	"sync"
	"time"

	bn256 "github.com/ethereum/go-ethereum/crypto/bn256/cloudflare"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/proto"

	"github.com/shutter-network/shutter/shlib/shcrypto"
	"github.com/shutter-network/shutter/shlib/shcrypto/shbls"
	"github.com/shutter-network/shutter/shuttermint/collator/client"
	"github.com/shutter-network/shutter/shuttermint/decryptor"
	"github.com/shutter-network/shutter/shuttermint/decryptor/dcrtopics"
	"github.com/shutter-network/shutter/shuttermint/keyper/kprtopics"
	"github.com/shutter-network/shutter/shuttermint/medley/bitfield"
	"github.com/shutter-network/shutter/shuttermint/p2p"
	"github.com/shutter-network/shutter/shuttermint/shdb"
	"github.com/shutter-network/shutter/shuttermint/shmsg"
)

var gossipTopicNames = [5]string{
	kprtopics.DecryptionTrigger,
	kprtopics.EonPublicKey,
	dcrtopics.CipherBatch,
	dcrtopics.DecryptionKey,
	dcrtopics.DecryptionSignature,
}

type MockNode struct {
	Config Config

	mux sync.Mutex

	collatorClient *client.Client
	p2p            *p2p.P2P

	eonSecretKeyShare *shcrypto.EonSecretKeyShare
	eonPublicKey      *shcrypto.EonPublicKey

	plainTxsSent  map[uint64][][]byte
	cipherTxsSent map[uint64][][]byte
}

func New(config Config) (*MockNode, error) {
	p2pConfig := p2p.Config{
		ListenAddr:     config.ListenAddress,
		PeerMultiaddrs: config.PeerMultiaddrs,
		PrivKey:        config.P2PKey,
	}
	p := p2p.New(p2pConfig)

	eonSecretKeyShare, eonPublicKey, err := computeEonKeys(config.EonKeySeed)
	if err != nil {
		return nil, err
	}

	collatorClient, err := client.NewClient("http://localhost:3000/v1")
	if err != nil {
		return nil, err
	}

	return &MockNode{
		Config: config,

		collatorClient: collatorClient,
		p2p:            p,

		eonSecretKeyShare: eonSecretKeyShare,
		eonPublicKey:      eonPublicKey,

		plainTxsSent:  make(map[uint64][][]byte),
		cipherTxsSent: make(map[uint64][][]byte),
	}, nil
}

func (m *MockNode) Run(ctx context.Context) error {
	if err := m.logStartupInfo(); err != nil {
		return err
	}

	g, errctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		return m.p2p.Run(errctx, gossipTopicNames[:], make(map[string]pubsub.Validator))
	})
	g.Go(func() error {
		return m.listen(errctx)
	})
	g.Go(func() error {
		return m.sendMessages(errctx)
	})
	g.Go(func() error {
		return m.sendTransactions(errctx)
	})
	return g.Wait()
}

func (m *MockNode) logStartupInfo() error {
	eonPublicKey := m.eonPublicKey.Marshal()
	log.Println("starting mocknode")
	log.Printf("eon public key: %X", eonPublicKey)
	return nil
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
	case dcrtopics.DecryptionSignature:
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
	case kprtopics.EonPublicKey:
		msg := shmsg.EonPublicKey{}
		if err := proto.Unmarshal(plainMsg.Message, &msg); err != nil {
			log.Printf(
				"received invalid message on topic %s from %s: %X",
				plainMsg.Topic,
				plainMsg.SenderID,
				plainMsg.Message,
			)
		}
		m.mux.Lock()
		defer m.mux.Unlock()
		if err := m.eonPublicKey.Unmarshal(msg.PublicKey); err != nil {
			log.Printf("error while unmarshalling eon public key: %s", err)
		}
		log.Printf("updated eon public key from messages to %s", (*bn256.G2)(m.eonPublicKey))
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
			"received not unmarshalable decryption signature in epoch %d from %s: %+v",
			msg.EpochID,
			senderID,
			msg,
		)
		return
	}

	signerIndices := bitfield.GetIndexes(msg.SignerBitfield)
	signerKeys := []*shbls.PublicKey{}
	for _, i := range signerIndices {
		signerKeys = append(signerKeys, m.Config.DecryptorPublicKeys[i])
	}
	aggregatedPublicKey := shbls.AggregatePublicKeys(signerKeys)
	validSignature := shbls.Verify(
		sig,
		aggregatedPublicKey,
		msg.SignedHash,
	)

	var expectedCipherBatch [][]byte
	var expectedDecryptedBatch [][]byte
	func() {
		m.mux.Lock()
		defer m.mux.Unlock()
		expectedCipherBatch = m.cipherTxsSent[msg.EpochID]
		expectedDecryptedBatch = m.plainTxsSent[msg.EpochID]
	}()
	expectedSigningData := decryptor.DecryptionSigningData{
		InstanceID:     m.Config.InstanceID,
		EpochID:        msg.EpochID,
		CipherBatch:    expectedCipherBatch,
		DecryptedBatch: expectedDecryptedBatch,
	}
	correctSignedHash := bytes.Equal(msg.SignedHash, expectedSigningData.Hash().Bytes())

	if validSignature && correctSignedHash {
		log.Printf("received valid decryption signature for epoch %d", msg.EpochID)
	} else {
		log.Printf(
			"received invalid decryption signature for epoch %d.\nValid signature: %t\nCorrect signed hash: %t\n%+v",
			msg.EpochID,
			validSignature,
			correctSignedHash,
			msg,
		)
	}
}

func (m *MockNode) sendTransactions(ctx context.Context) error {
	sleepDuration := time.Duration(1000/m.Config.Rate) * time.Millisecond

	epochID := uint64(0)
	for {
		select {
		case <-time.After(sleepDuration):
			httpResponse, err := m.collatorClient.SubmitTransaction(
				ctx,
				client.SubmitTransactionJSONRequestBody{
					EncryptedTx: []byte{'f', 'o', 'o'},
					Epoch:       shdb.EncodeUint64(epochID),
				})
			if err != nil {
				return err
			}
			response, err := client.ParseSubmitTransactionResponse(httpResponse)
			if err != nil {
				return err
			}
			if response.JSON200 != nil {
				log.Printf("Submitted tx %x", response.JSON200.Id)
			} else if response.JSONDefault != nil {
				log.Printf("Error submitting tx: %v", response.JSONDefault)
			} else {
				log.Printf("Error submitting tx: %s", response.Status())
			}
			epochID++
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

func computeEonKeys(seed int64) (*shcrypto.EonSecretKeyShare, *shcrypto.EonPublicKey, error) {
	r := rand.New(rand.NewSource(seed))
	p, err := shcrypto.RandomPolynomial(r, 0)
	if err != nil {
		return nil, nil, err
	}

	eonPublicKey := shcrypto.ComputeEonPublicKey([]*shcrypto.Gammas{p.Gammas()})

	v := p.EvalForKeyper(0)
	eonSecretKeyShare := shcrypto.ComputeEonSecretKeyShare([]*big.Int{v})
	return eonSecretKeyShare, eonPublicKey, nil
}

func computeEpochSecretKey(epochID uint64, eonSecretKeyShare *shcrypto.EonSecretKeyShare) (*shcrypto.EpochSecretKey, error) {
	epochIDG1 := shcrypto.ComputeEpochID(epochID)
	epochSecretKeyShare := shcrypto.ComputeEpochSecretKeyShare(eonSecretKeyShare, epochIDG1)
	return shcrypto.ComputeEpochSecretKey(
		[]int{0},
		[]*shcrypto.EpochSecretKeyShare{epochSecretKeyShare},
		1,
	)
}

func encryptRandomMessage(epochID uint64, eonPublicKey *shcrypto.EonPublicKey) ([]byte, []byte, error) {
	message := []byte("msgXXXXX")
	_, err := rand.Read(message[3:])
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to generate random batch data")
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	sigma, err := shcrypto.RandomSigma(r)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to generate random sigma")
	}

	epochIDG1 := shcrypto.ComputeEpochID(epochID)
	encryptedMessage := shcrypto.Encrypt(message, eonPublicKey, epochIDG1, sigma)

	return message, encryptedMessage.Marshal(), nil
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
	return m.p2p.Publish(ctx, kprtopics.DecryptionTrigger, msgBytes)
}

func (m *MockNode) sendCipherBatchMessage(ctx context.Context, epochID uint64) error {
	if _, ok := m.plainTxsSent[epochID]; ok {
		return errors.Errorf("cipher batch for epoch %d already sent", epochID)
	}
	log.Printf("sending cipher batch for epoch %d", epochID)

	plainTxs := [][]byte{}
	cipherTxs := [][]byte{}
	for i := 0; i < 3; i++ {
		plainTx, cipherTx, err := encryptRandomMessage(epochID, m.eonPublicKey)
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

	if err := m.p2p.Publish(ctx, dcrtopics.CipherBatch, msgBytes); err != nil {
		return err
	}

	m.mux.Lock()
	defer m.mux.Unlock()
	m.plainTxsSent[epochID] = plainTxs
	m.cipherTxsSent[epochID] = cipherTxs

	return nil
}

func (m *MockNode) sendDecryptionKey(ctx context.Context, epochID uint64) error {
	log.Printf("sending decryption key for epoch %d", epochID)

	epochSecretKey, err := computeEpochSecretKey(epochID, m.eonSecretKeyShare)
	if err != nil {
		return err
	}

	keyBytes := epochSecretKey.Marshal()

	msg := &shmsg.DecryptionKey{
		InstanceID: m.Config.InstanceID,
		EpochID:    epochID,
		Key:        keyBytes,
	}
	msgBytes, err := proto.Marshal(msg)
	if err != nil {
		return err
	}
	return m.p2p.Publish(ctx, dcrtopics.DecryptionKey, msgBytes)
}
