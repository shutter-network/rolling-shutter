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

	"github.com/shutter-network/rolling-shutter/rolling-shutter/chainobserver"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/collator/batcher"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/collator/cltrtopics"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/collator/config"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/collator/database"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/collator/l2client"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/collator/oapi"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/contract/deployment"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/db"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/identitypreimage"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/retry"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/service"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2p"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2pmsg"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/shdb"
)

var errTriggerAlreadySent error = errors.New("decryption-trigger already sent")

type signals struct {
	newDecryptionTrigger shdb.SignalFunc
	newDecryptionKey     shdb.SignalFunc
	newBatchTx           shdb.SignalFunc
}

type collator struct {
	Config *config.Config

	l1Client  *ethclient.Client
	l2Client  *rpc.Client
	contracts *deployment.Contracts
	batcher   *batcher.Batcher
	p2p       *p2p.P2PMessaging
	dbpool    *pgxpool.Pool
	submitter *Submitter
	signals   signals
}

func New(cfg *config.Config) service.Service {
	return &collator{Config: cfg}
}

func (c *collator) Start(ctx context.Context, runner service.Runner) error {
	var err error
	cfg := c.Config
	log.Info().Str("ethereum-address", cfg.Ethereum.PrivateKey.EthereumAddress().Hex()).Msg(
		"starting collator",
	)

	c.dbpool, err = db.Connect(ctx, runner, cfg.DatabaseURL, database.Definition.Name())
	if err != nil {
		return errors.Wrap(err, "failed to connect to database")
	}

	log.Info().Str("ethereum-url", cfg.Ethereum.EthereumURL).Msg("connecting to ethereum")
	l1Client, err := ethclient.Dial(cfg.Ethereum.EthereumURL)
	if err != nil {
		return err
	}
	log.Info().Str("contracts-url", cfg.Ethereum.ContractsURL).Msg("connecting contracts")
	contractsClient, err := ethclient.Dial(cfg.Ethereum.ContractsURL)
	if err != nil {
		return err
	}
	contracts, err := deployment.NewContracts(contractsClient, cfg.Ethereum.DeploymentDir)
	if err != nil {
		return err
	}

	err = c.dbpool.BeginFunc(db.WrapContext(ctx, database.Definition.Validate))
	if err != nil {
		return err
	}

	btchr, err := batcher.NewBatcher(ctx, cfg, c.dbpool)
	if err != nil {
		return err
	}

	log.Info().Str("sequencer-url", cfg.SequencerURL).Msg("connecting sequencer")
	l2RPCClient, err := rpc.Dial(cfg.SequencerURL)
	if err != nil {
		return err
	}
	submitter, err := NewSubmitter(ctx, cfg, c.dbpool)
	if err != nil {
		return err
	}

	c.l1Client = l1Client
	c.l2Client = l2RPCClient
	c.contracts = contracts

	c.p2p, err = p2p.New(cfg.P2P)
	if err != nil {
		return err
	}
	c.batcher = btchr
	c.submitter = submitter
	c.submitter.collator = c
	c.setupP2PHandler()

	chainobs := chainobserver.New(l1Client, c.dbpool)

	// FIXME:why doesn't the collator listen for the collator contract?
	err = chainobs.AddListenEvent(
		c.contracts.KeypersConfigsListNewConfig,
	)
	if err != nil {
		return err
	}

	httpServer := &http.Server{
		Addr:              c.Config.HTTPListenAddress,
		Handler:           c.setupRouter(),
		ReadHeaderTimeout: 10 * time.Second,
	}

	newSignal := func(signalName string, handler shdb.SignalHandler) shdb.SignalFunc {
		sig, loop := shdb.NewSignal(ctx, signalName, handler)
		runner.Go(loop)
		return sig
	}
	c.signals.newDecryptionTrigger = newSignal(newDecryptionTrigger, c.sendDecryptionTriggers)
	c.signals.newDecryptionKey = newSignal(newDecryptionKey, c.submitter.submitBatch)
	c.signals.newBatchTx = newSignal(newBatchtx, c.submitter.submitBatchTxToSequencer)

	runner.Go(httpServer.ListenAndServe)
	runner.Go(func() error {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		return httpServer.Shutdown(shutdownCtx)
	})
	err = runner.StartService(c.p2p, chainobs)
	if err != nil {
		return err
	}
	runner.Go(func() error {
		return c.handleDatabaseNotifications(ctx)
	})
	runner.Go(func() error {
		return c.closeBatchesTicker(ctx, c.Config.EpochDuration.Duration)
	})
	return nil
}

func (c *collator) setupP2PHandler() {
	c.p2p.AddMessageHandler(
		&eonPublicKeyHandler{config: c.Config, dbpool: c.dbpool},
		&decryptionKeyHandler{Config: c.Config, dbpool: c.dbpool},
	)

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

func getNextEpochID(ctx context.Context, queries *database.Queries) (identitypreimage.IdentityPreimage, error) {
	var nextIdentityPreimage identitypreimage.IdentityPreimage
	b, err := queries.GetNextBatch(ctx)
	if err != nil {
		return nextIdentityPreimage, err
	}
	return identitypreimage.IdentityPreimage(b.EpochID), nil
}

func (c *collator) getUnsentDecryptionTriggers(
	ctx context.Context,
	cfg *config.Config,
) (
	[]*p2pmsg.DecryptionTrigger,
	error,
) {
	var triggers []database.DecryptionTrigger
	err := c.dbpool.BeginFunc(ctx, func(dbtx pgx.Tx) error {
		var err error
		db := database.New(dbtx)
		triggers, err = db.GetUnsentTriggers(ctx)
		return err
	})
	if err != nil {
		return nil, err
	}
	trigMsgs := make([]*p2pmsg.DecryptionTrigger, len(triggers))
	for i, trig := range triggers {
		identityPreimage := identitypreimage.IdentityPreimage(trig.EpochID)
		trigMsg, err := p2pmsg.NewSignedDecryptionTrigger(
			cfg.InstanceID,
			identityPreimage,
			uint64(trig.L1BlockNumber),
			trig.BatchHash,
			cfg.Ethereum.PrivateKey.Key,
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
	blk, err := retry.FunctionCall(ctx, client.BlockNumber)
	if err != nil {
		return 0, err
	}
	return blk, nil
}
