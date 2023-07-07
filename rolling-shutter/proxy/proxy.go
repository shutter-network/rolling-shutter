// Package proxy contains a jsonrpc proxy implementation.
package proxy

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httputil"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/encodeable/url"
)

func NewConfig() *Config {
	c := &Config{}
	c.Init()
	return c
}

func (c *Config) Init() {
	c.SequencerURL = &url.URL{}
	c.CollatorURL = &url.URL{}
}

type Config struct {
	CollatorURL       *url.URL
	SequencerURL      *url.URL
	HTTPListenAddress string
}

func (c *Config) Validate() error {
	if c.CollatorURL == nil {
		return errors.Errorf("configuration value CollatorURL is missing")
	}
	if c.SequencerURL == nil {
		return errors.Errorf("configuration value SequencerURL is missing")
	}
	if c.HTTPListenAddress == "" {
		return errors.Errorf("configuration value HTTPListenAddress is missing")
	}
	return nil
}

func (c *Config) Name() string {
	return "proxy"
}

func (c *Config) SetDefaultValues() error {
	err := c.SequencerURL.UnmarshalText([]byte("http://127.0.0.1:8555/"))
	if err != nil {
		return err
	}
	err = c.CollatorURL.UnmarshalText([]byte("http://127.0.0.1:3000/"))
	if err != nil {
		return err
	}
	c.HTTPListenAddress = ":3001"
	return nil
}

func (c *Config) SetExampleValues() error {
	return c.SetDefaultValues()
}

func (c Config) TOMLWriteHeader(_ io.Writer) (int, error) {
	return 0, nil
}

type JSONRPCProxy struct {
	collator, sequencer *httputil.ReverseProxy
}

func (p *JSONRPCProxy) SelectReverseProxy(method string) *httputil.ReverseProxy {
	switch method {
	case "eth_sendTransaction":
		return p.collator
	case "eth_sendRawTransaction":
		return p.collator
	default:
		return p.sequencer
	}
}

func (p *JSONRPCProxy) HandleRequest(w http.ResponseWriter, r *http.Request) {
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
	p.SelectReverseProxy(rpcreq.Method).ServeHTTP(w, r)
}

// TODO also use the service interface here.
func Run(ctx context.Context, config *Config) error {
	p := JSONRPCProxy{
		collator:  httputil.NewSingleHostReverseProxy(config.CollatorURL.URL),
		sequencer: httputil.NewSingleHostReverseProxy(config.SequencerURL.URL),
	}
	router := chi.NewRouter()
	router.Post("/*", p.HandleRequest)

	httpServer := &http.Server{
		Addr:              config.HTTPListenAddress,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}
	errorgroup, errorctx := errgroup.WithContext(ctx)
	errorgroup.Go(httpServer.ListenAndServe)
	errorgroup.Go(func() error {
		<-errorctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		return httpServer.Shutdown(shutdownCtx)
	})
	return errorgroup.Wait()
}
