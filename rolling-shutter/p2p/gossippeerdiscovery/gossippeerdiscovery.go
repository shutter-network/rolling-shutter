package gossippeerdiscovery

import (
	"context"
	"fmt"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/rs/zerolog/log"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/encodeable/address"
	"google.golang.org/protobuf/proto"
)

const defaultTopic = "_peer-discovery._p2p._pubsub"

type GossipPeerDiscovery struct {
	PeerDiscoveryComponents
	Interval   int
	Topics     []string
	ListenOnly bool
}

type PeerDiscoveryComponents struct {
	peerId    address.P2PIdentifier
	pubsub    *pubsub.PubSub
	peerStore peerstore.Peerstore
}

func (pd *GossipPeerDiscovery) init(config PeerDiscoveryComponents, interval int, topics []string, listenOnly bool) {
	pd.Interval = interval
	if len(topics) > 0 {
		pd.Topics = topics
	} else {
		pd.Topics = []string{defaultTopic}
	}
	pd.ListenOnly = listenOnly
	pd.peerId = config.peerId
	pd.pubsub = config.pubsub
	pd.peerStore = config.peerStore
}

func (pd *GossipPeerDiscovery) broadcast() error {
	pubKey, err := pd.peerId.ExtractPublicKey()
	if err != nil {
		return fmt.Errorf("peerId was missing public key | err %v", err)
	}

	pubKeyBytes, err := pubKey.Raw()
	if err != nil || len(pubKeyBytes) == 0 {
		return fmt.Errorf("peerId was missing public key | err %v", err)
	}

	if pd.pubsub == nil {
		return fmt.Errorf("pubSub not configured | err %v", err)
	}

	addresses := make([][]byte, 0)

	for _, addr := range pd.peerStore.Addrs(pd.peerId.ID) {
		addresses = append(addresses, addr.Bytes())
	}

	peer := Peer{
		PublicKey: pubKeyBytes,
		Addrs:     addresses,
	}
	pbPeer, err := proto.Marshal(&peer)
	if err != nil {
		return fmt.Errorf("error marshalling message | err %v", err)
	}

	for _, topic := range pd.Topics {
		if len(pd.pubsub.ListPeers(topic)) == 0 {
			log.Info().Msgf("skipping broadcasting our peer data on topic %s because there are no peers present", topic)
			continue
		}
		log.Info().Msgf("broadcasting our peer data on topic %s", topic)
		topic, err := pd.pubsub.Join(topic)
		if err != nil {
			return fmt.Errorf("failed to join topic | err %v", err)
		}
		if err := topic.Publish(context.Background(), pbPeer); err != nil {
			return fmt.Errorf("failed to publish to topic | err %v", err)
		}
	}
	return nil
}
