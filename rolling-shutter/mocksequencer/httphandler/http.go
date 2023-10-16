package httphandler

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httputil"
	"strings"
	"time"

	ethrpc "github.com/ethereum/go-ethereum/rpc"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/encodeable/url"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/service"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/mocksequencer/rpc"
)

func (srv *server) rpcHandler() (http.Handler, error) {
	rpcServices := []rpc.RPCService{
		&rpc.ShutterService{},
	}
	if srv.config.EnableAdminService {
		rpcServices = append(rpcServices, &rpc.AdminService{})
	}

	rpcServer := ethrpc.NewServer()
	for _, service := range rpcServices {
		service.InjectProcessor(srv.processor)
		err := rpcServer.RegisterName(service.Name(), service)
		if err != nil {
			return nil, errors.Wrap(err, "error while trying to register RPCService")
		}
	}

	p := &JSONRPCProxy{
		l2Backend: httputil.NewSingleHostReverseProxy(srv.config.L2BackendURL.URL),
		sequencer: rpcServer,
	}
	// handler := injectHTTPLogger(p)
	return p, nil
}

type JSONRPCProxy struct {
	l2Backend http.Handler
	sequencer http.Handler
}

func (p *JSONRPCProxy) SelectHandler(method string) http.Handler {
	// route the eth_namespace to the l2-backend
	if strings.HasPrefix(method, "eth_") {
		return p.l2Backend
	}
	return p.sequencer
}

func (p *JSONRPCProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	rpcreq := medley.RPCRequest{}
	err = json.Unmarshal(body, &rpcreq)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	log.Info().Str("method", rpcreq.Method).Msg("dispatching")

	// make the body available again before letting reverse proxy handle the rest
	r.Body = io.NopCloser(bytes.NewBuffer(body))
	p.SelectHandler(rpcreq.Method).ServeHTTP(w, r)
}

type Config struct {
	L2BackendURL       *url.URL
	HTTPListenAddress  string
	EnableAdminService bool
}

type server struct {
	processor rpc.Sequencer
	config    *Config
}

func NewRPCService(processor rpc.Sequencer, config *Config) service.Service {
	return &server{
		processor: processor,
		config:    config,
	}
}

func (srv *server) setupRouter() (*chi.Mux, error) {
	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	// router.Mount("/v1", http.StripPrefix("/v1", srv.setupAPIRouter(swagger)))
	// apiJSON, _ := json.Marshal(swagger)
	// router.Get("/api.json", func(w http.ResponseWriter, r *http.Request) {
	// 	w.Header().Set("Access-Control-Allow-Origin", "*")
	// 	w.Header().Set("Access-Control-Allow-Headers", "*")
	// 	w.Header().Set("Access-Control-Allow-Methods", "POST, GET")
	// 	w.Header().Set("Content-Type", "application/json")
	// 	_, _ = w.Write(apiJSON)
	// })
	handler, err := srv.rpcHandler()
	if err != nil {
		return nil, err
	}
	router.Mount("/", handler)
	return router, nil
}

func (srv *server) Start(ctx context.Context, runner service.Runner) error {
	handler, err := srv.setupRouter()
	if err != nil {
		return err
	}
	httpServer := &http.Server{
		Addr:              srv.config.HTTPListenAddress,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}
	runner.Go(httpServer.ListenAndServe)
	runner.Go(func() error {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		return httpServer.Shutdown(shutdownCtx)
	})
	return nil
}
