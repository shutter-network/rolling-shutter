package snapshot

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/log"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
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

func (snp *Snapshot) runMetricsServer(ctx context.Context) error {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	addr := fmt.Sprintf("%s:%d", snp.Config.MetricsHost, snp.Config.MetricsPort)
	log.Info("Running metrics server at %s", addr)

	return http.ListenAndServe(addr, mux)
}
