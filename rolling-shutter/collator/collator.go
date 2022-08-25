package collator

import (
	"context"
	"encoding/json"
	"math"
	"net/http"
	"os"
	"time"

	chimiddleware "github.com/deepmap/oapi-codegen/pkg/chi-middleware"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/chainobserver"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/collator/batchhandler"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/collator/cltrdb"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/collator/cltrtopics"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/collator/config"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/collator/oapi"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/contract/deployment"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/epochid"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/eventsyncer"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2p"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/shdb"
)

type collator struct {
	Config config.Config

	l1Client     *ethclient.Client
	l2Client     *rpc.Client
	contracts    *deployment.Contracts
	batchHandler *batchhandler.BatchHandler

	p2p    *p2p.P2PHandler
	dbpool *pgxpool.Pool
}

func Run(ctx context.Context, cfg config.Config) error {
	log.Printf(
		"starting collator with ethereum address %s",
		cfg.EthereumAddress(),
	)

	dbpool, err := pgxpool.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		return errors.Wrap(err, "failed to connect to database")
	}
	defer dbpool.Close()
	log.Printf("Connected to database (%s)", shdb.ConnectionInfo(dbpool))

	l1Client, err := ethclient.Dial(cfg.EthereumURL)
	if err != nil {
		return err
	}
	l2Client, err := ethclient.Dial(cfg.SequencerURL)
	if err != nil {
		return err
	}
	contracts, err := deployment.NewContracts(l2Client, cfg.DeploymentDir)
	if err != nil {
		return err
	}

	err = cltrdb.ValidateDB(ctx, dbpool)
	if err != nil {
		return err
	}

	batchHandler, err := batchhandler.NewBatchHandler(cfg, dbpool)
	if err != nil {
		return err
	}

	l2RPCClient, err := rpc.Dial(cfg.SequencerURL)
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

		batchHandler: batchHandler,
		dbpool:       dbpool,
	}
	c.setupP2PHandler()

	return c.run(ctx)
}

// initializeNextBatch populates the next_batch table with a valid value if it is empty.
func initializeNextBatch(ctx context.Context, db *cltrdb.Queries, l1Client *ethclient.Client) error {
	_, err := db.GetNextBatch(ctx)
	if err == pgx.ErrNoRows {
		blk, err := getBlockNumber(ctx, l1Client)
		if err != nil {
			return err
		}
		if blk > math.MaxInt64 {
			return errors.Errorf("block number too big: %d", blk)
		}

		epochID, err := epochid.BigToEpochID(common.Big0)
		if err != nil {
			return err
		}
		return db.SetNextBatch(ctx, cltrdb.SetNextBatchParams{
			EpochID:       epochID.Bytes(),
			L1BlockNumber: int64(blk),
		})
	} else if err != nil {
		return err
	}
	return nil
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
	if swaggerUI != "" {
		log.Printf("Enabling the swagger ui at /ui/")
		fs := http.FileServer(http.Dir(os.Getenv("SWAGGER_UI")))
		router.Mount("/ui/", http.StripPrefix("/ui/", fs))
	}

	return router
}

func (c *collator) run(ctx context.Context) error {
	httpServer := &http.Server{
		Addr:    c.Config.HTTPListenAddress,
		Handler: c.setupRouter(),
	}

	errorgroup, errorctx := errgroup.WithContext(ctx)
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
		return c.batchHandler.Run(errorctx)
	})
	errorgroup.Go(func() error {
		return c.sendMessages(errorctx)
	})
	errorgroup.Go(func() error {
		return c.pollBatchConfirmations(errorctx)
	})

	return errorgroup.Wait()
}

func (c *collator) handleContractEvents(ctx context.Context) error {
	events := []*eventsyncer.EventType{
		c.contracts.KeypersConfigsListNewConfig,
	}
	return chainobserver.New(c.contracts, c.dbpool).Observe(ctx, events)
}

func (c *collator) sendMessages(ctx context.Context) error {
	messages := c.batchHandler.Messages()
	for msgOut := range messages {
		if err := c.p2p.SendMessage(ctx, msgOut); err != nil {
			log.Error().Err(err).Str("message", msgOut.LogInfo()).Msg("error sending message")
			continue
		}
	}
	return nil
}

func (c *collator) pollBatchConfirmations(ctx context.Context) error {
	var (
		batchIndex, newBatchIndex uint64
		epochID                   epochid.EpochID
	)

	// poll ten times during one epoch duration.
	// this is arbitrary for now, but allows to easily catch the next batch
	// confirmation and not wait too long until
	// this is noticed and progress is made
	pollTime := time.Duration(int64(c.Config.EpochDuration) / 10)
	ticker := time.NewTicker(pollTime)
	initial := true

	for {
		select {
		case <-ticker.C:
			var err error
			epochID, err = c.getBatchConfirmation(ctx)
			// for now only log errors but keep trying
			if err != nil {
				log.Error().Err(err).Msg("error retrieving the current batch-index from sequencer")
				continue
			}
			newBatchIndex = epochID.Uint64()

			if initial {
				initial = false
				batchIndex = newBatchIndex
				continue
			}

			delta := int(newBatchIndex - batchIndex)
			if delta > 1 {
				// this shouldn't happen because collator is needed to progress the batches
				log.Warn().Err(err).Msg("skipped batch-index")
			}
			for delta > 0 {
				// if we skipped indices, they still have to be pushed
				// to the batchhandler
				batchIndex++
				delta--

				nextEpochID, err := epochid.Uint64ToEpochID(batchIndex)
				if err != nil {
					log.Error().Err(err).Msg("can't decode batch-index to epochid")
					continue
				}
				select {
				case c.batchHandler.ConfirmedBatch() <- nextEpochID:
				case <-ctx.Done():
					return ctx.Err()
				}
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (c *collator) getBatchConfirmation(ctx context.Context) (epochid.EpochID, error) {
	var epochID epochid.EpochID

	f := func() (*string, error) {
		var result string
		log.Debug().Msg("polling batch-index from sequencer")
		err := c.l2Client.CallContext(ctx, &result, "shutter_getBatchIndex")
		if err != nil {
			return nil, err
		}
		return &result, nil
	}

	result, err := medley.Retry(ctx, f)
	if err != nil {
		return epochID, errors.Wrapf(err, "can't retrieve batch-index from sequencer")
	}

	epochID, err = epochid.HexToEpochID(*result)
	if err != nil {
		return epochID, errors.Wrap(err, "can't decode batch-index")
	}
	return epochID, nil
}

func getBlockNumber(ctx context.Context, client *ethclient.Client) (uint64, error) {
	blk, err := medley.Retry(ctx, func() (uint64, error) {
		return client.BlockNumber(ctx)
	})
	if err != nil {
		return 0, err
	}
	return blk, nil
}
