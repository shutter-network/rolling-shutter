package metrics

import "github.com/prometheus/client_golang/prometheus"

var TxPointer = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Namespace: "shutter",
		Subsystem: "gnosis",
		Name:      "tx_pointer",
		Help:      "Current value of the tx pointer",
	},
	[]string{"eon"},
)

var TxPointerAge = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Namespace: "shutter",
		Subsystem: "gnosis",
		Name:      "tx_pointer_age_blocks",
		Help:      "Current age of the tx pointer",
	},
	[]string{"eon"},
)

var LatestTxSubmittedEventIndex = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Namespace: "shutter",
		Subsystem: "gnosis",
		Name:      "latest_tx_submitted_event_index",
		Help:      "Index of the latest TxSubmitted event",
	},
	[]string{"eon"},
)

var TxSubmittedEventsSyncedUntil = prometheus.NewGauge(
	prometheus.GaugeOpts{
		Namespace: "shutter",
		Subsystem: "gnosis",
		Name:      "tx_submitted_events_synced_until",
		Help:      "Block number until which TxSubmitted events have been synced",
	},
)

var ValidatorRegistrationsSyncedUntil = prometheus.NewGauge(
	prometheus.GaugeOpts{
		Namespace: "shutter",
		Subsystem: "gnosis",
		Name:      "validator_registrations_synced_until",
		Help:      "Block number until which validator registration events have been synced",
	},
)

var NumValidatorRegistrations = prometheus.NewGauge(
	prometheus.GaugeOpts{
		Namespace: "shutter",
		Subsystem: "gnosis",
		Name:      "num_validator_registrations_total",
		Help:      "Number of synced validator registrations",
	},
)

var slotTimeDeltaBuckets = []float64{-5, -4.5, -4.0, -3.5, -3.0, -2.5, -2.0, -1.5, -1.0, -0.5, -0, 1.0, 100}

var KeysSentTimeDelta = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Namespace: "shutter",
		Subsystem: "gnosis",
		Name:      "keys_sent_time_delta_seconds",
		Help:      "Time at which keys are sent relative to slot",
		Buckets:   slotTimeDeltaBuckets,
	},
	[]string{"eon"},
)

var KeySharesSentTimeDelta = prometheus.NewHistogramVec(
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
	prometheus.MustRegister(TxPointer)
	prometheus.MustRegister(TxPointerAge)
	prometheus.MustRegister(LatestTxSubmittedEventIndex)
	prometheus.MustRegister(TxSubmittedEventsSyncedUntil)
	prometheus.MustRegister(ValidatorRegistrationsSyncedUntil)
	prometheus.MustRegister(NumValidatorRegistrations)
	prometheus.MustRegister(KeysSentTimeDelta)
	prometheus.MustRegister(KeySharesSentTimeDelta)
}
