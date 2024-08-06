package epochkghandler

import "github.com/prometheus/client_golang/prometheus"

var metricsEpochKGDecryptionKeysReceived = prometheus.NewCounter(
	prometheus.CounterOpts{
		Namespace: "shutter",
		Subsystem: "epochkg",
		Name:      "decryption_keys_received_total",
		Help:      "Number of received decryption keys",
	},
)

var metricsEpochKGDecryptionKeysGenerated = prometheus.NewCounter(
	prometheus.CounterOpts{
		Namespace: "shutter",
		Subsystem: "epochkg",
		Name:      "decryption_keys_generated_total",
		Help:      "Number of generated decryption keys",
	},
)

var metricsEpochKGDecryptionKeySharesReceived = prometheus.NewCounter(
	prometheus.CounterOpts{
		Namespace: "shutter",
		Subsystem: "epochkg",
		Name:      "decryption_keyshares_received_total",
		Help:      "Number of received decryption key shares",
	},
)

var metricsEpochKGDecryptionKeySharesSent = prometheus.NewCounter(
	prometheus.CounterOpts{
		Namespace: "shutter",
		Subsystem: "epochkg",
		Name:      "decryption_keyshares_sent_total",
		Help:      "Number of sent decryption key shares",
	},
)

var metricsEpochKGDecryptionTriggersReceived = prometheus.NewCounter(
	prometheus.CounterOpts{
		Namespace: "shutter",
		Subsystem: "epochkg",
		Name:      "decryption_triggers_received_total",
		Help:      "Number of received decryption triggers",
	},
)

func InitMetrics() {
	prometheus.MustRegister(metricsEpochKGDecryptionKeysReceived)
	prometheus.MustRegister(metricsEpochKGDecryptionKeysGenerated)
	prometheus.MustRegister(metricsEpochKGDecryptionKeySharesReceived)
	prometheus.MustRegister(metricsEpochKGDecryptionKeySharesSent)
	prometheus.MustRegister(metricsEpochKGDecryptionTriggersReceived)
}
