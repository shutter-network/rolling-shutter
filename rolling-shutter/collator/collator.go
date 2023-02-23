package collator

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"time"

	chimiddleware "github.com/deepmap/oapi-codegen/pkg/chi-middleware"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/chainobserver"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/collator/batcher"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/collator/cltrtopics"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/collator/config"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/collator/l2client"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/collator/oapi"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/contract/deployment"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/db/cltrdb"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/epochid"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/eventsyncer"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/retry"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2p"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/shdb"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/shmsg"
)

var errTriggerAlreadySent error = errors.New("decryption-trigger already sent")

type signals struct {
	newDecryptionTrigger shdb.SignalFunc
	newDecryptionKey     shdb.SignalFunc
	newBatchTx           shdb.SignalFunc
}

type collator struct {
	Config config.Config

	l1Client  *ethclient.Client
	l2Client  *rpc.Client
	contracts *deployment.Contracts
	batcher   *batcher.Batcher
	p2p       *p2p.P2PHandler
	dbpool    *pgxpool.Pool
	submitter *Submitter
	signals   signals
}

func Run(ctx context.Context, cfg config.Config) error {
	log.Info().Str("ethereum-address", cfg.EthereumAddress().Hex()).Msg(
		"starting collator",
	)

	dbpool, err := pgxpool.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		return errors.Wrap(err, "failed to connect to database")
	}
	defer dbpool.Close()
	shdb.AddConnectionInfo(log.Info(), dbpool).Msg("connected to database")

	l1Client, err := ethclient.Dial(cfg.EthereumURL)
	if err != nil {
		return err
	}
	contractsClient, err := ethclient.Dial(cfg.ContractsURL)
	if err != nil {
		return err
	}
	contracts, err := deployment.NewContracts(contractsClient, cfg.DeploymentDir)
	if err != nil {
		return err
	}

	err = cltrdb.ValidateDB(ctx, dbpool)
	if err != nil {
		return err
	}

	btchr, err := batcher.NewBatcher(ctx, cfg, dbpool)
	if err != nil {
		return err
	}

	l2RPCClient, err := rpc.Dial(cfg.SequencerURL)
	if err != nil {
		return err
	}
	submitter, err := NewSubmitter(ctx, cfg, dbpool)
	if err != nil {
		return err
	}

	c := collator{
		Config: cfg,

		l1Client:  l1Client,
		l2Client:  l2RPCClient,
		contracts: contracts,
		p2p: p2p.New(p2p.Config{
			ListenAddr:     cfg.ListenAddress,
			PeerMultiaddrs: cfg.PeerMultiaddrs,
			PrivKey:        cfg.P2PKey,
		}),
		batcher:   btchr,
		dbpool:    dbpool,
		submitter: submitter,
	}
	c.submitter.collator = &c
	c.setupP2PHandler()

	return c.run(ctx)
}

func (c *collator) setupP2PHandler() {
	p2p.AddValidator(c.p2p, c.validateEonPublicKey)
	p2p.AddHandlerFunc(c.p2p, c.handleEonPublicKey)

	p2p.AddValidator(c.p2p, c.validateDecryptionKey)
	p2p.AddHandlerFunc(c.p2p, c.handleDecryptionKey)

	c.p2p.AddGossipTopic(cltrtopics.DecryptionTrigger)
}

func (c *collator) setupAPIRouter(swagger *openapi3.T) http.Handler {
	router := chi.NewRouter()

	router.Use(chimiddleware.OapiRequestValidator(swagger))

	_ = oapi.HandlerFromMux(&server{c: c}, router)

	return router
}

func (c *collator) setupRouter() *chi.Mux {
	swagger, err := oapi.GetSwagger()
	if err != nil {
		panic(err)
	}
	swagger.Servers = nil

	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Mount("/v1", http.StripPrefix("/v1", c.setupAPIRouter(swagger)))
	apiJSON, _ := json.Marshal(swagger)
	router.Get("/api.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(apiJSON)
	})

	/*
	   The following enables the swagger ui. Run the following to use it:

	     npm pack swagger-ui-dist@4.1.2
	     tar -xf swagger-ui-dist-4.1.2.tgz
	     export SWAGGER_UI=$(pwd)/package
	*/
	swaggerUI := os.Getenv("SWAGGER_UI")
	UIPath := "/ui/"
	if swaggerUI != "" {
		log.Info().Str("path", UIPath).Msg("enabling the Swagger UI")
		fs := http.FileServer(http.Dir(os.Getenv("SWAGGER_UI")))
		router.Mount(UIPath, http.StripPrefix(UIPath, fs))
	}

	return router
}

func (c *collator) listenDatabaseNotifications(ctx context.Context) <-chan *pgconn.Notification {
	chann := make(chan *pgconn.Notification, 1)
	go func() {
		defer close(chann)
		defer log.Debug().Msg("stop listening database notifications")

		conn, err := c.dbpool.Acquire(ctx)
		if err != nil {
			log.Error().Err(err).Msg("error acquiring connection")
			return
		}
		defer conn.Release()

		err = shdb.ExecListenChannels(ctx, conn.Conn(), dbListenChannels)
		if err != nil {
			return
		}
		shdb.SlurpNotifications(ctx, conn.Conn(), chann)
	}()
	return chann
}

func (c *collator) handleDatabaseNotifications(ctx context.Context) error {
	notifications := c.listenDatabaseNotifications(ctx)
	log.Info().Msg("listening for notifications")
	for {
		select {
		case n := <-notifications:
			switch n.Channel {
			case newDecryptionTrigger:
				c.signals.newDecryptionTrigger()
			case newDecryptionKey:
				c.signals.newDecryptionKey()
			case newBatchtx:
				c.signals.newBatchTx()
			default:
				log.Error().
					Str("channel", n.Channel).
					Msg("ignoring database notification for unknown channel")
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (c *collator) run(ctx context.Context) error {
	httpServer := &http.Server{
		Addr:              c.Config.HTTPListenAddress,
		Handler:           c.setupRouter(),
		ReadHeaderTimeout: 10 * time.Second,
	}

	errorgroup, errorctx := errgroup.WithContext(ctx)
	newSignal := func(signalName string, handler shdb.SignalHandler) shdb.SignalFunc {
		sig, loop := shdb.NewSignal(errorctx, signalName, handler)
		errorgroup.Go(loop)
		return sig
	}
	c.signals.newDecryptionTrigger = newSignal(newDecryptionTrigger, c.sendDecryptionTriggers)
	c.signals.newDecryptionKey = newSignal(newDecryptionKey, c.submitter.submitBatch)
	c.signals.newBatchTx = newSignal(newBatchtx, c.submitter.submitBatchTxToSequencer)

	errorgroup.Go(httpServer.ListenAndServe)
	errorgroup.Go(func() error {
		<-errorctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		return httpServer.Shutdown(shutdownCtx)
	})
	errorgroup.Go(func() error {
		return c.handleContractEvents(errorctx)
	})
	errorgroup.Go(func() error {
		return c.p2p.Run(errorctx)
	})
	errorgroup.Go(func() error {
		return c.handleDatabaseNotifications(errorctx)
	})
	errorgroup.Go(func() error {
		return c.closeBatchesTicker(errorctx, c.Config.EpochDuration)
	})

	return errorgroup.Wait()
}

func (c *collator) handleContractEvents(ctx context.Context) error {
	events := []*eventsyncer.EventType{
		c.contracts.KeypersConfigsListNewConfig,
	}
	return chainobserver.New(c.contracts, c.dbpool).Observe(ctx, events)
}

func getNextEpochID(ctx context.Context, db *cltrdb.Queries) (epochid.EpochID, error) {
	var nextEpochID epochid.EpochID
	b, err := db.GetNextBatch(ctx)
	if err != nil {
		return nextEpochID, err
	}
	return epochid.BytesToEpochID(b.EpochID)
}

func (c *collator) getUnsentDecryptionTriggers(
	ctx context.Context,
	cfg config.Config,
) (
	[]*shmsg.DecryptionTrigger,
	error,
) {
	var triggers []cltrdb.DecryptionTrigger
	err := c.dbpool.BeginFunc(ctx, func(dbtx pgx.Tx) error {
		var err error
		db := cltrdb.New(dbtx)
		triggers, err = db.GetUnsentTriggers(ctx)
		return err
	})
	if err != nil {
		return nil, err
	}
	trigMsgs := make([]*shmsg.DecryptionTrigger, len(triggers))
	for i, trig := range triggers {
		epochID, err := epochid.BytesToEpochID(trig.EpochID)
		if err != nil {
			return nil, err
		}
		trigMsg, err := shmsg.NewSignedDecryptionTrigger(
			cfg.InstanceID,
			epochID,
			uint64(trig.L1BlockNumber),
			trig.BatchHash,
			cfg.EthereumKey,
		)
		if err != nil {
			return nil, err
		}
		trigMsgs[i] = trigMsg
	}
	return trigMsgs, nil
}

func (c *collator) getBatchConfirmation(ctx context.Context) (uint64, error) {
	return l2client.GetBatchIndex(ctx, c.l2Client)
}

func getBlockNumber(ctx context.Context, client *ethclient.Client) (uint64, error) {
	blk, err := retry.FunctionCall(ctx, func(ctx context.Context) (uint64, error) {
		return client.BlockNumber(ctx)
	})
	if err != nil {
		return 0, err
	}
	return blk, nil
}
