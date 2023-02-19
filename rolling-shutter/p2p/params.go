package p2p

import (
	"time"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/peer"
)

func peerScoreParams() *pubsub.PeerScoreParams {
	return &pubsub.PeerScoreParams{
		Topics:        make(map[string]*pubsub.TopicScoreParams),
		TopicScoreCap: 32.72,

		AppSpecificScore:  func(p peer.ID) float64 { return 0 },
		AppSpecificWeight: 1,

		IPColocationFactorWeight:    0, // -35.11 for non-testing environments
		IPColocationFactorThreshold: 10,
		IPColocationFactorWhitelist: nil,

		BehaviourPenaltyWeight:    -15.92,
		BehaviourPenaltyThreshold: 6,
		BehaviourPenaltyDecay:     0.928,

		DecayInterval: 12 * time.Second,
		DecayToZero:   0.01,

		RetainScore: 12 * time.Hour,
	}
}

func peerScoreThresholds() *pubsub.PeerScoreThresholds {
	return &pubsub.PeerScoreThresholds{
		GossipThreshold:             -4000,
		PublishThreshold:            -8000,
		GraylistThreshold:           -16000,
		AcceptPXThreshold:           100,
		OpportunisticGraftThreshold: 5,
	}
}

func topicScoreParams() *pubsub.TopicScoreParams {
	// Based on attestation topic in beacon chain network. The formula uses the number of
	// validators which we set to a fixed number which could be the number of keypers.
	n := float64(200)
	return &pubsub.TopicScoreParams{
		TopicWeight:                     1,
		TimeInMeshWeight:                0.0324,
		TimeInMeshQuantum:               12 * time.Second,
		TimeInMeshCap:                   300,
		FirstMessageDeliveriesWeight:    0.05,
		FirstMessageDeliveriesDecay:     0.631,
		FirstMessageDeliveriesCap:       n / 755.712,
		MeshMessageDeliveriesWeight:     -0.026,
		MeshMessageDeliveriesDecay:      0.631,
		MeshMessageDeliveriesCap:        n / 94.464,
		MeshMessageDeliveriesThreshold:  n / 377.856,
		MeshMessageDeliveriesWindow:     200 * time.Millisecond,
		MeshMessageDeliveriesActivation: 4 * 12 * time.Second,
		MeshFailurePenaltyWeight:        -0.0026,
		MeshFailurePenaltyDecay:         0.631,
		InvalidMessageDeliveriesWeight:  -99,
		InvalidMessageDeliveriesDecay:   0.9994,
	}
}
