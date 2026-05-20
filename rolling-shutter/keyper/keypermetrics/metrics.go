package keypermetrics

import (
	"context"
	"strconv"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog/log"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/database"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/kprconfig"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/chainsync"
)

const (
	DKGMessageTypePolyEval       = "poly_eval"
	DKGMessageTypePolyCommitment = "poly_commitment"
	DKGMessageTypeAccusation     = "accusation"
	DKGMessageTypeApology        = "apology"
)

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

var MetricsKeyperEonStartBlock = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Namespace: "shutter",
		Subsystem: "keyper",
		Name:      "eon_start_block",
		Help:      "Block at which the eon becomes active",
	},
	[]string{"eon"},
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

var MetricsKeyperDKGMessagesSent = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "shutter",
		Subsystem: "keyper",
		Name:      "dkg_messages_sent",
		Help:      "Number of DKG messages scheduled for sending for the given eon, partitioned by message type",
	},
	[]string{"eon", "message_type"},
)

var MetricsKeyperDKGMessagesReceived = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "shutter",
		Subsystem: "keyper",
		Name:      "dkg_messages_received",
		Help:      "Number of DKG messages received for the given eon, partitioned by message type",
	},
	[]string{"eon", "message_type"},
)

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
	prometheus.MustRegister(MetricsKeyperCurrentEon)
	prometheus.MustRegister(MetricsKeyperEonStartBlock)
	prometheus.MustRegister(MetricsKeyperCurrentPhase)
	prometheus.MustRegister(MetricsKeyperDKGStatus)
	prometheus.MustRegister(MetricsKeyperEthAddress)
	prometheus.MustRegister(MetricsExecutionClientVersion)
	prometheus.MustRegister(MetricsKeyperDKGMessagesSent)
	prometheus.MustRegister(MetricsKeyperDKGMessagesReceived)

	ctx := context.Background()
	queries := database.New(dbpool)

	MetricsKeyperEthAddress.WithLabelValues(config.GetAddress().Hex()).Set(1)

	if version, err := chainsync.GetClientVersion(ctx, config.Ethereum.EthereumURL); err != nil {
		log.Error().Err(err).Msg("keypermetrics | Failed to get execution client version")
	} else {
		MetricsExecutionClientVersion.WithLabelValues(version).Set(1)
	}

	eons, err := queries.GetAllEons(ctx)
	if err != nil {
		log.Error().Err(err).Msg("keypermetrics | Failed to fetch eons")
	} else if len(eons) == 0 {
		log.Warn().Msg("keypermetrics | No eons found")
	}

	if len(eons) > 0 {
		currentEon := eons[len(eons)-1]

		MetricsKeyperCurrentEon.Set(float64(currentEon.Eon))

		for _, eon := range eons {
			eonStr := strconv.FormatInt(eon.Eon, 10)
			MetricsKeyperEonStartBlock.WithLabelValues(eonStr).Set(float64(eon.ActivationBlockNumber))
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

	log.Info().Msg("keypermetrics | Metrics population completed")
}
