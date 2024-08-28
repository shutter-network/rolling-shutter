package floodsubpeerdiscovery

import (
	"context"
	"fmt"
	"time"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/rs/zerolog/log"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/encodeable/address"
	"google.golang.org/protobuf/proto"
)

const defaultTopic = "_peer-discovery._p2p._pubsub"

type FloodsubPeerDiscovery struct {
	PeerDiscoveryComponents
	Interval int
	Topics   []*pubsub.Topic
}

type PeerDiscoveryComponents struct {
	PeerId    address.P2PIdentifier
	Pubsub    *pubsub.PubSub
	PeerStore peerstore.Peerstore
}

func (pd *FloodsubPeerDiscovery) Init(config PeerDiscoveryComponents, interval int, topics []string) error {
	pd.Interval = interval
	pd.PeerId = config.PeerId
	pd.Pubsub = config.Pubsub
	pd.PeerStore = config.PeerStore

	if len(topics) > 0 {
		for _, topic := range topics {
			topic, err := pd.Pubsub.Join(topic)
			if err != nil {
				return fmt.Errorf("failed to join topic | err %w", err)
			}
			pd.Topics = append(pd.Topics, topic)
		}
	} else {
		topic, err := pd.Pubsub.Join(defaultTopic)
		if err != nil {
			return fmt.Errorf("failed to join topic | err %w", err)
		}
		pd.Topics = append(pd.Topics, topic)
	}
	return nil
}

func (pd *FloodsubPeerDiscovery) Start(ctx context.Context) error {
	timer := time.NewTicker(time.Duration(pd.Interval) * time.Second)

	for {
		select {
		case <-timer.C:
			err := pd.broadcast()
			if err != nil {
				log.Warn().Msgf("error in broadcasting floodsub msg | %v", err)
				return err
			}
		case <-ctx.Done():
			return nil
		}
	}
}

func (pd *FloodsubPeerDiscovery) broadcast() error {
	pubKey, err := pd.PeerId.ExtractPublicKey()
	if err != nil {
		return fmt.Errorf("peerId was missing public key | err %w", err)
	}

	pubKeyBytes, err := pubKey.Raw()
	if err != nil || len(pubKeyBytes) == 0 {
		return fmt.Errorf("peerId was missing public key | err %w", err)
	}

	if pd.Pubsub == nil {
		return fmt.Errorf("pubSub not configured | err %w", err)
	}

	addresses := make([][]byte, 0)

	for _, addr := range pd.PeerStore.Addrs(pd.PeerId.ID) {
		addresses = append(addresses, addr.Bytes())
	}

	peer := Peer{
		PublicKey: pubKeyBytes,
		Addrs:     addresses,
	}
	pbPeer, err := proto.Marshal(&peer)
	if err != nil {
		return fmt.Errorf("error marshalling message | err %w", err)
	}

	for _, topic := range pd.Topics {
		if len(pd.Pubsub.ListPeers(topic.String())) == 0 {
			log.Info().Msgf("skipping broadcasting our peer data on topic %s because there are no peers present", topic)
			continue
		}
		log.Info().Msgf("broadcasting our peer data on topic %s", topic)

		if err := topic.Publish(context.Background(), pbPeer); err != nil {
			return fmt.Errorf("failed to publish to topic | err %w", err)
		}
		defer topic.Close()
	}
	return nil
}
