package p2p

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/libp2p/go-libp2p"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	p2pcrypto "github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	basichost "github.com/libp2p/go-libp2p/p2p/host/basic"
	"github.com/libp2p/go-libp2p/p2p/net/connmgr"
	"github.com/multiformats/go-multiaddr"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
)

var DefaultBootstrapPeers []peer.AddrInfo

func init() {
	for _, s := range []string{} {
		addrInfo, err := peer.AddrInfoFromString(s)
		if err != nil {
			log.Info().Str("address", s).Msg("failed to cast address info, ignoring bootstrap peer")
		}
		DefaultBootstrapPeers = append(DefaultBootstrapPeers, *addrInfo)
	}
}

// Minimum number of peers in the routing table. If we drop below this and we
// messagesBufSize is the number of incoming messages to buffer for all of the rooms.
const (
	messagesBufSize = 128
	protocolVersion = "/shutter/0.1.0"
)

type Environment int

//go:generate stringer -type=Environment -output environment_string.gen.go
const (
	Staging Environment = iota
	Production
	Local
)

type P2P struct {
	Config Config

	connmngr         *connmgr.BasicConnMgr
	mux              sync.Mutex
	host             host.Host
	pubSub           *pubsub.PubSub
	gossipRooms      map[string]*gossipRoom
	GossipMessages   chan *Message
	logNumberOfPeers int // number of peers when we last logged them
}

type Config struct {
	ListenAddrs     []multiaddr.Multiaddr
	BootstrapPeers  []peer.AddrInfo
	PrivKey         p2pcrypto.PrivKey
	Environment     Environment
	IsBootstrapNode bool
}

func NewP2P(config Config) *P2P {
	p := P2P{
		Config:           config,
		connmngr:         nil,
		host:             nil,
		pubSub:           nil,
		gossipRooms:      make(map[string]*gossipRoom),
		GossipMessages:   make(chan *Message, messagesBufSize),
		logNumberOfPeers: -1,
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
		if err := p.init(ctx); err != nil {
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
		// TODO:
		// should we somehow allow the bootstrap nodes
		// to exchange the peers for topics?
		// Right now, they are "dumb" bootstrappers who
		// who will only stick to the peers that tried to
		// connect with them.
		// Maybe there are cases where this leads to network
		// compartmentalisation?
		if !p.Config.IsBootstrapNode {
			if err := p.connectBootstrapNodes(ctx); err != nil {
				return err
			}
		}

		// listen to gossip on all topics
		for _, room := range p.gossipRooms {
			room := room
			errorgroup.Go(func() error {
				return room.readLoop(errorgroupctx, p.GossipMessages)
			})
		}

		// in case there were no subscribed topics (e.g. bootstrap nodes),
		// just block the function until the context is canceled
		errorgroup.Go(func() error {
			<-errorgroupctx.Done()
			return ctx.Err()
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
		log.Info().Str("topic", topic).Msg("dropping message, not subscribed to topic")
		return nil
	}
	return room.Publish(ctx, message)
}

func (p *P2P) connectBootstrapNodes(ctx context.Context) error {
	if p.host == nil {
		return errors.New("Cannot connect to bootstrap nodes without existing host")
	}
	var (
		connectedNodes atomic.Uint32
		waitGroup      sync.WaitGroup
	)

	for _, addr := range p.Config.BootstrapPeers {
		waitGroup.Add(1)
		go func(ctx context.Context, a peer.AddrInfo) {
			defer waitGroup.Done()
			err := p.host.Connect(ctx, a)
			if err != nil {
				log.Error().Err(err).Str("peer", a.String()).Msg("couldn't connect to boostrap node")
				return
			}
			connectedNodes.Add(1)
		}(ctx, addr)
	}

	waitGroup.Wait()
	if connectedNodes.Load() == 0 {
		return errors.New("could not connect to any bootstrap node")
	}
	return nil
}

func (p *P2P) init(ctx context.Context) error {
	if p.host != nil {
		return errors.New("Cannot create host on p2p with existing host")
	}
	p2pHost, connectionManager, err := createHost(ctx, p.Config)
	if err != nil {
		return err
	}

	p2pPubSub, err := createPubSub(ctx, p2pHost, p.Config)
	if err != nil {
		return err
	}
	p.host = p2pHost
	p.connmngr = connectionManager
	p.pubSub = p2pPubSub
	log.Info().Str("address", p.p2pAddress()).Msg("created libp2p host")
	return nil
}

func createHost(ctx context.Context, config Config) (*basichost.BasicHost, *connmgr.BasicConnMgr, error) {
	var err error

	privKey := config.PrivKey
	if privKey == nil {
		privKey, _, err = p2pcrypto.GenerateKeyPair(
			p2pcrypto.Ed25519,
			-1,
		)
		if err != nil {
			return nil, nil, err
		}
	}

	connectionManager, err := connmgr.NewConnManager(
		10, // Lowwater
		40, // HighWater,
		connmgr.WithGracePeriod(time.Minute),
	)
	if err != nil {
		return nil, nil, err
	}

	options := []libp2p.Option{
		libp2p.Identity(privKey),
		libp2p.ListenAddrs(config.ListenAddrs...),
		libp2p.DefaultTransports,
		libp2p.DefaultSecurity,
		libp2p.ConnectionManager(connectionManager),
		libp2p.ProtocolVersion(protocolVersion),
	}

	localNetworking := bool(config.Environment == Local)
	if !localNetworking {
		options = append(options,
			// launch the server-side of AutoNAT too
			// in order to help determine other peer's NATted
			// peer-id (service is highly rate-limited)
			libp2p.EnableNATService(),
			// Attempt to open ports using uPNP for NATed hosts.
			libp2p.NATPortMap(),
		)
	}

	p2pHost, err := libp2p.New(options...)
	if err != nil {
		return nil, nil, err
	}
	// we don't want to work with an interface type here,
	// so explicitly cast it
	bhost, ok := p2pHost.(*basichost.BasicHost)
	if !ok {
		// this will happen as soon as libp2p.New does not
		// return a BasicHost ptr anymore
		// this will be noticed early in the p2p-integration test
		return nil, nil, errors.New("libp2p API did change")
	}
	return bhost, connectionManager, nil
}

func createPubSub(ctx context.Context, p2pHost host.Host, config Config) (*pubsub.PubSub, error) {
	localNetworking := bool(config.Environment == Local)
	gossipSubParams, peerScoreParams, peerScoreThresholds := makePubSubParams(pubSubParamsOptions{
		isBootstrapNode:   config.IsBootstrapNode,
		isLocalNetworking: localNetworking,
		bootstrapPeers:    config.BootstrapPeers,
	})
	pubsubOptions := []pubsub.Option{
		pubsub.WithGossipSubParams(*gossipSubParams),
		pubsub.WithPeerScore(peerScoreParams, peerScoreThresholds),
	}
	if config.IsBootstrapNode {
		// enables the pubsub v1.1 feature to handle discovery and
		// connection management over the PubSub protocol
		// This still needs an initial small set of connections,
		// to bootstrap the network,
		pubsubOptions = append(pubsubOptions, pubsub.WithPeerExchange(true))
	}
	pubSub, err := pubsub.NewGossipSub(ctx, p2pHost, pubsubOptions...)
	if err != nil {
		return nil, err
	}
	return pubSub, nil
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
