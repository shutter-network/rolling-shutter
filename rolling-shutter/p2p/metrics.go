package p2p

import (
	"github.com/prometheus/client_golang/prometheus"
)

var metricPeersTotal = prometheus.NewGauge(
	prometheus.GaugeOpts{
		Namespace: "shutter",
		Subsystem: "p2p",
		Name:      "peers_connected_total",
		Help:      "Number currently connected libp2p peers",
	},
)

var metricPeersMin = prometheus.NewGauge(
	prometheus.GaugeOpts{
		Namespace: "shutter",
		Subsystem: "p2p",
		Name:      "peers_required_total",
		Help:      "Configured number of required libp2p peers",
	},
)

var metricMessagesReceived = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "shutter",
		Subsystem: "p2p",
		Name:      "messages_received_total",
		Help:      "Number libp2p pubsub messages received",
	},
	[]string{"topic"},
)

var metricMessagesSent = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "shutter",
		Subsystem: "p2p",
		Name:      "messages_sent_total",
		Help:      "Number libp2p pubsub messages sent",
	},
	[]string{"topic"},
)

func (p *P2P) initMetrics() {
	prometheus.MustRegister(metricPeersTotal)
	prometheus.MustRegister(metricPeersMin)
	prometheus.MustRegister(metricMessagesReceived)
	prometheus.MustRegister(metricMessagesSent)
}
