package decryptor

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/libp2p/go-libp2p-core/peer"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/proto"

	"github.com/shutter-network/shutter/shlib/shcrypto/shbls"
	"github.com/shutter-network/shutter/shuttermint/decryptor/dcrdb"
	"github.com/shutter-network/shutter/shuttermint/decryptor/dcrtopics"
	"github.com/shutter-network/shutter/shuttermint/p2p"
	"github.com/shutter-network/shutter/shuttermint/shdb"
	"github.com/shutter-network/shutter/shuttermint/shmsg"
)

var gossipTopicNames = [3]string{dcrtopics.CipherBatch, dcrtopics.DecryptionKey, dcrtopics.DecryptionSignature}

type Decryptor struct {
	Config Config

	p2p        *p2p.P2P
	db         *dcrdb.Queries
	instanceID uint64
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

		p2p:        p,
		db:         nil,
		instanceID: config.InstanceID,
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
	case *shmsg.DecryptionKey:
		msgsOut, err = handleDecryptionKeyInput(ctx, d.Config, d.db, typedMsg)
	case *shmsg.CipherBatch:
		msgsOut, err = handleCipherBatchInput(ctx, d.Config, d.db, typedMsg)
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
	instanceIDValidator := makeInstanceIDValidator(d.instanceID)
	for _, topicName := range gossipTopicNames {
		validators[topicName] = instanceIDValidator
	}

	return validators
}

func makeInstanceIDValidator(instanceID uint64) pubsub.Validator {
	return func(ctx context.Context, peerID peer.ID, libp2pMessage *pubsub.Message) bool {
		p2pMessage := new(p2p.Message)
		if err := json.Unmarshal(libp2pMessage.Data, p2pMessage); err != nil {
			return false
		}
		msg, err := unmarshalP2PMessage(p2pMessage)
		if err != nil {
			return false
		}
		return msg.GetInstanceID() == instanceID
	}
}

func unmarshalP2PMessage(msg *p2p.Message) (shmsg.P2PMessage, error) {
	if msg == nil {
		return nil, nil
	}
	switch msg.Topic {
	case dcrtopics.DecryptionKey:
		decryptionKeyMsg := shmsg.DecryptionKey{}
		if err := proto.Unmarshal(msg.Message, &decryptionKeyMsg); err != nil {
			return nil, errors.Wrap(err, "failed to unmarshal decryption key message")
		}
		return &decryptionKeyMsg, nil
	case dcrtopics.CipherBatch:
		cipherBatchMsg := shmsg.CipherBatch{}
		if err := proto.Unmarshal(msg.Message, &cipherBatchMsg); err != nil {
			return nil, errors.Wrap(err, "failed to unmarshal cipher batch message")
		}
		return &cipherBatchMsg, nil
	case dcrtopics.DecryptionSignature:
		decryptionSignature := shmsg.AggregatedDecryptionSignature{}
		if err := proto.Unmarshal(msg.Message, &decryptionSignature); err != nil {
			return nil, errors.Wrap(err, "failed to unmarshal decryption signature message")
		}
		return &decryptionSignature, nil
	default:
		return nil, &unhandledTopicError{msg.Topic, "unhandled topic from message"}
	}
}

type unhandledTopicError struct {
	topic string
	msg   string
}

func (e *unhandledTopicError) Error() string {
	return fmt.Sprintf("%s: %s", e.msg, e.topic)
}
