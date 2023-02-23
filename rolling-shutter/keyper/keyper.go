// Package keyper contains the keyper implementation
package keyper

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"time"

	chimiddleware "github.com/deepmap/oapi-codegen/pkg/chi-middleware"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog/log"
	"github.com/tendermint/tendermint/rpc/client"
	tmhttp "github.com/tendermint/tendermint/rpc/client/http"
	"golang.org/x/sync/errgroup"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/chainobserver"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/contract/deployment"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/db/commondb"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/db/kprdb"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/db/metadb"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/fx"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/kproapi"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/eventsyncer"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/retry"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2p"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/shdb"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/shmsg"
)

type keyper struct {
	config            Config
	dbpool            *pgxpool.Pool
	db                *kprdb.Queries
	shuttermintClient client.Client
	messageSender     fx.RPCMessageSender
	l1Client          *ethclient.Client
	contracts         *deployment.Contracts

	shuttermintState *ShuttermintState
	p2p              *p2p.P2PHandler
}

// linkConfigToDB ensures that we use a database compatible with the given config. On first use
// it stores the config's ethereum address into the database. On subsequent uses it compares the
// stored value and raises an error if it doesn't match.
func linkConfigToDB(ctx context.Context, config Config, dbpool *pgxpool.Pool) error {
	const addressKey = "ethereum address"
	cfgAddress := config.Address().Hex()
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

func Run(ctx context.Context, config Config) error {
	dbpool, err := pgxpool.Connect(ctx, config.DatabaseURL)
	if err != nil {
		return errors.Wrap(err, "failed to connect to database")
	}
	defer dbpool.Close()
	shdb.AddConnectionInfo(log.Info(), dbpool).Msg("connected to database")
	db := kprdb.New(dbpool)

	l1Client, err := ethclient.Dial(config.EthereumURL)
	if err != nil {
		return err
	}
	l2Client, err := ethclient.Dial(config.ContractsURL)
	if err != nil {
		return err
	}
	contracts, err := deployment.NewContracts(l2Client, config.DeploymentDir)
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
	shuttermintClient, err := tmhttp.New(config.ShuttermintURL)
	if err != nil {
		return err
	}
	messageSender := fx.NewRPCMessageSender(shuttermintClient, config.SigningKey)

	p2pHandler := p2p.New(p2p.Config{
		ListenAddr:     config.ListenAddress,
		PeerMultiaddrs: config.PeerMultiaddrs,
		PrivKey:        config.P2PKey,
	})

	k := keyper{
		config:            config,
		dbpool:            dbpool,
		db:                db,
		shuttermintClient: shuttermintClient,
		messageSender:     messageSender,
		l1Client:          l1Client,
		contracts:         contracts,

		shuttermintState: NewShuttermintState(config),
		p2p:              p2pHandler,
	}

	k.setupP2PHandler()

	return k.run(ctx)
}

func (kpr *keyper) setupP2PHandler() {
	epochHandler := epochKGHandler{
		config: kpr.config,
		db:     kprdb.New(kpr.dbpool),
	}

	p2p.AddValidator(kpr.p2p, kpr.validateDecryptionKey)
	p2p.AddValidator(kpr.p2p, kpr.validateDecryptionKeyShare)
	p2p.AddValidator(kpr.p2p, kpr.validateEonPublicKey)
	p2p.AddValidator(kpr.p2p, kpr.validateDecryptionTrigger)

	p2p.AddHandlerFunc(kpr.p2p, epochHandler.handleDecryptionTrigger)
	p2p.AddHandlerFunc(kpr.p2p, epochHandler.handleDecryptionKeyShare)
	p2p.AddHandlerFunc(kpr.p2p, epochHandler.handleDecryptionKey)
}

func (kpr *keyper) setupAPIRouter(swagger *openapi3.T) http.Handler {
	router := chi.NewRouter()

	router.Use(chimiddleware.OapiRequestValidator(swagger))

	_ = kproapi.HandlerFromMux(&server{kpr: kpr}, router)

	return router
}

func (kpr *keyper) setupRouter() *chi.Mux {
	swagger, err := kproapi.GetSwagger()
	if err != nil {
		panic(err)
	}
	swagger.Servers = nil

	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Mount("/v1", http.StripPrefix("/v1", kpr.setupAPIRouter(swagger)))
	apiJSON, _ := json.Marshal(swagger)
	router.Get("/api.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(apiJSON)
	})
	router.Mount("/metrics", promhttp.Handler())
	/*
	   The following enables the swagger ui. Run the following to use it:

	     npm pack swagger-ui-dist@4.1.2
	     tar -xf swagger-ui-dist-4.1.2.tgz
	     export SWAGGER_UI=$(pwd)/package
	*/
	swaggerUI := os.Getenv("SWAGGER_UI")
	path := "/ui/"
	if swaggerUI != "" {
		log.Info().Str("path", path).Msg("enabling the swagger ui")
		fs := http.FileServer(http.Dir(os.Getenv("SWAGGER_UI")))
		router.Mount(path, http.StripPrefix(path, fs))
	}

	return router
}

func (kpr *keyper) run(ctx context.Context) error {
	group, ctx := errgroup.WithContext(ctx)

	if kpr.config.HTTPEnabled {
		httpServer := &http.Server{
			Addr:              kpr.config.HTTPListenAddress,
			Handler:           kpr.setupRouter(),
			ReadHeaderTimeout: 5 * time.Second,
		}
		group.Go(httpServer.ListenAndServe)
		group.Go(func() error {
			<-ctx.Done()
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			return httpServer.Shutdown(shutdownCtx)
		})
	}

	group.Go(func() error {
		return kpr.p2p.Run(ctx)
	})
	group.Go(func() error {
		return kpr.operateShuttermint(ctx)
	})
	group.Go(func() error {
		return kpr.broadcastEonPublicKeys(ctx)
	})
	group.Go(func() error {
		return kpr.handleContractEvents(ctx)
	})
	return group.Wait()
}

func (kpr *keyper) handleContractEvents(ctx context.Context) error {
	events := []*eventsyncer.EventType{
		kpr.contracts.KeypersConfigsListNewConfig,
		kpr.contracts.CollatorConfigsListNewConfig,
	}
	return chainobserver.New(kpr.contracts, kpr.dbpool).Observe(ctx, events)
}

func (kpr *keyper) handleOnChainChanges(ctx context.Context, tx pgx.Tx, l1BlockNumber uint64) error {
	err := kpr.handleOnChainKeyperSetChanges(ctx, tx)
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
	err = scheduleShutterMessage(ctx, q, "block seen", blockSeenMsg)
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
func (kpr *keyper) handleOnChainKeyperSetChanges(ctx context.Context, tx pgx.Tx) error {
	q := kprdb.New(tx)
	latestBatchConfig, err := q.GetLatestBatchConfig(ctx)
	if err == pgx.ErrNoRows {
		log.Print("no batch config found in tendermint")
		return nil
	} else if err != nil {
		return err
	}

	cq := commondb.New(tx)
	keyperSet, err := cq.GetKeyperSetByKeyperConfigIndex(ctx, int64(latestBatchConfig.KeyperConfigIndex)+1)
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

	lastBlock, err := q.GetLastBlockSeen(ctx)
	if err != nil {
		return err
	}

	if keyperSet.ActivationBlockNumber-lastBlock > kpr.config.DKGStartBlockDelta {
		log.Info().Interface("keyper-set", keyperSet).Msg("not yet submitting config")
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
	err = scheduleShutterMessage(ctx, q, "new batch config", batchConfigMsg)
	if err != nil {
		return err
	}
	return nil
}

func (kpr *keyper) operateShuttermint(ctx context.Context) error {
	for {
		l1BlockNumber, err := retry.FunctionCall(ctx, func(ctx context.Context) (uint64, error) {
			return kpr.l1Client.BlockNumber(ctx)
		})
		if err != nil {
			return err
		}

		err = SyncAppWithDB(ctx, kpr.shuttermintClient, kpr.dbpool, kpr.shuttermintState)
		if err != nil {
			return err
		}
		err = kpr.dbpool.BeginFunc(ctx, func(tx pgx.Tx) error {
			return kpr.handleOnChainChanges(ctx, tx, l1BlockNumber)
		})
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

func (kpr *keyper) broadcastEonPublicKeys(ctx context.Context) error {
	for {
		eonPublicKeys, err := kpr.db.GetAndDeleteEonPublicKeys(ctx)
		if err != nil {
			return err
		}
		for _, eonPublicKey := range eonPublicKeys {
			_, exists := kprdb.GetKeyperIndex(kpr.config.Address(), eonPublicKey.Keypers)
			if !exists {
				return errors.Errorf("own keyper index not found for Eon=%d", eonPublicKey.Eon)
			}
			msg, err := shmsg.NewSignedEonPublicKey(
				kpr.config.InstanceID,
				eonPublicKey.EonPublicKey,
				uint64(eonPublicKey.ActivationBlockNumber),
				uint64(eonPublicKey.KeyperConfigIndex),
				uint64(eonPublicKey.Eon),
				kpr.config.SigningKey,
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
