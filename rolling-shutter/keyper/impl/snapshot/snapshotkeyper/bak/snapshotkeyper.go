// Package snapshotkeyper contains the snapshot specific keyper implementation
package snapshotkeyper

import (
	"context"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/tendermint/tendermint/rpc/client"
	tmhttp "github.com/tendermint/tendermint/rpc/client/http"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/contract/deployment"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/db/kprdb"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/epochkghandler"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/fx"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/kprapi"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/smobserver"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/metricsserver"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/retry"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/service"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2p"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/shdb"
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

	err = kprdb.ValidateKeyperDB(ctx, dbpool)
	if err != nil {
		return err
	}
	err = keyper.LinkConfigToDB(ctx, config, dbpool)
	if err != nil {
		return err
	}
	shuttermintClient, err := tmhttp.New(config.Shuttermint.ShuttermintURL)
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

		err = fx.SendShutterMessages(ctx, kprdb.New(snkpr.dbpool), &snkpr.messageSender)
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
