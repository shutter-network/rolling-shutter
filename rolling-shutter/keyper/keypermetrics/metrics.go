package keypermetrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var MetricsKeyperCurrentBlockL1 = prometheus.NewGauge(
	prometheus.GaugeOpts{
		Namespace: "shutter",
		Subsystem: "keyper",
		Name:      "current_block_l1",
		Help:      "Current L1 block number",
	},
)

var MetricsKeyperCurrentBlockShuttermint = prometheus.NewGauge(
	prometheus.GaugeOpts{
		Namespace: "shutter",
		Subsystem: "keyper",
		Name:      "current_block_shuttermint",
		Help:      "Current shuttermint block number",
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

var MetricsKeyperEonStartBlock = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Namespace: "shutter",
		Subsystem: "keyper",
		Name:      "eon_start_block",
		Help:      "Block at which the eon becomes active",
	},
	[]string{"eon"},
)

var MetricsKeyperIsKeyper = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Namespace: "shutter",
		Subsystem: "keyper",
		Name:      "is_keyper",
		Help:      "Is this node a Keyper in the respective batch config",
	},
	[]string{"batch_config_index"},
)

var MetricsKeyperCurrentPhase = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Namespace: "shutter",
		Subsystem: "keyper",
		Name:      "current_phase",
		Help:      "Current DKG phase of this Keyper node",
	},
	[]string{"eon", "phase"},
)

var MetricsKeyperCurrentBatchConfigIndex = prometheus.NewGauge(
	prometheus.GaugeOpts{
		Namespace: "shutter",
		Subsystem: "keyper",
		Name:      "current_batch_config_index",
		Help:      "Current batch config index",
	},
)

var MetricsKeyperBatchConfigInfo = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Namespace: "shutter",
		Subsystem: "keyper",
		Name:      "batch_config_info",
		Help:      "Information about the batch configuration in use",
	},
	[]string{"batch_config_index", "keyper_addresses"})

var MetricsKeyperSuccessfulDKG = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Namespace: "shutter",
		Subsystem: "keyper",
		Name:      "successful_dkg",
		Help:      "Is DKG successful",
	},
	[]string{"eon"},
)

func InitMetrics() {
	prometheus.MustRegister(MetricsKeyperCurrentBlockL1)
	prometheus.MustRegister(MetricsKeyperCurrentBlockShuttermint)
	prometheus.MustRegister(MetricsKeyperCurrentEon)
	prometheus.MustRegister(MetricsKeyperEonStartBlock)
	prometheus.MustRegister(MetricsKeyperIsKeyper)
	prometheus.MustRegister(MetricsKeyperCurrentPhase)
	prometheus.MustRegister(MetricsKeyperCurrentBatchConfigIndex)
	prometheus.MustRegister(MetricsKeyperBatchConfigInfo)
	prometheus.MustRegister(MetricsKeyperSuccessfulDKG)
}
