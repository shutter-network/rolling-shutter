// Package proxy contains a jsonrpc proxy implementation.
package proxy

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/mitchellh/mapstructure"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"golang.org/x/sync/errgroup"

	"github.com/shutter-network/shutter/shuttermint/medley"
)

type RPCRequest struct {
	Version string      `json:"jsonrpc"`
	Method  string      `json:"method,omitempty"`
	Params  interface{} `json:"params,omitempty"`
	ID      interface{} `json:"id,omitempty"`
}

type Config struct {
	CollatorURL, SequencerURL *url.URL
	HTTPListenAddress         string
}

func (config *Config) Unmarshal(v *viper.Viper) error {
	err := v.Unmarshal(config, viper.DecodeHook(
		mapstructure.ComposeDecodeHookFunc(
			medley.StringToURL,
		)))
	return err
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
	rpcreq := RPCRequest{}
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

func Run(ctx context.Context, config Config) error {
	p := JSONRPCProxy{
		collator:  httputil.NewSingleHostReverseProxy(config.CollatorURL),
		sequencer: httputil.NewSingleHostReverseProxy(config.SequencerURL),
	}
	router := chi.NewRouter()
	router.Post("/*", p.HandleRequest)

	httpServer := &http.Server{
		Addr:    config.HTTPListenAddress,
		Handler: router,
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
