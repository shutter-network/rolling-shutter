package snapshot

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/ethereum/go-ethereum/log"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/service"
)

var metricKeysGenerated = prometheus.NewCounter(
	prometheus.CounterOpts{
		Namespace: "shutter",
		Subsystem: "snapshot",
		Name:      "proposal_keys_generated_total",
		Help:      "Number of generated proposal keys",
	},
)

var metricEons = prometheus.NewCounter(
	prometheus.CounterOpts{
		Namespace: "shutter",
		Subsystem: "snapshot",
		Name:      "eons_total",
		Help:      "Number of eons",
	},
)

func (snp *Snapshot) initMetrics(ctx context.Context) error {
	prometheus.MustRegister(metricEons)
	prometheus.MustRegister(metricKeysGenerated)

	eonCount, err := snp.db.GetEonCount(ctx)
	if err != nil {
		return err
	}
	metricEons.Add(float64(eonCount))

	keyCount, err := snp.db.GetDecryptionKeyCount(ctx)
	if err != nil {
		return err
	}
	metricKeysGenerated.Add(float64(keyCount))

	return nil
}

type MetricsServer struct {
	mux    *http.ServeMux
	config *Config
}

func NewMetricsServer(config Config) service.Service {
	return &MetricsServer{config: &config, mux: http.NewServeMux()}
}

func (srv *MetricsServer) Start(_ context.Context, _ service.Runner) error {
	srv.mux.Handle("/metrics", promhttp.Handler())
	addr := fmt.Sprintf("%s:%d", srv.config.MetricsHost, srv.config.MetricsPort)
	server := &http.Server{
		Addr:         addr,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		Handler:      srv.mux,
	}

	log.Info("Running metrics server at %s", addr)
	return server.ListenAndServe()
}
