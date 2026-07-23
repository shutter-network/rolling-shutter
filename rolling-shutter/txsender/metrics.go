package txsender

import "github.com/prometheus/client_golang/prometheus"

var (
	metricsTxOutboxPending = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "shutter",
			Subsystem: "keyper",
			Name:      "tx_outbox_pending",
			Help:      "Number of tx_outbox rows awaiting submission",
		},
	)

	metricsTxOutboxSubmitted = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "shutter",
			Subsystem: "keyper",
			Name:      "tx_outbox_submitted",
			Help:      "Number of tx_outbox rows submitted and awaiting confirmation",
		},
	)

	metricsTxOutboxFailed = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "shutter",
			Subsystem: "keyper",
			Name:      "tx_outbox_failed",
			Help:      "Number of tx_outbox rows that terminally failed",
		},
	)
)

func init() {
	prometheus.MustRegister(metricsTxOutboxPending)
	prometheus.MustRegister(metricsTxOutboxSubmitted)
	prometheus.MustRegister(metricsTxOutboxFailed)
}
