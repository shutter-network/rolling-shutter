package decryptor

import (
	"context"
	"log"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/proto"

	"github.com/shutter-network/shutter/shuttermint/decryptor/dcrdb"
	"github.com/shutter-network/shutter/shuttermint/p2p"
	"github.com/shutter-network/shutter/shuttermint/shmsg"
)

var gossipTopicNames = [3]string{"cipherBatch", "decryptionKey", "decryptionSignature"}

type Decryptor struct {
	Config Config

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

		p2p: p,
		db:  nil,
	}
}

func (d *Decryptor) Run(ctx context.Context) error {
	dbpool, err := pgxpool.Connect(ctx, d.Config.DatabaseURL)
	if err != nil {
		return errors.Wrap(err, "failed to connect to database")
	}
	defer dbpool.Close()

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
		return d.p2p.Run(errorctx, gossipTopicNames[:])
	})
	return errorgroup.Wait()
}

func (d *Decryptor) handleMessages(ctx context.Context) error {
	for {
		select {
		case msg := <-d.p2p.GossipMessages:
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

	switch msg.Topic {
	case "decryptionKey":
		decryptionKeyMsg := shmsg.DecryptionKey{}
		if err := proto.Unmarshal(msg.Message, &decryptionKeyMsg); err != nil {
			return errors.Wrap(err, "failed to unmarshal decryption key message")
		}
		msgsOut, err = handleDecryptionKeyInput(ctx, d.db, &decryptionKeyMsg)
	case "cipherBatch":
		cipherBatchMsg := shmsg.CipherBatch{}
		if err := proto.Unmarshal(msg.Message, &cipherBatchMsg); err != nil {
			return errors.Wrap(err, "failed to unmarshal cipher batch message")
		}
		msgsOut, err = handleCipherBatchInput(ctx, d.db, &cipherBatchMsg)
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
	var err error
	var topic string
	var msgBytes []byte

	switch msgTyped := msg.(type) {
	case *shmsg.AggregatedDecryptionSignature:
		topic = "decryptionSignature"
		msgBytes, err = proto.Marshal(msgTyped)
		if err != nil {
			return errors.Wrap(err, "failed to marshal decryption signature message")
		}
	default:
		return errors.Errorf("received output message of unknown type: %T", msgTyped)
	}

	return d.p2p.Publish(ctx, topic, msgBytes)
}
