package keypermetrics

import (
	"context"
	"strconv"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog/log"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/database"
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

var MetricsKeyperDKGstatus = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Namespace: "shutter",
		Subsystem: "keyper",
		Name:      "dkg_status",
		Help:      "Is DKG successful",
	},
	[]string{"eon"},
)

func InitMetrics(dbpool *pgxpool.Pool, config kprconfig.Config) {
	prometheus.MustRegister(MetricsKeyperCurrentBlockL1)
	prometheus.MustRegister(MetricsKeyperCurrentBlockShuttermint)
	prometheus.MustRegister(MetricsKeyperCurrentEon)
	prometheus.MustRegister(MetricsKeyperEonStartBlock)
	prometheus.MustRegister(MetricsKeyperIsKeyper)
	prometheus.MustRegister(MetricsKeyperCurrentPhase)
	prometheus.MustRegister(MetricsKeyperCurrentBatchConfigIndex)
	prometheus.MustRegister(MetricsKeyperBatchConfigInfo)
	prometheus.MustRegister(MetricsKeyperDKGstatus)

	queries := database.New(dbpool)
	eons, err := queries.GetAllEons(context.Background())
	if err != nil {
		log.Error().Err(err).Msg("keypermetrics | Failed to get all eons")
		return
	}
	keyperIndex, isKeyper, err := queries.GetKeyperIndex(context.Background(), eons[len(eons)-1].KeyperConfigIndex, config.GetAddress())
	if err != nil {
		log.Error().Err(err).Msg("keypermetrics | Failed to get keyper index")
		return
	}
	if isKeyper {
		MetricsKeyperIsKeyper.WithLabelValues(strconv.FormatInt(keyperIndex, 10)).Set(1)
	} else {
		MetricsKeyperIsKeyper.WithLabelValues(strconv.FormatInt(keyperIndex, 10)).Set(0)
	}

	dkgResult, err := queries.GetDKGResultForKeyperConfigIndex(context.Background(), eons[len(eons)-1].KeyperConfigIndex)
	if err != nil {
		MetricsKeyperDKGstatus.WithLabelValues(strconv.FormatInt(eons[len(eons)-1].Eon, 10)).Set(0)
		log.Error().Err(err).Msg("keypermetrics | Failed to get dkg result")
		return
	}
	if dkgResult.Success {
		MetricsKeyperDKGstatus.WithLabelValues(strconv.FormatInt(eons[len(eons)-1].Eon, 10)).Set(1)
	} else {
		MetricsKeyperDKGstatus.WithLabelValues(strconv.FormatInt(eons[len(eons)-1].Eon, 10)).Set(0)
	}
  
  version, err := chainsync.GetClientVersion(context.Background(), config.Ethereum.EthereumURL)
	if err != nil {
		log.Error().Err(err).Msg("execution_client_version metrics | Failed to get execution client version")
		return
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
	metricsKeyperEthAddress := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "shutter",
			Subsystem: "keyper",
			Name:      "address",
			Help:      "Ethereum address of the Keyper",
			ConstLabels: prometheus.Labels{
				"address": config.GetAddress().Hex(),
			},
		},
	)
	metricsKeyperEthAddress.Set(1)

	prometheus.MustRegister(metricsKeyperEthAddress)
}
