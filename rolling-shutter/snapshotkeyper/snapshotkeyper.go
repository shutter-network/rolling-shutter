// Package snapshotkeyper contains the snapshot specific keyper implementation
package snapshotkeyper

import (
	"context"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/tendermint/tendermint/rpc/client"
	tmhttp "github.com/tendermint/tendermint/rpc/client/http"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/chainobserver"
	chainobskprdb "github.com/shutter-network/rolling-shutter/rolling-shutter/chainobserver/db/keyper"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/contract/deployment"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/database"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/epochkghandler"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/fx"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/kprapi"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/smobserver"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/db"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/eventsyncer"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/metricsserver"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/retry"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/service"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2p"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2pmsg"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/shdb"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/shmsg"
)

type snapshotkeyper struct {
	config            *keyper.Config
	dbpool            *pgxpool.Pool
	shuttermintClient client.Client
	messageSender     fx.RPCMessageSender
	l1Client          *ethclient.Client
	contracts         *deployment.Contracts

	shuttermintState *smobserver.ShuttermintState
	p2p              *p2p.P2PHandler
	metricsServer    *metricsserver.MetricsServer
}

func New(config *keyper.Config) service.Service {
	return &snapshotkeyper{config: config}
}

func (snkpr *snapshotkeyper) Start(ctx context.Context, runner service.Runner) error {
	config := snkpr.config
	dbpool, err := pgxpool.Connect(ctx, config.DatabaseURL)
	if err != nil {
		return errors.Wrap(err, "failed to connect to database")
	}
	runner.Defer(dbpool.Close)
	shdb.AddConnectionInfo(log.Info(), dbpool).Msg("connected to database")

	l1Client, err := ethclient.Dial(config.Ethereum.EthereumURL)
	if err != nil {
		return err
	}
	contracts, err := deployment.NewContracts(l1Client, config.Ethereum.DeploymentDir)
	if err != nil {
		return err
	}

	err = snkpr.dbpool.BeginFunc(db.WrapContext(ctx, database.Definition.Validate))
	if err != nil {
		return err
	}
	err = keyper.LinkConfigToDB(ctx, config, dbpool)
	if err != nil {
		return err
	}
	shuttermintClient, err := tmhttp.New(config.Shuttermint.ShuttermintURL, "/websocket")
	if err != nil {
		return err
	}
	messageSender := fx.NewRPCMessageSender(shuttermintClient, config.Ethereum.PrivateKey.Key)

	p2pHandler, err := p2p.New(config.P2P)
	if err != nil {
		return err
	}

	if snkpr.config.Metrics.Enabled {
		epochkghandler.InitMetrics()
		snkpr.metricsServer = metricsserver.New(snkpr.config.Metrics)
	}

	snkpr.dbpool = dbpool
	snkpr.shuttermintClient = shuttermintClient
	snkpr.messageSender = messageSender
	snkpr.l1Client = l1Client
	snkpr.contracts = contracts
	snkpr.shuttermintState = smobserver.NewShuttermintState(config)
	snkpr.p2p = p2pHandler

	snkpr.setupP2PHandler()
	return runner.StartService(snkpr.getServices()...)
}

func (snkpr *snapshotkeyper) setupP2PHandler() {
	snkpr.p2p.AddMessageHandler(
		epochkghandler.NewDecryptionKeyHandler(snkpr.config, snkpr.dbpool),
		epochkghandler.NewDecryptionKeyShareHandler(snkpr.config, snkpr.dbpool),
		epochkghandler.NewDecryptionTriggerHandler(snkpr.config, snkpr.dbpool),
		epochkghandler.NewEonPublicKeyHandler(snkpr.config, snkpr.dbpool),
	)
}

func (snkpr *snapshotkeyper) getServices() []service.Service {
	services := []service.Service{
		snkpr.p2p,
		service.ServiceFn{Fn: snkpr.operateShuttermint},
		service.ServiceFn{Fn: snkpr.broadcastEonPublicKeys},
		service.ServiceFn{Fn: snkpr.handleContractEvents},
	}

	if snkpr.config.HTTPEnabled {
		services = append(services, kprapi.NewHTTPService(snkpr.dbpool, snkpr.config, snkpr.p2p))
	}
	if snkpr.config.Metrics.Enabled {
		services = append(services, snkpr.metricsServer)
	}
	return services
}

func (snkpr *snapshotkeyper) handleContractEvents(ctx context.Context) error {
	events := []*eventsyncer.EventType{
		snkpr.contracts.KeypersConfigsListNewConfig,
		snkpr.contracts.CollatorConfigsListNewConfig,
	}
	return chainobserver.New(snkpr.contracts, snkpr.dbpool).Observe(ctx, events)
}

func (snkpr *snapshotkeyper) handleOnChainChanges(ctx context.Context, tx pgx.Tx, l1BlockNumber uint64) error {
	log.Debug().Uint64("l1-block-number", l1BlockNumber).Msg("handle on chain changes")
	err := snkpr.handleOnChainKeyperSetChanges(ctx, tx, l1BlockNumber)
	if err != nil {
		return err
	}
	err = snkpr.sendNewBlockSeen(ctx, tx, l1BlockNumber)
	if err != nil {
		return err
	}
	return nil
}

// sendNewBlockSeen sends shmsg.NewBlockSeen messages to the shuttermint chain. This function sends
// NewBlockSeen messages to the shuttermint chain, so that the chain can start new batch configs if
// enough keypers have seen a block past the start block of some BatchConfig. We only send messages
// when the current block we see, could lead to a batch config being started.
func (snkpr *snapshotkeyper) sendNewBlockSeen(
	ctx context.Context,
	tx pgx.Tx,
	l1BlockNumber uint64,
) error {
	q := database.New(tx)
	lastBlock, err := q.GetLastBlockSeen(ctx)
	if err != nil {
		return err
	}

	count, err := q.CountBatchConfigsInBlockRange(ctx,
		database.CountBatchConfigsInBlockRangeParams{
			StartBlock: lastBlock,
			EndBlock:   int64(l1BlockNumber),
		})
	if err != nil {
		return err
	}
	if count == 0 {
		return nil
	}

	blockSeenMsg := shmsg.NewBlockSeen(l1BlockNumber)
	err = q.ScheduleShutterMessage(
		ctx,
		fmt.Sprintf("block seen (block=%d)", l1BlockNumber),
		blockSeenMsg,
	)
	if err != nil {
		return err
	}
	err = q.SetLastBlockSeen(ctx, int64(l1BlockNumber))
	if err != nil {
		return err
	}
	log.Info().Uint64("block-number", l1BlockNumber).Msg("block seen")
	return nil
}

// handleOnChainKeyperSetChanges looks for changes in the keyper_set table.
func (snkpr *snapshotkeyper) handleOnChainKeyperSetChanges(ctx context.Context, tx pgx.Tx, l1BlockNumber uint64) error {
	q := database.New(tx)
	latestBatchConfig, err := q.GetLatestBatchConfig(ctx)
	if err == pgx.ErrNoRows {
		log.Print("no batch config found in tendermint")
		return nil
	} else if err != nil {
		return err
	}

	cq := chainobskprdb.New(tx)
	keyperSet, err := cq.GetKeyperSetByKeyperConfigIndex(
		ctx,
		int64(latestBatchConfig.KeyperConfigIndex)+1,
	)
	if err == pgx.ErrNoRows {
		return nil
	}

	if err != nil {
		return err
	}

	lastSent, err := q.GetLastBatchConfigSent(ctx)
	if err != nil {
		return err
	}
	if lastSent == keyperSet.KeyperConfigIndex {
		return nil
	}

	activationBlockNumber, err := medley.Int64ToUint64Safe(keyperSet.ActivationBlockNumber)
	if err != nil {
		return err
	}
	// We *MUST* check if the l1BlockNumber is smaller than the activationBlockNumber since both are uint64 and therefore subtraction can never result in negative numbers.
	// This means that if we missed the activationBlockNumber we will never submit the config.
	if l1BlockNumber < activationBlockNumber && activationBlockNumber-l1BlockNumber > snkpr.config.Shuttermint.DKGStartBlockDelta {
		log.Info().Interface("keyper-set", keyperSet).
			Uint64("l1-block-number", l1BlockNumber).
			Uint64("dkg-start-delta", snkpr.config.Shuttermint.DKGStartBlockDelta).
			Msg("not yet submitting config")
		return nil
	}

	err = q.SetLastBatchConfigSent(ctx, keyperSet.KeyperConfigIndex)
	if err != nil {
		return nil
	}

	keypers, err := shdb.DecodeAddresses(keyperSet.Keypers)
	if err != nil {
		return err
	}
	log.Info().Interface("keyper-set", keyperSet).Msg("have a new config to be scheduled")
	batchConfigMsg := shmsg.NewBatchConfig(
		uint64(keyperSet.ActivationBlockNumber),
		keypers,
		uint64(keyperSet.Threshold),
		uint64(keyperSet.KeyperConfigIndex),
	)
	err = q.ScheduleShutterMessage(
		ctx,
		fmt.Sprintf("new batch config (activation-block-number=%d, config-index=%d)",
			keyperSet.ActivationBlockNumber, keyperSet.KeyperConfigIndex),
		batchConfigMsg,
	)
	if err != nil {
		return err
	}
	return nil
}

func (snkpr *snapshotkeyper) operateShuttermint(ctx context.Context) error {
	for {
		l1BlockNumber, err := retry.FunctionCall(ctx, snkpr.l1Client.BlockNumber)
		if err != nil {
			log.Err(err).Msg("Error when getting block")
			return err
		}

		err = smobserver.SyncAppWithDB(
			ctx,
			snkpr.shuttermintClient,
			snkpr.dbpool,
			snkpr.shuttermintState,
		)
		if err != nil {
			log.Err(err).Msg("Error on syncing app with db")
			return err
		}
		err = snkpr.dbpool.BeginFunc(ctx, func(tx pgx.Tx) error {
			return snkpr.handleOnChainChanges(ctx, tx, l1BlockNumber)
		})
		if err != nil {
			log.Err(err).Msg("Error on handling onChainChanges")
			return err
		}

		err = fx.SendShutterMessages(ctx, database.New(snkpr.dbpool), &snkpr.messageSender)
		if err != nil {
			log.Err(err).Msg("Error sending shutter messages")
			return err
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(2 * time.Second):
		}
	}
}

func (snkpr *snapshotkeyper) broadcastEonPublicKeys(ctx context.Context) error {
	for {
		eonPublicKeys, err := database.New(snkpr.dbpool).GetAndDeleteEonPublicKeys(ctx)
		if err != nil {
			return err
		}
		for _, eonPublicKey := range eonPublicKeys {
			_, exists := database.GetKeyperIndex(snkpr.config.GetAddress(), eonPublicKey.Keypers)
			if !exists {
				return errors.Errorf("own keyper index not found for Eon=%d", eonPublicKey.Eon)
			}
			msg, err := p2pmsg.NewSignedEonPublicKey(
				snkpr.config.InstanceID,
				eonPublicKey.EonPublicKey,
				uint64(eonPublicKey.ActivationBlockNumber),
				uint64(eonPublicKey.KeyperConfigIndex),
				uint64(eonPublicKey.Eon),
				snkpr.config.Ethereum.PrivateKey.Key,
			)
			if err != nil {
				return errors.Wrap(err, "error while signing EonPublicKey")
			}

			err = snkpr.p2p.SendMessage(ctx, msg)
			if err != nil {
				return errors.Wrap(err, "error while broadcasting EonPublicKey")
			}
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(2 * time.Second):
		}
	}
}
