package collator

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	chimiddleware "github.com/deepmap/oapi-codegen/pkg/chi-middleware"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v4/pgxpool"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/pkg/errors"
	"golang.org/x/crypto/sha3"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/proto"

	"github.com/shutter-network/shutter/shuttermint/collator/cltrdb"
	"github.com/shutter-network/shutter/shuttermint/collator/cltrtopics"
	"github.com/shutter-network/shutter/shuttermint/collator/oapi"
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

type Server struct {
	c *Collator
}

func sendError(w http.ResponseWriter, code int, message string) {
	e := oapi.Error{
		Code:    int32(code),
		Message: message,
	}
	w.WriteHeader(code)
	w.Header().Set("Content-Type", "application/json")

	_ = json.NewEncoder(w).Encode(e)
}

func (srv *Server) Ping(w http.ResponseWriter, _ *http.Request) {
	w.Write([]byte("pong"))
}

func (srv *Server) SubmitTransaction(w http.ResponseWriter, r *http.Request) {
	var x oapi.SubmitTransactionJSONBody
	if err := json.NewDecoder(r.Body).Decode(&x); err != nil {
		sendError(w, http.StatusBadRequest, "Invalid format for SubmitTransaction")
		return
	}
	ctx := r.Context()

	hash := sha3.New256()
	fmt.Fprintf(hash, "%d\n", len(x.Epoch))
	hash.Write(x.Epoch)
	hash.Write(x.EncryptedTx)
	txid := hash.Sum(nil)

	err := srv.c.db.InsertTx(ctx, cltrdb.InsertTxParams{
		TxID:        txid,
		EpochID:     x.Epoch,
		EncryptedTx: x.EncryptedTx,
	})
	if err != nil {
		sendError(w, http.StatusConflict, err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(oapi.TransactionId{Id: txid})
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

func (c *Collator) setupAPIRouter(swagger *openapi3.T) http.Handler {
	router := chi.NewRouter()

	router.Use(chimiddleware.OapiRequestValidator(swagger))

	_ = oapi.HandlerFromMux(&Server{c: c}, router)

	return router
}

func (c *Collator) setupRouter() *chi.Mux {
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
		w.Write(apiJSON)
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
