package collator

import (
	"context"
	"encoding/json"
	"log"
	"math"
	"net/http"
	"os"
	"time"

	chimiddleware "github.com/deepmap/oapi-codegen/pkg/chi-middleware"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/chainobserver"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/collator/batch"
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
	"github.com/shutter-network/rolling-shutter/rolling-shutter/shmsg"
)

type collator struct {
	Config config.Config

	contracts    *deployment.Contracts
	batchHandler *batch.BatchHandler

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

	ethereumClient, err := ethclient.Dial(cfg.EthereumURL)
	if err != nil {
		return err
	}
	contracts, err := deployment.NewContracts(ethereumClient, cfg.DeploymentDir)
	if err != nil {
		return err
	}

	err = cltrdb.ValidateDB(ctx, dbpool)
	if err != nil {
		return err
	}

	err = initializeNextBatch(ctx, cltrdb.New(dbpool), contracts)
	if err != nil {
		return err
	}

	batchHandler, err := batch.NewBatchHandler(cfg, dbpool)
	if err != nil {
		return err
	}

	c := collator{
		Config: cfg,

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
func initializeNextBatch(ctx context.Context, db *cltrdb.Queries, contracts *deployment.Contracts) error {
	_, err := db.GetNextBatch(ctx)
	if err == pgx.ErrNoRows {
		blk, err := getBlockNumber(ctx, contracts.Client)
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
		return c.processEpochLoop(errorctx)
	})
	errorgroup.Go(func() error {
		return c.submitBatches(errorctx)
	})

	return errorgroup.Wait()
}

func (c *collator) handleContractEvents(ctx context.Context) error {
	events := []*eventsyncer.EventType{
		c.contracts.KeypersConfigsListNewConfig,
	}
	return chainobserver.New(c.contracts, c.dbpool).Observe(ctx, events)
}

func (c *collator) processEpochLoop(ctx context.Context) error {
	sleepDuration := c.Config.EpochDuration

	for {
		select {
		case <-time.After(sleepDuration):
			if err := c.newEpoch(ctx); err != nil {
				log.Printf("error creating new epoch: %s", err)
				return err
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (c *collator) newEpoch(ctx context.Context) error {
	var outMessages []shmsg.P2PMessage

	blockNumber, err := getBlockNumber(ctx, c.contracts.Client)
	if err != nil {
		return err
	}
	if blockNumber > math.MaxUint32 {
		return errors.Errorf("block number too big: %d", blockNumber)
	}

	outMessages, err = c.batchHandler.StartNextEpoch(ctx, uint32(blockNumber))
	if err != nil {
		return err
	}
	for _, msgOut := range outMessages {
		if err := c.p2p.SendMessage(ctx, msgOut); err != nil {
			log.Printf("error sending message %+v: %s", msgOut, err)
			continue
		}
	}
	return nil
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
