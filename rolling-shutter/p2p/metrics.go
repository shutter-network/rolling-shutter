package p2p

import (
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/prometheus/client_golang/prometheus"
)

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

var metricsP2PPeerTuples = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Namespace: "shutter",
		Subsystem: "p2p",
		Name:      "dialed_peer",
		Help:      "Collection of the encountered peer tuples",
	},
	[]string{"peer_id", "peer_ip"})

func collectPeerAddresses(peer peer.AddrInfo) {
	for _, multiAddr := range peer.Addrs {
		metricsP2PPeerTuples.WithLabelValues(peer.ID.String(), multiAddr.String()).Add(1)
	}
}

func init() {
	prometheus.MustRegister(metricsP2PMessageValidationTime)
	prometheus.MustRegister(metricsP2PMessageHandlingTime)
	prometheus.MustRegister(metricsP2PPeerTuples)
}
