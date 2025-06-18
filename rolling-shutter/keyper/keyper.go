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

	obskeyper "github.com/shutter-network/rolling-shutter/rolling-shutter/chainobserver/db/keyper"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/contract/deployment"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/database"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/epochkghandler"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/fx"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/keypermetrics"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/kprapi"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/kprconfig"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/smobserver"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/broker"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/channel"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/db"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/metricsserver"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/retry"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/service"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2p"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/shdb"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/shmsg"
)

type KeyperCore struct {
	trigger <-chan *broker.Event[*epochkghandler.DecryptionTrigger]
	opts    *options
	config  *kprconfig.Config

	dbpool            *pgxpool.Pool
	shuttermintClient client.Client
	messaging         p2p.Messaging
	messageSender     fx.RPCMessageSender
	blockSyncClient   *ethclient.Client

	shuttermintState *smobserver.ShuttermintState
	metricsServer    *metricsserver.MetricsServer
}

func New(
	config *kprconfig.Config,
	trigger <-chan *broker.Event[*epochkghandler.DecryptionTrigger],
	options ...Option,
) (*KeyperCore, error) {
	opts := newDefaultOptions()
	for _, option := range options {
		err := option(opts)
		if err != nil {
			return nil, err
		}
	}
	sender := opts.messaging
	if sender == nil {
		var err error
		sender, err = p2p.New(config.P2P)
		if err != nil {
			return nil, err
		}
	}
	return &KeyperCore{config: config, trigger: trigger, messaging: sender, opts: opts}, nil
}

// LinkConfigToDB ensures that we use a database compatible with the given config. On first use
// it stores the config's ethereum address into the database. On subsequent uses it compares the
// stored value and raises an error if it doesn't match.
func LinkConfigToDB(ctx context.Context, config *kprconfig.Config, dbpool *pgxpool.Pool) error {
	const addressKey = "ethereum address"
	cfgAddress := config.GetAddress().String()
	queries := db.New(dbpool)
	dbAddr, err := queries.GetMeta(ctx, addressKey)
	if err == pgx.ErrNoRows {
		return queries.InsertMeta(ctx, db.InsertMetaParams{
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

func (kpr *KeyperCore) initOptions(ctx context.Context, runner service.Runner) error {
	err := validateOptions(kpr.opts)
	if err != nil {
		return err
	}
	if kpr.opts.dbpool == nil {
		// connect, but don't validate any database version.
		// If that is desired, it should be done in the keyper-implementation
		kpr.dbpool, err = db.Connect(ctx, runner, kpr.config.DatabaseURL, database.Definition.Name())
		if err != nil {
			return err
		}
		runner.Defer(kpr.dbpool.Close)
	} else {
		kpr.dbpool = kpr.opts.dbpool
	}
	if kpr.opts.blockSyncClient == nil {
		var err error
		kpr.blockSyncClient, err = ethclient.DialContext(ctx, kpr.config.Ethereum.EthereumURL)
		if err != nil {
			return err
		}
	} else {
		kpr.blockSyncClient = kpr.opts.blockSyncClient
	}
	return nil
}

func (kpr *KeyperCore) Start(ctx context.Context, runner service.Runner) error {
	config := kpr.config
	err := kpr.initOptions(ctx, runner)
	if err != nil {
		return err
	}

	err = kpr.dbpool.BeginFunc(db.WrapContext(ctx, database.Definition.Validate))
	if err != nil {
		return err
	}
	err = LinkConfigToDB(ctx, config, kpr.dbpool)
	if err != nil {
		return err
	}
	shuttermintClient, err := tmhttp.New(config.Shuttermint.ShuttermintURL, "/websocket")
	if err != nil {
		return err
	}
	messageSender := fx.NewRPCMessageSender(shuttermintClient, config.Ethereum.PrivateKey.Key)

	if kpr.config.Metrics.Enabled {
		keypermetrics.InitMetrics(kpr.dbpool, *kpr.config)
		epochkghandler.InitMetrics()
		deployment.InitMetrics()
		kpr.metricsServer = metricsserver.New(kpr.config.Metrics)
	}

	kpr.shuttermintClient = shuttermintClient
	kpr.messageSender = messageSender
	kpr.shuttermintState = smobserver.NewShuttermintState(config)

	kpr.messaging.AddMessageHandler(
		epochkghandler.NewDecryptionKeyHandler(kpr.config, kpr.dbpool),
		epochkghandler.NewDecryptionKeyShareHandler(kpr.config, kpr.dbpool),
		// this is purely used to subscribe to the public key topic for broadcast
		epochkghandler.NewEonPublicKeyHandler(kpr.config, kpr.dbpool),
	)
	kpr.messaging.AddMessageHandler(kpr.opts.messageHandler...)
	return runner.StartService(kpr.getServices()...)
}

func (kpr *KeyperCore) getServices() []service.Service {
	services := []service.Service{
		kpr.messaging,
		service.Function{Func: kpr.operateShuttermint},
		newEonPubKeyHandler(kpr),
	}
	keyTrigger := kpr.trigger
	if kpr.config.HTTPEnabled {
		httpServer := kprapi.NewHTTPService(kpr.dbpool, kpr.config, kpr.messaging)
		services = append(services, httpServer)
		// combine two sources of decryption triggers
		// and spawn the fan-in routine
		apiDecrTrig := httpServer.GetDecryptionTriggerChannel()
		fanIn := channel.NewFanInService(kpr.trigger, apiDecrTrig)
		services = append(services, fanIn)
		keyTrigger = fanIn.C
	}
	keyShareHandler := &epochkghandler.KeyShareHandler{
		InstanceID:           kpr.config.GetInstanceID(),
		KeyperAddress:        kpr.config.GetAddress(),
		MaxNumKeysPerMessage: kpr.config.GetMaxNumKeysPerMessage(),
		DBPool:               kpr.dbpool,
		Messaging:            kpr.messaging,
		Trigger:              keyTrigger,
	}
	services = append(services, keyShareHandler)
	if kpr.config.Metrics.Enabled {
		services = append(services, kpr.metricsServer)
	}
	return services
}

func (kpr *KeyperCore) handleOnChainChanges(
	ctx context.Context,
	tx pgx.Tx,
	syncBlockNumber uint64,
) error {
	log.Debug().Uint64("sync-block-number", syncBlockNumber).Msg("handle on chain changes")
	err := kpr.handleOnChainKeyperSetChanges(ctx, tx, syncBlockNumber)
	if err != nil {
		return err
	}
	err = kpr.sendNewBlockSeen(ctx, tx, syncBlockNumber)
	if err != nil {
		return err
	}
	return nil
}

// sendNewBlockSeen sends shmsg.NewBlockSeen messages to the shuttermint chain. This function sends
// NewBlockSeen messages to the shuttermint chain, so that the chain can start new batch configs if
// enough keypers have seen a block past the start block of some BatchConfig. We only send messages
// when the current block we see, could lead to a batch config being started.
func (kpr *KeyperCore) sendNewBlockSeen(ctx context.Context, tx pgx.Tx, l1BlockNumber uint64) error {
	q := database.New(tx)
	lastBlock, err := q.GetLastBlockSeen(ctx)
	if err != nil {
		return err
	}

	count, err := q.CountBatchConfigsInBlockRangeWithKeyper(ctx,
		database.CountBatchConfigsInBlockRangeWithKeyperParams{
			KeyperAddress: []string{kpr.config.GetAddress().String()},
			StartBlock:    lastBlock,
			EndBlock:      int64(l1BlockNumber),
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
func (kpr *KeyperCore) handleOnChainKeyperSetChanges(
	ctx context.Context,
	tx pgx.Tx,
	blockNumber uint64,
) error {
	q := database.New(tx)
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
		log.Debug().
			Int64("keyper-config-index", keyperSet.KeyperConfigIndex).
			Msg("batch config already sent (scheduled).")
		return nil
	}

	activationBlockNumber, err := medley.Int64ToUint64Safe(keyperSet.ActivationBlockNumber)
	if err != nil {
		return err
	}
	// We *MUST* check if the blockNumber is smaller than the activationBlockNumber since both are
	// uint64 and therefore subtraction can never result in negative numbers.
	// This means that if we missed the activationBlockNumber we will never submit the config.
	if blockNumber < activationBlockNumber && activationBlockNumber-blockNumber > kpr.config.Shuttermint.DKGStartBlockDelta {
		log.Info().Interface("keyper-set", keyperSet).
			Uint64("l1-block-number", blockNumber).
			Uint64("dkg-start-delta", kpr.config.Shuttermint.DKGStartBlockDelta).
			Msg("not yet submitting config")
		return nil
	}

	err = q.SetLastBatchConfigSent(ctx, keyperSet.KeyperConfigIndex)
	if err != nil {
		log.Warn().Err(err).
			Interface("keyper-set", keyperSet).
			Int64("keyper-config-index", keyperSet.KeyperConfigIndex).
			Msg("error when setting last batch config sent. Returning nil.")
		return nil
	}

	keypers, err := shdb.DecodeAddresses(keyperSet.Keypers)
	if err != nil {
		return err
	}
	log.Info().Interface("keyper-set", keyperSet).
		Uint64("sync-block-number", blockNumber).
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

// TODO: we need a better block syncing mechanism!
// Also this is doing too much work sequentially in one routine.
func (kpr *KeyperCore) operateShuttermint(ctx context.Context, _ service.Runner) error {
	for {
		syncBlockNumber, err := retry.FunctionCall(ctx, kpr.blockSyncClient.BlockNumber)
		if err != nil {
			return err
		}
		keypermetrics.MetricsKeyperCurrentBlockL1.Set(float64(syncBlockNumber))

		err = smobserver.SyncAppWithDB(ctx, kpr.shuttermintClient, kpr.dbpool, kpr.shuttermintState)
		if err != nil {
			return err
		}
		err = kpr.dbpool.BeginFunc(ctx, func(tx pgx.Tx) error {
			return kpr.handleOnChainChanges(ctx, tx, syncBlockNumber)
		})
		if err != nil {
			return err
		}
		err = fx.SendShutterMessages(ctx, database.New(kpr.dbpool), &kpr.messageSender)
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
