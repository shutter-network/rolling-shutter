package snapshot

import (
	"context"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/db/snpdb"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/epochid"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/service"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2p"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2pmsg"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/shdb"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/snapshot/hubapi"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/snapshot/snpjrpc"
)

// FIXME: Needs to be in DB.
var (
	seenEons      = make(map[uint64]struct{})
	seenProposals = make(map[string]struct{})
	zeroTXHash    = make([]byte, 32)
)

type Snapshot struct {
	Config *Config

	p2p      *p2p.P2PHandler
	dbpool   *pgxpool.Pool
	db       *snpdb.Queries
	l1Client *ethclient.Client
	hubapi   *hubapi.HubAPI
}

func New(config *Config) (service.Service, error) {
	p2pInstance, err := p2p.New(config.P2P)
	return &Snapshot{
		Config: config,
		p2p:    p2pInstance,
	}, err
}

func (snp *Snapshot) Start(ctx context.Context, runner service.Runner) error {
	log.Printf(
		"starting Snapshot Hub interface",
	)
	l1Client, err := ethclient.Dial(snp.Config.EthereumURL)
	if err != nil {
		return err
	}
	snp.l1Client = l1Client

	dbpool, err := pgxpool.Connect(ctx, snp.Config.DatabaseURL)
	if err != nil {
		return errors.Wrap(err, "failed to connect to database")
	}
	runner.Defer(dbpool.Close)

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

	snp.setupP2PHandler()
	return runner.StartService(snp.getServices()...)
}

func (snp *Snapshot) getServices() []service.Service {
	services := []service.Service{
		snp.p2p,
		snpjrpc.New(
			snp.Config.JSONRPCHost,
			snp.Config.JSONRPCPort,
			snp.handleDecryptionKeyRequest,
			snp.handleRequestEonKey,
		),
	}
	if snp.Config.MetricsEnabled {
		services = append(services, NewMetricsServer(snp.Config))
	}
	return services
}

func (snp *Snapshot) setupP2PHandler() {
	snp.p2p.AddMessageHandler(
		NewEonPublicKeyHandler(snp.Config, snp),
		NewDecryptionKeyHandler(snp.Config, snp),
		// We need the decryption trigger handler in order to be subscribed to the topic mesh.
		NewDecryptionTriggerHandler(),
	)
}

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

func (snp *Snapshot) handleDecryptionKeyRequest(ctx context.Context, epochID []byte) error {
	blockNumber, err := snp.l1Client.BlockNumber(ctx)
	if err != nil {
		return err
	}
	convEpoch, err := epochid.BytesToEpochID(epochID)
	if err != nil {
		return err
	}
	trigMsg, err := p2pmsg.NewSignedDecryptionTrigger(
		snp.Config.InstanceID,
		convEpoch,
		blockNumber,
		zeroTXHash,
		snp.Config.EthereumKey,
	)
	if err != nil {
		return err
	}

	err = snp.SendMessage(ctx, trigMsg)
	if err != nil {
		return err
	}
	log.Printf("Trigger decryption for proposal %X", epochID)
	return nil
}

func (snp *Snapshot) SendMessage(ctx context.Context, msg p2pmsg.Message) error {
	log.Printf("sending %s", msg.LogInfo())

	return snp.p2p.SendMessage(ctx, msg)
}
