package keypermetrics

import (
	"context"

	"github.com/ethereum/go-ethereum/rpc"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog/log"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/kprconfig"
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

func InitMetrics(config *kprconfig.Config) {
	prometheus.MustRegister(MetricsKeyperCurrentBlockL1)
	prometheus.MustRegister(MetricsKeyperCurrentBlockShuttermint)
	prometheus.MustRegister(MetricsKeyperCurrentEon)
	prometheus.MustRegister(MetricsKeyperEonStartBlock)
	prometheus.MustRegister(MetricsKeyperIsKeyper)
	prometheus.MustRegister(MetricsKeyperCurrentPhase)
	prometheus.MustRegister(MetricsKeyperCurrentBatchConfigIndex)
	prometheus.MustRegister(MetricsKeyperBatchConfigInfo)

	version, err := getClientVersion(config.Ethereum.EthereumURL)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get execution client version")
	}

	executionClientVersion := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "shutter",
			Subsystem: "keyper",
			Name:      "execution_client_version",
			Help:      "Version of the execution client",
			ConstLabels: prometheus.Labels{
				"version": version,
			},
		},
	)
	prometheus.MustRegister(executionClientVersion)
}

func getClientVersion(rpcURL string) (string, error) {
	client, err := rpc.DialContext(context.Background(), rpcURL)
	if err != nil {
		return "", err
	}
	var version string
	err = client.CallContext(context.Background(), &version, "web3_clientVersion")
	return version, err
}
