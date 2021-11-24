package collator

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v4/pgxpool"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/proto"

	"github.com/shutter-network/shutter/shuttermint/collator/cltrdb"
	"github.com/shutter-network/shutter/shuttermint/collator/cltrtopics"
	"github.com/shutter-network/shutter/shuttermint/contract/deployment"
	"github.com/shutter-network/shutter/shuttermint/p2p"
	"github.com/shutter-network/shutter/shuttermint/shdb"
	"github.com/shutter-network/shutter/shuttermint/shmsg"
)

type Collator struct {
	Config Config

	contracts *deployment.Contracts

	p2p *p2p.P2P
	db  *cltrdb.Queries
}

var gossipTopicNames = [2]string{
	cltrtopics.CipherBatch,
	cltrtopics.DecryptionTrigger,
}

func New(config Config) *Collator {
	p2pConfig := p2p.Config{
		ListenAddr:     config.ListenAddress,
		PeerMultiaddrs: config.PeerMultiaddrs,
		PrivKey:        config.P2PKey,
	}
	p := p2p.New(p2pConfig)

	return &Collator{
		Config: config,

		contracts: nil,

		p2p: p,
		db:  nil,
	}
}

func (c *Collator) setupRouter() *chi.Mux {
	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("pong"))
	})
	router.Get("/slow", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(5 * time.Second)
		w.Write([]byte("slow"))
	})
	return router
}

func (c *Collator) Run(ctx context.Context) error {
	log.Printf(
		"starting collator with ethereum address %X",
		c.Config.EthereumAddress(),
	)

	dbpool, err := pgxpool.Connect(ctx, c.Config.DatabaseURL)
	if err != nil {
		return errors.Wrap(err, "failed to connect to database")
	}
	defer dbpool.Close()
	log.Printf("Connected to database (%s)", shdb.ConnectionInfo(dbpool))

	ethereumClient, err := ethclient.Dial(c.Config.EthereumURL)
	if err != nil {
		return err
	}
	contracts, err := deployment.NewContracts(ethereumClient, c.Config.DeploymentDir)
	if err != nil {
		return err
	}
	c.contracts = contracts

	err = cltrdb.ValidateDB(ctx, dbpool)
	if err != nil {
		return err
	}
	db := cltrdb.New(dbpool)
	c.db = db

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

func (c *Collator) processEpochLoop(ctx context.Context) error {
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

func (c *Collator) newEpoch(ctx context.Context) error {
	outMessages, err := handleEpoch(ctx, c.Config, c.db)
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

func (c *Collator) sendMessage(ctx context.Context, msg shmsg.P2PMessage) error {
	msgBytes, err := proto.Marshal(msg)
	if err != nil {
		return errors.Wrap(err, "failed to marshal p2p message")
	}
	log.Printf("sending message %v", msg)

	return c.p2p.Publish(ctx, msg.Topic(), msgBytes)
}
