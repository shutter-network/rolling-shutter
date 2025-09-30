package p2p

import (
	"context"
	"fmt"
	"sync"

	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/discovery/routing"
	"github.com/libp2p/go-libp2p/p2p/discovery/util"
	rcmgr "github.com/libp2p/go-libp2p/p2p/host/resource-manager"
	rhost "github.com/libp2p/go-libp2p/p2p/host/routed"
	"github.com/libp2p/go-libp2p/p2p/net/connmgr"
	"github.com/multiformats/go-multiaddr"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/cmd/shversion"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/encodeable/address"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/encodeable/env"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/encodeable/keys"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/service"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2p/floodsubpeerdiscovery"
)

var DefaultBootstrapPeers []*address.P2PAddress

func init() {
	// NOTE: currently there are no stable production / staging
	//   bootstrap nodes running.
	//   As soon as this is the case, DNS and IP based multiaddrs
	//   for our bootstrap peers will be hardcoded here
	//   for the production/staging environments
	for _, s := range []string{} {
		addr := address.MustP2PAddress(s)
		DefaultBootstrapPeers = append(DefaultBootstrapPeers, addr)
	}
}

const (
	// messagesBufSize is the number of incoming messages to buffer for all of the rooms.
	messagesBufSize = 128
	protocolVersion = "/shutter/0.1.0"
)

type Notifee interface {
	NewPeer()
}

type P2PNode struct {
	config p2pNodeConfig

	mux         sync.Mutex
	host        host.Host
	dht         *dht.IpfsDHT
	discovery   *routing.RoutingDiscovery
	pubSub      *pubsub.PubSub
	gossipRooms map[string]*gossipRoom

	GossipMessages    chan *pubsub.Message
	FloodSubDiscovery *floodsubpeerdiscovery.FloodsubPeerDiscovery
}

type p2pNodeConfig struct {
	ListenAddrs        []multiaddr.Multiaddr
	AdvertiseAddrs     []multiaddr.Multiaddr
	BootstrapPeers     []peer.AddrInfo
	PrivKey            keys.Libp2pPrivate
	Environment        env.Environment
	IsBootstrapNode    bool
	IsAccessNode       bool
	DiscoveryNamespace string
	FloodsubDiscovery  FloodsubDiscoveryConfig
}

type FloodsubDiscoveryConfig struct {
	Enabled  bool
	Interval int
	Topics   []string
}

func NewP2PNode(config p2pNodeConfig) *P2PNode {
	p := P2PNode{
		config:         config,
		host:           nil,
		pubSub:         nil,
		gossipRooms:    make(map[string]*gossipRoom),
		GossipMessages: make(chan *pubsub.Message, messagesBufSize),
	}
	return &p
}

func (p *P2PNode) Run(
	ctx context.Context,
	runner service.Runner,
	topicNames []string,
	topicValidators ValidatorRegistry,
) error {
	p.mux.Lock()
	defer p.mux.Unlock()

	if err := p.init(ctx); err != nil {
		return err
	}

	runner.Go(func() error {
		<-ctx.Done()
		log.Debug().Msg("stopping host when context is done")
		if err := p.host.Close(); err != nil {
			log.Error().Err(err).Msg("error closing host")
		}
		if err := p.dht.Close(); err != nil {
			log.Error().Err(err).Msg("error closing dht")
		}
		close(p.GossipMessages)
		log.Debug().Msg("host closed")
		return nil
	})

	for topicName := range topicValidators {
		validator := topicValidators.GetCombinedValidator(topicName)
		if err := p.pubSub.RegisterTopicValidator(topicName, validator); err != nil {
			return err
		}
	}

	if err := p.joinTopics(topicNames); err != nil {
		return err
	}

	err := bootstrap(ctx, p.host, p.config, p.dht)
	if err != nil {
		return err
	}
	// listen to gossip on all topics
	for _, room := range p.gossipRooms {
		room := room
		runner.Go(func() error {
			return room.readLoop(ctx, p.GossipMessages)
		})
	}
	if p.config.FloodsubDiscovery.Enabled {
		log.Info().Msg("floodsub peer discovery is enabled")
		p.FloodSubDiscovery = &floodsubpeerdiscovery.FloodsubPeerDiscovery{}
		peerDiscoveryComponents := floodsubpeerdiscovery.PeerDiscoveryComponents{
			PeerID:    address.P2PIdentifier{ID: p.host.ID()},
			PeerStore: p.host.Peerstore(),
			Pubsub:    p.pubSub,
		}
		err := p.FloodSubDiscovery.Init(peerDiscoveryComponents, p.config.FloodsubDiscovery.Interval, p.config.FloodsubDiscovery.Topics)
		if err != nil {
			return err
		}
		runner.Go(func() error {
			return p.FloodSubDiscovery.Start(ctx)
		})

		for _, subs := range p.FloodSubDiscovery.Subscription {
			runner.Go(func() error {
				return p.FloodSubDiscovery.ReadLoop(ctx, subs)
			})
		}
	}
	runner.Go(func() error {
		log.Info().Str("namespace", p.config.DiscoveryNamespace).Msg("starting advertizing discovery node")
		util.Advertise(ctx, p.discovery, p.config.DiscoveryNamespace)
		return nil
	})
	runner.Go(func() error {
		return findPeers(ctx, p.host, p.discovery, p.config.DiscoveryNamespace)
	})
	return nil
}

func (p *P2PNode) Publish(ctx context.Context, topic string, message []byte) error {
	p.mux.Lock()
	room, ok := p.gossipRooms[topic]
	p.mux.Unlock()

	if !ok {
		log.Info().Str("topic", topic).Msg("dropping message, not subscribed to topic")
		return nil
	}
	return room.Publish(ctx, message)
}

func (p *P2PNode) init(ctx context.Context) error {
	if p.host != nil {
		return errors.New("Cannot create host on p2p with existing host")
	}
	p2pHost, hashTable, err := createHost(ctx, p.config)
	if err != nil {
		return err
	}
	discovery := routing.NewRoutingDiscovery(hashTable)
	p2pPubSub, err := createPubSub(ctx, p2pHost, p.config, discovery)
	if err != nil {
		return err
	}

	p.host = p2pHost
	p.dht = hashTable
	p.discovery = discovery
	p.pubSub = p2pPubSub
	log.Info().Str("address", p.p2pAddress()).Msg("created libp2p host")
	return nil
}

func createConnectionManager() (*connmgr.BasicConnMgr, error) {
	// TODO: This starts a background goroutine. It works, but it's better to do that later in
	// P2PNode.Run() when we have a proper context.
	m, err := connmgr.NewConnManager(peerLow, peerHigh)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create connection manager")
	}

	return m, nil
}

func createHost(
	ctx context.Context,
	config p2pNodeConfig,
) (host.Host, *dht.IpfsDHT, error) {
	var err error

	// NOTE:
	// Upon initialization, we are seeing log warnings:
	// "rcmgr limit conflicts with connmgr limit: conn manager high watermark limit: 192, exceeds the system connection limit of: 1"
	//
	// This was a bug in the check function, reading the wrong config value to check against:
	// https://github.com/libp2p/go-libp2p/issues/2628

	connectionManager, err := createConnectionManager()
	if err != nil {
		return nil, nil, err
	}

	options := []libp2p.Option{
		libp2p.Identity(&config.PrivKey.Key),
		libp2p.ListenAddrs(config.ListenAddrs...),
		libp2p.UserAgent(fmt.Sprintf("shutter-network/%s", shversion.VersionShort())),
		libp2p.ConnectionManager(connectionManager),
		libp2p.ProtocolVersion(protocolVersion),
		libp2p.EnableRelay(),
		libp2p.Ping(true),
	}

	localNetworking := bool(config.Environment == env.EnvironmentLocal)
	if !localNetworking {
		options = append(options,
			// launch the server-side of AutoNAT too
			// in order to help determine other peer's NATted
			// peer-id (service is highly rate-limited)
			libp2p.EnableNATService(),
			// Attempt to open ports using uPNP for NATed hosts.
			libp2p.NATPortMap(),
		)
		if len(config.AdvertiseAddrs) > 0 {
			// If advertise addresses are set, only advertise those
			options = append(options,
				libp2p.AddrsFactory(func(addrs []multiaddr.Multiaddr) []multiaddr.Multiaddr {
					return config.AdvertiseAddrs
				}),
			)
		}
		if len(config.BootstrapPeers) > 0 {
			options = append(options,
				libp2p.EnableAutoRelayWithStaticRelays(config.BootstrapPeers),
			)
		}
		if config.IsBootstrapNode {
			// Enable the Relay service on bootstrap nodes so other peers can connect through us
			options = append(options,
				libp2p.EnableRelayService(),
			)
		}
		if config.IsBootstrapNode || config.IsAccessNode {
			// No resource limits on boot and access nodes for now
			mgr, err := rcmgr.NewResourceManager(rcmgr.NewFixedLimiter(rcmgr.InfiniteLimits))
			if err != nil {
				return nil, nil, err
			}
			options = append(options, libp2p.ResourceManager(mgr))
		}
	}

	p2pHost, err := libp2p.New(options...)
	if err != nil {
		return nil, nil, err
	}

	opts := dhtRoutingOptions(&config)
	idht, err := dht.New(ctx, p2pHost, opts...)
	if err != nil {
		return nil, nil, err
	}
	// the wrapped host will try to query the routing table (dht)/
	// whenever it doesn't have the full routed address for a peer id
	routedHost := rhost.Wrap(p2pHost, idht)

	return routedHost, idht, nil
}

func createPubSub(
	ctx context.Context,
	p2pHost host.Host,
	config p2pNodeConfig,
	discovery *routing.RoutingDiscovery,
) (*pubsub.PubSub, error) {
	gossipSubParams, peerScoreParams, peerScoreThresholds := makePubSubParams(pubSubParamsOptions{
		isBootstrapNode: config.IsBootstrapNode,
		bootstrapPeers:  config.BootstrapPeers,
		isAccessNode:    config.IsAccessNode,
	})

	pubsubOptions := []pubsub.Option{
		pubsub.WithGossipSubParams(*gossipSubParams),
		pubsub.WithPeerScore(peerScoreParams, peerScoreThresholds),
		pubsub.WithDiscovery(discovery),
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

func (p *P2PNode) p2pAddress() string {
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
func (p *P2PNode) P2PAddress() string {
	p.mux.Lock()
	defer p.mux.Unlock()
	return p.p2pAddress()
}

func (p *P2PNode) joinTopics(topicNames []string) error {
	for _, topicName := range topicNames {
		if err := p.joinTopic(topicName); err != nil {
			return err
		}
	}
	return nil
}

// JoinTopic tries to subscribe to the PubSub topic.
func (p *P2PNode) joinTopic(topicName string) error {
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

func (p *P2PNode) GetMultiaddr() (multiaddr.Multiaddr, error) {
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

func (p *P2PNode) HostID() string {
	p.mux.Lock()
	defer p.mux.Unlock()
	return p.host.ID().String()
}
