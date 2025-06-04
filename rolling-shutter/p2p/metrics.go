package p2p

import (
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog/log"
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
		Name:      "peer_candidate_info",
		Help:      "Collection of the encountered peer tuples.",
	},
	[]string{"peer_id", "peer_ip"})

var metricsP2PPeerConnectedness = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Namespace: "shutter",
		Subsystem: "p2p",
		Name:      "peer_connectedness",
		Help:      "Collection of the connectedness (0=NotConnected; 1=Connected; 2=CanConnect; 3=CannotConnect; 4=Limited) to a peer ID.",
	},
	[]string{"our_id", "peer_id"})

var metricsP2PPeerPing = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Namespace: "shutter",
		Subsystem: "p2p",
		Name:      "peer_ping_time_seconds",
		Help:      "Collection of the ping time to a peer ID.",
	},
	[]string{"our_id", "peer_id"},
)

var metricsP2PPeerUserAgent = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Namespace: "shutter",
		Subsystem: "p2p",
		Name:      "peer_user_agent",
		Help:      "Collection of the user agent of a peer ID.",
	},
	[]string{"peer_id", "user_agent"},
)

func collectPeerAddresses(p peer.AddrInfo) {
	for _, multiAddr := range p.Addrs {
		metricsP2PPeerTuples.WithLabelValues(p.ID.String(), multiAddr.String()).Set(1)
	}
}

func init() {
	prometheus.MustRegister(metricsP2PMessageValidationTime)
	prometheus.MustRegister(metricsP2PMessageHandlingTime)
	prometheus.MustRegister(metricsP2PPeerTuples)
	prometheus.MustRegister(metricsP2PPeerConnectedness)
	prometheus.MustRegister(metricsP2PPeerPing)
	prometheus.MustRegister(metricsP2PPeerUserAgent)
}

func updatePeersMetrics(h host.Host, peerIds mapset.Set[peer.ID]) {
	ourID := h.ID().String()
	for p := range peerIds.Iterator().C {
		connectedness := h.Network().Connectedness(p)
		metricsP2PPeerConnectedness.WithLabelValues(ourID, p.String()).Set(float64(connectedness))
		peerPing := h.Peerstore().LatencyEWMA(p)
		if peerPing != 0 {
			metricsP2PPeerPing.WithLabelValues(ourID, p.String()).Set(peerPing.Seconds())
		}
		ua, err := h.Peerstore().Get(p, "AgentVersion")
		if err != nil {
			log.Warn().Str("peer", p.String()).Msg("Can't get user agent for peer")
			continue
		}
		if ua != nil {
			metricsP2PPeerUserAgent.WithLabelValues(p.String(), ua.(string)).Set(1)
		}
	}
}
