package p2p

import (
	"time"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/peer"
)

type pubSubParamsOptions struct {
	isBootstrapNode bool
	bootstrapPeers  []peer.AddrInfo
}

func makePubSubParams(
	options pubSubParamsOptions,
) (*pubsub.GossipSubParams, *pubsub.PeerScoreParams, *pubsub.PeerScoreThresholds) {
	gsDefault := pubsub.DefaultGossipSubParams()
	gossipSubParams := &gsDefault

	// modified defaults from ethereum consensus spec
	// https://github.com/ethereum/consensus-specs/blob/5d80b1954a4b7a121aa36143d50b366727b66cbc/\
	//   specs/phase0/p2p-interface.md#why-are-these-specific-gossip-parameters-chosen //nolint:lll
	gossipSubParams.HeartbeatInterval = 700 * time.Millisecond
	gossipSubParams.HistoryLength = 6

	// From the spec:
	// to allow bootstrapping via PeerExchange (PX),
	// the bootstrappers should not form a mesh, thus D=D_lo=D_hi=D_out=0
	if options.isBootstrapNode {
		gossipSubParams.D = 0
		gossipSubParams.Dlo = 0
		gossipSubParams.Dhi = 0
		gossipSubParams.Dout = 0
		gossipSubParams.Dscore = 0
	}

	peerScoreThresholds := &pubsub.PeerScoreThresholds{
		GossipThreshold:             -4000,
		PublishThreshold:            -8000,
		GraylistThreshold:           -16000,
		AcceptPXThreshold:           100,
		OpportunisticGraftThreshold: 5,
	}

	bootstrapSet := make(map[peer.ID]bool, 0)
	for _, bs := range options.bootstrapPeers {
		bootstrapSet[bs.ID] = true
	}

	// NOTE: loosely from the gossipsub spec:
	// Only the bootstrappers / highly trusted PX'ing nodes
	// should reach the AcceptPXThreshold thus they need
	// to be treated differently in the scoring function.
	appSpecificScoringFn := func(p peer.ID) float64 {
		_, ok := bootstrapSet[p]
		if !ok {
			return 0.
		}
		// In order to be able to participate in the gossipsub,
		// a peer has to be PX'ed by a bootstrap node - this is only
		// possible if the AcceptPXThreshold peer-score is reached.

		// NOTE: we have yet to determine a value that is
		// sufficient to reach the AcceptPXThreshold most of the times,
		// but don't overshoot and trust the bootstrap peers
		// unconditionally - they should still be punishable
		// for malicous behavior
		return 200.
	}
	peerScoreParams := &pubsub.PeerScoreParams{
		// Topics score-map will be filled later while subscribing to topics.
		Topics:        make(map[string]*pubsub.TopicScoreParams),
		TopicScoreCap: 32.72,

		AppSpecificScore:  appSpecificScoringFn,
		AppSpecificWeight: 1,

		IPColocationFactorWeight:    -35.11,
		IPColocationFactorThreshold: 10,
		IPColocationFactorWhitelist: nil,

		BehaviourPenaltyWeight:    -15.92,
		BehaviourPenaltyThreshold: 6,
		BehaviourPenaltyDecay:     0.928,

		DecayInterval: 12 * time.Second,
		DecayToZero:   0.01,

		RetainScore: 12 * time.Hour,
	}
	return gossipSubParams, peerScoreParams, peerScoreThresholds
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
