// Package keyper contains the keyper implementation
package keyper

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
	"github.com/shutter-network/rolling-shutter/rolling-shutter/contract/deployment"
	obscollator "github.com/shutter-network/rolling-shutter/rolling-shutter/db/chainobsdb/collator"
	obskeyper "github.com/shutter-network/rolling-shutter/rolling-shutter/db/chainobsdb/keyper"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/db/kprdb"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/db/metadb"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/epochkghandler"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/fx"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/kprapi"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/smobserver"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/broker"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/eventsyncer"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/metricsserver"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/retry"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/service"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2p"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2pmsg"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/shdb"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/shmsg"
)

type keyper struct {
	trigger chan broker.Event[*epochkghandler.DecryptionTrigger]
	// TODO don't save the whole thing here
	opts              *options
	config            *Config
	dbpool            *pgxpool.Pool
	shuttermintClient client.Client
	messageSender     fx.RPCMessageSender
	l1Client          *ethclient.Client
	contracts         *deployment.Contracts

	shuttermintState *smobserver.ShuttermintState
	p2p              *p2p.P2PHandler
	metricsServer    *metricsserver.MetricsServer
}

func New(config *Config, options ...Option) service.Service {
	opts := newDefaultOptions()
	for _, option := range options {
		err := option(opts)
		if err != nil {
			// TODO should we panic here?
			panic(err)
		}
	}
	return &keyper{config: config, trigger: opts.trigger}
}

// LinkConfigToDB ensures that we use a database compatible with the given config. On first use
// it stores the config's ethereum address into the database. On subsequent uses it compares the
// stored value and raises an error if it doesn't match.
func LinkConfigToDB(ctx context.Context, config *Config, dbpool *pgxpool.Pool) error {
	const addressKey = "ethereum address"
	cfgAddress := config.GetAddress().String()
	queries := metadb.New(dbpool)
	dbAddr, err := queries.GetMeta(ctx, addressKey)
	if err == pgx.ErrNoRows {
		return queries.InsertMeta(ctx, metadb.InsertMetaParams{
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

func (kpr *keyper) Start(ctx context.Context, runner service.Runner) error {
	config := kpr.config
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
	l2Client, err := ethclient.Dial(config.Ethereum.ContractsURL)
	if err != nil {
		return err
	}
	contracts, err := deployment.NewContracts(l2Client, config.Ethereum.DeploymentDir)
	if err != nil {
		return err
	}

	err = kprdb.ValidateKeyperDB(ctx, dbpool)
	if err != nil {
		return err
	}
	err = LinkConfigToDB(ctx, config, dbpool)
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

	if kpr.config.Metrics.Enabled {
		epochkghandler.InitMetrics()
		kpr.metricsServer = metricsserver.New(kpr.config.Metrics)
	}

	kpr.dbpool = dbpool
	kpr.shuttermintClient = shuttermintClient
	kpr.messageSender = messageSender
	kpr.l1Client = l1Client
	kpr.contracts = contracts
	kpr.shuttermintState = smobserver.NewShuttermintState(config)
	kpr.p2p = p2pHandler

	kpr.setupP2PHandler()
	return runner.StartService(kpr.getServices()...)
}

func (kpr *keyper) setupP2PHandler() {
	kpr.p2p.AddMessageHandler(
		epochkghandler.NewDecryptionKeyHandler(kpr.config, kpr.dbpool),
		epochkghandler.NewDecryptionKeyShareHandler(kpr.config, kpr.dbpool),
		epochkghandler.NewEonPublicKeyHandler(kpr.config, kpr.dbpool),

		// TODO factor this out, this is the rollup keyper mechanism.
		// here we need:
		//  - DB to get the collator for the block number.
		//    --> the "collator" concept should be exclusive to the rollup keyper,
		//        so we should not have to expose the internal db here

		//  - P2P to register the very handler here (AddMessageHandler).
		//    the p2p should be somehow available from the outside.
		//    Also the concept of DecryptionTrigger p2p messages shouldn't
		//    be known in the common keyper
		epochkghandler.NewDecryptionTriggerHandler(kpr.config, kpr.dbpool),
	)
}

func (kpr *keyper) getServices() []service.Service {
	services := []service.Service{
		kpr.p2p,
		service.ServiceFn{Fn: kpr.operateShuttermint},
		service.ServiceFn{Fn: kpr.broadcastEonPublicKeys},
		service.ServiceFn{Fn: kpr.handleContractEvents},
		service.ServiceFn{Fn: kpr.listenDecryptionTriggers},
	}

	if kpr.config.HTTPEnabled {
		services = append(services, kprapi.NewHTTPService(kpr.dbpool, kpr.config, kpr.p2p))
	}
	if kpr.config.Metrics.Enabled {
		services = append(services, kpr.metricsServer)
	}
	return services
}

func (kpr *keyper) handleContractEvents(ctx context.Context) error {
	kprHandler := &obskeyper.Handler{
		KeyperContract: kpr.contracts.Keypers,
	}
	cltHandler := &obscollator.Handler{
		CollatorContract: kpr.contracts.Collators,
	}
	events := map[*eventsyncer.EventType]chainobserver.EventHandlerFunc{
		kpr.contracts.KeypersConfigsListNewConfig:  chainobserver.MakeHandler(kprHandler.HandleKeypersConfigsListNewConfigEvent),
		kpr.contracts.CollatorConfigsListNewConfig: chainobserver.MakeHandler(cltHandler.HandleCollatorConfigsListNewConfigEvent),
	}
	return chainobserver.New(kpr.contracts, kpr.dbpool).Observe(ctx, events)
}

func (kpr *keyper) handleOnChainChanges(
	ctx context.Context,
	tx pgx.Tx,
	l1BlockNumber uint64,
) error {
	log.Debug().Uint64("l1-block-number", l1BlockNumber).Msg("handle on chain changes")
	err := kpr.handleOnChainKeyperSetChanges(ctx, tx, l1BlockNumber)
	if err != nil {
		return err
	}
	err = kpr.sendNewBlockSeen(ctx, tx, l1BlockNumber)
	if err != nil {
		return err
	}
	return nil
}

// sendNewBlockSeen sends shmsg.NewBlockSeen messages to the shuttermint chain. This function sends
// NewBlockSeen messages to the shuttermint chain, so that the chain can start new batch configs if
// enough keypers have seen a block past the start block of some BatchConfig. We only send messages
// when the current block we see, could lead to a batch config being started.
func (kpr *keyper) sendNewBlockSeen(ctx context.Context, tx pgx.Tx, l1BlockNumber uint64) error {
	q := kprdb.New(tx)
	lastBlock, err := q.GetLastBlockSeen(ctx)
	if err != nil {
		return err
	}

	count, err := q.CountBatchConfigsInBlockRange(ctx,
		kprdb.CountBatchConfigsInBlockRangeParams{
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
func (kpr *keyper) handleOnChainKeyperSetChanges(
	ctx context.Context,
	tx pgx.Tx,
	l1BlockNumber uint64,
) error {
	q := kprdb.New(tx)
	latestBatchConfig, err := q.GetLatestBatchConfig(ctx)
	if err == pgx.ErrNoRows {
		log.Print("no batch config found in tendermint")
		return nil
	} else if err != nil {
		return err
	}

	cq := obskeyper.New(tx)
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
	if l1BlockNumber < activationBlockNumber && activationBlockNumber-l1BlockNumber > kpr.config.Shuttermint.DKGStartBlockDelta {
		log.Info().Interface("keyper-set", keyperSet).
			Uint64("l1-block-number", l1BlockNumber).
			Uint64("dkg-start-delta", kpr.config.Shuttermint.DKGStartBlockDelta).
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
	log.Info().Interface("keyper-set", keyperSet).
		Uint64("l1-block-number", l1BlockNumber).
		Uint64("dkg-start-delta", kpr.config.Shuttermint.DKGStartBlockDelta).
		Msg("have a new config to be scheduled")
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

func (kpr *keyper) operateShuttermint(ctx context.Context) error {
	for {
		l1BlockNumber, err := retry.FunctionCall(ctx, kpr.l1Client.BlockNumber)
		if err != nil {
			return err
		}

		err = smobserver.SyncAppWithDB(ctx, kpr.shuttermintClient, kpr.dbpool, kpr.shuttermintState)
		if err != nil {
			return err
		}
		err = kpr.dbpool.BeginFunc(ctx, func(tx pgx.Tx) error {
			return kpr.handleOnChainChanges(ctx, tx, l1BlockNumber)
		})
		if err != nil {
			return err
		}

		err = fx.SendShutterMessages(ctx, kprdb.New(kpr.dbpool), &kpr.messageSender)
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

// FIXME name
func (kpr *keyper) listenDecryptionTriggers(ctx context.Context) error {
	// the sender has to do this before
	// eon, err := db.GetEonForBlockNumber(ctx, blockNumber)
	// if err != nil {
	// 	return nil, errors.Wrapf(err, "failed to get eon for block %d from db", blockNumber)
	// }

	for {
		select {
		case triggerEvent, ok := <-kpr.trigger:
			if !ok {
				return nil
			}
			keySharesMsg, err := epochkghandler.ConstructDecryptionKeyShare(
				ctx,
				kpr.config,
				kprdb.New(kpr.dbpool),
				triggerEvent.Value,
			)
			if err != nil {
				// FIXME how to handle
				continue
			}

			// TODO retry options
			if err := kpr.p2p.SendMessage(ctx, keySharesMsg); err != nil {
				// FIXME how to handle?
				continue
			}
		case <-ctx.Done():
			return ctx.Err()
		}
		return nil

	}
}

func (kpr *keyper) broadcastEonPublicKeys(ctx context.Context) error {
	for {
		eonPublicKeys, err := kprdb.New(kpr.dbpool).GetAndDeleteEonPublicKeys(ctx)
		if err != nil {
			return err
		}
		for _, eonPublicKey := range eonPublicKeys {
			_, exists := kprdb.GetKeyperIndex(kpr.config.GetAddress(), eonPublicKey.Keypers)
			if !exists {
				return errors.Errorf("own keyper index not found for Eon=%d", eonPublicKey.Eon)
			}
			msg, err := p2pmsg.NewSignedEonPublicKey(
				kpr.config.InstanceID,
				eonPublicKey.EonPublicKey,
				uint64(eonPublicKey.ActivationBlockNumber),
				uint64(eonPublicKey.KeyperConfigIndex),
				uint64(eonPublicKey.Eon),
				kpr.config.Ethereum.PrivateKey.Key,
			)
			if err != nil {
				return errors.Wrap(err, "error while signing EonPublicKey")
			}

			err = kpr.p2p.SendMessage(ctx, msg)
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
