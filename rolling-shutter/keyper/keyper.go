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

	obskeyper "github.com/shutter-network/rolling-shutter/rolling-shutter/db/chainobsdb/keyper"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/db/kprdb"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/db/metadb"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/epochkghandler"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/fx"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/kprapi"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/kprconfig"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/smobserver"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/broker"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/metricsserver"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/retry"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/service"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2p"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2pmsg"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/shdb"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/shmsg"
)

type KeyperCore struct {
	P2P *p2p.P2PHandler

	trigger <-chan *broker.Event[*epochkghandler.DecryptionTrigger]
	opts    *options
	config  *kprconfig.Config

	dbpool            *pgxpool.Pool
	shuttermintClient client.Client
	messageSender     fx.RPCMessageSender
	l1Client          *ethclient.Client

	shuttermintState *smobserver.ShuttermintState
	metricsServer    *metricsserver.MetricsServer
}

// TODO get rid of the config and use options instead
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
	err := validateOptions(opts)
	if err != nil {
		return nil, err
	}
	return &KeyperCore{config: config, trigger: trigger}, nil
}

// LinkConfigToDB ensures that we use a database compatible with the given config. On first use
// it stores the config's ethereum address into the database. On subsequent uses it compares the
// stored value and raises an error if it doesn't match.
func LinkConfigToDB(ctx context.Context, config *kprconfig.Config, dbpool *pgxpool.Pool) error {
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

func (kpr *KeyperCore) Start(ctx context.Context, runner service.Runner) error {
	config := kpr.config
	dbpool, err := pgxpool.Connect(ctx, config.DatabaseURL)
	if err != nil {
		return errors.Wrap(err, "failed to connect to database")
	}
	runner.Defer(dbpool.Close)
	shdb.AddConnectionInfo(log.Info(), dbpool).Msg("connected to database")

	l1Client, err := ethclient.Dial(config.EthereumURL)
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
	messageSender := fx.NewRPCMessageSender(shuttermintClient, config.EthereumPrivateKey.Key)

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
	kpr.shuttermintState = smobserver.NewShuttermintState(config)
	kpr.P2P = p2pHandler

	kpr.setupP2PHandler()
	return runner.StartService(kpr.getServices()...)
}

func (kpr *KeyperCore) setupP2PHandler() {
	kpr.P2P.AddMessageHandler(
		epochkghandler.NewDecryptionKeyHandler(kpr.config, kpr.dbpool),
		epochkghandler.NewDecryptionKeyShareHandler(kpr.config, kpr.dbpool),
		epochkghandler.NewEonPublicKeyHandler(kpr.config, kpr.dbpool),
	)
}

func (kpr *KeyperCore) getServices() []service.Service {
	services := []service.Service{
		kpr.P2P,
		service.ServiceFn{Fn: kpr.operateShuttermint},
		service.ServiceFn{Fn: kpr.handleNewEonPublicKeys},
		service.ServiceFn{Fn: kpr.listenDecryptionTriggers},
	}
	if kpr.config.HTTPEnabled {
		services = append(services, kprapi.NewHTTPService(kpr.dbpool, kpr.config, kpr.P2P))
	}
	if kpr.config.Metrics.Enabled {
		services = append(services, kpr.metricsServer)
	}
	return services
}

func (kpr *KeyperCore) handleOnChainChanges(
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
func (kpr *KeyperCore) sendNewBlockSeen(ctx context.Context, tx pgx.Tx, l1BlockNumber uint64) error {
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
func (kpr *KeyperCore) handleOnChainKeyperSetChanges(
	ctx context.Context,
	tx pgx.Tx,
	blockNumber uint64,
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
	if blockNumber < activationBlockNumber && activationBlockNumber-blockNumber > kpr.config.Shuttermint.DKGStartBlockDelta {
		log.Info().Interface("keyper-set", keyperSet).
			Uint64("l1-block-number", blockNumber).
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
		Uint64("l1-block-number", blockNumber).
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

// TODO we need a better block syncing mechanism!
// Also this is doing too much work synchronously.
func (kpr *KeyperCore) operateShuttermint(ctx context.Context) error {
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

func (kpr *KeyperCore) getEonForBlockNumber(ctx context.Context, blockNumber uint64) (kprdb.Eon, error) {
	var (
		eon kprdb.Eon
		err error
	)
	db := kprdb.New(kpr.dbpool)
	block, err := medley.Uint64ToInt64Safe(blockNumber)
	if err != nil {
		return eon, errors.Wrap(err, "invalid blocknumber")
	}
	eon, err = db.GetEonForBlockNumber(ctx, block)
	// TODO wrap error
	return eon, err
}

// FIXME name
func (kpr *KeyperCore) listenDecryptionTriggers(ctx context.Context) error {
	for {
		select {
		case triggerEvent, ok := <-kpr.trigger:
			if !ok {
				return nil
			}
			// TODO expose publicly?
			// epochkghandler.metricsEpochKGDectyptionTriggersReceived.Inc()
			trigger := triggerEvent.Value
			eon, err := kpr.getEonForBlockNumber(ctx, trigger.BlockNumber)
			if err != nil {
				// FIXME how to handle
				continue
			}
			keySharesMsg, err := epochkghandler.ConstructDecryptionKeyShare(
				ctx,
				kpr.config,
				kprdb.New(kpr.dbpool),
				eon,
				trigger.EpochIDs,
			)
			if err != nil {
				// FIXME how to handle
				continue
			}

			// TODO retry options
			if err := kpr.P2P.SendMessage(ctx, keySharesMsg); err != nil {
				// FIXME how to handle?
				continue
			}
		case <-ctx.Done():
			return ctx.Err()
		}
		return nil

	}
}

type EonPublicKey struct {
	PublicKey         []byte
	ActivationBlock   uint64
	KeyperConfigIndex uint64
	Eon               uint64
}

// TODO this mechanism could be different as well!
// Not all implementations will expect the keys to be broadcasted
// on the p2pnetwork.
// I think it should be default in any case, and can be turned off with an option
func (kpr *KeyperCore) handleNewEonPublicKeys(ctx context.Context) error {
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
			// FIXME careful with returning errors, since this function is registered as a Service.
			// This will shut down the whole keyper
			activationBlock, err := medley.Int64ToUint64Safe(eonPublicKey.ActivationBlockNumber)
			if err != nil {
				return errors.Wrap(err, "failed safe int cast")
			}
			keyperIndex, err := medley.Int32ToUint64Safe(eonPublicKey.KeyperConfigIndex)
			if err != nil {
				return errors.Wrap(err, "failed safe int cast")
			}
			eon, err := medley.Int64ToUint64Safe(eonPublicKey.Eon)
			if err != nil {
				return errors.Wrap(err, "failed safe int cast")
			}
			eonPubKey := EonPublicKey{
				PublicKey:         eonPublicKey.EonPublicKey,
				ActivationBlock:   activationBlock,
				KeyperConfigIndex: keyperIndex,
				Eon:               eon,
			}
			_ = eonPubKey
			if kpr.opts.broadcastEonPubKey {
				err := kpr.broadcastEonPublicKey(ctx, eonPubKey)
				return errors.Wrap(err, "failed to broadcast eon public key")
			}
			if kpr.opts.eonPubkeyHandler != nil {
				err := kpr.opts.eonPubkeyHandler(ctx, eonPubKey)
				return errors.Wrap(err, "failed to handle eon public key")
			}
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(2 * time.Second):
		}
	}
}

func (kpr *KeyperCore) broadcastEonPublicKey(ctx context.Context, eonPubKey EonPublicKey) error {
	msg, err := p2pmsg.NewSignedEonPublicKey(
		kpr.config.InstanceID,
		eonPubKey.PublicKey,
		eonPubKey.ActivationBlock,
		eonPubKey.KeyperConfigIndex,
		eonPubKey.Eon,
		kpr.config.EthereumPrivateKey.Key,
	)
	if err != nil {
		return errors.Wrap(err, "error while signing EonPublicKey")
	}

	err = kpr.P2P.SendMessage(ctx, msg)
	if err != nil {
		return errors.Wrap(err, "error while broadcasting EonPublicKey")
	}
	return nil
}
