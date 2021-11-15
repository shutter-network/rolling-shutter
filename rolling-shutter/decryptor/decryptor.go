package decryptor

import (
	"context"
	"encoding/json"
	"log"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/libp2p/go-libp2p-core/peer"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/proto"

	"github.com/shutter-network/shutter/shlib/shcrypto"
	"github.com/shutter-network/shutter/shlib/shcrypto/shbls"
	"github.com/shutter-network/shutter/shuttermint/contract/deployment"
	"github.com/shutter-network/shutter/shuttermint/decryptor/dcrdb"
	"github.com/shutter-network/shutter/shuttermint/decryptor/dcrtopics"
	"github.com/shutter-network/shutter/shuttermint/medley/bitfield"
	"github.com/shutter-network/shutter/shuttermint/p2p"
	"github.com/shutter-network/shutter/shuttermint/shdb"
	"github.com/shutter-network/shutter/shuttermint/shmsg"
)

var gossipTopicNames = [4]string{
	dcrtopics.CipherBatch,
	dcrtopics.DecryptionKey,
	dcrtopics.DecryptionSignature,
	dcrtopics.AggregatedDecryptionSignature,
}

type Decryptor struct {
	Config Config

	contracts *deployment.Contracts

	p2p *p2p.P2P
	db  *dcrdb.Queries
}

func New(config Config) *Decryptor {
	p2pConfig := p2p.Config{
		ListenAddr:     config.ListenAddress,
		PeerMultiaddrs: config.PeerMultiaddrs,
		PrivKey:        config.P2PKey,
	}
	p := p2p.New(p2pConfig)

	return &Decryptor{
		Config: config,

		contracts: nil,

		p2p: p,
		db:  nil,
	}
}

func (d *Decryptor) Run(ctx context.Context) error {
	log.Printf(
		"starting keyper with signing public key %X",
		shbls.SecretToPublicKey(d.Config.SigningKey).Marshal(),
	)

	dbpool, err := pgxpool.Connect(ctx, d.Config.DatabaseURL)
	if err != nil {
		return errors.Wrap(err, "failed to connect to database")
	}
	defer dbpool.Close()
	log.Printf("Connected to database (%s)", shdb.ConnectionInfo(dbpool))

	ethereumClient, err := ethclient.Dial(d.Config.EthereumURL)
	if err != nil {
		return err
	}
	contracts, err := deployment.NewContracts(ethereumClient, d.Config.DeploymentDir)
	if err != nil {
		return err
	}
	d.contracts = contracts

	err = dcrdb.ValidateDecryptorDB(ctx, dbpool)
	if err != nil {
		return err
	}
	db := dcrdb.New(dbpool)
	d.db = db

	errorgroup, errorctx := errgroup.WithContext(ctx)
	errorgroup.Go(func() error {
		return d.handleMessages(errorctx)
	})
	errorgroup.Go(func() error {
		return d.handleContractEvents(errorctx)
	})

	topicValidators := d.makeMessagesValidators()

	errorgroup.Go(func() error {
		return d.p2p.Run(errorctx, gossipTopicNames[:], topicValidators)
	})
	return errorgroup.Wait()
}

func (d *Decryptor) handleMessages(ctx context.Context) error {
	for {
		select {
		case msg, ok := <-d.p2p.GossipMessages:
			if !ok {
				return nil
			}
			if err := d.handleMessage(ctx, msg); err != nil {
				log.Printf("error handling message %+v: %s", msg, err)
				continue
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (d *Decryptor) handleMessage(ctx context.Context, msg *p2p.Message) error {
	var msgsOut []shmsg.P2PMessage
	var err error

	unmarshalled, err := unmarshalP2PMessage(msg)
	if topicError, ok := err.(*unhandledTopicError); ok {
		log.Println(topicError.Error())
	} else if err != nil {
		return err
	}

	switch typedMsg := unmarshalled.(type) {
	case *decryptionKey:
		msgsOut, err = handleDecryptionKeyInput(ctx, d.Config, d.db, typedMsg)
	case *cipherBatch:
		msgsOut, err = handleCipherBatchInput(ctx, d.Config, d.db, typedMsg)
	case *decryptionSignature:
		msgsOut, err = handleSignatureInput(ctx, d.Config, d.db, typedMsg)
	case *aggregatedDecryptionSignature:
		return nil
	default:
		log.Println("ignoring message received on topic", msg.Topic)
		return nil
	}

	if err != nil {
		return err
	}
	for _, msgOut := range msgsOut {
		if err := d.sendMessage(ctx, msgOut); err != nil {
			log.Printf("error sending message %+v: %s", msgOut, err)
			continue
		}
	}
	return nil
}

func (d *Decryptor) sendMessage(ctx context.Context, msg shmsg.P2PMessage) error {
	msgBytes, err := proto.Marshal(msg)
	if err != nil {
		return errors.Wrap(err, "failed to marshal p2p message")
	}

	return d.p2p.Publish(ctx, msg.Topic(), msgBytes)
}

func (d *Decryptor) makeMessagesValidators() map[string]pubsub.Validator {
	validators := make(map[string]pubsub.Validator)
	validators[dcrtopics.DecryptionSignature] = d.makeDecryptionSignatureValidator()
	validators[dcrtopics.AggregatedDecryptionSignature] = d.makeAggregatedDecryptionSignatureValidator()
	validators[dcrtopics.CipherBatch] = d.makeInstanceIDValidator()
	validators[dcrtopics.DecryptionKey] = d.makeDecryptionKeyValidator()

	return validators
}

func (d *Decryptor) makeInstanceIDValidator() pubsub.Validator {
	return func(ctx context.Context, peerID peer.ID, libp2pMessage *pubsub.Message) bool {
		p2pMessage := new(p2p.Message)
		if err := json.Unmarshal(libp2pMessage.Data, p2pMessage); err != nil {
			return false
		}
		msg, err := unmarshalP2PMessage(p2pMessage)
		if err != nil {
			return false
		}
		return msg.GetInstanceID() == d.Config.InstanceID
	}
}

func (d *Decryptor) makeDecryptionKeyValidator() pubsub.Validator {
	return func(ctx context.Context, peerID peer.ID, libp2pMessage *pubsub.Message) bool {
		p2pMessage := new(p2p.Message)
		if err := json.Unmarshal(libp2pMessage.Data, p2pMessage); err != nil {
			return false
		}
		msg, err := unmarshalP2PMessage(p2pMessage)
		if err != nil {
			return false
		}
		if msg.GetInstanceID() != d.Config.InstanceID {
			return false
		}

		key, ok := msg.(*decryptionKey)
		if !ok {
			panic("unmarshalled non decryption key message in decryption key validator")
		}

		eonPublicKeyBytes, err := d.db.GetEonPublicKey(ctx, shdb.EncodeUint64(key.epochID))
		if err == pgx.ErrNoRows {
			log.Printf("received decryption key for epoch %d for which we don't have an eon public key", key.epochID)
			return false
		}
		if err != nil {
			log.Printf("error while getting eon public key from database for epoch ID %v", key.epochID)
			return false
		}
		eonPublicKey := new(shcrypto.EonPublicKey)
		err = eonPublicKey.Unmarshal(eonPublicKeyBytes)
		if err != nil {
			log.Printf("error while unmarshalling eon public key for epoch %v", key.epochID)
			return false
		}
		ok, err = shcrypto.VerifyEpochSecretKey(key.key, eonPublicKey, key.epochID)
		if err != nil {
			log.Printf("error while checking epoch secret key for epoch %v", key.epochID)
			return false
		}
		return ok
	}
}

func (d *Decryptor) makeDecryptionSignatureValidator() pubsub.Validator {
	return func(ctx context.Context, peerID peer.ID, libp2pMessage *pubsub.Message) bool {
		p2pMessage := new(p2p.Message)
		if err := json.Unmarshal(libp2pMessage.Data, p2pMessage); err != nil {
			return false
		}
		msg, err := unmarshalP2PMessage(p2pMessage)
		if err != nil {
			return false
		}

		if msg.GetInstanceID() != d.Config.InstanceID {
			return false
		}

		signature, ok := msg.(*decryptionSignature)
		if !ok {
			panic("unmarshalled non signature message in signature validator")
		}

		decryptorIndexes := bitfield.GetIndexes(signature.SignerBitfield)
		if len(decryptorIndexes) != 1 {
			return false
		}
		dbKey, err := d.db.GetDecryptorKey(ctx, dcrdb.GetDecryptorKeyParams{
			Index:        decryptorIndexes[0],
			StartEpochID: shdb.EncodeUint64(signature.epochID),
		})
		if err == pgx.ErrNoRows {
			return false
		}
		if err != nil {
			log.Printf("error while getting decryption key from database: %s", err)
			return false
		}
		key := new(shbls.PublicKey)
		if err := key.Unmarshal(dbKey); err != nil {
			return false
		}

		return shbls.Verify(signature.signature, key, signature.signedHash.Bytes())
	}
}

func (d *Decryptor) makeAggregatedDecryptionSignatureValidator() pubsub.Validator {
	return func(ctx context.Context, peerID peer.ID, libp2pMessage *pubsub.Message) bool {
		p2pMessage := new(p2p.Message)
		if err := json.Unmarshal(libp2pMessage.Data, p2pMessage); err != nil {
			return false
		}
		msg, err := unmarshalP2PMessage(p2pMessage)
		if err != nil {
			return false
		}

		if msg.GetInstanceID() != d.Config.InstanceID {
			return false
		}

		signature, ok := msg.(*aggregatedDecryptionSignature)
		if !ok {
			panic("unmarshalled non signature message in aggregated signature validator")
		}

		decryptorIndexes := bitfield.GetIndexes(signature.signerBitfield)
		if len(decryptorIndexes) == 0 {
			return false
		}
		keys := make([]*shbls.PublicKey, 0, len(decryptorIndexes))
		for _, decryptorIndex := range decryptorIndexes {
			dbKey, err := d.db.GetDecryptorKey(ctx, dcrdb.GetDecryptorKeyParams{
				Index:        decryptorIndex,
				StartEpochID: shdb.EncodeUint64(signature.epochID),
			})
			if err == pgx.ErrNoRows {
				return false
			}
			if err != nil {
				log.Printf("error while getting decryption key from database: %s", err)
				return false
			}
			key := new(shbls.PublicKey)
			if err := key.Unmarshal(dbKey); err != nil {
				return false
			}
			keys = append(keys, key)
		}

		aggregatedKey := shbls.AggregatePublicKeys(keys)
		return shbls.Verify(signature.signature, aggregatedKey, signature.signedHash.Bytes())
	}
}
