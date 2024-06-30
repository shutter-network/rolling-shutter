package p2p

import "github.com/prometheus/client_golang/prometheus"

var metricsP2PMessageValidationTime = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Namespace: "shutter",
		Subsystem: "p2p",
		Name:      "message_validation_time_seconds",
		Help:      "Histogram of the time it takes to validate a P2P message.",
		Buckets:   prometheus.DefBuckets,
	},
	[]string{"topic"},
)

var metricsP2PMessageHandlingTime = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Namespace: "shutter",
		Subsystem: "p2p",
		Name:      "message_handling_time_seconds",
		Help:      "Histogram of the time it takes to handle a P2P message.",
		Buckets:   prometheus.DefBuckets,
	},
	[]string{"topic"},
)

func init() {
	prometheus.MustRegister(metricsP2PMessageValidationTime)
	prometheus.MustRegister(metricsP2PMessageHandlingTime)
}
