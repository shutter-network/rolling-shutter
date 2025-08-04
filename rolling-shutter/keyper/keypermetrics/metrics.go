package keypermetrics

import (
	"context"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog/log"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/database"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/kprconfig"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/chainsync"
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

var MetricsKeyperDKGStatus = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Namespace: "shutter",
		Subsystem: "keyper",
		Name:      "dkg_status",
		Help:      "Is DKG successful",
	},
	[]string{"eon"},
)

var MetricsKeyperEthAddress = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Namespace: "shutter",
		Subsystem: "keyper",
		Name:      "address",
		Help:      "Ethereum address of the Keyper",
	}, []string{"address"})

var MetricsExecutionClientVersion = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Namespace: "shutter",
		Subsystem: "keyper",
		Name:      "execution_client_version",
		Help:      "Version of the execution client",
	},
	[]string{"version"})

func InitMetrics(dbpool *pgxpool.Pool, config kprconfig.Config) {
	prometheus.MustRegister(MetricsKeyperCurrentBlockL1)
	prometheus.MustRegister(MetricsKeyperCurrentBlockShuttermint)
	prometheus.MustRegister(MetricsKeyperCurrentEon)
	prometheus.MustRegister(MetricsKeyperEonStartBlock)
	prometheus.MustRegister(MetricsKeyperIsKeyper)
	prometheus.MustRegister(MetricsKeyperCurrentPhase)
	prometheus.MustRegister(MetricsKeyperCurrentBatchConfigIndex)
	prometheus.MustRegister(MetricsKeyperBatchConfigInfo)
	prometheus.MustRegister(MetricsKeyperDKGStatus)
	prometheus.MustRegister(MetricsKeyperEthAddress)
	prometheus.MustRegister(MetricsExecutionClientVersion)

	ctx := context.Background()
	queries := database.New(dbpool)

	MetricsKeyperEthAddress.WithLabelValues(config.GetAddress().Hex()).Set(1)

	if version, err := chainsync.GetClientVersion(ctx, config.Ethereum.EthereumURL); err != nil {
		log.Error().Err(err).Msg("keypermetrics | Failed to get execution client version")
	} else {
		MetricsExecutionClientVersion.WithLabelValues(version).Set(1)
	}

	eons, err := queries.GetAllEons(ctx)
	if err != nil || len(eons) == 0 {
		log.Error().Err(err).Msg("keypermetrics | No eons found or failed to fetch eons")
		return
	}

	currentEon := eons[len(eons)-1]

	MetricsKeyperCurrentEon.Set(float64(currentEon.Eon))

	MetricsKeyperCurrentBatchConfigIndex.Set(float64(currentEon.KeyperConfigIndex))

	for _, eon := range eons {
		eonStr := strconv.FormatInt(eon.Eon, 10)
		MetricsKeyperEonStartBlock.WithLabelValues(eonStr).Set(float64(eon.ActivationBlockNumber))
	}

	// Populate MetricsKeyperBatchConfigInfo && MetricsKeyperIsKeyper
	batchConfigs, err := queries.GetBatchConfigs(ctx)
	if err != nil {
		log.Error().Err(err).Msg("keypermetrics | Failed to fetch batch configs")
	} else {
		currentAddress := config.GetAddress().Hex()

		for _, batchConfig := range batchConfigs {
			batchConfigIndexStr := strconv.Itoa(int(batchConfig.KeyperConfigIndex))

			// Join keyper addresses for the label
			keyperAddresses := strings.Join(batchConfig.Keypers, ",")
			MetricsKeyperBatchConfigInfo.WithLabelValues(batchConfigIndexStr, keyperAddresses).Set(1)

			// Check if current node is a keyper in this batch config
			isKeyper := false
			for _, keyperAddr := range batchConfig.Keypers {
				if strings.EqualFold(keyperAddr, currentAddress) {
					isKeyper = true
					break
				}
			}

			var isKeyperValue float64
			if isKeyper {
				isKeyperValue = 1
			}
			MetricsKeyperIsKeyper.WithLabelValues(batchConfigIndexStr).Set(isKeyperValue)
		}
	}

	// Populate MetricsKeyperDKGStatus
	dkgResults, err := queries.GetAllDKGResults(ctx)
	if err != nil {
		log.Error().Err(err).Msg("keypermetrics | Failed to fetch DKG results")
	} else {
		dkgResultMap := make(map[int64]database.DkgResult)
		for _, result := range dkgResults {
			dkgResultMap[result.Eon] = result
		}

		// Set DKG status for all eons
		for _, eon := range eons {
			eonStr := strconv.FormatInt(eon.Eon, 10)

			if dkgResult, exists := dkgResultMap[eon.Eon]; exists {
				var dkgStatusValue float64
				if dkgResult.Success {
					dkgStatusValue = 1
				}
				MetricsKeyperDKGStatus.WithLabelValues(eonStr).Set(dkgStatusValue)
			} else {
				// No DKG result found for this eon, set to 0
				MetricsKeyperDKGStatus.WithLabelValues(eonStr).Set(0)
			}
		}
	}
}