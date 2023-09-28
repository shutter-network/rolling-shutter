package keypermetrics

import "github.com/prometheus/client_golang/prometheus"

var MetricsKeyperCurrentBlockL1 = prometheus.NewGauge(
	prometheus.GaugeOpts{
		Namespace: "shutter",
		Subsystem: "keyper",
		Name:      "current_block_l1",
		Help:      "Current L1 block number",
	},
)

var MetricsKeyperCurrentEon = prometheus.NewGauge(
	prometheus.GaugeOpts{
		Namespace: "shutter",
		Subsystem: "keyper",
		Name:      "current_eon",
		Help:      "Current eon ID",
	},
)

var MetricsKeyperIsKeyper = prometheus.NewGauge(
	prometheus.GaugeOpts{
		Namespace: "shutter",
		Subsystem: "keyper",
		Name:      "is_keyper",
		Help:      "Is this node a keyper in the current batch config",
	},
)

func InitMetrics() {
	prometheus.MustRegister(MetricsKeyperCurrentBlockL1)
	prometheus.MustRegister(MetricsKeyperCurrentEon)
	prometheus.MustRegister(MetricsKeyperIsKeyper)
}
