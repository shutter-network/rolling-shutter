package metricsserver

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog/log"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/service"
)

type MetricsServer struct {
	mux        *http.ServeMux
	config     *MetricsConfig
	httpServer *http.Server
}

func New(config *MetricsConfig) *MetricsServer {
	return &MetricsServer{config: config, mux: http.NewServeMux()}
}

func (srv *MetricsServer) Start(ctx context.Context, group service.Runner) error { //nolint:unparam
	group.Go(func() error {
		srv.mux.Handle("/metrics", promhttp.Handler())
		addr := fmt.Sprintf("%s:%d", srv.config.Host, srv.config.Port)
		srv.httpServer = &http.Server{
			Addr:         addr,
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 5 * time.Second,
			Handler:      srv.mux,
		}

		log.Info().Str("address", addr).Msg("Running metrics server at")
		if err := srv.httpServer.ListenAndServe(); err != http.ErrServerClosed {
			return err
		}
		return nil
	})
	group.Go(func() error {
		<-ctx.Done()
		srv.Shutdown()
		return ctx.Err()
	})
	return nil
}

func (srv *MetricsServer) Shutdown() {
	timeoutCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	connectionsClosed := make(chan struct{})
	go func() {
		if err := srv.httpServer.Shutdown(timeoutCtx); err != nil {
			log.Error().Err(err).Msg("Error shutting down metrics server")
		}
		close(connectionsClosed)
	}()
	<-connectionsClosed
	log.Debug().Msg("Metrics server shut down")
}
