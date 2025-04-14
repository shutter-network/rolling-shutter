package shutterservice

import "github.com/prometheus/client_golang/prometheus"

var metricsRegistryEventsSyncedUntil = prometheus.NewGauge(
	prometheus.GaugeOpts{
		Namespace: "shutter",
		Subsystem: "shutter_api",
		Name:      "registry_events_syned_until",
		Help:      "Current value of the latest block fetched",
	},
)

func init() {
	prometheus.MustRegister(metricsRegistryEventsSyncedUntil)
}
