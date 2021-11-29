// Package keyper contains the keyper implementation
package keyper

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/libp2p/go-libp2p-core/peer"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/pkg/errors"
	"github.com/tendermint/tendermint/rpc/client"
	"github.com/tendermint/tendermint/rpc/client/http"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/proto"

	"github.com/shutter-network/shutter/shlib/shcrypto"
	"github.com/shutter-network/shutter/shuttermint/contract/deployment"
	"github.com/shutter-network/shutter/shuttermint/keyper/fx"
	"github.com/shutter-network/shutter/shuttermint/keyper/kprdb"
	"github.com/shutter-network/shutter/shuttermint/keyper/kprtopics"
	"github.com/shutter-network/shutter/shuttermint/medley"
	"github.com/shutter-network/shutter/shuttermint/p2p"
	"github.com/shutter-network/shutter/shuttermint/shdb"
	"github.com/shutter-network/shutter/shuttermint/shmsg"
)

var GossipTopicNames = []string{
	kprtopics.DecryptionTrigger,
	kprtopics.DecryptionKeyShare,
	kprtopics.DecryptionKey,
	kprtopics.EonPublicKey,
}

type keyper struct {
	config            Config
	dbpool            *pgxpool.Pool
	db                *kprdb.Queries
	shuttermintClient client.Client
	messageSender     fx.RPCMessageSender
	contracts         *deployment.Contracts

	shuttermintState *ShuttermintState
	p2p              *p2p.P2P
}

// linkConfigToDB ensures that we use a database compatible with the given config. On first use
// it stores the config's ethereum address into the database. On subsequent uses it compares the
// stored value and raises an error if it doesn't match.
func linkConfigToDB(ctx context.Context, config Config, dbpool *pgxpool.Pool) error {
	const addressKey = "ethereum address"
	cfgAddress := config.Address().Hex()
	queries := kprdb.New(dbpool)
	dbAddr, err := queries.GetMeta(ctx, addressKey)
	if err == pgx.ErrNoRows {
		return queries.InsertMeta(ctx, kprdb.InsertMetaParams{
			Key:   addressKey,
			Value: cfgAddress,
		})
	} else if err != nil {
		return err
	}

	if dbAddr != cfgAddress {
		return errors.Errorf(
			"database linked to wrong address %s, config address is %s",
			dbAddr, cfgAddress)
	}
	return nil
}

func Run(ctx context.Context, config Config) error {
	dbpool, err := pgxpool.Connect(ctx, config.DatabaseURL)
	if err != nil {
		return errors.Wrap(err, "failed to connect to database")
	}
	defer dbpool.Close()
	log.Printf("Connected to database (%s)", shdb.ConnectionInfo(dbpool))
	db := kprdb.New(dbpool)

	ethereumClient, err := ethclient.Dial(config.EthereumURL)
	if err != nil {
		return err
	}
	contracts, err := deployment.NewContracts(ethereumClient, config.DeploymentDir)
	if err != nil {
		return err
	}

	err = kprdb.ValidateKeyperDB(ctx, dbpool)
	if err != nil {
		return err
	}
	err = linkConfigToDB(ctx, config, dbpool)
	if err != nil {
		return err
	}
	shuttermintClient, err := http.New(config.ShuttermintURL, "/websocket")
	if err != nil {
		return err
	}
	messageSender := fx.NewRPCMessageSender(shuttermintClient, config.SigningKey)

	k := keyper{
		config:            config,
		dbpool:            dbpool,
		db:                db,
		shuttermintClient: shuttermintClient,
		messageSender:     messageSender,
		contracts:         contracts,

		shuttermintState: NewShuttermintState(config),
		p2p: p2p.New(p2p.Config{
			ListenAddr:     config.ListenAddress,
			PeerMultiaddrs: config.PeerMultiaddrs,
			PrivKey:        config.P2PKey,
		}),
	}
	return k.run(ctx)
}

func (kpr *keyper) run(ctx context.Context) error {
	group, ctx := errgroup.WithContext(ctx)

	topicValidators := kpr.makeMessagesValidators()

	group.Go(func() error {
		return kpr.p2p.Run(ctx, GossipTopicNames, topicValidators)
	})
	group.Go(func() error {
		return kpr.operateShuttermint(ctx)
	})
	group.Go(func() error {
		return kpr.operateP2P(ctx)
	})
	group.Go(func() error {
		return kpr.broadcastEonPublicKeys(ctx)
	})
	group.Go(func() error {
		return kpr.handleContractEvents(ctx)
	})
	return group.Wait()
}

func (kpr *keyper) operateShuttermint(ctx context.Context) error {
	for {
		err := SyncAppWithDB(ctx, kpr.shuttermintClient, kpr.dbpool, kpr.shuttermintState)
		if err != nil {
			return err
		}
		err = SendShutterMessages(ctx, kprdb.New(kpr.dbpool), &kpr.messageSender)
		if err != nil {
			return err
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(2 * time.Second):
		}
	}
}

func (kpr *keyper) operateP2P(ctx context.Context) error {
	for {
		select {
		case msg, ok := <-kpr.p2p.GossipMessages:
			if !ok {
				return nil
			}
			if err := kpr.handleP2PMessage(ctx, msg); err != nil {
				log.Printf("error handling message %+v: %s", msg, err)
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (kpr *keyper) broadcastEonPublicKeys(ctx context.Context) error {
	for {
		eonPublicKeys, err := kpr.db.GetAndDeleteEonPublicKeys(ctx)
		if err != nil {
			return err
		}
		for _, eonPublicKey := range eonPublicKeys {
			err := kpr.sendMessage(ctx, &shmsg.EonPublicKey{
				PublicKey:  eonPublicKey.EonPublicKey,
				Eon:        uint64(eonPublicKey.Eon),
				InstanceID: kpr.config.InstanceID,
			})
			if err != nil {
				return err
			}
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(2 * time.Second):
		}
	}
}

func (kpr *keyper) handleP2PMessage(ctx context.Context, msg *p2p.Message) error {
	var msgsOut []shmsg.P2PMessage
	var err error

	unmarshalled, err := unmarshalP2PMessage(msg)
	if err != nil {
		return err
	}

	epochKGHandler := epochKGHandler{
		config: kpr.config,
		db:     kprdb.New(kpr.dbpool),
	}

	switch typedMsg := unmarshalled.(type) {
	case *decryptionTrigger:
		msgsOut, err = epochKGHandler.handleDecryptionTrigger(ctx, typedMsg)
	case *decryptionKeyShare:
		msgsOut, err = epochKGHandler.handleDecryptionKeyShare(ctx, typedMsg)
	case *decryptionKey:
		msgsOut, err = epochKGHandler.handleDecryptionKey(ctx, typedMsg)
	default:
		log.Println("ignoring message received on topic", msg.Topic)
		return nil
	}

	if err != nil {
		return err
	}
	for _, msgOut := range msgsOut {
		if err := kpr.sendMessage(ctx, msgOut); err != nil {
			log.Printf("error sending message %+v: %s", msgOut, err)
			continue
		}
	}
	return nil
}

func (kpr *keyper) sendMessage(ctx context.Context, msg shmsg.P2PMessage) error {
	msgBytes, err := proto.Marshal(msg)
	if err != nil {
		return errors.Wrap(err, "failed to marshal p2p message")
	}

	return kpr.p2p.Publish(ctx, msg.Topic(), msgBytes)
}

func (kpr *keyper) makeMessagesValidators() map[string]pubsub.Validator {
	db := kprdb.New(kpr.dbpool)
	validators := make(map[string]pubsub.Validator)
	validators[kprtopics.DecryptionKey] = kpr.makeDecryptionKeyValidator(db)
	validators[kprtopics.DecryptionKeyShare] = kpr.makeKeyShareValidator(db)
	validators[kprtopics.EonPublicKey] = kpr.makeEonPublicKeyValidator()

	return validators
}

func (kpr *keyper) makeDecryptionKeyValidator(db *kprdb.Queries) pubsub.Validator {
	return func(ctx context.Context, peerID peer.ID, libp2pMessage *pubsub.Message) bool {
		p2pMessage := new(p2p.Message)
		if err := json.Unmarshal(libp2pMessage.Data, p2pMessage); err != nil {
			return false
		}
		msg, err := unmarshalP2PMessage(p2pMessage)
		if err != nil {
			return false
		}
		if msg.GetInstanceID() != kpr.config.InstanceID {
			return false
		}

		key, ok := msg.(*decryptionKey)
		if !ok {
			panic("unmarshalled non decryption key message in decryption key validator")
		}

		activationBlockNumber := medley.ActivationBlockNumberFromEpochID(key.epochID)
		dkgResultDB, err := db.GetDKGResultForBlockNumber(ctx, int64(activationBlockNumber))
		if err == pgx.ErrNoRows {
			return false
		}
		if err != nil {
			log.Printf("failed to get dkg result for epoch %d from db", key.epochID)
			return false
		}
		if !dkgResultDB.Success {
			return false
		}
		pureDKGResult, err := shdb.DecodePureDKGResult(dkgResultDB.PureResult)
		if err != nil {
			log.Printf("error while decoding pure DKG result for epoch %d", key.epochID)
			return false
		}

		ok, err = shcrypto.VerifyEpochSecretKey(key.key, pureDKGResult.PublicKey, key.epochID)
		if err != nil {
			log.Printf("error while checking epoch secret key for epoch %v", key.epochID)
			return false
		}
		return ok
	}
}

func (kpr *keyper) makeKeyShareValidator(db *kprdb.Queries) pubsub.Validator {
	return func(ctx context.Context, peerID peer.ID, libp2pMessage *pubsub.Message) bool {
		p2pMessage := new(p2p.Message)
		if err := json.Unmarshal(libp2pMessage.Data, p2pMessage); err != nil {
			return false
		}
		msg, err := unmarshalP2PMessage(p2pMessage)
		if err != nil {
			return false
		}
		if msg.GetInstanceID() != kpr.config.InstanceID {
			return false
		}

		keyShare, ok := msg.(*decryptionKeyShare)
		if !ok {
			panic("unmarshalled non decryption key share message in decryption key share validator")
		}

		activationBlockNumber := medley.ActivationBlockNumberFromEpochID(keyShare.epochID)
		dkgResultDB, err := db.GetDKGResultForBlockNumber(ctx, int64(activationBlockNumber))
		if err == pgx.ErrNoRows {
			return false
		}
		if err != nil {
			log.Printf("failed to get dkg result for epoch %d from db", keyShare.epochID)
			return false
		}
		if !dkgResultDB.Success {
			return false
		}
		pureDKGResult, err := shdb.DecodePureDKGResult(dkgResultDB.PureResult)
		if err != nil {
			log.Printf("error while decoding pure DKG result for epoch %d", keyShare.epochID)
			return false
		}

		ok = shcrypto.VerifyEpochSecretKeyShare(
			keyShare.share,
			pureDKGResult.PublicKeyShares[keyShare.keyperIndex],
			shcrypto.ComputeEpochID(keyShare.epochID),
		)
		return ok
	}
}

func (kpr *keyper) makeEonPublicKeyValidator() pubsub.Validator {
	return func(ctx context.Context, peerID peer.ID, libp2pMessage *pubsub.Message) bool {
		p2pMessage := new(p2p.Message)
		if err := json.Unmarshal(libp2pMessage.Data, p2pMessage); err != nil {
			return false
		}
		msg, err := unmarshalP2PMessage(p2pMessage)
		if err != nil {
			return false
		}
		return msg.GetInstanceID() == kpr.config.InstanceID
	}
}
