package decryptor

import (
	"context"
	"log"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/proto"

	"github.com/shutter-network/shutter/shlib/shcrypto/shbls"
	"github.com/shutter-network/shutter/shuttermint/chainobserver"
	"github.com/shutter-network/shutter/shuttermint/contract/deployment"
	"github.com/shutter-network/shutter/shuttermint/decryptor/dcrdb"
	"github.com/shutter-network/shutter/shuttermint/decryptor/dcrtopics"
	"github.com/shutter-network/shutter/shuttermint/medley/eventsyncer"
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

	p2p    *p2p.P2P
	dbpool *pgxpool.Pool
	db     *dcrdb.Queries
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

		p2p:    p,
		dbpool: nil,
		db:     nil,
	}
}

func (d *Decryptor) Run(ctx context.Context) error {
	log.Printf(
		"starting decryptor with signing public key %X",
		shbls.SecretToPublicKey(d.Config.SigningKey).Marshal(),
	)

	dbpool, err := pgxpool.Connect(ctx, d.Config.DatabaseURL)
	if err != nil {
		return errors.Wrap(err, "failed to connect to database")
	}
	defer dbpool.Close()
	d.dbpool = dbpool
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

func (d *Decryptor) handleContractEvents(ctx context.Context) error {
	events := []*eventsyncer.EventType{
		d.contracts.KeypersConfigsListNewConfig,
		d.contracts.DecryptorsConfigsListNewConfig,
		d.contracts.BLSRegistryRegistered,
		d.contracts.CollatorConfigsListNewConfig,
	}
	return chainobserver.New(d.contracts, d.dbpool).Observe(ctx, events)
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
	log.Printf("sending %s", msg.ProtoReflect().Descriptor().FullName().Name())

	return d.p2p.Publish(ctx, msg.Topic(), msgBytes)
}
