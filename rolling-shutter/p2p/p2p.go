package p2p

import (
	"context"
	"crypto/rand"
	"fmt"
	"log"
	"sync"

	"github.com/libp2p/go-libp2p"
	p2pcrypto "github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/multiformats/go-multiaddr"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

// messagesBufSize is the number of incoming messages to buffer for all of the rooms.
const messagesBufSize = 128

type P2P struct {
	Config Config

	mux            sync.Mutex
	host           host.Host
	pubSub         *pubsub.PubSub
	gossipRooms    map[string]*gossipRoom
	GossipMessages chan *Message
}

type Config struct {
	ListenAddr     multiaddr.Multiaddr
	PeerMultiaddrs []multiaddr.Multiaddr
	PrivKey        p2pcrypto.PrivKey
}

func NewP2P(config Config) *P2P {
	p := P2P{
		Config:         config,
		host:           nil,
		pubSub:         nil,
		gossipRooms:    make(map[string]*gossipRoom),
		GossipMessages: make(chan *Message, messagesBufSize),
	}
	return &p
}

func (p *P2P) Run(ctx context.Context, topicNames []string, topicValidators ValidatorRegistry) error {
	defer func() {
		close(p.GossipMessages)
	}()

	errorgroup, errorgroupctx := errgroup.WithContext(ctx)
	errorgroup.Go(func() error {
		p.mux.Lock()
		defer p.mux.Unlock()
		if err := p.createHost(ctx); err != nil {
			return err
		}

		for topicName, validator := range topicValidators {
			if err := p.pubSub.RegisterTopicValidator(topicName, validator); err != nil {
				return err
			}
		}

		if err := p.joinTopics(topicNames); err != nil {
			return err
		}

		// listen to gossip on all topics
		for _, room := range p.gossipRooms {
			room := room
			errorgroup.Go(func() error {
				return room.readLoop(errorgroupctx, p.GossipMessages)
			})
		}

		errorgroup.Go(func() error {
			return p.managePeers(errorgroupctx)
		})
		return nil
	})
	return errorgroup.Wait()
}

func (p *P2P) Publish(ctx context.Context, topic string, message []byte) error {
	p.mux.Lock()
	room, ok := p.gossipRooms[topic]
	p.mux.Unlock()

	if !ok {
		log.Printf("dropping message to not (yet) subscribed topic %s", topic)
		return nil
	}
	return room.Publish(ctx, message)
}

func (p *P2P) createHost(ctx context.Context) error {
	var err error
	if p.host != nil {
		return errors.New("Cannot create host on p2p with existing host")
	}
	privKey := p.Config.PrivKey
	if privKey == nil {
		privKey, _, err = p2pcrypto.GenerateEd25519Key(rand.Reader)
		if err != nil {
			return err
		}
	}

	// create a new libp2p Host
	p.host, err = libp2p.New(ctx, libp2p.ListenAddrs(p.Config.ListenAddr), libp2p.Identity(privKey))
	if err != nil {
		return err
	}
	// print the node's PeerInfo in multiaddr format
	log.Println("libp2p node address:", p.p2pAddress())

	// create a new PubSub service using the GossipSub router
	options := []pubsub.Option{
		pubsub.WithPeerScore(peerScoreParams(), peerScoreThresholds()),
	}
	pubSub, err := pubsub.NewGossipSub(ctx, p.host, options...)
	if err != nil {
		return err
	}
	p.pubSub = pubSub

	return nil
}

func (p *P2P) p2pAddress() string {
	if p.host == nil {
		return "<not connected yet>"
	}
	peerInfo := peer.AddrInfo{
		ID:    p.host.ID(),
		Addrs: p.host.Addrs(),
	}
	addrs, err := peer.AddrInfoToP2pAddrs(&peerInfo)
	if err != nil {
		return fmt.Sprintf("<error: %s>", err)
	}
	return addrs[0].String()
}

// P2PAddress returns the node's PeerInfo in multiaddr format.
func (p *P2P) P2PAddress() string {
	p.mux.Lock()
	defer p.mux.Unlock()
	return p.p2pAddress()
}

func (p *P2P) joinTopics(topicNames []string) error {
	for _, topicName := range topicNames {
		if err := p.joinTopic(topicName); err != nil {
			return err
		}
	}
	return nil
}

// JoinTopic tries to subscribe to the PubSub topic.
func (p *P2P) joinTopic(topicName string) error {
	if _, ok := p.gossipRooms[topicName]; ok {
		return errors.New("Cannot join new topic if already joined")
	}
	// join the pubsub topic
	topic, err := p.pubSub.Join(topicName)
	if err != nil {
		return err
	}

	// set peer scoring parameters
	err = topic.SetScoreParams(topicScoreParams())
	if err != nil {
		return errors.Wrapf(err, "failed to set peer scoring parameters")
	}

	// and subscribe to it
	sub, err := topic.Subscribe()
	if err != nil {
		return err
	}

	p.gossipRooms[topicName] = &gossipRoom{
		pubSub:       p.pubSub,
		topic:        topic,
		subscription: sub,
		self:         p.host.ID(),
		topicName:    topicName,
	}
	return nil
}

func (p *P2P) GetMultiaddr() (multiaddr.Multiaddr, error) {
	p.mux.Lock()
	defer p.mux.Unlock()
	peerInfo := peer.AddrInfo{
		ID:    p.host.ID(),
		Addrs: p.host.Addrs(),
	}
	addrs, err := peer.AddrInfoToP2pAddrs(&peerInfo)
	if len(addrs) != 0 {
		return addrs[0], err
	}
	return nil, err
}

func (p *P2P) HostID() string {
	p.mux.Lock()
	defer p.mux.Unlock()
	return p.host.ID().Pretty()
}
