package mocknode

import (
	"context"
	cryptorand "crypto/rand"
	"math/big"
	"math/rand"
	"sync"
	"time"

	bn256 "github.com/ethereum/go-ethereum/crypto/bn256/cloudflare"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	txtypes "github.com/shutter-network/txtypes/types"
	"golang.org/x/sync/errgroup"

	"github.com/shutter-network/shutter/shlib/shcrypto"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/collator/client"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/kprtopics"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/epochid"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2p"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/shmsg"
)

type MockNode struct {
	Config Config

	mux sync.Mutex

	collatorClient *client.Client
	p2p            *p2p.P2PHandler

	eonSecretKeyShare *shcrypto.EonSecretKeyShare
	eonPublicKey      *shcrypto.EonPublicKey

	plainTxsSent  map[uint64][][]byte
	cipherTxsSent map[uint64][][]byte
}

func New(config Config) (*MockNode, error) {
	eonSecretKeyShare, eonPublicKey, err := computeEonKeys(config.EonKeySeed)
	if err != nil {
		return nil, err
	}

	collatorClient, err := client.NewClient("http://localhost:3000/v1")
	if err != nil {
		return nil, err
	}

	node := &MockNode{
		Config: config,

		collatorClient: collatorClient,
		p2p: p2p.New(p2p.Config{
			ListenAddrs:    config.ListenAddresses,
			BootstrapPeers: config.CustomBootstrapAddresses,
			PrivKey:        config.P2PKey,
			Environment:    p2p.Production,
		}),

		eonSecretKeyShare: eonSecretKeyShare,
		eonPublicKey:      eonPublicKey,

		plainTxsSent:  make(map[uint64][][]byte),
		cipherTxsSent: make(map[uint64][][]byte),
	}
	node.setupP2PHandler()
	return node, nil
}

func (m *MockNode) Run(ctx context.Context) error {
	if err := m.logStartupInfo(); err != nil {
		return err
	}

	g, errctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		return m.p2p.Run(errctx)
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

func (m *MockNode) setupP2PHandler() {
	p2p.AddHandlerFunc(m.p2p, m.handleEonPublicKey)

	m.p2p.AddGossipTopic(kprtopics.DecryptionTrigger)
	m.p2p.AddGossipTopic(kprtopics.DecryptionKey)
}

func (m *MockNode) logStartupInfo() error {
	eonPublicKey := m.eonPublicKey.Marshal()
	log.Info().Hex("eon-public-key", eonPublicKey).Msg("starting mocknode")
	return nil
}

func (m *MockNode) handleEonPublicKey(_ context.Context, key *shmsg.EonPublicKey) ([]shmsg.P2PMessage, error) {
	m.mux.Lock()
	defer m.mux.Unlock()
	if err := m.eonPublicKey.Unmarshal(key.PublicKey); err != nil {
		log.Info().Err(err).Msg("failed to unmarshal eon public key")
	}
	log.Info().Str("eon-public-key", (*bn256.G2)(m.eonPublicKey).String()).
		Msg("updated eon public key from messages to %s")
	return make([]shmsg.P2PMessage, 0), nil
}

func (m *MockNode) sendTransactions(ctx context.Context) error {
	for {
		sleepDuration := time.Duration(rand.ExpFloat64() / m.Config.Rate * float64(time.Second))
		select {
		case <-time.After(sleepDuration):
			httpResponse, err := m.collatorClient.GetNextEpoch(ctx)
			if err != nil {
				log.Error().Err(err).Msg("failed to get next epoch from collator")
				continue
			}
			nextEpochResponse, err := client.ParseGetNextEpochResponse(httpResponse)
			if err != nil {
				log.Error().Err(err).Msg("failed to parse response from collator")
				continue
			}
			if nextEpochResponse.JSONDefault != nil {
				jsonDefault := nextEpochResponse.JSONDefault
				log.Error().Str("message", jsonDefault.Message).Int32("status", jsonDefault.Code).
					Msg("received error response from collator")
				continue
			}

			if nextEpochResponse.JSON200 == nil {
				log.Error().Interface("response", nextEpochResponse).
					Msg("collator response did not contain epoch-id")
				continue
			}

			epochID, err := epochid.BigToEpochID(new(big.Int).SetBytes(nextEpochResponse.JSON200.Id))
			if err != nil {
				log.Error().Msg("error converting epoch-id")
			}
			_, encryptedTx, err := encryptRandomMessage(epochID, m.eonPublicKey)
			if err != nil {
				return err
			}
			httpResponse, err = m.collatorClient.SubmitTransaction(
				ctx,
				client.SubmitTransactionJSONRequestBody{
					EncryptedTx: encryptedTx,
					Epoch:       epochID.Bytes(),
				})
			if err != nil {
				return err
			}
			response, err := client.ParseSubmitTransactionResponse(httpResponse)
			if err != nil {
				return err
			}
			if response.JSON200 != nil {
				log.Info().Str("epoch-id", epochID.Hex()).Hex("transaction-id", response.JSON200.Id).
					Msg("submitted transaction")
			} else if response.JSONDefault != nil {
				jsonDefault := response.JSONDefault
				log.Error().Str("epoch-id", epochID.Hex()).Str("message", jsonDefault.Message).Int32("status", jsonDefault.Code).
					Msg("failed to submit transaction")
			} else {
				log.Info().Str("epoch-id", epochID.Hex()).Str("status", response.Status()).
					Msg("failed to submit transcation")
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (m *MockNode) sendMessages(ctx context.Context) error {
	sleepDuration := time.Duration(1000/m.Config.Rate) * time.Millisecond

	epochIDUint64 := uint64(0)
	for {
		select {
		case <-time.After(sleepDuration):
			epochID := epochid.Uint64ToEpochID(epochIDUint64)
			if err := m.sendMessagesForEpoch(ctx, epochID); err != nil {
				return err
			}
			epochIDUint64++
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func computeEonKeys(seed int64) (*shcrypto.EonSecretKeyShare, *shcrypto.EonPublicKey, error) {
	r := rand.New(rand.NewSource(seed)) //nolint:gosec // we need the seed for testing and this is a mock function
	p, err := shcrypto.RandomPolynomial(r, 0)
	if err != nil {
		return nil, nil, err
	}

	eonPublicKey := shcrypto.ComputeEonPublicKey([]*shcrypto.Gammas{p.Gammas()})

	v := p.EvalForKeyper(0)
	eonSecretKeyShare := shcrypto.ComputeEonSecretKeyShare([]*big.Int{v})
	return eonSecretKeyShare, eonPublicKey, nil
}

func computeEpochSecretKey(epochID epochid.EpochID, eonSecretKeyShare *shcrypto.EonSecretKeyShare) (*shcrypto.EpochSecretKey, error) {
	epochIDG1 := shcrypto.ComputeEpochID(epochID.Bytes())
	epochSecretKeyShare := shcrypto.ComputeEpochSecretKeyShare(eonSecretKeyShare, epochIDG1)
	return shcrypto.ComputeEpochSecretKey(
		[]int{0},
		[]*shcrypto.EpochSecretKeyShare{epochSecretKeyShare},
		1,
	)
}

func encryptRandomMessage(epochID epochid.EpochID, eonPublicKey *shcrypto.EonPublicKey) ([]byte, []byte, error) {
	message := []byte("msgXXXXX")
	_, err := cryptorand.Read(message[3:])
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to generate random batch data")
	}
	encryptedMessage, err := EncryptMessage(message, epochID, eonPublicKey)
	return message, encryptedMessage, err
}

func (m *MockNode) sendMessagesForEpoch(ctx context.Context, epochID epochid.EpochID) error {
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

func (m *MockNode) sendDecryptionTrigger(ctx context.Context, epochID epochid.EpochID) error {
	log.Info().Str("epoch-id", epochID.Hex()).Msg("sending decryption trigger")
	msg := &shmsg.DecryptionTrigger{
		InstanceID: m.Config.InstanceID,
		EpochID:    epochID.Bytes(),
	}
	return m.p2p.SendMessage(ctx, msg)
}

func (m *MockNode) sendDecryptionKey(ctx context.Context, epochID epochid.EpochID) error {
	log.Info().Str("epoch-id", epochID.Hex()).Msg("sending decryption key")

	epochSecretKey, err := computeEpochSecretKey(epochID, m.eonSecretKeyShare)
	if err != nil {
		return err
	}

	keyBytes := epochSecretKey.Marshal()

	msg := &shmsg.DecryptionKey{
		InstanceID: m.Config.InstanceID,
		EpochID:    epochID.Bytes(),
		Key:        keyBytes,
	}
	return m.p2p.SendMessage(ctx, msg)
}

func EncryptShutterPayload(payload *txtypes.ShutterPayload, epoch epochid.EpochID, eonPubKey *shcrypto.EonPublicKey) ([]byte, error) {
	var encryptedMessage []byte
	message, err := payload.Encode()
	if err != nil {
		return encryptedMessage, err
	}
	encryptedMessage, err = EncryptMessage(message, epoch, eonPubKey)
	return encryptedMessage, err
}

func EncryptMessage(message []byte, epochID epochid.EpochID, eonPublicKey *shcrypto.EonPublicKey) ([]byte, error) {
	sigma, err := shcrypto.RandomSigma(cryptorand.Reader)
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate random sigma")
	}
	epochIDG1 := shcrypto.ComputeEpochID(epochID.Bytes())
	encryptedMessage := shcrypto.Encrypt(message, eonPublicKey, epochIDG1, sigma)
	return encryptedMessage.Marshal(), nil
}
