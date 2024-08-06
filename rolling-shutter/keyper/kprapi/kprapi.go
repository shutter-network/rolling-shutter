package kprapi

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"time"

	chimiddleware "github.com/deepmap/oapi-codegen/pkg/chi-middleware"
	"github.com/ethereum/go-ethereum/common"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog/log"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/epochkghandler"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/kproapi"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/broker"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/retry"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/service"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2pmsg"
)

type P2PMessageSender interface {
	SendMessage(ctx context.Context, msg p2pmsg.Message, retryOpts ...retry.Option) error
}

type Config interface {
	GetHTTPListenAddress() string
	GetAddress() common.Address
	GetInstanceID() uint64
}

type Server struct {
	dbpool      *pgxpool.Pool
	config      Config
	p2p         P2PMessageSender
	trigger     chan *broker.Event[*epochkghandler.DecryptionTrigger]
	shutdownSig chan struct{}
}

// Decryption triggering is blocking for now.
const decrTrigChanBufferSize = 0

func NewHTTPService(
	dbpool *pgxpool.Pool,
	config Config,
	p2p P2PMessageSender,
) *Server {
	trigger := make(
		chan *broker.Event[*epochkghandler.DecryptionTrigger],
		decrTrigChanBufferSize,
	)
	return &Server{
		dbpool:      dbpool,
		config:      config,
		p2p:         p2p,
		trigger:     trigger,
		shutdownSig: make(chan struct{}),
	}
}

func (srv *Server) setupRouter() *chi.Mux {
	swagger, err := kproapi.GetSwagger()
	if err != nil {
		panic(err)
	}
	swagger.Servers = nil

	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Mount("/v1", http.StripPrefix("/v1", srv.setupAPIRouter(swagger)))
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

func (srv *Server) GetDecryptionTriggerChannel() <-chan *broker.Event[*epochkghandler.DecryptionTrigger] {
	return srv.trigger
}

func (srv *Server) Start(ctx context.Context, runner service.Runner) error { //nolint:unparam
	httpServer := &http.Server{
		Addr:              srv.config.GetHTTPListenAddress(),
		Handler:           srv.setupRouter(),
		ReadHeaderTimeout: 5 * time.Second,
	}
	runner.Defer(func() { close(srv.shutdownSig) })

	runner.Go(httpServer.ListenAndServe)
	runner.Go(func() error {
		return srv.waitShutdown(ctx)
	})
	runner.Go(func() error {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		close(srv.trigger)
		return httpServer.Shutdown(shutdownCtx)
	})
	return nil
}

func (srv *Server) waitShutdown(ctx context.Context) error {
	for {
		select {
		case _, ok := <-srv.shutdownSig:
			if !ok {
				// channel close without a send
				// means we want to stop the shutdown waiter
				// but not stop execution
				return nil
			}
			return medley.ErrShutdownRequested
		case <-ctx.Done():
			// we canceled somewhere else
			return nil
		}
	}
}

func (srv *Server) setupAPIRouter(swagger *openapi3.T) http.Handler {
	router := chi.NewRouter()

	router.Use(chimiddleware.OapiRequestValidator(swagger))
	_ = kproapi.HandlerFromMux(srv, router)

	return router
}
