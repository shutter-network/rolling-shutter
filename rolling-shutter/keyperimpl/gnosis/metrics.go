package gnosis

import "github.com/prometheus/client_golang/prometheus"

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

func init() {
	prometheus.MustRegister(metricsTxPointer)
	prometheus.MustRegister(metricsTxPointerAge)
	prometheus.MustRegister(metricsLatestTxSubmittedEventIndex)
	prometheus.MustRegister(metricsTxSubmittedEventsSyncedUntil)
	prometheus.MustRegister(metricsValidatorRegistrationsSyncedUntil)
	prometheus.MustRegister(metricsNumValidatorRegistrations)
}
