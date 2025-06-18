package gnosis

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog/log"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/beaconapiclient"
)

var metricsTxPointer = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Namespace: "shutter",
		Subsystem: "gnosis",
		Name:      "tx_pointer",
		Help:      "Current value of the tx pointer",
	},
	[]string{"eon"},
)

var metricsTxPointerAge = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Namespace: "shutter",
		Subsystem: "gnosis",
		Name:      "tx_pointer_age_blocks",
		Help:      "Current age of the tx pointer",
	},
	[]string{"eon"},
)

var metricsLatestTxSubmittedEventIndex = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Namespace: "shutter",
		Subsystem: "gnosis",
		Name:      "latest_tx_submitted_event_index",
		Help:      "Index of the latest TxSubmitted event",
	},
	[]string{"eon"},
)

var metricsTxSubmittedEventsSyncedUntil = prometheus.NewGauge(
	prometheus.GaugeOpts{
		Namespace: "shutter",
		Subsystem: "gnosis",
		Name:      "tx_submitted_events_synced_until",
		Help:      "Block number until which TxSubmitted events have been synced",
	},
)

var metricsValidatorRegistrationsSyncedUntil = prometheus.NewGauge(
	prometheus.GaugeOpts{
		Namespace: "shutter",
		Subsystem: "gnosis",
		Name:      "validator_registrations_synced_until",
		Help:      "Block number until which validator registration events have been synced",
	},
)

var metricsNumValidatorRegistrations = prometheus.NewGauge(
	prometheus.GaugeOpts{
		Namespace: "shutter",
		Subsystem: "gnosis",
		Name:      "num_validator_registrations_total",
		Help:      "Number of synced validator registrations",
	},
)

var slotTimeDeltaBuckets = []float64{-5, -4.5, -4.0, -3.5, -3.0, -2.5, -2.0, -1.5, -1.0, -0.5, -0, 1.0, 100}

var metricsKeysSentTimeDelta = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Namespace: "shutter",
		Subsystem: "gnosis",
		Name:      "keys_sent_time_delta_seconds",
		Help:      "Time at which keys are sent relative to slot",
		Buckets:   slotTimeDeltaBuckets,
	},
	[]string{"eon"},
)

var metricsKeySharesSentTimeDelta = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Namespace: "shutter",
		Subsystem: "gnosis",
		Name:      "key_shares_sent_time_delta_seconds",
		Help:      "Time at which key shares are sent relative to slot",
		Buckets:   slotTimeDeltaBuckets,
	},
	[]string{"eon"},
)

func init() {
	prometheus.MustRegister(metricsTxPointer)
	prometheus.MustRegister(metricsTxPointerAge)
	prometheus.MustRegister(metricsLatestTxSubmittedEventIndex)
	prometheus.MustRegister(metricsTxSubmittedEventsSyncedUntil)
	prometheus.MustRegister(metricsValidatorRegistrationsSyncedUntil)
	prometheus.MustRegister(metricsNumValidatorRegistrations)
	prometheus.MustRegister(metricsKeysSentTimeDelta)
	prometheus.MustRegister(metricsKeySharesSentTimeDelta)
}

func InitMetrics(beaconClient *beaconapiclient.Client) {
	version, err := beaconClient.GetBeaconNodeVersion(context.Background())
	if err != nil {
		log.Error().Err(err).Msg("Failed to get beacon node version")
	}
	beaconClientVersion := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "shutter",
			Subsystem: "gnosis",
			Name:      "beacon_client_version",
			Help:      "Version of the beacon client",
			ConstLabels: prometheus.Labels{
				"version": version,
			},
		},
	)
	prometheus.MustRegister(beaconClientVersion)
}
