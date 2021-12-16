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
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/proto"

	"github.com/shutter-network/shutter/shuttermint/collator/cltrdb"
	"github.com/shutter-network/shutter/shuttermint/collator/cltrtopics"
	"github.com/shutter-network/shutter/shuttermint/collator/oapi"
	"github.com/shutter-network/shutter/shuttermint/contract/deployment"
	"github.com/shutter-network/shutter/shuttermint/medley"
	"github.com/shutter-network/shutter/shuttermint/medley/epochid"
	"github.com/shutter-network/shutter/shuttermint/p2p"
	"github.com/shutter-network/shutter/shuttermint/shdb"
	"github.com/shutter-network/shutter/shuttermint/shmsg"
)

type collator struct {
	Config Config

	contracts *deployment.Contracts

	p2p    *p2p.P2P
	dbpool *pgxpool.Pool
}

var gossipTopicNames = [2]string{
	cltrtopics.CipherBatch,
	cltrtopics.DecryptionTrigger,
}

func Run(ctx context.Context, config Config) error {
	log.Printf(
		"starting collator with ethereum address %s",
		config.EthereumAddress(),
	)

	dbpool, err := pgxpool.Connect(ctx, config.DatabaseURL)
	if err != nil {
		return errors.Wrap(err, "failed to connect to database")
	}
	defer dbpool.Close()
	log.Printf("Connected to database (%s)", shdb.ConnectionInfo(dbpool))

	ethereumClient, err := ethclient.Dial(config.EthereumURL)
	if err != nil {
		return err
	}
	contracts, err := deployment.NewContracts(ethereumClient, config.DeploymentDir)
	if err != nil {
		return err
	}

	err = cltrdb.ValidateDB(ctx, dbpool)
	if err != nil {
		return err
	}

	err = initializeEpochID(ctx, cltrdb.New(dbpool), contracts)
	if err != nil {
		return err
	}

	c := collator{
		Config: config,

		contracts: contracts,

		p2p: p2p.New(p2p.Config{
			ListenAddr:     config.ListenAddress,
			PeerMultiaddrs: config.PeerMultiaddrs,
			PrivKey:        config.P2PKey,
		}),

		dbpool: dbpool,
	}
	return c.run(ctx)
}

// initializeEpochID populate the epoch_id table with a valid value if it is empty.
func initializeEpochID(ctx context.Context, db *cltrdb.Queries, contracts *deployment.Contracts) error {
	_, err := db.GetNextEpochID(ctx)
	if err == pgx.ErrNoRows {
		blkUntyped, err := medley.Retry(ctx, func() (interface{}, error) {
			return contracts.Client.BlockNumber(ctx)
		})
		if err != nil {
			return err
		}
		blk := blkUntyped.(uint64)
		if blk > math.MaxUint32 {
			return errors.Errorf("block number too big: %d", blk)
		}

		epochID := epochid.New(0, uint32(blk))
		return db.SetNextEpochID(ctx, shdb.EncodeUint64(epochID))
	}
	return err
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
		return c.p2p.Run(errorctx, gossipTopicNames[:], map[string]pubsub.Validator{})
	})
	errorgroup.Go(func() error {
		return c.processEpochLoop(errorctx)
	})

	return errorgroup.Wait()
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

	blockNumberUntyped, err := medley.Retry(ctx, func() (interface{}, error) {
		return c.contracts.Client.BlockNumber(ctx)
	})
	if err != nil {
		return err
	}
	blockNumber := blockNumberUntyped.(uint64)
	if blockNumber > math.MaxUint32 {
		return errors.Errorf("block number too big: %d", blockNumber)
	}

	err = c.dbpool.BeginFunc(ctx, func(tx pgx.Tx) error {
		// Disallow submitting transactions at the same time.
		_, err = tx.Exec(ctx, "LOCK TABLE collator.decryption_trigger IN SHARE ROW EXCLUSIVE MODE")
		if err != nil {
			return err
		}

		db := cltrdb.New(tx)
		outMessages, err = startNextEpoch(ctx, c.Config, db, uint32(blockNumber))
		return err
	})
	if err != nil {
		return err
	}
	for _, msgOut := range outMessages {
		if err := c.sendMessage(ctx, msgOut); err != nil {
			log.Printf("error sending message %+v: %s", msgOut, err)
			continue
		}
	}
	return nil
}

func (c *collator) sendMessage(ctx context.Context, msg shmsg.P2PMessage) error {
	msgBytes, err := proto.Marshal(msg)
	if err != nil {
		return errors.Wrap(err, "failed to marshal p2p message")
	}
	log.Printf("sending %s", msg.ProtoReflect().Descriptor().FullName().Name())

	return c.p2p.Publish(ctx, msg.Topic(), msgBytes)
}

// getNextEpochID gets the epochID that will be used for the next decryption trigger or cipher batch.
func getNextEpochID(ctx context.Context, db *cltrdb.Queries) (uint64, error) {
	epochID, err := db.GetNextEpochID(ctx)
	if err != nil {
		// There should already be an epochID in the database so not finding a row is an error
		return 0, err
	}
	return shdb.DecodeUint64(epochID), nil
}
