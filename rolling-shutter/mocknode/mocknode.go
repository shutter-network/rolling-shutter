package mocknode

import (
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
	"github.com/shutter-network/shutter/shuttermint/collator/client"
	"github.com/shutter-network/shutter/shuttermint/keyper/kprtopics"
	"github.com/shutter-network/shutter/shuttermint/p2p"
	"github.com/shutter-network/shutter/shuttermint/shdb"
	"github.com/shutter-network/shutter/shuttermint/shmsg"
)

var gossipTopicNames = []string{
	kprtopics.DecryptionTrigger,
	kprtopics.EonPublicKey,
	kprtopics.DecryptionKey,
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
		return m.p2p.Run(errctx, gossipTopicNames, make(map[string]pubsub.Validator))
	})
	g.Go(func() error {
		return m.listen(errctx)
	})
	g.Go(func() error {
		return m.sendMessages(errctx)
	})

	if m.Config.SendTransactions {
		g.Go(func() error {
			return m.sendTransactions(errctx)
		})
	}
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
		case msg, ok := <-m.p2p.GossipMessages:
			if !ok {
				return nil
			}
			m.handleMessage(msg)
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (m *MockNode) handleMessage(plainMsg *p2p.Message) {
	switch plainMsg.Topic {
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

func (m *MockNode) sendTransactions(ctx context.Context) error {
	for {
		sleepDuration := time.Duration(rand.ExpFloat64() / m.Config.Rate * float64(time.Second))
		select {
		case <-time.After(sleepDuration):
			httpResponse, err := m.collatorClient.GetNextEpoch(ctx)
			if err != nil {
				log.Printf("Error while calling next-epoch: %s", err)
				continue
			}
			nextEpochResponse, err := client.ParseGetNextEpochResponse(httpResponse)
			if err != nil {
				log.Printf("111 Error while calling next-epoch: %s", err)
				continue
			}
			if nextEpochResponse.JSONDefault != nil {
				log.Printf("222 Error while calling next-epoch: %v", nextEpochResponse.JSONDefault)
				continue
			}

			if nextEpochResponse.JSON200 == nil {
				log.Printf("333 Error getting next-epoch: %+v", nextEpochResponse)
				continue
			}

			epochID := shdb.DecodeUint64(nextEpochResponse.JSON200.Id)
			_, encryptedTx, err := encryptRandomMessage(epochID, m.eonPublicKey)
			if err != nil {
				return err
			}
			httpResponse, err = m.collatorClient.SubmitTransaction(
				ctx,
				client.SubmitTransactionJSONRequestBody{
					EncryptedTx: encryptedTx,
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
				log.Printf("Submitted tx for epoch %d: %x", epochID, response.JSON200.Id)
			} else if response.JSONDefault != nil {
				log.Printf("Error submitting tx for epoch %d: %v", epochID, response.JSONDefault)
			} else {
				log.Printf("Error submitting tx for epoch %d: %s", epochID, response.Status())
			}
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
	return m.p2p.Publish(ctx, msg.Topic(), msgBytes)
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
	return m.p2p.Publish(ctx, msg.Topic(), msgBytes)
}
