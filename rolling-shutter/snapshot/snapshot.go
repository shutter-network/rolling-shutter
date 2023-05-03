package snapshot

import (
	"context"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/proto"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/db/snpdb"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2p"
	shmsg "github.com/shutter-network/rolling-shutter/rolling-shutter/p2pmsg"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/shdb"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/snapshot/hubapi"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/snapshot/snpjrpc"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/snapshot/snptopics"
)

var gossipTopicNames = [3]string{
	snptopics.DecryptionKey,
	snptopics.EonPublicKey,
	snptopics.TimedEpoch,
}

// FIXME: Needs to be in DB
var seenEons = make(map[uint64]struct{})
var seenProposals = make(map[string]struct{})

type Snapshot struct {
	Config Config

	p2p    *p2p.P2P
	dbpool *pgxpool.Pool
	db     *snpdb.Queries

	hubapi *hubapi.HubAPI
}

func New(config Config) *Snapshot {
	p2pConfig := p2p.Config{
		ListenAddrs:       config.ListenAddresses,
		BootstrapPeers:    config.CustomBootstrapAddresses, // FIXME: add to own config
		PrivKey:           config.P2PKey,
		DisableTopicDHT:   true,
		DisableRoutingDHT: true,
	}
	p2p_instance := p2p.New(p2pConfig)

	return &Snapshot{
		Config: config,
		p2p:    p2p_instance,
	}
}

func (snp *Snapshot) Run(ctx context.Context) error {
	log.Printf(
		"starting Snapshot Hub interface",
	)

	dbpool, err := pgxpool.Connect(ctx, snp.Config.DatabaseURL)
	if err != nil {
		return errors.Wrap(err, "failed to connect to database")
	}
	defer dbpool.Close()
	snp.dbpool = dbpool
	shdb.AddConnectionInfo(log.Info(), dbpool).Msg("connected to database")

	err = snpdb.ValidateSnapshotDB(ctx, dbpool)
	if err != nil {
		return err
	}
	db := snpdb.New(dbpool)
	snp.db = db

	if snp.Config.MetricsEnabled {
		err = snp.initMetrics(ctx)
		if err != nil {
			return err
		}
	}

	hub := hubapi.New(snp.Config.SnapshotHubURL)
	snp.hubapi = hub

	jrpc := snpjrpc.New(
		snp.Config.JSONRPCHost,
		snp.Config.JSONRPCPort,
		snp.handleDecryptionKeyRequest,
		snp.handleRequestEonKey,
	)

	errorgroup, errorctx := errgroup.WithContext(ctx)

	topicValidators := snp.makeMessagesValidators()

	errorgroup.Go(
		func() error {
			return snp.p2p.Run(errorctx, gossipTopicNames[:], topicValidators)
		},
	)
	errorgroup.Go(
		func() error {
			return snp.handleMessages(errorctx)
		},
	)
	errorgroup.Go(
		func() error {
			jrpc.Server.Start()
			return nil
		},
	)
	if snp.Config.MetricsEnabled {
		errorgroup.Go(
			func() error {
				return snp.runMetricsServer(errorctx)
			},
		)
	}
	return errorgroup.Wait()
}

/*
func (snp *Snapshot) handleMessages(ctx context.Context) error {
	for {
		select {
		case msg, ok := <-snp.p2p.GossipMessages:
			if !ok {
				return nil
			}
			if err := snp.handleMessage(ctx, msg); err != nil {
				log.Printf("error handling message %+v: %s", msg, err)
				continue
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (snp *Snapshot) handleMessage(ctx context.Context, msg *p2p.Message) error {
	var msgsOut []shmsg.P2PMessage
	var err error

	unmarshalled, err := unmarshalP2PMessage(msg)
	if topicError, ok := err.(*unhandledTopicError); ok {
		log.Print(topicError.Error())
	} else if err != nil {
		return err
	}

	switch typedMsg := unmarshalled.(type) {
	case *decryptionKey:
		err = snp.handleDecryptionKeyInput(ctx, snp.Config, snp.db, typedMsg)
	case *eonPublicKey:
		err = snp.handleEonPublicKeyInput(ctx, typedMsg.Eon, typedMsg.PublicKey)
	default:
		log.Print("ignoring message received on topic", msg.Topic)
		return nil
	}

	if err != nil {
		return err
	}
	for _, msgOut := range msgsOut {
		if err := snp.SendMessage(ctx, msgOut); err != nil {
			log.Printf("error sending message %+v: %s", msgOut, err)
			continue
		}
	}
	return nil
}
*/

/* func (snp *Snapshot) handleDecryptionKeyInput(
	ctx context.Context,
	config Config,
	db *snpdb.Queries,
	key *decryptionKey,
) error {
	_, seen := seenProposals[string(key.EpochID)]
	if seen {
		return nil
	}
	log.Printf("Sending key %X for proposal %X to hub", key.Key, key.EpochID)

	metricKeysGenerated.Inc()

	err := snp.hubapi.SubmitProposalKey(key.EpochID, key.Key)
	if err != nil {
		return err
	}
	// FIXME: Apart from needing to be in DB we need to keep track of the proposals better
	seenProposals[string(key.EpochID)] = struct{}{}
	return nil
} */

func (snp *Snapshot) handleRequestEonKey(ctx context.Context) error {
	row, err := snp.db.GetEonPublicKeyLatest(ctx)
	if err == pgx.ErrNoRows {
		return errors.Errorf("No Eon key found: %v", err)
	}
	err = snp.hubapi.SubmitEonKey(uint64(row.EonID), row.EonPublicKey)
	if err != nil {
		return err
	}
	return nil
}

/*
func (snp *Snapshot) handleEonPublicKeyInput(

	ctx context.Context,
	eonId uint64,
	key []byte,

	) error {
		err := snp.db.InsertEonPublicKey(
			ctx, snpdb.InsertEonPublicKeyParams{
				EonID:        int64(eonId),
				EonPublicKey: key,
			},
		)
		if err != nil {
			return err
		}
		_, seen := seenEons[eonId]
		if seen {
			return nil
		}

		metricEons.Inc()

		log.Printf("Sending Eon %d public key to hub", eonId)
		err = snp.hubapi.SubmitEonKey(eonId, key)
		if err != nil {
			return err
		}
		seenEons[eonId] = struct{}{}
		return nil
	}
*/
func (snp *Snapshot) handleDecryptionKeyRequest(ctx context.Context, epochId []byte) error {
	msg := &shmsg.TimedEpoch{
		InstanceID: 0,
		EpochID:    epochId,
		NotBefore:  0,
	}
	err := snp.SendMessage(ctx, msg)
	if err != nil {
		return err
	}
	log.Printf("Trigger decryption for proposal %X", epochId)
	return nil
}

func (snp *Snapshot) SendMessage(ctx context.Context, msg shmsg.P2PMessage) error {
	msgBytes, err := proto.Marshal(msg)
	if err != nil {
		return errors.Wrap(err, "failed to marshal p2p message")
	}
	log.Printf("sending %s", msg.LogInfo())

	return snp.p2p.Publish(ctx, msg.Topic(), msgBytes)
}
